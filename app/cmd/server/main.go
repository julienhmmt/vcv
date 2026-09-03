package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vcv/internal/metrics"

	"vcv/internal/config"
	"vcv/internal/handlers"
	"vcv/internal/logger"
	"vcv/internal/middleware"
	"vcv/internal/notify"
	"vcv/internal/vault"
	"vcv/internal/version"
	"vcv/web"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const serverReadHeaderTimeout time.Duration = 5 * time.Second
const serverMaxHeaderBytes int = 1 << 20
const routerMaxBodyBytes int64 = 1 << 20
const routerRateLimitMaxRequests int = 300
const notifyCheckInterval time.Duration = 15 * time.Minute
const routerRateLimitWindow time.Duration = 1 * time.Minute

// publicVaultStatusError maps internal vault connection errors to stable,
// non-sensitive strings for the public /api/status response.
func publicVaultStatusError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, vault.ErrVaultNotConfigured) {
		return "vault not configured"
	}
	return "vault unavailable"
}

// vaultStatusEntry is the per-vault section of the /api/status response.
type vaultStatusEntry struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Connected   bool   `json:"connected"`
	Error       string `json:"error,omitempty"`
}

// statusResponse is the /api/status payload.
type statusResponse struct {
	Version         string             `json:"version"`
	VaultConnected  bool               `json:"vault_connected"`
	VaultError      string             `json:"vault_error,omitempty"`
	AdminAPIEnabled bool               `json:"admin_api_enabled"`
	Vaults          []vaultStatusEntry `json:"vaults"`
}

// checkPrimaryVault checks the legacy primary connection in parallel-friendly
// form and maps failures to a sanitized error string.
func checkPrimaryVault(ctx context.Context, client vault.Client) (bool, string) {
	results := vault.CheckInstances(ctx, []string{"primary"}, map[string]vault.Client{"primary": client}, 5*time.Second)
	var primaryErr error
	if len(results) > 0 {
		if results[0].Connected {
			return true, ""
		}
		primaryErr = results[0].Error
	}
	return false, publicVaultStatusError(primaryErr)
}

// appendVaultStatuses checks each configured vault and appends status entries.
func appendVaultStatuses(ctx context.Context, response *statusResponse, instances []config.VaultInstance, clients map[string]vault.Client) {
	ordered := make([]string, 0, len(instances))
	displayNames := make(map[string]string, len(instances))
	for _, instance := range instances {
		ordered = append(ordered, instance.ID)
		displayNames[instance.ID] = instance.DisplayName
	}
	if clients == nil {
		clients = map[string]vault.Client{}
	}
	checked := vault.CheckInstances(ctx, ordered, clients, 5*time.Second)
	for _, item := range checked {
		entry := vaultStatusEntry{ID: item.ID, DisplayName: displayNames[item.ID], Connected: item.Connected}
		if _, ok := clients[item.ID]; !ok || clients[item.ID] == nil {
			entry.Connected = false
			entry.Error = "missing vault status client"
		} else if !item.Connected {
			entry.Error = publicVaultStatusError(item.Error)
		}
		response.Vaults = append(response.Vaults, entry)
	}
}

func newStatusHandler(cfg config.Config, primaryVaultClient vault.Client, statusClients map[string]vault.Client, adminAPIEnabled bool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		response := statusResponse{Version: version.Version, AdminAPIEnabled: adminAPIEnabled, Vaults: make([]vaultStatusEntry, 0, len(cfg.Vaults))}
		// Primary connection (historical field) checked via helper.
		response.VaultConnected, response.VaultError = checkPrimaryVault(ctx, primaryVaultClient)
		appendVaultStatuses(ctx, &response, cfg.Vaults, statusClients)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}

// routerDeps bundles the collaborators buildRouter needs, keeping the
// function signature within the parameter limit.
type routerDeps struct {
	cfg           config.Config
	primaryVault  vault.Client
	statusClients map[string]vault.Client
	multiVault    vault.Client
	vaultRegistry *vault.Registry
	promRegistry  *prometheus.Registry
	webFS         fs.FS
	settingsPath  string
}

func buildRouter(deps routerDeps) (*chi.Mux, error) {
	cfg := deps.cfg
	r := chi.NewRouter()
	distFS, distError := fs.Sub(deps.webFS, "dist")
	if distError != nil {
		return nil, distError
	}
	corsConfig := middleware.DefaultCORSConfig()
	corsConfig.AllowedOrigins = cfg.CORS.AllowedOrigins
	corsConfig.AllowCredentials = cfg.CORS.AllowCredentials

	// Middleware must be registered before any routes
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CORS(corsConfig))
	rateLimitConfig := middleware.DefaultRateLimitConfig()
	rateLimitConfig.MaxRequests = routerRateLimitMaxRequests
	rateLimitConfig.Window = routerRateLimitWindow
	rateLimitConfig.ExemptPaths = []string{"/api/health", "/api/ready", "/metrics"}
	rateLimitConfig.ExemptPathPrefixes = []string{"/assets/"}
	rateLimitConfig.TrustProxy = cfg.TrustProxy
	r.Use(middleware.RateLimit(rateLimitConfig))
	r.Use(middleware.BodyLimit(routerMaxBodyBytes))
	r.Use(middleware.CSRFProtectionWithTrust(cfg.TrustProxy))

	handlers.RegisterStaticRoutes(r, distFS)

	// Public read APIs: assume network ACL. See app/README.md "Security & deployment assumptions".
	// Health and readiness probes
	r.Get("/api/health", handlers.HealthCheck)
	r.Get("/api/ready", handlers.ReadinessCheck)
	// Admin is optional; process stays up. Surface enablement on /api/status (not fail-ready).
	adminAPIEnabled := handlers.RegisterAdminRoutes(r, deps.settingsPath, cfg.Env, deps.vaultRegistry, deps.statusClients, deps.multiVault, cfg.TrustProxy)
	r.Get("/api/status", newStatusHandler(cfg, deps.primaryVault, deps.statusClients, adminAPIEnabled))
	r.Get("/api/version", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(version.Info())
	})
	r.Get("/api/config", handlers.GetConfig(cfg, deps.vaultRegistry))
	r.Get("/metrics", promhttp.HandlerFor(deps.promRegistry, promhttp.HandlerOpts{}).ServeHTTP)
	handlers.RegisterI18nRoutes(r)
	handlers.RegisterCertRoutes(r, deps.multiVault)

	return r, nil
}

// initVaultClients creates clients for ALL vaults (including disabled) so
// they can be toggled at runtime via the admin panel without a restart, and
// returns them plus the primary client. Sets cfg.Vault to the primary
// instance's config for logging/legacy use.
func initVaultClients(cfg *config.Config, log *logger.Logger) (map[string]vault.Client, vault.Client) {
	allClients := make(map[string]vault.Client, len(cfg.AllVaults))
	var primaryVaultClient vault.Client
	for i, instance := range cfg.AllVaults {
		if instance.ID == "" {
			continue
		}
		vaultCfg := config.VaultConfig{Addr: instance.Address, PKIMounts: instance.PKIMounts, ReadToken: instance.Token, TLSCACertBase64: instance.TLSCACertBase64, TLSCACert: instance.TLSCACert, TLSCAPath: instance.TLSCAPath, TLSServerName: instance.TLSServerName, TLSInsecure: instance.TLSInsecure}
		client, err := vault.NewClientFromConfig(vaultCfg)
		if err != nil {
			log.Error().Err(err).
				Str("vault_id", instance.ID).
				Msg("Failed to initialize Vault client, skipping")
			continue
		}
		allClients[instance.ID] = client
		if i == 0 {
			primaryVaultClient = client
			cfg.Vault = config.VaultConfig{Addr: instance.Address, PKIMounts: instance.PKIMounts, ReadToken: instance.Token, TLSCACertBase64: instance.TLSCACertBase64, TLSCACert: instance.TLSCACert, TLSCAPath: instance.TLSCAPath, TLSServerName: instance.TLSServerName, TLSInsecure: instance.TLSInsecure}
		}
	}
	if primaryVaultClient == nil {
		primaryVaultClient = vault.NewDisabledClient()
	}
	return allClients, primaryVaultClient
}

// shutdownVaultClients shuts down each unique client once.
func shutdownVaultClients(allClients map[string]vault.Client) {
	uniqueClients := make(map[vault.Client]struct{})
	for _, client := range allClients {
		if client == nil {
			continue
		}
		uniqueClients[client] = struct{}{}
	}
	for client := range uniqueClients {
		client.Shutdown()
	}
}

func main() {
	cfg, cfgErr := config.Load()
	if cfgErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", cfgErr)
		os.Exit(1)
	}

	// Initialize structured logger from config
	logger.Init(cfg.LogLevel)
	log := logger.Get()

	log.Info().
		Str("version", version.Version).
		Msg("VaultCertsViewer starting")

	log.Info().
		Str("env", string(cfg.Env)).
		Str("log_level", cfg.LogLevel).
		Str("log_format", cfg.LogFormat).
		Msg("Configuration loaded")
	allClients, primaryVaultClient := initVaultClients(&cfg, log)

	vaultRegistry := vault.NewRegistry(cfg.AllVaults)
	multiVaultClient := vault.NewMultiClient(cfg.AllVaults, allClients, vaultRegistry)

	log.Info().
		Str("vault_addr", cfg.Vault.Addr).
		Strs("vault_mounts", cfg.Vault.PKIMounts).
		Int("vault_instances_total", len(cfg.AllVaults)).
		Int("vault_instances_enabled", len(cfg.Vaults)).
		Msg("Vault client initialized")

	promRegistry := prometheus.NewRegistry()
	promRegistry.MustRegister(collectors.NewGoCollector())
	promRegistry.MustRegister(metrics.NewCertificateCollectorWithVaults(multiVaultClient, allClients, cfg.ExpirationThresholds, cfg.Metrics, cfg.AllVaults))

	webFS, fsError := fs.Sub(web.EmbeddedFS, ".")
	if fsError != nil {
		log.Fatal().Err(fsError).
			Msg("Failed to initialize embedded web filesystem")
	}

	settingsPath := cfg.SettingsPath

	log.Info().
		Str("settings_path", settingsPath).
		Msg("Using admin settings file")

	router, buildErr := buildRouter(routerDeps{
		cfg:           cfg,
		primaryVault:  primaryVaultClient,
		statusClients: allClients,
		multiVault:    multiVaultClient,
		vaultRegistry: vaultRegistry,
		promRegistry:  promRegistry,
		webFS:         webFS,
		settingsPath:  settingsPath,
	})
	if buildErr != nil {
		log.Fatal().Err(buildErr).
			Msg("Failed to initialize router")
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: serverReadHeaderTimeout,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    serverMaxHeaderBytes,
	}

	go func() {
		log.Info().Str("port", cfg.Port).Msg("Server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server error")
		}
	}()

	notifier := notify.New(multiVaultClient, config.Load)
	notifyCtx, notifyCancel := context.WithCancel(context.Background())
	go func() {
		// Check once shortly after startup so an already-crossed threshold
		// notifies promptly, then on a fixed interval.
		notifier.Check(notifyCtx)
		ticker := time.NewTicker(notifyCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				notifier.Check(notifyCtx)
			case <-notifyCtx.Done():
				return
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	notifyCancel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	shutdownVaultClients(allClients)

	log.Info().Msg("Server stopped")
}
