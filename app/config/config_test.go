package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Create a minimal settings file for testing
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")
	settingsContent := `{
		"app": {
			"env": "dev",
			"port": 52000,
			"logging": {
				"level": "debug",
				"format": "console",
				"output": "stdout"
			}
		},
		"vaults": []
	}`

	err := os.WriteFile(settingsFile, []byte(settingsContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test settings file: %v", err)
	}

	// Change to temp directory to ensure the settings file is found
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

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

func TestLoadFromSettingsFile(t *testing.T) {
	tmpDir := t.TempDir()
	settingsFile := filepath.Join(tmpDir, "settings.json")
	settingsContent := `{
		"app": {
			"env": "prod",
			"port": 1234,
			"logging": {
				"level": "warn",
				"format": "json",
				"output": "stdout"
			}
		},
		"cors": {
			"allowed_origins": ["http://example.com"],
			"allow_credentials": true
		},
		"vaults": [
			{
				"id": "test-vault",
				"address": "http://vault:8200",
				"token": "test-token",
				"pki_mount": "pki"
			}
		],
		"certificates": {
			"expiration_thresholds": {
				"critical": 14,
				"warning": 60
			}
		},
		"metrics": {
			"per_certificate": true,
			"enhanced_metrics": false
		}
	}`

	err := os.WriteFile(settingsFile, []byte(settingsContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test settings file: %v", err)
	}

	// Change to temp directory to ensure the settings file is found
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

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
	if cfg.LogLevel != "warn" {
		t.Fatalf("expected log level warn, got %s", cfg.LogLevel)
	}
	if cfg.LogFormat != "json" {
		t.Fatalf("expected log format json, got %s", cfg.LogFormat)
	}
	if len(cfg.CORS.AllowedOrigins) != 1 || cfg.CORS.AllowedOrigins[0] != "http://example.com" {
		t.Fatalf("expected CORS origin http://example.com, got %v", cfg.CORS.AllowedOrigins)
	}
	if !cfg.CORS.AllowCredentials {
		t.Fatalf("expected CORS allow credentials true, got %v", cfg.CORS.AllowCredentials)
	}
	if len(cfg.Vaults) != 1 {
		t.Fatalf("expected 1 vault, got %d", len(cfg.Vaults))
	}
	if cfg.Vaults[0].ID != "test-vault" {
		t.Fatalf("expected vault ID test-vault, got %s", cfg.Vaults[0].ID)
	}
	if cfg.ExpirationThresholds.Critical != 14 {
		t.Fatalf("expected critical threshold 14, got %d", cfg.ExpirationThresholds.Critical)
	}
	if cfg.ExpirationThresholds.Warning != 60 {
		t.Fatalf("expected warning threshold 60, got %d", cfg.ExpirationThresholds.Warning)
	}
	if !cfg.Metrics.PerCertificate {
		t.Fatalf("expected per certificate metrics true, got %v", cfg.Metrics.PerCertificate)
	}
	if cfg.Metrics.EnhancedMetrics {
		t.Fatalf("expected enhanced metrics false, got %v", cfg.Metrics.EnhancedMetrics)
	}
}

func TestLoadSettingsFile_MissingFile(t *testing.T) {
	// Change to temp directory with no settings file
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalWd) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	_, err = Load()
	if err == nil {
		t.Fatalf("expected error for missing settings file, got none")
	}
}

func TestBuildConfigFromSettings(t *testing.T) {
	settings := SettingsFile{
		App: AppSettings{
			Env:  "prod",
			Port: 1234,
			Logging: LoggingSettings{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
		},
		CORS: CORSSettings{
			AllowedOrigins:   []string{"http://example.com"},
			AllowCredentials: true,
		},
		Certificates: CertificateSettings{
			ExpirationThresholds: ExpirationThresholds{
				Critical: 14,
				Warning:  60,
			},
		},
		Metrics: MetricsSettings{
			PerCertificate:  &[]bool{true}[0],
			EnhancedMetrics: &[]bool{false}[0],
		},
	}

	cfg := buildConfigFromSettings(settings)

	if cfg.Env != EnvProd {
		t.Fatalf("expected prod env, got %s", cfg.Env)
	}
	if cfg.Port != "1234" {
		t.Fatalf("expected port 1234, got %s", cfg.Port)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("expected log level info, got %s", cfg.LogLevel)
	}
	if cfg.LogFormat != "json" {
		t.Fatalf("expected log format json, got %s", cfg.LogFormat)
	}
	if len(cfg.CORS.AllowedOrigins) != 1 || cfg.CORS.AllowedOrigins[0] != "http://example.com" {
		t.Fatalf("expected CORS origin http://example.com, got %v", cfg.CORS.AllowedOrigins)
	}
	if !cfg.CORS.AllowCredentials {
		t.Fatalf("expected CORS allow credentials true, got %v", cfg.CORS.AllowCredentials)
	}
	if cfg.ExpirationThresholds.Critical != 14 {
		t.Fatalf("expected critical threshold 14, got %d", cfg.ExpirationThresholds.Critical)
	}
	if cfg.ExpirationThresholds.Warning != 60 {
		t.Fatalf("expected warning threshold 60, got %d", cfg.ExpirationThresholds.Warning)
	}
	if !cfg.Metrics.PerCertificate {
		t.Fatalf("expected per certificate metrics true, got %v", cfg.Metrics.PerCertificate)
	}
	if cfg.Metrics.EnhancedMetrics {
		t.Fatalf("expected enhanced metrics false, got %v", cfg.Metrics.EnhancedMetrics)
	}
}

func TestParseEnv(t *testing.T) {
	tests := []struct {
		input    string
		expected Environment
	}{
		{"dev", EnvDev},
		{"development", EnvDev},
		{"DEV", EnvDev},
		{"prod", EnvProd},
		{"production", EnvProd},
		{"PROD", EnvProd},
		{"", EnvDev},
		{"invalid", EnvDev},
	}

	for _, tt := range tests {
		result := parseEnv(tt.input)
		if result != tt.expected {
			t.Fatalf("expected %v for input %q, got %v", tt.expected, tt.input, result)
		}
	}
}

func TestDefaultLogLevel(t *testing.T) {
	if defaultLogLevel(EnvDev) != "debug" {
		t.Fatalf("expected debug log level for dev, got %s", defaultLogLevel(EnvDev))
	}
	if defaultLogLevel(EnvProd) != "info" {
		t.Fatalf("expected info log level for prod, got %s", defaultLogLevel(EnvProd))
	}
}

func TestDefaultLogFormat(t *testing.T) {
	if defaultLogFormat(EnvDev) != "console" {
		t.Fatalf("expected console log format for dev, got %s", defaultLogFormat(EnvDev))
	}
	if defaultLogFormat(EnvProd) != "json" {
		t.Fatalf("expected json log format for prod, got %s", defaultLogFormat(EnvProd))
	}
}

func TestLoadCORSConfig(t *testing.T) {
	devConfig := loadCORSConfig(EnvDev)
	if len(devConfig.AllowedOrigins) != 2 {
		t.Fatalf("expected 2 default origins for dev, got %d", len(devConfig.AllowedOrigins))
	}
	if !devConfig.AllowCredentials {
		t.Fatalf("expected allow credentials true for dev, got %v", devConfig.AllowCredentials)
	}

	prodConfig := loadCORSConfig(EnvProd)
	if len(prodConfig.AllowedOrigins) != 0 {
		t.Fatalf("expected no origins for prod, got %v", prodConfig.AllowedOrigins)
	}
	if !prodConfig.AllowCredentials {
		t.Fatalf("expected allow credentials true for prod, got %v", prodConfig.AllowCredentials)
	}
}

func TestIsDevIsProd(t *testing.T) {
	devConfig := Config{Env: EnvDev}
	if !devConfig.IsDev() {
		t.Fatalf("expected IsDev true for dev env")
	}
	if devConfig.IsProd() {
		t.Fatalf("expected IsProd false for dev env")
	}

	prodConfig := Config{Env: EnvProd}
	if prodConfig.IsDev() {
		t.Fatalf("expected IsDev false for prod env")
	}
	if !prodConfig.IsProd() {
		t.Fatalf("expected IsProd true for prod env")
	}
}
