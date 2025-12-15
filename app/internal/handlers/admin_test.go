package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vcv/config"
)

func newAdminWebFS() fstest.MapFS {
	return fstest.MapFS{
		"templates/admin-page.html":           &fstest.MapFile{Data: []byte("<html><body><div id=\"admin-root\"></div></body></html>")},
		"templates/admin-login-fragment.html": &fstest.MapFile{Data: []byte("<div>login</div>")},
		"templates/admin-panel-fragment.html": &fstest.MapFile{Data: []byte("<div>panel</div>")},
		"templates/admin-vault-item.html":     &fstest.MapFile{Data: []byte("<div>vault</div>")},
	}
}

func TestRegisterAdminRoutes_DisabledWithoutPassword(t *testing.T) {
	router := chi.NewRouter()
	t.Setenv("VCV_ADMIN_PASSWORD", "")
	RegisterAdminRoutes(router, newAdminWebFS(), t.TempDir()+"/settings.json", config.EnvDev)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminLoginAndSettingsRoundtrip(t *testing.T) {
	settingsPath := filepath.Join(t.TempDir(), "settings.json")
	router := chi.NewRouter()
	t.Setenv("VCV_ADMIN_PASSWORD", "secret")
	RegisterAdminRoutes(router, newAdminWebFS(), settingsPath, config.EnvDev)
	loginPayload, err := json.Marshal(map[string]string{"username": "admin", "password": "secret"})
	require.NoError(t, err)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewReader(loginPayload))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	require.Equal(t, http.StatusNoContent, loginRec.Code)
	cookies := loginRec.Result().Cookies()
	require.NotEmpty(t, cookies)

	getReq := httptest.NewRequest(http.MethodGet, "/api/admin/settings", nil)
	getReq.AddCookie(cookies[0])
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	require.Equal(t, http.StatusOK, getRec.Code)
	var initial config.SettingsFile
	require.NoError(t, json.NewDecoder(getRec.Body).Decode(&initial))
	assert.Equal(t, "dev", initial.App.Env)

	updated := initial
	updated.App.Env = "prod"
	updated.Certificates.ExpirationThresholds.Critical = 5
	updated.Certificates.ExpirationThresholds.Warning = 25
	enabled := true
	updated.Vaults = []config.VaultInstance{{
		ID:          "v1",
		Address:     "https://vault.example.com:8200",
		Token:       "token",
		PKIMount:    "pki",
		PKIMounts:   []string{"pki"},
		DisplayName: "vault",
		TLSInsecure: false,
		Enabled:     &enabled,
	}}
	payload, err := json.Marshal(updated)
	require.NoError(t, err)
	putReq := httptest.NewRequest(http.MethodPut, "/api/admin/settings", bytes.NewReader(payload))
	putReq.Header.Set("Content-Type", "application/json")
	putReq.AddCookie(cookies[0])
	putRec := httptest.NewRecorder()
	router.ServeHTTP(putRec, putReq)
	require.Equal(t, http.StatusNoContent, putRec.Code)

	getReq2 := httptest.NewRequest(http.MethodGet, "/api/admin/settings", nil)
	getReq2.AddCookie(cookies[0])
	getRec2 := httptest.NewRecorder()
	router.ServeHTTP(getRec2, getReq2)
	require.Equal(t, http.StatusOK, getRec2.Code)
	var after config.SettingsFile
	require.NoError(t, json.NewDecoder(getRec2.Body).Decode(&after))
	assert.Equal(t, "prod", after.App.Env)
	assert.Equal(t, 5, after.Certificates.ExpirationThresholds.Critical)
	assert.Equal(t, 25, after.Certificates.ExpirationThresholds.Warning)
	require.Len(t, after.Vaults, 1)
	assert.Equal(t, "v1", after.Vaults[0].ID)

	fileBytes, readErr := os.ReadFile(settingsPath)
	require.NoError(t, readErr)
	assert.Contains(t, string(fileBytes), "\"vaults\"")
}
