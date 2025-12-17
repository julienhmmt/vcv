package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"vcv/config"
	"vcv/internal/vault"
)

type statusResponse struct {
	Version        string `json:"version"`
	VaultConnected bool   `json:"vault_connected"`
	VaultError     string `json:"vault_error,omitempty"`
	Vaults         []struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
		Connected   bool   `json:"connected"`
		Error       string `json:"error,omitempty"`
	} `json:"vaults"`
}

func TestNewStatusHandler_PrimaryDisconnectedAndMissingClient(t *testing.T) {
	cfg := config.Config{Vaults: []config.VaultInstance{{ID: "v1", DisplayName: "Vault 1"}}}
	primary := &vault.MockClient{}
	primary.On("CheckConnection", mock.Anything).Return(errors.New("primary down"))
	statusClients := map[string]vault.Client{}
	h := newStatusHandler(cfg, primary, statusClients)
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var payload statusResponse
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
	assert.False(t, payload.VaultConnected)
	assert.Contains(t, payload.VaultError, "primary down")
	assert.Len(t, payload.Vaults, 1)
	assert.Equal(t, "v1", payload.Vaults[0].ID)
	assert.Equal(t, "Vault 1", payload.Vaults[0].DisplayName)
	assert.False(t, payload.Vaults[0].Connected)
	assert.Equal(t, "missing vault status client", payload.Vaults[0].Error)
	primary.AssertExpectations(t)
}

func TestNewStatusHandler_AllConnected(t *testing.T) {
	cfg := config.Config{Vaults: []config.VaultInstance{{ID: "v1", DisplayName: "Vault 1"}, {ID: "v2", DisplayName: "Vault 2"}}}
	primary := &vault.MockClient{}
	primary.On("CheckConnection", mock.Anything).Return(nil)
	client1 := &vault.MockClient{}
	client1.On("CheckConnection", mock.Anything).Return(nil)
	client2 := &vault.MockClient{}
	client2.On("CheckConnection", mock.Anything).Return(nil)
	statusClients := map[string]vault.Client{"v1": client1, "v2": client2}
	h := newStatusHandler(cfg, primary, statusClients)
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var payload statusResponse
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
	assert.True(t, payload.VaultConnected)
	assert.Equal(t, "", payload.VaultError)
	assert.Len(t, payload.Vaults, 2)
	assert.True(t, payload.Vaults[0].Connected)
	assert.True(t, payload.Vaults[1].Connected)
	primary.AssertExpectations(t)
	client1.AssertExpectations(t)
	client2.AssertExpectations(t)
}
