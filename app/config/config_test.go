package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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

func TestConfig_IsDev_IsProd(t *testing.T) {
	tests := []struct {
		name       string
		env        Environment
		expectDev  bool
		expectProd bool
	}{
		{name: "dev", env: EnvDev, expectDev: true, expectProd: false},
		{name: "prod", env: EnvProd, expectDev: false, expectProd: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Env: tt.env}
			if cfg.IsDev() != tt.expectDev {
				t.Fatalf("unexpected IsDev result")
			}
			if cfg.IsProd() != tt.expectProd {
				t.Fatalf("unexpected IsProd result")
			}
		})
	}
}

func TestLoadSettingsFile_SettingsPathMissing(t *testing.T) {
	clearEnv(t)
	missingPath := filepath.Join(t.TempDir(), "missing.json")
	if err := os.Setenv("SETTINGS_PATH", missingPath); err != nil {
		t.Fatalf("failed to set SETTINGS_PATH: %v", err)
	}
	settings, settingsPath, err := loadSettingsFile()
	if err == nil {
		t.Fatalf("expected error")
	}
	if settings != nil {
		t.Fatalf("expected nil settings")
	}
	if settingsPath != missingPath {
		t.Fatalf("expected settingsPath %q, got %q", missingPath, settingsPath)
	}
}

func TestLoadSettingsFile_SettingsPathInvalidJSON(t *testing.T) {
	clearEnv(t)
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0o644); err != nil {
		t.Fatalf("failed to write settings file: %v", err)
	}
	if err := os.Setenv("SETTINGS_PATH", path); err != nil {
		t.Fatalf("failed to set SETTINGS_PATH: %v", err)
	}
	settings, settingsPath, err := loadSettingsFile()
	if err == nil {
		t.Fatalf("expected error")
	}
	if settings != nil {
		t.Fatalf("expected nil settings")
	}
	if settingsPath != path {
		t.Fatalf("expected settingsPath %q, got %q", path, settingsPath)
	}
}

func TestLoadSettingsFile_ErrorCases(t *testing.T) {
	// Test non-existent file by setting SETTINGS_PATH to non-existent path
	originalPath := os.Getenv("SETTINGS_PATH")
	defer func() {
		if originalPath != "" {
			_ = os.Setenv("SETTINGS_PATH", originalPath)
		} else {
			_ = os.Unsetenv("SETTINGS_PATH")
		}
	}()

	_ = os.Setenv("SETTINGS_PATH", "/non/existent/path.json")
	settings, settingsPath, err := loadSettingsFile()
	if err == nil {
		t.Fatalf("expected error for non-existent file, got nil")
	}
	if settings != nil {
		t.Fatalf("expected nil settings for error case")
	}
	if settingsPath != "/non/existent/path.json" {
		t.Fatalf("expected settingsPath to be the requested path even on error")
	}
}

func TestBuildConfigFromSettings_ErrorCases(t *testing.T) {
	clearEnv(t)

	// Test with empty settings
	emptySettings := SettingsFile{}
	cfg := buildConfigFromSettings(emptySettings)
	if cfg.Env == "" {
		t.Fatalf("expected env to have default value")
	}

	// Test with minimal settings
	minimalSettings := SettingsFile{
		App: AppSettings{
			Env:  "prod",
			Port: 8080,
		},
		Vaults: []VaultInstance{
			{ID: "", Address: ""}, // Invalid empty instance
		},
	}
	cfg = buildConfigFromSettings(minimalSettings)
	if cfg.Env != "prod" {
		t.Fatalf("expected prod env, got %s", cfg.Env)
	}
	if cfg.Port != "8080" {
		t.Fatalf("expected port 8080, got %s", cfg.Port)
	}
}

func TestLoadCORSConfig_EdgeCases(t *testing.T) {
	clearEnv(t)

	// Test with various CORS configurations
	testCases := []struct {
		name            string
		allowedOrigins  string
		allowCreds      string
		expectedOrigins int
	}{
		{
			name:            "single origin",
			allowedOrigins:  "http://localhost:3000",
			allowCreds:      "true",
			expectedOrigins: 1,
		},
		{
			name:            "multiple origins",
			allowedOrigins:  "http://localhost:3000,https://example.com",
			allowCreds:      "false",
			expectedOrigins: 2,
		},
		{
			name:            "empty origins",
			allowedOrigins:  "",
			allowCreds:      "true",
			expectedOrigins: 2, // Default origins are applied when empty
		},
		{
			name:            "comma separated with spaces",
			allowedOrigins:  "http://localhost:3000, https://example.com ,http://test.com",
			allowCreds:      "true",
			expectedOrigins: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_ = os.Setenv("CORS_ALLOWED_ORIGINS", tc.allowedOrigins)
			_ = os.Setenv("CORS_ALLOW_CREDENTIALS", tc.allowCreds)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(cfg.CORS.AllowedOrigins) != tc.expectedOrigins {
				t.Fatalf("expected %d origins, got %d", tc.expectedOrigins, len(cfg.CORS.AllowedOrigins))
			}
		})
	}
}

func TestConvertVaultInstanceToLegacy_EdgeCases(t *testing.T) {
	// Test with various vault instances
	testCases := []struct {
		name     string
		instance VaultInstance
		expected VaultConfig
	}{
		{
			name: "minimal instance",
			instance: VaultInstance{
				ID:      "test",
				Address: "http://vault:8200",
			},
			expected: VaultConfig{
				Addr:            "http://vault:8200",
				PKIMounts:       []string{"pki"}, // Default PKI mount is applied
				ReadToken:       "",
				TLSCACertBase64: "",
				TLSCACert:       "",
				TLSCAPath:       "",
				TLSServerName:   "",
				TLSInsecure:     false,
			},
		},
		{
			name: "full instance",
			instance: VaultInstance{
				ID:              "full",
				Address:         "https://vault:8200",
				Token:           "root",
				PKIMounts:       []string{"pki1", "pki2"},
				TLSInsecure:     true,
				TLSCACertBase64: "base64cert",
				TLSCACert:       "certdata",
				TLSCAPath:       "/path/to/ca",
				TLSServerName:   "vault.example.com",
			},
			expected: VaultConfig{
				Addr:            "https://vault:8200",
				PKIMounts:       []string{"pki1", "pki2"},
				ReadToken:       "root", // Token is copied to ReadToken
				TLSCACertBase64: "base64cert",
				TLSCACert:       "certdata",
				TLSCAPath:       "/path/to/ca",
				TLSServerName:   "vault.example.com",
				TLSInsecure:     true,
			},
		},
		{
			name: "instance with single PKI mount",
			instance: VaultInstance{
				ID:       "single-pki",
				Address:  "http://vault:8200",
				PKIMount: "pki",
			},
			expected: VaultConfig{
				Addr:            "http://vault:8200",
				PKIMounts:       []string{"pki"},
				ReadToken:       "",
				TLSCACertBase64: "",
				TLSCACert:       "",
				TLSCAPath:       "",
				TLSServerName:   "",
				TLSInsecure:     false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := convertVaultInstanceToLegacy(tc.instance)
			if result.Addr != tc.expected.Addr {
				t.Fatalf("expected addr %q, got %q", tc.expected.Addr, result.Addr)
			}
			if len(result.PKIMounts) != len(tc.expected.PKIMounts) {
				t.Fatalf("expected %d PKI mounts, got %d", len(tc.expected.PKIMounts), len(result.PKIMounts))
			}
			if result.ReadToken != tc.expected.ReadToken {
				t.Fatalf("expected read token %q, got %q", tc.expected.ReadToken, result.ReadToken)
			}
			if result.TLSInsecure != tc.expected.TLSInsecure {
				t.Fatalf("expected TLS insecure %v, got %v", tc.expected.TLSInsecure, result.TLSInsecure)
			}
			if result.TLSCACertBase64 != tc.expected.TLSCACertBase64 {
				t.Fatalf("expected TLS CA cert base64 %q, got %q", tc.expected.TLSCACertBase64, result.TLSCACertBase64)
			}
			if result.TLSCACert != tc.expected.TLSCACert {
				t.Fatalf("expected TLS CA cert %q, got %q", tc.expected.TLSCACert, result.TLSCACert)
			}
			if result.TLSCAPath != tc.expected.TLSCAPath {
				t.Fatalf("expected TLS CA path %q, got %q", tc.expected.TLSCAPath, result.TLSCAPath)
			}
			if result.TLSServerName != tc.expected.TLSServerName {
				t.Fatalf("expected TLS server name %q, got %q", tc.expected.TLSServerName, result.TLSServerName)
			}
		})
	}
}
