package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_VaultsFromSettings(t *testing.T) {
	tempDir := t.TempDir()
	changeWorkingDirectory(t, tempDir)
	writeSettingsFile(t, tempDir, `{
  "app": {"env": "dev", "port": 52000},
  "vaults": [
    {"id": "prod", "address": "https://vault-prod", "token": "tok-prod", "pki_mount": "pki", "display_name": "Production"},
    {"id": "stg", "address": "https://vault-stg", "token": "tok-stg"}
  ]
}`)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected settings to load, got error: %v", err)
	}
	if len(cfg.Vaults) != 2 {
		t.Fatalf("expected 2 vaults, got %d", len(cfg.Vaults))
	}
	if cfg.Vaults[0].DisplayName != "Production" {
		t.Fatalf("expected display name to be kept from settings")
	}
	if cfg.Vaults[1].PKIMount != defaultPKIMount {
		t.Fatalf("expected default pki mount to be applied")
	}
	if cfg.Vaults[1].DisplayName != "stg" {
		t.Fatalf("expected display name fallback to id, got %s", cfg.Vaults[1].DisplayName)
	}
}

func TestLoad_EmptyVaults(t *testing.T) {
	tempDir := t.TempDir()
	changeWorkingDirectory(t, tempDir)
	writeSettingsFile(t, tempDir, `{"app": {"env": "dev"}, "vaults": []}`)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error for empty vaults, got %v", err)
	}
	if len(cfg.Vaults) != 0 {
		t.Fatalf("expected 0 vaults, got %d", len(cfg.Vaults))
	}
}

func TestLoad_NoSettingsFile(t *testing.T) {
	tempDir := t.TempDir()
	changeWorkingDirectory(t, tempDir)
	_, err := Load()
	if err == nil {
		t.Fatalf("expected error for missing settings file, got none")
	}
}

func TestResolveSettingsPath_FindsExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	changeWorkingDirectory(t, tempDir)
	writeSettingsFile(t, tempDir, `{"app":{"env":"dev"}}`)
	resolved := ResolveSettingsPath()
	realDir, _ := filepath.EvalSymlinks(tempDir)
	expected := filepath.Join(realDir, "settings.json")
	if resolved != expected {
		t.Fatalf("expected %s, got %s", expected, resolved)
	}
}

func TestResolveSettingsPath_ReturnsDefaultWhenMissing(t *testing.T) {
	tempDir := t.TempDir()
	changeWorkingDirectory(t, tempDir)
	resolved := ResolveSettingsPath()
	realDir, _ := filepath.EvalSymlinks(tempDir)
	expected := filepath.Join(realDir, "settings.json")
	if resolved != expected {
		t.Fatalf("expected default %s, got %s", expected, resolved)
	}
}

func TestSettingsCandidates(t *testing.T) {
	candidates := settingsCandidates()
	if len(candidates) < 4 {
		t.Fatalf("expected at least 4 candidates, got %d", len(candidates))
	}
	if candidates[0] != "settings.dev.json" {
		t.Fatalf("expected first candidate settings.dev.json, got %s", candidates[0])
	}
	if candidates[1] != "settings.prod.json" {
		t.Fatalf("expected second candidate settings.prod.json, got %s", candidates[1])
	}
}

func TestNormalizeVaultInstances(t *testing.T) {
	tests := []struct {
		name      string
		instances []VaultInstance
		expected  int
		error     bool
	}{
		{
			name: "valid instances",
			instances: []VaultInstance{
				{ID: "vault1", Address: "https://vault1:8200", Token: "token1"},
				{ID: "vault2", Address: "https://vault2:8200", Token: "token2"},
			},
			expected: 2,
		},
		{
			name: "disabled instance",
			instances: []VaultInstance{
				{ID: "vault1", Address: "https://vault1:8200", Token: "token1"},
				{ID: "vault2", Address: "https://vault2:8200", Token: "token2", Enabled: &[]bool{false}[0]},
			},
			expected: 1,
		},
		{
			name: "duplicate IDs",
			instances: []VaultInstance{
				{ID: "vault1", Address: "https://vault1:8200", Token: "token1"},
				{ID: "vault1", Address: "https://vault2:8200", Token: "token2"},
			},
			error: true,
		},
		{
			name: "invalid address",
			instances: []VaultInstance{
				{ID: "vault1", Address: "invalid-url", Token: "token1"},
			},
			error: true,
		},
		{
			name: "empty token",
			instances: []VaultInstance{
				{ID: "vault1", Address: "https://vault1:8200", Token: ""},
			},
			error: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeVaultInstances(tt.instances)
			if tt.error {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != tt.expected {
				t.Fatalf("expected %d instances, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestNormalizeVaultInstance(t *testing.T) {
	tests := []struct {
		name     string
		instance VaultInstance
		expected VaultInstance
		error    bool
	}{
		{
			name: "minimal valid instance",
			instance: VaultInstance{
				Address: "https://vault:8200",
				Token:   "token",
			},
			expected: VaultInstance{
				ID:          "vault-8200",
				Address:     "https://vault:8200",
				Token:       "token",
				PKIMount:    defaultPKIMount,
				PKIMounts:   []string{defaultPKIMount},
				DisplayName: "vault-8200",
				Enabled:     &[]bool{true}[0],
			},
		},
		{
			name: "complete instance",
			instance: VaultInstance{
				ID:              "my-vault",
				Address:         "https://vault.example.com:8200",
				Token:           "my-token",
				PKIMount:        "my-pki",
				PKIMounts:       []string{"pki1", "pki2"},
				DisplayName:     "My Vault",
				TLSInsecure:     true,
				TLSCACertBase64: "base64-cert",
				Enabled:         &[]bool{true}[0],
			},
			expected: VaultInstance{
				ID:              "my-vault",
				Address:         "https://vault.example.com:8200",
				Token:           "my-token",
				PKIMount:        "pki1",
				PKIMounts:       []string{"pki1", "pki2"},
				DisplayName:     "My Vault",
				TLSInsecure:     true,
				TLSCACertBase64: "base64-cert",
				Enabled:         &[]bool{true}[0],
			},
		},
		{
			name: "empty address",
			instance: VaultInstance{
				ID:      "vault",
				Address: "",
				Token:   "token",
			},
			error: true,
		},
		{
			name: "empty token",
			instance: VaultInstance{
				ID:      "vault",
				Address: "https://vault:8200",
				Token:   "",
			},
			error: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeVaultInstance(tt.instance)
			if tt.error {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ID != tt.expected.ID {
				t.Fatalf("expected ID %s, got %s", tt.expected.ID, result.ID)
			}
			if result.Address != tt.expected.Address {
				t.Fatalf("expected Address %s, got %s", tt.expected.Address, result.Address)
			}
			if result.Token != tt.expected.Token {
				t.Fatalf("expected Token %s, got %s", tt.expected.Token, result.Token)
			}
			if result.DisplayName != tt.expected.DisplayName {
				t.Fatalf("expected DisplayName %s, got %s", tt.expected.DisplayName, result.DisplayName)
			}
			if result.TLSInsecure != tt.expected.TLSInsecure {
				t.Fatalf("expected TLSInsecure %v, got %v", tt.expected.TLSInsecure, result.TLSInsecure)
			}
		})
	}
}

func TestIsVaultEnabled(t *testing.T) {
	tests := []struct {
		name     string
		instance VaultInstance
		expected bool
	}{
		{
			name: "enabled nil defaults to true",
			instance: VaultInstance{
				ID:      "vault",
				Address: "https://vault:8200",
				Token:   "token",
			},
			expected: true,
		},
		{
			name: "enabled true",
			instance: VaultInstance{
				ID:      "vault",
				Address: "https://vault:8200",
				Token:   "token",
				Enabled: &[]bool{true}[0],
			},
			expected: true,
		},
		{
			name: "enabled false",
			instance: VaultInstance{
				ID:      "vault",
				Address: "https://vault:8200",
				Token:   "token",
				Enabled: &[]bool{false}[0],
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsVaultEnabled(tt.instance)
			if result != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDeriveVaultID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://vault.example.com:8200", "vault-example-com-8200"},
		{"http://vault:8200", "vault-8200"},
		{"vault.example.com", "vault-example-com"},
		{"https://vault.example.com/", "vault-example-com"},
		{"https://vault.example.com/api/v1/", "vault-example-com-api-v1"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := deriveVaultID(tt.input)
			if result != tt.expected {
				t.Fatalf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestConvertVaultInstanceToLegacy(t *testing.T) {
	instance := VaultInstance{
		ID:              "test-vault",
		Address:         "https://vault:8200",
		Token:           "token",
		PKIMount:        "custom-pki",
		PKIMounts:       []string{"pki1", "pki2"},
		DisplayName:     "Test Vault",
		TLSInsecure:     true,
		TLSCACertBase64: "base64-cert",
		TLSCACert:       "pem-cert",
		TLSCAPath:       "/path/to/ca",
		TLSServerName:   "vault.example.com",
	}

	legacy := convertVaultInstanceToLegacy(instance)

	if legacy.Addr != instance.Address {
		t.Fatalf("expected Addr %s, got %s", instance.Address, legacy.Addr)
	}
	if legacy.ReadToken != instance.Token {
		t.Fatalf("expected ReadToken %s, got %s", instance.Token, legacy.ReadToken)
	}
	if len(legacy.PKIMounts) != 2 {
		t.Fatalf("expected 2 PKI mounts, got %d", len(legacy.PKIMounts))
	}
	if legacy.PKIMounts[0] != "pki1" {
		t.Fatalf("expected first PKI mount pki1, got %s", legacy.PKIMounts[0])
	}
	if !legacy.TLSInsecure {
		t.Fatalf("expected TLSInsecure true, got %v", legacy.TLSInsecure)
	}
	if legacy.TLSCACertBase64 != instance.TLSCACertBase64 {
		t.Fatalf("expected TLSCACertBase64 %s, got %s", instance.TLSCACertBase64, legacy.TLSCACertBase64)
	}
}

// Helper functions

func changeWorkingDirectory(t *testing.T, dir string) {
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWd)
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
}

func writeSettingsFile(t *testing.T, dir, content string) {
	settingsPath := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write settings file: %v", err)
	}
}
