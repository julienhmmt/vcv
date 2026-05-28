package main

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"vcv/internal/certs"
	"vcv/internal/config"
	"vcv/internal/vault"
)

type configResponse struct {
	ExpirationThresholds struct {
		Critical int `json:"critical"`
		Warning  int `json:"warning"`
	} `json:"expirationThresholds"`
	PKIMounts []string `json:"pkiMounts"`
}

func newServerWebFS() fs.FS {
	return fstest.MapFS{
		"dist/index.html":     &fstest.MapFile{Data: []byte("<!doctype html><html><body>ok</body></html>")},
		"dist/admin.html":     &fstest.MapFile{Data: []byte("<!doctype html><html><body>admin</body></html>")},
		"dist/assets/app.js":  &fstest.MapFile{Data: []byte("console.log('ok')")},
		"dist/favicon.ico":    &fstest.MapFile{Data: []byte("\x00")},
	}
}

func TestBuildRouter_BasicEndpoints(t *testing.T) {
	cfg := config.Config{
		Env:                  config.EnvDev,
		Port:                 "52000",
		ExpirationThresholds: config.ExpirationThresholds{Critical: 7, Warning: 30},
		Vault:                config.VaultConfig{PKIMounts: []string{"pki"}},
		Vaults:               []config.VaultInstance{},
	}
	primary := &vault.MockClient{}
	primary.On("CheckConnection", mock.Anything).Return(nil)
	multi := &vault.MockClient{}
	multi.On("ListCertificates", mock.Anything).Return([]certs.Certificate{}, nil)
	registry := prometheus.NewRegistry()
	webFS := newServerWebFS()
	statusClients := map[string]vault.Client{}
	router, err := buildRouter(cfg, primary, statusClients, multi, registry, webFS, "", nil)
	assert.NoError(t, err)

	t.Run("serves index", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
		assert.Contains(t, rec.Body.String(), "ok")
	})

	t.Run("serves admin html", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "admin")
	})

	t.Run("serves asset", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "console.log")
	})

	t.Run("serves favicon", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("ready", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/ready", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
		primary.AssertExpectations(t)
	})

	t.Run("config", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		var resp configResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.Equal(t, 7, resp.ExpirationThresholds.Critical)
		assert.Equal(t, 30, resp.ExpirationThresholds.Warning)
		assert.Equal(t, []string{"pki"}, resp.PKIMounts)
	})

	t.Run("version", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
	})

	t.Run("metrics", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestBuildRouter_MissingAssets_Returns404(t *testing.T) {
	cfg := config.Config{Env: config.EnvDev}
	primary := &vault.MockClient{}
	multi := &vault.MockClient{}
	registry := prometheus.NewRegistry()
	webFS := fstest.MapFS{
		"dist/index.html": &fstest.MapFile{Data: []byte("ok")},
	}
	router, err := buildRouter(cfg, primary, map[string]vault.Client{}, multi, registry, webFS, "", nil)
	assert.NotNil(t, router)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/assets/missing.js", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestBuildRouter_MissingIndex_Returns404(t *testing.T) {
	cfg := config.Config{Env: config.EnvDev}
	primary := &vault.MockClient{}
	multi := &vault.MockClient{}
	registry := prometheus.NewRegistry()
	webFS := fstest.MapFS{
		"dist/assets/app.js": &fstest.MapFile{Data: []byte("console.log('ok')")},
	}
	router, err := buildRouter(cfg, primary, map[string]vault.Client{}, multi, registry, webFS, "", nil)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
