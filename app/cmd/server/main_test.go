package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"vcv/internal/config"
	"vcv/internal/vault"
)

func TestNewStatusHandler_PrimaryConnected(t *testing.T) {
	primary := &vault.MockClient{}
	primary.On("CheckConnection", mock.Anything).Return(nil)

	cfg := config.Config{Vaults: []config.VaultInstance{}}
	handler := newStatusHandler(cfg, primary, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["vault_connected"].(bool))
	assert.Nil(t, resp["vault_error"])
	primary.AssertExpectations(t)
}

func TestNewStatusHandler_PrimaryDisconnected(t *testing.T) {
	primary := &vault.MockClient{}
	primary.On("CheckConnection", mock.Anything).Return(errors.New("connection refused"))

	cfg := config.Config{Vaults: []config.VaultInstance{}}
	handler := newStatusHandler(cfg, primary, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.False(t, resp["vault_connected"].(bool))
	assert.Equal(t, "vault unavailable", resp["vault_error"])
	primary.AssertExpectations(t)
}

func TestNewStatusHandler_StatusClients(t *testing.T) {
	primary := &vault.MockClient{}
	primary.On("CheckConnection", mock.Anything).Return(nil)

	v1 := &vault.MockClient{}
	v1.On("CheckConnection", mock.Anything).Return(nil)
	v2 := &vault.MockClient{}
	v2.On("CheckConnection", mock.Anything).Return(errors.New("down"))

	cfg := config.Config{
		Vaults: []config.VaultInstance{
			{ID: "v1", DisplayName: "Vault 1"},
			{ID: "v2", DisplayName: "Vault 2"},
			{ID: "v3", DisplayName: "Vault 3"},
		},
	}
	statusClients := map[string]vault.Client{
		"v1": v1,
		"v2": v2,
		// v3 intentionally missing
	}

	handler := newStatusHandler(cfg, primary, statusClients)
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	vaults := resp["vaults"].([]any)
	assert.Len(t, vaults, 3)

	v1Status := vaults[0].(map[string]any)
	assert.Equal(t, "v1", v1Status["id"])
	assert.True(t, v1Status["connected"].(bool))

	v2Status := vaults[1].(map[string]any)
	assert.Equal(t, "v2", v2Status["id"])
	assert.False(t, v2Status["connected"].(bool))
	assert.Equal(t, "vault unavailable", v2Status["error"])

	v3Status := vaults[2].(map[string]any)
	assert.Equal(t, "v3", v3Status["id"])
	assert.False(t, v3Status["connected"].(bool))
	assert.Equal(t, "missing vault status client", v3Status["error"])

	primary.AssertExpectations(t)
	v1.AssertExpectations(t)
	v2.AssertExpectations(t)
}

func TestBuildRouter_Dev(t *testing.T) {
	primary := &vault.MockClient{}
	primary.On("CheckConnection", mock.Anything).Return(nil)
	multi := &vault.MockClient{}
	multi.On("ListCertificates", mock.Anything).Return(nil, nil)

	cfg := config.Config{
		Env: config.EnvDev,
		Vaults: []config.VaultInstance{
			{ID: "v1", DisplayName: "Vault 1"},
		},
	}
	webFS := fstest.MapFS{
		"dist/index.html": &fstest.MapFile{Data: []byte("<html></html>")},
	}
	registry := prometheus.NewRegistry()
	vaultRegistry := vault.NewRegistry(cfg.Vaults)

	router, err := buildRouter(cfg, primary, map[string]vault.Client{"v1": primary}, multi, registry, webFS, "/tmp/settings.json", vaultRegistry)
	require.NoError(t, err)
	assert.NotNil(t, router)

	// Test a few routes exist
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBuildRouter_Prod(t *testing.T) {
	primary := &vault.MockClient{}
	primary.On("CheckConnection", mock.Anything).Return(nil)
	multi := &vault.MockClient{}

	cfg := config.Config{
		Env:    config.EnvProd,
		Vaults: []config.VaultInstance{},
	}
	webFS := fstest.MapFS{
		"dist/index.html": &fstest.MapFile{Data: []byte("<html></html>")},
	}
	registry := prometheus.NewRegistry()
	vaultRegistry := vault.NewRegistry(cfg.Vaults)

	router, err := buildRouter(cfg, primary, map[string]vault.Client{}, multi, registry, webFS, "/tmp/settings.json", vaultRegistry)
	require.NoError(t, err)
	assert.NotNil(t, router)
}

type errFS struct{}

func (errFS) Open(name string) (fs.File, error) { return nil, errors.New("fs error") }
func (errFS) Sub(dir string) (fs.FS, error)     { return nil, errors.New("no dist dir") }

func TestBuildRouter_MissingDist(t *testing.T) {
	primary := &vault.MockClient{}
	multi := &vault.MockClient{}
	cfg := config.Config{Env: config.EnvDev}
	registry := prometheus.NewRegistry()
	vaultRegistry := vault.NewRegistry(nil)

	_, err := buildRouter(cfg, primary, nil, multi, registry, errFS{}, "/tmp/settings.json", vaultRegistry)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no dist dir")
}

func TestBuildRouter_Routes(t *testing.T) {
	primary := &vault.MockClient{}
	primary.On("CheckConnection", mock.Anything).Return(nil)
	multi := &vault.MockClient{}

	cfg := config.Config{
		Env:    config.EnvDev,
		Vaults: []config.VaultInstance{},
	}
	webFS := fstest.MapFS{
		"dist/index.html": &fstest.MapFile{Data: []byte("<html></html>")},
	}
	registry := prometheus.NewRegistry()
	vaultRegistry := vault.NewRegistry(cfg.Vaults)

	router, err := buildRouter(cfg, primary, map[string]vault.Client{}, multi, registry, webFS, "/tmp/settings.json", vaultRegistry)
	require.NoError(t, err)

	// Test /api/version
	req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var versionResp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &versionResp))
	assert.NotEmpty(t, versionResp["version"])

	// Test /api/ready
	req = httptest.NewRequest(http.MethodGet, "/api/ready", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test /api/config
	req = httptest.NewRequest(http.MethodGet, "/api/config", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test / (static)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test /metrics
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
