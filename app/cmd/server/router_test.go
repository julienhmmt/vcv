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

	"vcv/config"
	"vcv/internal/certs"
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
		"index.html":                           &fstest.MapFile{Data: []byte(`<!DOCTYPE html><html lang="{{.Language}}"><head><title>{{.Messages.AppTitle}}</title></head><body>ok</body></html>`)},
		"assets/app.js":                        &fstest.MapFile{Data: []byte("console.log('ok')")},
		"templates/cert-details.html":          &fstest.MapFile{Data: []byte("<div></div>")},
		"templates/footer-status.html":         &fstest.MapFile{Data: []byte("<div></div>")},
		"templates/certs-fragment.html":        &fstest.MapFile{Data: []byte("{{template \"certs-rows\" .}}{{template \"dashboard-fragment\" .}}{{template \"certs-state\" .}}{{template \"certs-pagination\" .}}{{template \"certs-sort\" .}}")},
		"templates/certs-rows.html":            &fstest.MapFile{Data: []byte("{{define \"certs-rows\"}}{{end}}")},
		"templates/dashboard-fragment.html":    &fstest.MapFile{Data: []byte("{{define \"dashboard-fragment\"}}{{end}}")},
		"templates/theme-toggle-fragment.html": &fstest.MapFile{Data: []byte("<div></div>")},
		"templates/certs-state.html":           &fstest.MapFile{Data: []byte("{{define \"certs-state\"}}{{end}}")},
		"templates/certs-pagination.html":      &fstest.MapFile{Data: []byte("{{define \"certs-pagination\"}}{{end}}")},
		"templates/certs-sort.html":            &fstest.MapFile{Data: []byte("{{define \"certs-sort\"}}{{end}}")},
	}
}

func TestBuildRouter_BasicEndpoints(t *testing.T) {
	t.Setenv("VCV_ADMIN_PASSWORD", "")
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
	router, err := buildRouter(cfg, primary, statusClients, multi, registry, webFS, "")
	assert.NoError(t, err)

	t.Run("serves index", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
		assert.Contains(t, rec.Body.String(), "ok")
	})

	t.Run("serves asset", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "console.log")
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
	t.Setenv("VCV_ADMIN_PASSWORD", "")
	cfg := config.Config{Env: config.EnvDev}
	primary := &vault.MockClient{}
	multi := &vault.MockClient{}
	registry := prometheus.NewRegistry()
	webFS := fstest.MapFS{
		"index.html":                   &fstest.MapFile{Data: []byte("ok")},
		"templates/footer-status.html": &fstest.MapFile{Data: []byte("<div></div>")},
	}
	router, err := buildRouter(cfg, primary, map[string]vault.Client{}, multi, registry, webFS, "")
	assert.NotNil(t, router)
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/assets/missing.js", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestBuildRouter_MissingIndex_Returns500(t *testing.T) {
	t.Setenv("VCV_ADMIN_PASSWORD", "")
	cfg := config.Config{Env: config.EnvDev}
	primary := &vault.MockClient{}
	multi := &vault.MockClient{}
	registry := prometheus.NewRegistry()
	webFS := fstest.MapFS{
		"assets/app.js":                        &fstest.MapFile{Data: []byte("console.log('ok')")},
		"templates/cert-details.html":          &fstest.MapFile{Data: []byte("<div></div>")},
		"templates/footer-status.html":         &fstest.MapFile{Data: []byte("<div></div>")},
		"templates/certs-fragment.html":        &fstest.MapFile{Data: []byte("{{define \"certs-fragment\"}}{{end}}")},
		"templates/certs-rows.html":            &fstest.MapFile{Data: []byte("{{define \"certs-rows\"}}{{end}}")},
		"templates/dashboard-fragment.html":    &fstest.MapFile{Data: []byte("{{define \"dashboard-fragment\"}}{{end}}")},
		"templates/theme-toggle-fragment.html": &fstest.MapFile{Data: []byte("<div></div>")},
		"templates/certs-state.html":           &fstest.MapFile{Data: []byte("{{define \"certs-state\"}}{{end}}")},
		"templates/certs-pagination.html":      &fstest.MapFile{Data: []byte("{{define \"certs-pagination\"}}{{end}}")},
		"templates/certs-sort.html":            &fstest.MapFile{Data: []byte("{{define \"certs-sort\"}}{{end}}")},
	}
	router, err := buildRouter(cfg, primary, map[string]vault.Client{}, multi, registry, webFS, "")
	assert.NoError(t, err)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
