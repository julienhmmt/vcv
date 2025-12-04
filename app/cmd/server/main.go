package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vcv/config"
	h "vcv/internal/handlers"
	"vcv/internal/logger"
	"vcv/internal/vault"
	"vcv/middleware"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.Load()

	// Initialize structured logger from config
	logger.Init(cfg.LogLevel)
	log := logger.Get()

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

	// Middleware must be registered before any routes
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CORS(middleware.CORSConfig{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           86400,
	}))

	// Static frontend
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "./web/index.html")
	})
	staticHandler := http.StripPrefix("/assets/", http.FileServer(http.Dir("./web/assets")))
	r.Handle("/assets/*", staticHandler)

	// Health and readiness probes
	r.Get("/api/health", h.HealthCheck)
	r.Get("/api/ready", h.ReadinessCheck)
	h.RegisterCertRoutes(r, vaultClient, cfg.Vault.EnableRevoke)

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
	log.Info().Msg("Server stopped")
}
