package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)
	cfg := Load()
	if cfg.Port == "" {
		t.Fatalf("expected default port, got empty")
	}
	if cfg.Env != EnvDev {
		t.Fatalf("expected dev env by default, got %s", cfg.Env)
	}
	if cfg.LogLevel == "" || cfg.LogFormat == "" {
		t.Fatalf("expected log defaults")
	}
}

func TestLoadFromEnv(t *testing.T) {
	clearEnv(t)
	_ = os.Setenv("APP_ENV", "prod")
	_ = os.Setenv("PORT", "1234")
	_ = os.Setenv("LOG_LEVEL", "warn")
	_ = os.Setenv("LOG_FORMAT", "json")
	_ = os.Setenv("LOG_OUTPUT", "stdout")
	_ = os.Setenv("VAULT_ADDR", "http://vault")
	_ = os.Setenv("VAULT_PKI_MOUNT", "pki")
	_ = os.Setenv("VAULT_READ_TOKEN", "token")
	_ = os.Setenv("VAULT_TLS_INSECURE", "true")
	_ = os.Setenv("CORS_ALLOWED_ORIGINS", "http://example.com")
	_ = os.Setenv("CORS_ALLOW_CREDENTIALS", "true")

	cfg := Load()

	if cfg.Env != EnvProd {
		t.Fatalf("expected prod env, got %s", cfg.Env)
	}
	if cfg.Port != "1234" {
		t.Fatalf("expected port 1234, got %s", cfg.Port)
	}
	if cfg.LogLevel != "warn" || cfg.LogFormat != "json" || cfg.LogOutput != "stdout" {
		t.Fatalf("expected log env values to be applied")
	}
	if cfg.Vault.Addr != "http://vault" || len(cfg.Vault.PKIMounts) != 1 || cfg.Vault.PKIMounts[0] != "pki" || cfg.Vault.ReadToken != "token" || !cfg.Vault.TLSInsecure {
		t.Fatalf("expected vault env values to be applied")
	}
	if len(cfg.CORS.AllowedOrigins) != 1 || cfg.CORS.AllowedOrigins[0] != "http://example.com" || !cfg.CORS.AllowCredentials {
		t.Fatalf("expected cors env values to be applied")
	}
}

func clearEnv(t *testing.T) {
	t.Helper()
	envs := []string{
		"APP_ENV", "PORT", "LOG_LEVEL", "LOG_FORMAT", "LOG_OUTPUT",
		"VAULT_ADDR", "VAULT_PKI_MOUNT", "VAULT_READ_TOKEN", "VAULT_TLS_INSECURE",
		"CORS_ALLOWED_ORIGINS", "CORS_ALLOW_CREDENTIALS",
	}
	for _, key := range envs {
		_ = os.Unsetenv(key)
	}
}
