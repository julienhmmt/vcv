package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Environment represents the application environment.
type Environment string

const (
	EnvDev  Environment = "dev"
	EnvProd Environment = "prod"
)

// Config holds application configuration.
type Config struct {
	Env                  Environment
	Port                 string
	LogLevel             string
	LogFormat            string
	LogOutput            string
	LogFilePath          string
	SettingsPath         string
	CORS                 CORSConfig
	Vault                VaultConfig
	Vaults               []VaultInstance
	AllVaults            []VaultInstance
	ExpirationThresholds ExpirationThresholds
	Metrics              MetricsConfig
}

// CORSConfig holds CORS-specific configuration.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowCredentials bool
}

type VaultConfig struct {
	Addr            string
	PKIMounts       []string
	ReadToken       string
	TLSCACertBase64 string
	TLSCACert       string
	TLSCAPath       string
	TLSServerName   string
	TLSInsecure     bool
}

// ExpirationThresholds holds certificate expiration alert thresholds (in days).
type ExpirationThresholds struct {
	Critical int `json:"critical"`
	Warning  int `json:"warning"`
}

// MetricsConfig holds metrics collection configuration.
type MetricsConfig struct {
	PerCertificate  bool `json:"per_certificate"`
	EnhancedMetrics bool `json:"enhanced_metrics"`
}

type SettingsFile struct {
	App          AppSettings         `json:"app"`
	Admin        AdminSettings       `json:"admin,omitempty"`
	Certificates CertificateSettings `json:"certificates"`
	Metrics      MetricsSettings     `json:"metrics"`
	CORS         CORSSettings        `json:"cors"`
	Vaults       []VaultInstance     `json:"vaults"`
}

type AppSettings struct {
	Env     string          `json:"env"`
	Logging LoggingSettings `json:"logging"`
	Port    int             `json:"port"`
}

type LoggingSettings struct {
	Level    string `json:"level"`
	Format   string `json:"format"`
	Output   string `json:"output"`
	FilePath string `json:"file_path"`
}

type CORSSettings struct {
	AllowedOrigins   []string `json:"allowed_origins"`
	AllowCredentials bool     `json:"allow_credentials"`
}

type CertificateSettings struct {
	ExpirationThresholds ExpirationThresholds `json:"expiration_thresholds"`
}

type MetricsSettings struct {
	PerCertificate  *bool `json:"per_certificate"`
	EnhancedMetrics *bool `json:"enhanced_metrics"`
}

type AdminSettings struct {
	Password string `json:"password,omitempty"`
}

// Load reads configuration from settings file only.
func Load() (Config, error) {
	settings, settingsPath, settingsErr := loadSettingsFile()
	if settingsErr != nil {
		return Config{}, fmt.Errorf("failed to load settings file: %w", settingsErr)
	}
	if settings == nil {
		return Config{}, fmt.Errorf("no settings file found")
	}

	cfg := buildConfigFromSettings(*settings)
	cfg.SettingsPath = settingsPath
	allVaults, allNormalizeErr := NormalizeAllVaultInstances(settings.Vaults)
	if allNormalizeErr != nil {
		return Config{}, fmt.Errorf("invalid settings file %s: %w", settingsPath, allNormalizeErr)
	}
	cfg.AllVaults = allVaults
	vaults, normalizeErr := normalizeVaultInstances(settings.Vaults)
	if normalizeErr != nil {
		return Config{}, fmt.Errorf("invalid settings file %s: %w", settingsPath, normalizeErr)
	}
	cfg.Vaults = vaults
	if len(vaults) > 0 {
		cfg.Vault = convertVaultInstanceToLegacy(vaults[0])
	}

	return cfg, nil
}

// settingsCandidates returns the ordered list of settings file paths to try.
func settingsCandidates() []string {
	return []string{"settings.dev.json", "settings.prod.json", "settings.json", "./settings.json", "/app/settings.json"}
}

// ResolveSettingsPath returns the absolute path of the first settings file found on disk.
// If no file is found it returns a default path.
func ResolveSettingsPath() string {
	for _, candidate := range settingsCandidates() {
		absPath, absErr := filepath.Abs(candidate)
		if absErr != nil {
			continue
		}
		if _, statErr := os.Stat(absPath); statErr != nil {
			continue
		}
		return absPath
	}
	absDefault, _ := filepath.Abs("settings.json")
	return absDefault
}

func loadSettingsFile() (*SettingsFile, string, error) {
	for _, candidate := range settingsCandidates() {
		absPath, absErr := filepath.Abs(candidate)
		if absErr != nil {
			continue
		}
		if _, statErr := os.Stat(absPath); statErr != nil {
			continue
		}
		data, readErr := os.ReadFile(absPath)
		if readErr != nil {
			return nil, absPath, readErr
		}
		var settings SettingsFile
		if err := json.Unmarshal(data, &settings); err != nil {
			return nil, absPath, err
		}
		return &settings, absPath, nil
	}
	return nil, "", os.ErrNotExist
}

func buildConfigFromSettings(settings SettingsFile) Config {
	envValue := strings.TrimSpace(settings.App.Env)
	if envValue == "" {
		envValue = "dev"
	}
	env := parseEnv(envValue)
	port := "52000"
	if settings.App.Port > 0 {
		port = strconv.Itoa(settings.App.Port)
	}
	logLevel := strings.TrimSpace(settings.App.Logging.Level)
	if logLevel == "" {
		logLevel = defaultLogLevel(env)
	}
	logFormat := strings.TrimSpace(settings.App.Logging.Format)
	if logFormat == "" {
		logFormat = defaultLogFormat(env)
	}
	logOutput := strings.TrimSpace(settings.App.Logging.Output)
	if logOutput == "" {
		logOutput = "stdout"
	}
	logFilePath := strings.TrimSpace(settings.App.Logging.FilePath)
	cors := loadCORSConfig(env)
	if len(settings.CORS.AllowedOrigins) > 0 {
		cors.AllowedOrigins = settings.CORS.AllowedOrigins
		cors.AllowCredentials = settings.CORS.AllowCredentials
	}
	expirations := ExpirationThresholds{Critical: 7, Warning: 30}
	if settings.Certificates.ExpirationThresholds.Critical > 0 {
		expirations.Critical = settings.Certificates.ExpirationThresholds.Critical
	}
	if settings.Certificates.ExpirationThresholds.Warning > 0 {
		expirations.Warning = settings.Certificates.ExpirationThresholds.Warning
	}
	metrics := MetricsConfig{PerCertificate: false, EnhancedMetrics: true}
	// Check if metrics section exists in settings (non-nil pointers)
	if settings.Metrics.PerCertificate != nil {
		metrics.PerCertificate = *settings.Metrics.PerCertificate
	}
	if settings.Metrics.EnhancedMetrics != nil {
		metrics.EnhancedMetrics = *settings.Metrics.EnhancedMetrics
	}
	// Otherwise, keep defaults (PerCertificate: false, EnhancedMetrics: true)
	return Config{
		Env:                  env,
		Port:                 port,
		LogLevel:             logLevel,
		LogFormat:            logFormat,
		LogOutput:            logOutput,
		LogFilePath:          logFilePath,
		CORS:                 cors,
		Vault:                VaultConfig{},
		Vaults:               []VaultInstance{},
		ExpirationThresholds: expirations,
		Metrics:              metrics,
	}
}

// IsDev returns true if the environment is development.
func (c Config) IsDev() bool {
	return c.Env == EnvDev
}

// IsProd returns true if the environment is production.
func (c Config) IsProd() bool {
	return c.Env == EnvProd
}

func parseEnv(s string) Environment {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "prod", "production":
		return EnvProd
	default:
		return EnvDev
	}
}

func defaultLogLevel(env Environment) string {
	switch env {
	case EnvProd:
		return "info"
	default:
		return "debug"
	}
}

func defaultLogFormat(env Environment) string {
	switch env {
	case EnvProd:
		return "json"
	default:
		return "console"
	}
}

func loadCORSConfig(env Environment) CORSConfig {
	switch env {
	case EnvProd:
		return CORSConfig{
			AllowedOrigins:   []string{},
			AllowCredentials: true,
		}
	default:
		return CORSConfig{
			AllowedOrigins:   []string{"http://localhost:4321", "http://localhost:3000"},
			AllowCredentials: true,
		}
	}
}

func convertVaultInstanceToLegacy(instance VaultInstance) VaultConfig {
	defaultMount := defaultPKIMount
	if strings.TrimSpace(instance.PKIMount) != "" {
		defaultMount = strings.TrimSpace(instance.PKIMount)
	}
	pkiMounts := instance.PKIMounts
	if len(pkiMounts) == 0 {
		pkiMounts = []string{defaultMount}
	}
	return VaultConfig{
		Addr:            instance.Address,
		PKIMounts:       pkiMounts,
		ReadToken:       instance.Token,
		TLSCACertBase64: instance.TLSCACertBase64,
		TLSCACert:       instance.TLSCACert,
		TLSCAPath:       instance.TLSCAPath,
		TLSServerName:   instance.TLSServerName,
		TLSInsecure:     instance.TLSInsecure,
	}
}
