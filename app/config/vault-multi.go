package config

import (
	"fmt"
	"net/url"
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

// NormalizeAllVaultInstances normalizes all vault instances including disabled ones.
// This is used to create clients for every vault so they can be toggled at runtime.
func NormalizeAllVaultInstances(instances []VaultInstance) ([]VaultInstance, error) {
	normalized := make([]VaultInstance, 0, len(instances))
	seen := make(map[string]bool)
	for index, instance := range instances {
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
