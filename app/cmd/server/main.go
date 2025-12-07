package main

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
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
	vaultClient, vaultError := vault.NewClientFromConfig(cfg.Vault)
	if vaultError != nil {
		log.Fatal().Err(vaultError).
			Msg("Failed to initialize Vault client")
	}

	log.Info().
		Str("vault_addr", cfg.Vault.Addr).
		Str("vault_mount", cfg.Vault.PKIMount).
		Msg("Vault client initialized")

	registry := prometheus.NewRegistry()
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(metrics.NewCertificateCollector(vaultClient))

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
		status := map[string]interface{}{
			"version": version.Version,
		}
		if err := vaultClient.CheckConnection(ctx); err != nil {
			status["vault_connected"] = false
			status["vault_error"] = err.Error()
		} else {
			status["vault_connected"] = true
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(status)
	})
	r.Get("/api/version", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(version.Info())
	})
	r.Get("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP)
	handlers.RegisterI18nRoutes(r)
	handlers.RegisterCertRoutes(r, vaultClient)

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

	// Shutdown Vault client (stops background goroutines)
	vaultClient.Shutdown()

	log.Info().Msg("Server stopped")
}
