package config

import (
	"os"
	"path/filepath"
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

func TestLoadFromVaultAddrsEnv(t *testing.T) {
	clearEnv(t)
	_ = os.Setenv("VAULT_ADDR", "http://legacy")
	_ = os.Setenv("VAULT_READ_TOKEN", "legacy-token")
	_ = os.Setenv("VAULT_ADDRS", "prod@http://vault-1:8200#token1#pki,stg@http://vault-2:8200#token2#pki2")
	_ = os.Setenv("VAULT_TLS_INSECURE", "true")

	cfg := Load()

	if len(cfg.Vaults) != 2 {
		t.Fatalf("expected 2 vaults, got %d", len(cfg.Vaults))
	}
	if cfg.Vault.Addr != "http://vault-1:8200" {
		t.Fatalf("expected primary vault addr to be set from VAULT_ADDRS, got %s", cfg.Vault.Addr)
	}
}

func TestLoadFromSettingsFile(t *testing.T) {
	clearEnv(t)
	tempDir := t.TempDir()
	changeWorkingDirectory(t, tempDir)
	_ = os.Setenv("APP_ENV", "dev")

	settings := `{
  "app": {
    "env": "dev",
    "port": 52000,
    "logging": {
      "level": "debug",
      "format": "json",
      "output": "both",
      "file_path": "/var/log/app/vcv.log"
    }
  },
  "cors": {
    "allowed_origins": ["http://localhost:4321", "http://localhost:3000"],
    "allow_credentials": true
  },
  "certificates": {
    "expiration_thresholds": {
      "critical": 2,
      "warning": 10
    }
  },
  "vaults": [
    {
      "id": "dev",
      "enabled": true,
      "address": "http://vault:8200",
      "token": "root",
      "pki_mount": "pki",
      "pki_mounts": ["pki", "pki_dev"],
      "tls_insecure": true
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(tempDir, "settings.dev.json"), []byte(settings), 0o644); err != nil {
		t.Fatalf("failed to write settings.dev.json: %v", err)
	}

	cfg := Load()

	if cfg.Env != EnvDev {
		t.Fatalf("expected env to be dev, got %s", cfg.Env)
	}
	if cfg.Port != "52000" {
		t.Fatalf("expected port 52000, got %s", cfg.Port)
	}
	if cfg.LogLevel != "debug" || cfg.LogFormat != "json" || cfg.LogOutput != "both" {
		t.Fatalf("expected logging settings from file, got level=%s format=%s output=%s", cfg.LogLevel, cfg.LogFormat, cfg.LogOutput)
	}
	if cfg.LogFilePath != "/var/log/app/vcv.log" {
		t.Fatalf("expected log file path from file, got %s", cfg.LogFilePath)
	}
	if os.Getenv("LOG_OUTPUT") != "both" || os.Getenv("LOG_FORMAT") != "json" || os.Getenv("LOG_FILE_PATH") != "/var/log/app/vcv.log" {
		t.Fatalf("expected logging env vars to be applied from file")
	}
	if cfg.ExpirationThresholds.Critical != 2 || cfg.ExpirationThresholds.Warning != 10 {
		t.Fatalf("expected expiration thresholds from file")
	}
	if len(cfg.Vaults) != 1 {
		t.Fatalf("expected 1 vault from file, got %d", len(cfg.Vaults))
	}
	if cfg.Vault.Addr != "http://vault:8200" {
		t.Fatalf("expected primary vault addr from file, got %s", cfg.Vault.Addr)
	}
	if len(cfg.Vault.PKIMounts) != 2 {
		t.Fatalf("expected PKI mounts from file, got %d", len(cfg.Vault.PKIMounts))
	}
}

func TestLoadFromSettingsFile_PrimarySkipsDisabledVaults(t *testing.T) {
	clearEnv(t)
	tempDir := t.TempDir()
	changeWorkingDirectory(t, tempDir)
	_ = os.Setenv("APP_ENV", "dev")

	settings := `{
  "app": {"env": "dev", "port": 52000, "logging": {"level": "debug", "format": "json", "output": "stdout", "file_path": ""}},
  "cors": {"allowed_origins": [], "allow_credentials": true},
  "certificates": {"expiration_thresholds": {"critical": 2, "warning": 10}},
  "vaults": [
    {"id": "disabled", "enabled": false, "address": "://bad", "token": "tok"},
    {"id": "enabled", "enabled": true, "address": "http://vault:8200", "token": "root", "pki_mount": "pki", "tls_insecure": true}
  ]
}`
	if err := os.WriteFile(filepath.Join(tempDir, "settings.dev.json"), []byte(settings), 0o644); err != nil {
		t.Fatalf("failed to write settings.dev.json: %v", err)
	}

	cfg := Load()
	if len(cfg.Vaults) != 1 {
		t.Fatalf("expected 1 enabled vault after filtering, got %d", len(cfg.Vaults))
	}
	if cfg.Vault.Addr != "http://vault:8200" {
		t.Fatalf("expected primary vault to be first enabled, got %s", cfg.Vault.Addr)
	}
}

func TestLoadFromSettingsFile_AllDisabledDoesNotPanic(t *testing.T) {
	clearEnv(t)
	tempDir := t.TempDir()
	changeWorkingDirectory(t, tempDir)
	_ = os.Setenv("APP_ENV", "dev")

	settings := `{
  "app": {"env": "dev", "port": 52000, "logging": {"level": "debug", "format": "json", "output": "stdout", "file_path": ""}},
  "cors": {"allowed_origins": [], "allow_credentials": true},
  "certificates": {"expiration_thresholds": {"critical": 2, "warning": 10}},
  "vaults": [
    {"id": "disabled", "enabled": false, "address": "http://vault:8200", "token": "root", "pki_mount": "pki", "tls_insecure": true}
  ]
}`
	if err := os.WriteFile(filepath.Join(tempDir, "settings.dev.json"), []byte(settings), 0o644); err != nil {
		t.Fatalf("failed to write settings.dev.json: %v", err)
	}

	cfg := Load()
	if len(cfg.Vaults) != 0 {
		t.Fatalf("expected 0 enabled vaults, got %d", len(cfg.Vaults))
	}
	if cfg.Vault.Addr != "" {
		t.Fatalf("expected empty primary vault when all vaults are disabled")
	}
}

func clearEnv(t *testing.T) {
	t.Helper()
	envs := []string{
		"APP_ENV", "PORT", "LOG_LEVEL", "LOG_FORMAT", "LOG_OUTPUT",
		"SETTINGS_PATH",
		"LOG_FILE_PATH",
		"VAULT_ADDR", "VAULT_ADDRS", "VAULT_PKI_MOUNT", "VAULT_PKI_MOUNTS", "VAULT_READ_TOKEN", "VAULT_TLS_INSECURE",
		"CORS_ALLOWED_ORIGINS", "CORS_ALLOW_CREDENTIALS",
	}
	for _, key := range envs {
		_ = os.Unsetenv(key)
	}
}
