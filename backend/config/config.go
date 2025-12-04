package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Environment represents the application environment.
type Environment string

const (
	EnvDev   Environment = "dev"
	EnvStage Environment = "stage"
	EnvProd  Environment = "prod"
)

// Config holds application configuration.
type Config struct {
	Env       Environment
	Port      string
	LogLevel  string
	LogFormat string
	LogOutput string
	CORS      CORSConfig
	Vault     VaultConfig
}

// CORSConfig holds CORS-specific configuration.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowCredentials bool
}

type VaultConfig struct {
	Addr         string
	PKIMount     string
	EnableRevoke bool
}

// Load reads configuration from environment variables.
func Load() Config {
	_ = godotenv.Load()

	env := parseEnv(getEnv("APP_ENV", "dev"))

	cfg := Config{
		Env:       env,
		Port:      getEnv("PORT", "52000"),
		LogLevel:  getEnv("LOG_LEVEL", defaultLogLevel(env)),
		LogFormat: getEnv("LOG_FORMAT", defaultLogFormat(env)),
		LogOutput: getEnv("LOG_OUTPUT", "stdout"),
		CORS:      loadCORSConfig(env),
		Vault:     loadVaultConfig(),
	}

	return cfg
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
	case "stage", "staging":
		return EnvStage
	default:
		return EnvDev
	}
}

func defaultLogLevel(env Environment) string {
	switch env {
	case EnvProd:
		return "info"
	case EnvStage:
		return "debug"
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
	enableRevoke := strings.ToLower(strings.TrimSpace(getEnv("ENABLE_REVOKE", "false"))) == "true"
	return VaultConfig{
		Addr:         getEnv("VAULT_ADDR", ""),
		PKIMount:     getEnv("VAULT_PKI_MOUNT", "pki"),
		EnableRevoke: enableRevoke,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
