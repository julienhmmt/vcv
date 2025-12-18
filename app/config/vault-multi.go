package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const defaultPKIMount = "pki"

type VaultInstance struct {
	ID              string   `json:"id"`
	Address         string   `json:"address"`
	Token           string   `json:"token"`
	PKIMount        string   `json:"pki_mount"`
	PKIMounts       []string `json:"pki_mounts,omitempty"`
	DisplayName     string   `json:"display_name"`
	TLSInsecure     bool     `json:"tls_insecure"`
	TLSCACertBase64 string   `json:"tls_ca_cert_base64,omitempty"`
	TLSCACert       string   `json:"tls_ca_cert,omitempty"`
	TLSCAPath       string   `json:"tls_ca_path,omitempty"`
	TLSServerName   string   `json:"tls_server_name,omitempty"`
	Enabled         *bool    `json:"enabled,omitempty"`
}

type MultiVaultConfig struct {
	Instances []VaultInstance `json:"vaults"`
}

func LoadVaultInstances() ([]VaultInstance, error) {
	settingsInstances, settingsErr := loadVaultInstancesFromSettings()
	envInstances := parseVaultAddrsEnv()
	merged := mergeVaultInstances(settingsInstances, envInstances)
	if len(merged) == 0 {
		fallback := fallbackInstanceFromEnv()
		if fallback != nil {
			merged = append(merged, *fallback)
		}
	}
	if settingsErr != nil && !errors.Is(settingsErr, os.ErrNotExist) {
		return nil, settingsErr
	}
	if len(merged) == 0 {
		return []VaultInstance{}, nil
	}
	normalized, normalizeErr := normalizeVaultInstances(merged)
	if normalizeErr != nil {
		return nil, normalizeErr
	}
	return normalized, nil
}

func loadVaultInstancesFromSettings() ([]VaultInstance, error) {
	settingsPath := strings.TrimSpace(getEnv("SETTINGS_PATH", ""))
	if settingsPath != "" {
		content, readErr := os.ReadFile(settingsPath)
		if readErr != nil {
			return nil, readErr
		}
		var config MultiVaultConfig
		if jsonErr := json.Unmarshal(content, &config); jsonErr != nil {
			return nil, fmt.Errorf("invalid settings.json content: %w", jsonErr)
		}
		return config.Instances, nil
	}

	envName := strings.ToLower(strings.TrimSpace(getEnv("APP_ENV", "dev")))
	paths := []string{fmt.Sprintf("settings.%s.json", envName), "settings.json", "./settings.json", "/etc/vcv/settings.json"}
	for _, candidate := range paths {
		absPath, absErr := filepath.Abs(candidate)
		if absErr != nil {
			continue
		}
		if _, statErr := os.Stat(absPath); statErr != nil {
			continue
		}
		content, readErr := os.ReadFile(absPath)
		if readErr != nil {
			return nil, readErr
		}
		var config MultiVaultConfig
		if jsonErr := json.Unmarshal(content, &config); jsonErr != nil {
			return nil, fmt.Errorf("invalid settings.json content: %w", jsonErr)
		}
		return config.Instances, nil
	}
	return nil, os.ErrNotExist
}

func parseVaultAddrsEnv() []VaultInstance {
	raw := strings.TrimSpace(os.Getenv("VAULT_ADDRS"))
	if raw == "" {
		return []VaultInstance{}
	}
	entries := strings.Split(raw, ",")
	instances := make([]VaultInstance, 0, len(entries))
	for index, entry := range entries {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			continue
		}
		instance := parseVaultAddrsEntry(trimmed, index)
		if instance != nil {
			instances = append(instances, *instance)
		}
	}
	return instances
}

func parseVaultAddrsEntry(entry string, index int) *VaultInstance {
	var id, address, token, pkiMount string
	if atIdx := strings.Index(entry, "@"); atIdx > 0 {
		id = entry[:atIdx]
		entry = entry[atIdx+1:]
	}
	parts := strings.Split(entry, "#")
	switch len(parts) {
	case 1:
		address = parts[0]
		token = os.Getenv("VAULT_READ_TOKEN")
		pkiMount = defaultPKIMount
	case 2:
		address = parts[0]
		token = parts[1]
		pkiMount = defaultPKIMount
	case 3:
		address = parts[0]
		token = parts[1]
		pkiMount = parts[2]
	default:
		return nil
	}
	if id == "" {
		id = fmt.Sprintf("vault-%d", index+1)
	}
	if address == "" || token == "" {
		return nil
	}
	tlsInsecure := strings.ToLower(os.Getenv("VAULT_TLS_INSECURE")) == "true"
	tlsCACertBase64 := strings.TrimSpace(os.Getenv("VAULT_TLS_CA_CERT_BASE64"))
	if tlsCACertBase64 == "" {
		tlsCACertBase64 = strings.TrimSpace(os.Getenv("VAULT_CACERT_BYTES"))
	}
	tlsCACert := strings.TrimSpace(os.Getenv("VAULT_TLS_CA_CERT"))
	if tlsCACert == "" {
		tlsCACert = strings.TrimSpace(os.Getenv("VAULT_CACERT"))
	}
	tlsCAPath := strings.TrimSpace(os.Getenv("VAULT_TLS_CA_PATH"))
	if tlsCAPath == "" {
		tlsCAPath = strings.TrimSpace(os.Getenv("VAULT_CAPATH"))
	}
	tlsServerName := strings.TrimSpace(os.Getenv("VAULT_TLS_SERVER_NAME"))
	return &VaultInstance{
		ID:              id,
		Address:         address,
		Token:           token,
		PKIMount:        pkiMount,
		PKIMounts:       []string{pkiMount},
		DisplayName:     "",
		TLSInsecure:     tlsInsecure,
		TLSCACertBase64: tlsCACertBase64,
		TLSCACert:       tlsCACert,
		TLSCAPath:       tlsCAPath,
		TLSServerName:   tlsServerName,
	}
}

func mergeVaultInstances(primary []VaultInstance, secondary []VaultInstance) []VaultInstance {
	seen := make(map[string]bool)
	result := make([]VaultInstance, 0, len(primary)+len(secondary))
	for _, instance := range primary {
		if instance.ID == "" {
			continue
		}
		seen[instance.ID] = true
		result = append(result, instance)
	}
	for _, instance := range secondary {
		if instance.ID == "" {
			continue
		}
		if seen[instance.ID] {
			continue
		}
		result = append(result, instance)
	}
	return result
}

func fallbackInstanceFromEnv() *VaultInstance {
	single := loadVaultConfig()
	if single.Addr == "" || single.ReadToken == "" {
		return nil
	}
	pkiMount := defaultPKIMount
	if len(single.PKIMounts) > 0 && strings.TrimSpace(single.PKIMounts[0]) != "" {
		pkiMount = strings.TrimSpace(single.PKIMounts[0])
	}
	return &VaultInstance{
		ID:              "default",
		Address:         single.Addr,
		Token:           single.ReadToken,
		PKIMount:        pkiMount,
		PKIMounts:       single.PKIMounts,
		DisplayName:     "default",
		TLSInsecure:     single.TLSInsecure,
		TLSCACertBase64: single.TLSCACertBase64,
		TLSCACert:       single.TLSCACert,
		TLSCAPath:       single.TLSCAPath,
		TLSServerName:   single.TLSServerName,
	}
}

func normalizeVaultInstances(instances []VaultInstance) ([]VaultInstance, error) {
	normalized := make([]VaultInstance, 0, len(instances))
	seen := make(map[string]bool)
	for index, instance := range instances {
		if !IsVaultEnabled(instance) {
			continue
		}
		normalizedInstance, normalizeErr := normalizeVaultInstance(instance)
		if normalizeErr != nil {
			return nil, fmt.Errorf("vault %d: %w", index, normalizeErr)
		}
		if seen[normalizedInstance.ID] {
			return nil, fmt.Errorf("duplicate vault id: %s", normalizedInstance.ID)
		}
		seen[normalizedInstance.ID] = true
		normalized = append(normalized, normalizedInstance)
	}
	return normalized, nil
}

func normalizeVaultInstance(instance VaultInstance) (VaultInstance, error) {
	id := strings.TrimSpace(instance.ID)
	address := strings.TrimSpace(instance.Address)
	token := strings.TrimSpace(instance.Token)
	pkiMount := strings.TrimSpace(instance.PKIMount)
	pkiMounts := instance.PKIMounts
	displayName := strings.TrimSpace(instance.DisplayName)
	tlsCACertBase64 := strings.TrimSpace(instance.TLSCACertBase64)
	tlsCACert := strings.TrimSpace(instance.TLSCACert)
	tlsCAPath := strings.TrimSpace(instance.TLSCAPath)
	tlsServerName := strings.TrimSpace(instance.TLSServerName)
	if instance.Enabled == nil {
		value := true
		instance.Enabled = &value
	}
	if id == "" {
		id = deriveVaultID(address)
	}
	if id == "" {
		return VaultInstance{}, fmt.Errorf("vault id is empty")
	}
	if address == "" {
		return VaultInstance{}, fmt.Errorf("vault address is empty")
	}
	if _, parseErr := url.ParseRequestURI(address); parseErr != nil {
		return VaultInstance{}, fmt.Errorf("invalid vault address: %w", parseErr)
	}
	if token == "" {
		return VaultInstance{}, fmt.Errorf("vault token is empty")
	}
	if len(pkiMounts) == 0 {
		if pkiMount != "" {
			pkiMounts = []string{pkiMount}
		}
	}
	if len(pkiMounts) == 0 {
		pkiMounts = []string{defaultPKIMount}
	}
	if pkiMount == "" {
		pkiMount = strings.TrimSpace(pkiMounts[0])
	}
	if pkiMount == "" {
		pkiMount = defaultPKIMount
	}
	if displayName == "" {
		displayName = id
	}
	return VaultInstance{
		ID:              id,
		Address:         address,
		Token:           token,
		PKIMount:        pkiMount,
		PKIMounts:       pkiMounts,
		DisplayName:     displayName,
		TLSInsecure:     instance.TLSInsecure,
		TLSCACertBase64: tlsCACertBase64,
		TLSCACert:       tlsCACert,
		TLSCAPath:       tlsCAPath,
		TLSServerName:   tlsServerName,
		Enabled:         instance.Enabled,
	}, nil
}

func IsVaultEnabled(instance VaultInstance) bool {
	if instance.Enabled == nil {
		return true
	}
	return *instance.Enabled
}

func deriveVaultID(address string) string {
	cleanAddress := strings.TrimSpace(address)
	cleanAddress = strings.TrimPrefix(cleanAddress, "https://")
	cleanAddress = strings.TrimPrefix(cleanAddress, "http://")
	cleanAddress = strings.TrimSuffix(cleanAddress, "/")
	cleanAddress = strings.ReplaceAll(cleanAddress, "/", "-")
	cleanAddress = strings.ReplaceAll(cleanAddress, ":", "-")
	cleanAddress = strings.ReplaceAll(cleanAddress, ".", "-")
	return strings.Trim(cleanAddress, "-")
}
