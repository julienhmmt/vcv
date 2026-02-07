package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
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
	CORS                 CORSConfig
	Vault                VaultConfig
	Vaults               []VaultInstance
	ExpirationThresholds ExpirationThresholds
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

type SettingsFile struct {
	App          AppSettings         `json:"app"`
	Certificates CertificateSettings `json:"certificates"`
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

// Load reads configuration from environment variables or settings file.
func Load() (Config, error) {
	_ = godotenv.Load()
	settings, settingsPath, settingsErr := loadSettingsFile()
	if settingsErr == nil && settings != nil {
		cfg := buildConfigFromSettings(*settings)
		vaults, normalizeErr := normalizeVaultInstances(settings.Vaults)
		if normalizeErr != nil {
			return Config{}, fmt.Errorf("invalid settings file %s: %w", settingsPath, normalizeErr)
		}
		cfg.Vaults = vaults
		if len(vaults) > 0 {
			cfg.Vault = convertVaultInstanceToLegacy(vaults[0])
		}
		applyLoggingEnv(cfg)
		return cfg, nil
	}

	env := parseEnv(getEnv("APP_ENV", "dev"))
	legacyVault := loadVaultConfig()

	cfg := Config{
		Env:                  env,
		Port:                 getEnv("PORT", "52000"),
		LogLevel:             getEnv("LOG_LEVEL", defaultLogLevel(env)),
		LogFormat:            getEnv("LOG_FORMAT", defaultLogFormat(env)),
		LogOutput:            getEnv("LOG_OUTPUT", "stdout"),
		LogFilePath:          getEnv("LOG_FILE_PATH", ""),
		CORS:                 loadCORSConfig(env),
		Vault:                legacyVault,
		ExpirationThresholds: loadExpirationThresholds(),
	}

	vaultInstances, vaultErr := LoadVaultInstances()
	if vaultErr == nil && len(vaultInstances) > 0 {
		cfg.Vaults = vaultInstances
		cfg.Vault = convertVaultInstanceToLegacy(vaultInstances[0])
	}

	return cfg, nil
}

func loadSettingsFile() (*SettingsFile, string, error) {
	settingsPath := strings.TrimSpace(getEnv("SETTINGS_PATH", ""))
	if settingsPath != "" {
		data, readErr := os.ReadFile(settingsPath)
		if readErr != nil {
			return nil, settingsPath, readErr
		}
		var settings SettingsFile
		if err := json.Unmarshal(data, &settings); err != nil {
			return nil, settingsPath, err
		}
		return &settings, settingsPath, nil
	}

	envName := strings.ToLower(strings.TrimSpace(getEnv("APP_ENV", "dev")))
	candidates := []string{fmt.Sprintf("settings.%s.json", envName), "settings.json", "./settings.json", "/etc/vcv/settings.json"}
	for _, candidate := range candidates {
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
	expirationThresholds := ExpirationThresholds{Critical: 7, Warning: 30}
	if settings.Certificates.ExpirationThresholds.Critical > 0 {
		expirationThresholds.Critical = settings.Certificates.ExpirationThresholds.Critical
	}
	if settings.Certificates.ExpirationThresholds.Warning > 0 {
		expirationThresholds.Warning = settings.Certificates.ExpirationThresholds.Warning
	}
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
		ExpirationThresholds: expirationThresholds,
	}
}

func applyLoggingEnv(cfg Config) {
	if strings.TrimSpace(cfg.LogOutput) != "" {
		_ = os.Setenv("LOG_OUTPUT", cfg.LogOutput)
	}
	if strings.TrimSpace(cfg.LogFormat) != "" {
		_ = os.Setenv("LOG_FORMAT", cfg.LogFormat)
	}
	if strings.TrimSpace(cfg.LogFilePath) != "" {
		_ = os.Setenv("LOG_FILE_PATH", cfg.LogFilePath)
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
	originsEnv := getEnv("CORS_ALLOWED_ORIGINS", "")
	if originsEnv != "" {
		origins := strings.Split(originsEnv, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
		return CORSConfig{
			AllowedOrigins:   origins,
			AllowCredentials: true,
		}
	}
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

func loadVaultConfig() VaultConfig {
	skipVerifyDefault := getEnv("VAULT_SKIP_VERIFY", "false")
	tlsInsecure := strings.ToLower(getEnv("VAULT_TLS_INSECURE", skipVerifyDefault)) == "true"
	tlsCACertBase64 := strings.TrimSpace(getEnv("VAULT_TLS_CA_CERT_BASE64", getEnv("VAULT_CACERT_BYTES", "")))
	tlsCACert := strings.TrimSpace(getEnv("VAULT_TLS_CA_CERT", getEnv("VAULT_CACERT", "")))
	tlsCAPath := strings.TrimSpace(getEnv("VAULT_TLS_CA_PATH", getEnv("VAULT_CAPATH", "")))
	tlsServerName := strings.TrimSpace(getEnv("VAULT_TLS_SERVER_NAME", ""))

	// Support both new VAULT_PKI_MOUNTS (comma-separated) and legacy VAULT_PKI_MOUNT
	pkiMountsStr := getEnv("VAULT_PKI_MOUNTS", "")
	if pkiMountsStr == "" {
		// Fallback to legacy single mount for backward compatibility
		legacyMount := getEnv("VAULT_PKI_MOUNT", defaultPKIMount)
		pkiMountsStr = legacyMount
	}

	var pkiMounts []string
	if pkiMountsStr != "" {
		pkiMounts = strings.Split(pkiMountsStr, ",")
		for i := range pkiMounts {
			pkiMounts[i] = strings.TrimSpace(pkiMounts[i])
		}
	}

	return VaultConfig{
		Addr:            getEnv("VAULT_ADDR", ""),
		PKIMounts:       pkiMounts,
		ReadToken:       getEnv("VAULT_READ_TOKEN", ""),
		TLSCACertBase64: tlsCACertBase64,
		TLSCACert:       tlsCACert,
		TLSCAPath:       tlsCAPath,
		TLSServerName:   tlsServerName,
		TLSInsecure:     tlsInsecure,
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

func loadExpirationThresholds() ExpirationThresholds {
	critical := getEnvInt("VCV_EXPIRE_CRITICAL", 7)
	warning := getEnvInt("VCV_EXPIRE_WARNING", 30)
	return ExpirationThresholds{
		Critical: critical,
		Warning:  warning,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
