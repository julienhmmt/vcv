package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadVaultInstances_FromSettings(t *testing.T) {
	clearEnv(t)
	tempDir := t.TempDir()
	changeWorkingDirectory(t, tempDir)
	writeSettingsFile(t, tempDir, `{
  "vaults": [
    {"id": "prod", "address": "https://vault-prod", "token": "tok-prod", "pki_mount": "pki", "display_name": "Production"},
    {"id": "stg", "address": "https://vault-stg", "token": "tok-stg"}
  ]
}`)
	instances, err := LoadVaultInstances()
	if err != nil {
		t.Fatalf("expected settings to load, got error: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}
	if instances[0].DisplayName != "Production" {
		t.Fatalf("expected display name to be kept from settings")
	}
	if instances[1].PKIMount != defaultPKIMount {
		t.Fatalf("expected default pki mount to be applied")
	}
	if instances[1].DisplayName != "stg" {
		t.Fatalf("expected display name fallback to id, got %s", instances[1].DisplayName)
	}
}

func TestLoadVaultInstances_SettingsPathWins(t *testing.T) {
	clearEnv(t)
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "my-settings.json")
	settingsContent := `{
  "vaults": [
    {"id": "from-path", "address": "https://vault-from-path", "token": "tok"}
  ]
}`
	if err := os.WriteFile(settingsPath, []byte(settingsContent), 0o644); err != nil {
		t.Fatalf("failed to write settings file: %v", err)
	}
	setEnv(t, "SETTINGS_PATH", settingsPath)
	setEnv(t, "APP_ENV", "dev")

	instances, err := LoadVaultInstances()
	if err != nil {
		t.Fatalf("expected settings from SETTINGS_PATH to load, got error: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(instances))
	}
	if instances[0].ID != "from-path" {
		t.Fatalf("expected instance to come from SETTINGS_PATH, got %+v", instances[0])
	}
}

func TestLoadVaultInstances_FromEnvSpecificSettingsFile(t *testing.T) {
	clearEnv(t)
	tempDir := t.TempDir()
	changeWorkingDirectory(t, tempDir)
	setEnv(t, "APP_ENV", "dev")
	settingsPath := filepath.Join(tempDir, "settings.dev.json")
	settingsContent := `{
  "vaults": [
    {"id": "dev", "address": "https://vault-dev", "token": "tok"}
  ]
}`
	if err := os.WriteFile(settingsPath, []byte(settingsContent), 0o644); err != nil {
		t.Fatalf("failed to write settings file: %v", err)
	}

	instances, err := LoadVaultInstances()
	if err != nil {
		t.Fatalf("expected settings.dev.json to load, got error: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(instances))
	}
	if instances[0].ID != "dev" {
		t.Fatalf("expected env-specific settings file instance, got %+v", instances[0])
	}
}

func TestLoadVaultInstances_FromEnv(t *testing.T) {
	clearEnv(t)
	envValue := "prod@https://vault-prod:8200#tok-prod#pki,https://vault-stg:8200#tok-stg"
	setEnv(t, "VAULT_ADDRS", envValue)
	instances, err := LoadVaultInstances()
	if err != nil {
		t.Fatalf("expected env to load, got error: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}
	if instances[0].ID != "prod" || instances[0].PKIMount != "pki" {
		t.Fatalf("unexpected first instance: %+v", instances[0])
	}
	if instances[1].ID == "" {
		t.Fatalf("expected derived id for second instance")
	}
}

func TestLoadVaultInstances_FallbackLegacy(t *testing.T) {
	clearEnv(t)
	setEnv(t, "VAULT_ADDR", "https://legacy")
	setEnv(t, "VAULT_READ_TOKEN", "legacy-token")
	setEnv(t, "VAULT_PKI_MOUNT", "legacy-pki")
	setEnv(t, "VAULT_TLS_INSECURE", "true")
	instances, err := LoadVaultInstances()
	if err != nil {
		t.Fatalf("expected fallback to load, got error: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 fallback instance, got %d", len(instances))
	}
	if instances[0].ID != "default" || instances[0].PKIMount != "legacy-pki" || !instances[0].TLSInsecure {
		t.Fatalf("unexpected fallback instance: %+v", instances[0])
	}
}

func TestLoadVaultInstances_DuplicateID(t *testing.T) {
	clearEnv(t)
	tempDir := t.TempDir()
	changeWorkingDirectory(t, tempDir)
	writeSettingsFile(t, tempDir, `{
  "vaults": [
    {"id": "dup", "address": "https://one", "token": "a"},
    {"id": "dup", "address": "https://two", "token": "b"}
  ]
}`)
	_, err := LoadVaultInstances()
	if err == nil {
		t.Fatalf("expected duplicate id error")
	}
}

func TestLoadVaultInstances_InvalidAddress(t *testing.T) {
	clearEnv(t)
	tempDir := t.TempDir()
	changeWorkingDirectory(t, tempDir)
	writeSettingsFile(t, tempDir, `{
  "vaults": [
    {"id": "bad", "address": "://bad", "token": "tok"}
  ]
}`)
	_, err := LoadVaultInstances()
	if err == nil {
		t.Fatalf("expected invalid address error")
	}
}

func TestLoadVaultInstances_DisabledVaultsAreSkipped(t *testing.T) {
	clearEnv(t)
	tempDir := t.TempDir()
	changeWorkingDirectory(t, tempDir)
	writeSettingsFile(t, tempDir, `{
  "vaults": [
    {"id": "disabled", "enabled": false, "address": "://bad", "token": "tok"},
    {"id": "enabled", "enabled": true, "address": "https://vault-ok", "token": "tok-ok"}
  ]
}`)
	instances, err := LoadVaultInstances()
	if err != nil {
		t.Fatalf("expected disabled vault to be skipped, got error: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 enabled instance, got %d", len(instances))
	}
	if instances[0].ID != "enabled" {
		t.Fatalf("expected enabled instance to remain, got %+v", instances[0])
	}
}

func TestLoadVaultInstances_AllDisabledIsError(t *testing.T) {
	clearEnv(t)
	tempDir := t.TempDir()
	changeWorkingDirectory(t, tempDir)
	writeSettingsFile(t, tempDir, `{
  "vaults": [
    {"id": "disabled", "enabled": false, "address": "https://vault", "token": "tok"}
  ]
}`)
	_, err := LoadVaultInstances()
	if err == nil {
		t.Fatalf("expected error when all vaults are disabled")
	}
}

func changeWorkingDirectory(t *testing.T, dir string) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(original)
	})
}

func writeSettingsFile(t *testing.T, dir string, content string) {
	t.Helper()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write settings file: %v", err)
	}
}

func setEnv(t *testing.T, key string, value string) {
	t.Helper()
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed to set env %s: %v", key, err)
	}
	t.Cleanup(func() {
		_ = os.Unsetenv(key)
	})
}
