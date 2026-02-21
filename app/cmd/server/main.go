package main

import (
	"context"
	"encoding/json"
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
const routerRateLimitWindow time.Duration = 1 * time.Minute

func newStatusHandler(cfg config.Config, primaryVaultClient vault.Client, statusClients map[string]vault.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		type vaultStatusEntry struct {
			ID          string `json:"id"`
			DisplayName string `json:"display_name"`
			Connected   bool   `json:"connected"`
			Error       string `json:"error,omitempty"`
		}
		type statusResponse struct {
			Version        string             `json:"version"`
			VaultConnected bool               `json:"vault_connected"`
			VaultError     string             `json:"vault_error,omitempty"`
			Vaults         []vaultStatusEntry `json:"vaults"`
		}
		response := statusResponse{Version: version.Version, Vaults: make([]vaultStatusEntry, 0, len(cfg.Vaults))}
		if err := primaryVaultClient.CheckConnection(ctx); err != nil {
			response.VaultConnected = false
			response.VaultError = err.Error()
		} else {
			response.VaultConnected = true
		}
		for _, instance := range cfg.Vaults {
			name := instance.DisplayName
			client := statusClients[instance.ID]
			entry := vaultStatusEntry{ID: instance.ID, DisplayName: name}
			if client == nil {
				entry.Connected = false
				entry.Error = "missing vault status client"
				response.Vaults = append(response.Vaults, entry)
				continue
			}
			if err := client.CheckConnection(ctx); err != nil {
				entry.Connected = false
				entry.Error = err.Error()
			} else {
				entry.Connected = true
			}
			response.Vaults = append(response.Vaults, entry)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}

func buildRouter(cfg config.Config, primaryVaultClient vault.Client, statusClients map[string]vault.Client, multiVaultClient vault.Client, registry *prometheus.Registry, webFS fs.FS, settingsPath string, vaultRegistry *vault.Registry) (*chi.Mux, error) {
	r := chi.NewRouter()
	assetsFS, assetsError := fs.Sub(webFS, "assets")
	if assetsError != nil {
		return nil, assetsError
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
	if cfg.Env == config.EnvProd {
		rateLimitConfig := middleware.DefaultRateLimitConfig()
		rateLimitConfig.MaxRequests = routerRateLimitMaxRequests
		rateLimitConfig.Window = routerRateLimitWindow
		rateLimitConfig.ExemptPaths = []string{"/api/health", "/api/ready", "/metrics"}
		rateLimitConfig.ExemptPathPrefixes = []string{"/assets/"}
		r.Use(middleware.RateLimit(rateLimitConfig))
	}
	r.Use(middleware.BodyLimit(routerMaxBodyBytes))
	r.Use(middleware.CSRFProtection)

	staticHandler := http.StripPrefix("/assets/", http.FileServer(http.FS(assetsFS)))
	r.Handle("/assets/*", staticHandler)

	// Health and readiness probes
	r.Get("/api/health", handlers.HealthCheck)
	r.Get("/api/ready", handlers.ReadinessCheck)
	r.Get("/api/status", newStatusHandler(cfg, primaryVaultClient, statusClients))
	r.Get("/api/version", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(version.Info())
	})
	r.Get("/api/config", handlers.GetConfig(cfg, vaultRegistry))
	r.Get("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP)
	handlers.RegisterI18nRoutes(r)
	handlers.RegisterCertRoutes(r, multiVaultClient)
	handlers.RegisterUIRoutes(r, multiVaultClient, cfg.AllVaults, statusClients, webFS, cfg.ExpirationThresholds, vaultRegistry)
	handlers.RegisterAdminRoutes(r, webFS, settingsPath, cfg.Env, vaultRegistry, statusClients)

	return r, nil
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
	// Create clients for ALL vaults (including disabled) so they can be
	// toggled at runtime via the admin panel without a restart.
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

	router, buildErr := buildRouter(cfg, primaryVaultClient, allClients, multiVaultClient, promRegistry, webFS, settingsPath, vaultRegistry)
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

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

	log.Info().Msg("Server stopped")
}
