package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"vcv/internal/metrics"

	"vcv/config"
	"vcv/internal/handlers"
	"vcv/internal/logger"
	"vcv/internal/vault"
	"vcv/internal/version"
	"vcv/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.Load()

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

	r := chi.NewRouter()
	primaryVaultClient, vaultError := vault.NewClientFromConfig(cfg.Vault)
	if vaultError != nil {
		log.Fatal().Err(vaultError).
			Msg("Failed to initialize Vault client")
	}

	statusClients := make(map[string]vault.Client, len(cfg.Vaults))
	primaryID := ""
	if len(cfg.Vaults) > 0 {
		primaryID = cfg.Vaults[0].ID
	}
	for _, instance := range cfg.Vaults {
		if instance.ID == "" {
			continue
		}
		if primaryID != "" && instance.ID == primaryID {
			statusClients[instance.ID] = primaryVaultClient
			continue
		}
		statusCfg := config.VaultConfig{Addr: instance.Address, PKIMounts: instance.PKIMounts, ReadToken: instance.Token, TLSInsecure: instance.TLSInsecure}
		client, err := vault.NewClientFromConfig(statusCfg)
		if err != nil {
			log.Fatal().Err(err).
				Str("vault_id", instance.ID).
				Msg("Failed to initialize Vault status client")
		}
		statusClients[instance.ID] = client
	}

	multiVaultClient := vault.NewMultiClient(cfg.Vaults, statusClients)

	log.Info().
		Str("vault_addr", cfg.Vault.Addr).
		Strs("vault_mounts", cfg.Vault.PKIMounts).
		Msg("Vault client initialized")

	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(metrics.NewCertificateCollectorWithVaults(multiVaultClient, statusClients, cfg.ExpirationThresholds, cfg.Vaults))

	webFS, fsError := fs.Sub(embeddedWeb, "web")
	if fsError != nil {
		log.Fatal().Err(fsError).
			Msg("Failed to initialize embedded web filesystem")
	}
	assetsFS, assetsError := fs.Sub(webFS, "assets")
	if assetsError != nil {
		log.Fatal().Err(assetsError).
			Msg("Failed to initialize embedded assets filesystem")
	}

	settingsPath := strings.TrimSpace(os.Getenv("SETTINGS_PATH"))
	if settingsPath == "" {
		candidates := []string{fmt.Sprintf("settings.%s.json", string(cfg.Env)), "settings.json", "./settings.json", "/etc/vcv/settings.json"}
		for _, candidate := range candidates {
			absPath, absErr := filepath.Abs(candidate)
			if absErr != nil {
				continue
			}
			if _, statErr := os.Stat(absPath); statErr != nil {
				continue
			}
			settingsPath = absPath
			break
		}
		if settingsPath == "" {
			settingsPath = filepath.Join(".", fmt.Sprintf("settings.%s.json", string(cfg.Env)))
		}
	}

	// Middleware must be registered before any routes
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SecurityHeaders)

	// Static frontend from embedded filesystem
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		data, readError := fs.ReadFile(webFS, "index.html")
		if readError != nil {
			log.Error().Err(readError).
				Str("path", "/").
				Msg("Failed to read embedded index.html")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	})
	staticHandler := http.StripPrefix("/assets/", http.FileServer(http.FS(assetsFS)))
	r.Handle("/assets/*", staticHandler)

	// Health and readiness probes
	r.Get("/api/health", handlers.HealthCheck)
	r.Get("/api/ready", handlers.ReadinessCheck)
	r.Get("/api/status", func(w http.ResponseWriter, req *http.Request) {
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
	})
	r.Get("/api/version", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(version.Info())
	})
	r.Get("/api/config", handlers.GetConfig(cfg))
	r.Get("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP)
	handlers.RegisterI18nRoutes(r)
	handlers.RegisterCertRoutes(r, multiVaultClient)
	handlers.RegisterUIRoutes(r, multiVaultClient, cfg.Vaults, statusClients, webFS, cfg.ExpirationThresholds)
	handlers.RegisterAdminRoutes(r, webFS, settingsPath, cfg.Env)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
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
	for _, client := range statusClients {
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
