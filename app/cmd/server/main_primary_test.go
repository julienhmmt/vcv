package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vcv/internal/config"
	"vcv/internal/logger"
)

func TestInitVaultClients_PrimaryIsFirstEnabledVault(t *testing.T) {
	disabled := false
	cfg := &config.Config{AllVaults: []config.VaultInstance{
		{ID: "off", Address: "https://off.example.test:8200", Token: "t", Enabled: &disabled},
		{ID: "on", Address: "https://on.example.test:8200", Token: "t"},
	}}

	clients, primary := initVaultClients(cfg, logger.Get())

	require.Len(t, clients, 2)
	assert.Equal(t, clients["on"], primary)
	assert.Equal(t, "https://on.example.test:8200", cfg.Vault.Addr)
}

func TestInitVaultClients_NoEnabledVaultUsesDisabledClient(t *testing.T) {
	disabled := false
	cfg := &config.Config{AllVaults: []config.VaultInstance{
		{ID: "off", Address: "https://off.example.test:8200", Token: "t", Enabled: &disabled},
	}}

	_, primary := initVaultClients(cfg, logger.Get())

	require.NotNil(t, primary)
	assert.Empty(t, cfg.Vault.Addr)
}
