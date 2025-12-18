package handlers

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"testing/fstest"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"vcv/config"
)

func newAdminWebFS() fstest.MapFS {
	return fstest.MapFS{
		"templates/admin-page.html":           &fstest.MapFile{Data: []byte("<html><body><div id=\"admin-root\"></div></body></html>")},
		"templates/admin-login-fragment.html": &fstest.MapFile{Data: []byte("<div>login</div>")},
		"templates/admin-panel-fragment.html": &fstest.MapFile{Data: []byte("<div>panel {{.ErrorText}} {{.SuccessText}} {{.CorsOriginsText}}{{range .VaultViews}}<span class=\"vault\">{{.Vault.ID}}</span>{{end}}</div>")},
		"templates/admin-vault-item.html":     &fstest.MapFile{Data: []byte("<div>vault {{.Key}}</div>")},
	}
}

func mustBcryptPasswordHash(t *testing.T, value string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(hash)
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

func TestRegisterAdminRoutes_DisabledWithPlaintextPassword(t *testing.T) {
	router := chi.NewRouter()
	t.Setenv("VCV_ADMIN_PASSWORD", "secret")
	RegisterAdminRoutes(router, newAdminWebFS(), t.TempDir()+"/settings.json", config.EnvDev)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAdminLoginAndSettingsRoundtrip(t *testing.T) {
	settingsPath := filepath.Join(t.TempDir(), "settings.json")
	router := chi.NewRouter()
	t.Setenv("VCV_ADMIN_PASSWORD", mustBcryptPasswordHash(t, "secret"))
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
		ID:              "v1",
		Address:         "https://vault.example.com:8200",
		Token:           "token",
		PKIMount:        "pki",
		PKIMounts:       []string{"pki"},
		DisplayName:     "vault",
		TLSCACertBase64: "ZHVtbXk",
		TLSCACert:       "/etc/vcv/tls/vault-ca.pem",
		TLSCAPath:       "/etc/vcv/tls/ca",
		TLSServerName:   "vault.service.consul",
		TLSInsecure:     false,
		Enabled:         &enabled,
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
	assert.Equal(t, "ZHVtbXk", after.Vaults[0].TLSCACertBase64)
	assert.Equal(t, "/etc/vcv/tls/vault-ca.pem", after.Vaults[0].TLSCACert)
	assert.Equal(t, "/etc/vcv/tls/ca", after.Vaults[0].TLSCAPath)
	assert.Equal(t, "vault.service.consul", after.Vaults[0].TLSServerName)

	fileBytes, readErr := os.ReadFile(settingsPath)
	require.NoError(t, readErr)
	assert.Contains(t, string(fileBytes), "\"vaults\"")
	assert.Contains(t, string(fileBytes), "\"tls_ca_cert_base64\": \"ZHVtbXk\"")
	assert.Contains(t, string(fileBytes), "\"tls_ca_cert\": \"/etc/vcv/tls/vault-ca.pem\"")
}

func TestAdminSessionStore_LoginFromForm_SetsCookie(t *testing.T) {
	store := newAdminSessionStore(mustBcryptPasswordHash(t, "secret"), false)
	req := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader("username=admin&password=secret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ok, _ := store.loginFromForm(rec, req)
	assert.True(t, ok)
	result := rec.Result()
	cookies := result.Cookies()
	require.NotEmpty(t, cookies)
	assert.Equal(t, adminCookieName, cookies[0].Name)
	assert.NotEmpty(t, cookies[0].Value)
}

func TestAdminSessionStore_LogoutJSON_ClearsCookie(t *testing.T) {
	store := newAdminSessionStore(mustBcryptPasswordHash(t, "secret"), false)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/logout", nil)
	rec := httptest.NewRecorder()
	store.logoutJSON(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	result := rec.Result()
	cookies := result.Cookies()
	require.NotEmpty(t, cookies)
	assert.Equal(t, adminCookieName, cookies[0].Name)
	assert.Equal(t, "", cookies[0].Value)
}

func TestAdminSessionStore_RequireAuth_Unauthorized_WhenMissingCookie(t *testing.T) {
	store := newAdminSessionStore(mustBcryptPasswordHash(t, "secret"), false)
	h := store.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAdminSessionStore_IsAuthed_ExpiredSession(t *testing.T) {
	store := newAdminSessionStore(mustBcryptPasswordHash(t, "secret"), false)
	token := "tok"
	store.sessions[token] = time.Now().Add(-1 * time.Minute)
	req := httptest.NewRequest(http.MethodGet, "/admin/panel", nil)
	req.AddCookie(&http.Cookie{Name: adminCookieName, Value: token})
	ok := store.isAuthed(req)
	assert.False(t, ok)
	_, stillThere := store.sessions[token]
	assert.False(t, stillThere)
}

func TestParseSettingsUpdateForm_InvalidThresholds(t *testing.T) {
	base := config.SettingsFile{App: config.AppSettings{Env: "dev", Port: 52000}}
	reqCritical := httptest.NewRequest(http.MethodPost, "/admin/settings", strings.NewReader("expire_critical=abc&expire_warning=10"))
	reqCritical.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_, errCritical := parseSettingsUpdateForm(reqCritical, base)
	require.Error(t, errCritical)
	assert.Contains(t, errCritical.Error(), "invalid critical")
	reqWarning := httptest.NewRequest(http.MethodPost, "/admin/settings", strings.NewReader("expire_critical=1&expire_warning=abc"))
	reqWarning.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_, errWarning := parseSettingsUpdateForm(reqWarning, base)
	require.Error(t, errWarning)
	assert.Contains(t, errWarning.Error(), "invalid warning")
}

func TestSplitAndTrim(t *testing.T) {
	result := splitAndTrim(" a, ,b ,  ,c")
	assert.Equal(t, []string{"a", "b", "c"}, result)
}

func TestExtractVaultKeys_Sorted(t *testing.T) {
	form := url.Values{}
	form.Set("vault_id_2", "a")
	form.Set("vault_id_10", "b")
	form.Set("vault_id_1", "c")
	keys := extractVaultKeys(form)
	assert.Equal(t, []string{"1", "10", "2"}, keys)
}

func TestParseVaultsFromForm_Defaults(t *testing.T) {
	form := url.Values{}
	form.Set("vault_id_1", "v1")
	form.Set("vault_address_1", "https://vault.example.com")
	form.Set("vault_token_1", "tok")
	form.Set("vault_tls_1", "on")
	form.Set("vault_id_2", "v2")
	form.Set("vault_address_2", "https://vault2.example.com")
	form.Set("vault_token_2", "tok2")
	form.Set("vault_enabled_2", "on")
	vaults := parseVaultsFromForm(form)
	require.Len(t, vaults, 2)
	assert.Equal(t, "pki", vaults[0].PKIMount)
	assert.Equal(t, []string{}, vaults[0].PKIMounts)
	assert.True(t, vaults[0].TLSInsecure)
	require.NotNil(t, vaults[0].Enabled)
	assert.False(t, *vaults[0].Enabled)
	require.NotNil(t, vaults[1].Enabled)
	assert.True(t, *vaults[1].Enabled)
}

func TestShouldFallbackToDirectWrite(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		expect bool
	}{
		{name: "nil", err: nil, expect: false},
		{name: "permission", err: &os.PathError{Op: "write", Path: "/nope", Err: fs.ErrPermission}, expect: true},
		{name: "rofs", err: syscall.EROFS, expect: true},
		{name: "eperm", err: syscall.EPERM, expect: true},
		{name: "eacces", err: syscall.EACCES, expect: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, shouldFallbackToDirectWrite(tt.err))
		})
	}
}

func TestFallbackWriteSettings_WritesWhenFallbackAllowed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	payload := []byte("{\"vaults\":[]}")
	err := fallbackWriteSettings(path, payload, &os.PathError{Op: "rename", Path: path, Err: syscall.EPERM})
	require.NoError(t, err)
	content, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, string(payload), string(content))
}

func TestRenderAdminTemplate_WritesHTMLResponse(t *testing.T) {
	templates, err := template.New("root").Parse(`{{define "hello"}}<div>hello</div>{{end}}`)
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	writeErr := renderAdminTemplate(rec, templates, "hello", nil)
	require.NoError(t, writeErr)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, rec.Body.String(), "hello")
}

func TestBuildAdminPanelData_BuildsVaultViewsAndCORS(t *testing.T) {
	enabled := true
	settings := config.SettingsFile{
		CORS: config.CORSSettings{AllowedOrigins: []string{"http://a", "http://b"}},
		Vaults: []config.VaultInstance{
			{ID: "v1", Address: "https://vault.example.com", Token: "tok", PKIMount: "pki", PKIMounts: []string{"pki", "pki_dev"}, DisplayName: "Vault One", Enabled: &enabled},
			{ID: "v2", Address: "https://vault2.example.com", Token: "tok2", PKIMount: "", PKIMounts: []string{}, DisplayName: "", Enabled: &enabled},
		},
	}
	data := buildAdminPanelData(settings, "ok", "")
	assert.Equal(t, "http://a,http://b", data.CorsOriginsText)
	assert.Equal(t, "ok", data.SuccessText)
	require.Len(t, data.VaultViews, 2)
	assert.Equal(t, "0", data.VaultViews[0].Key)
	assert.Equal(t, "pki,pki_dev", data.VaultViews[0].MountsText)
	assert.Equal(t, "1", data.VaultViews[1].Key)
	assert.Equal(t, "", data.VaultViews[1].MountsText)
}

func TestNewVaultKey_GeneratesValue(t *testing.T) {
	key, err := newVaultKey()
	require.NoError(t, err)
	assert.NotEmpty(t, key)
}

func TestAdminRoutes_HTMXPanelLoginLogoutAndVaultActions(t *testing.T) {
	settingsPath := filepath.Join(t.TempDir(), "settings.json")
	router := chi.NewRouter()
	t.Setenv("VCV_ADMIN_PASSWORD", mustBcryptPasswordHash(t, "secret"))
	RegisterAdminRoutes(router, newAdminWebFS(), settingsPath, config.EnvDev)

	panelReq := httptest.NewRequest(http.MethodGet, "/admin/panel", nil)
	panelRec := httptest.NewRecorder()
	router.ServeHTTP(panelRec, panelReq)
	assert.Equal(t, http.StatusOK, panelRec.Code)
	assert.Contains(t, panelRec.Body.String(), "login")

	loginReq := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader("username=admin&password=secret"))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	assert.Equal(t, http.StatusOK, loginRec.Code)
	assert.Contains(t, loginRec.Body.String(), "panel")
	cookies := loginRec.Result().Cookies()
	require.NotEmpty(t, cookies)

	authedPanelReq := httptest.NewRequest(http.MethodGet, "/admin/panel", nil)
	authedPanelReq.AddCookie(cookies[0])
	authedPanelRec := httptest.NewRecorder()
	router.ServeHTTP(authedPanelRec, authedPanelReq)
	assert.Equal(t, http.StatusOK, authedPanelRec.Code)
	assert.Contains(t, authedPanelRec.Body.String(), "panel")

	addReq := httptest.NewRequest(http.MethodPost, "/admin/vault/add", nil)
	addReq.AddCookie(cookies[0])
	addRec := httptest.NewRecorder()
	router.ServeHTTP(addRec, addReq)
	assert.Equal(t, http.StatusOK, addRec.Code)
	assert.Contains(t, addRec.Header().Get("HX-Trigger-After-Swap"), "adminVaultAdded")
	assert.Contains(t, addRec.Body.String(), "vault")

	removeReqEmpty := httptest.NewRequest(http.MethodPost, "/admin/vault/remove", strings.NewReader("vaultId="))
	removeReqEmpty.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	removeReqEmpty.AddCookie(cookies[0])
	removeRecEmpty := httptest.NewRecorder()
	router.ServeHTTP(removeRecEmpty, removeReqEmpty)
	assert.Equal(t, http.StatusOK, removeRecEmpty.Code)

	logoutReq := httptest.NewRequest(http.MethodPost, "/admin/logout", nil)
	logoutReq.AddCookie(cookies[0])
	logoutRec := httptest.NewRecorder()
	router.ServeHTTP(logoutRec, logoutReq)
	assert.Equal(t, http.StatusOK, logoutRec.Code)
	assert.Contains(t, logoutRec.Body.String(), "login")
}

func TestAdminRoutes_SettingsPost_ErrorsAndSuccess(t *testing.T) {
	settingsPath := filepath.Join(t.TempDir(), "settings.json")
	router := chi.NewRouter()
	t.Setenv("VCV_ADMIN_PASSWORD", mustBcryptPasswordHash(t, "secret"))
	RegisterAdminRoutes(router, newAdminWebFS(), settingsPath, config.EnvDev)

	loginReq := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader("username=admin&password=secret"))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	cookies := loginRec.Result().Cookies()
	require.NotEmpty(t, cookies)

	badFormReq := httptest.NewRequest(http.MethodPost, "/admin/settings", strings.NewReader("expire_critical=abc&expire_warning=10"))
	badFormReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	badFormReq.AddCookie(cookies[0])
	badFormRec := httptest.NewRecorder()
	router.ServeHTTP(badFormRec, badFormReq)
	assert.Equal(t, http.StatusOK, badFormRec.Code)
	assert.Contains(t, badFormRec.Body.String(), "invalid critical")

	invalidVaultReq := httptest.NewRequest(http.MethodPost, "/admin/settings", strings.NewReader("expire_critical=1&expire_warning=2&vault_id_1=&vault_address_1=https://vault.example.com&vault_token_1=tok&vault_mounts_1=pki"))
	invalidVaultReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	invalidVaultReq.AddCookie(cookies[0])
	invalidVaultRec := httptest.NewRecorder()
	router.ServeHTTP(invalidVaultRec, invalidVaultReq)
	assert.Equal(t, http.StatusOK, invalidVaultRec.Code)
	assert.Contains(t, invalidVaultRec.Body.String(), "vault id is empty")

	goodReq := httptest.NewRequest(http.MethodPost, "/admin/settings", strings.NewReader("expire_critical=1&expire_warning=2&cors_origins=http://a,http://b&vault_id_1=v1&vault_address_1=https://vault.example.com&vault_token_1=tok&vault_mounts_1=pki&vault_tls_ca_cert_base64_1=ZHVtbXk&vault_tls_ca_cert_1=/etc/vcv/tls/vault-ca.pem&vault_tls_ca_path_1=/etc/vcv/tls/ca&vault_tls_server_name_1=vault.service.consul&vault_enabled_1=on"))
	goodReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	goodReq.AddCookie(cookies[0])
	goodRec := httptest.NewRecorder()
	router.ServeHTTP(goodRec, goodReq)
	assert.Equal(t, http.StatusOK, goodRec.Code)
	assert.Contains(t, goodRec.Body.String(), "Settings saved")

	fileBytes, err := os.ReadFile(settingsPath)
	require.NoError(t, err)
	assert.Contains(t, string(fileBytes), "\"id\": \"v1\"")
	assert.Contains(t, string(fileBytes), "\"tls_ca_cert_base64\": \"ZHVtbXk\"")
	assert.Contains(t, string(fileBytes), "\"tls_ca_cert\": \"/etc/vcv/tls/vault-ca.pem\"")
	assert.Contains(t, string(fileBytes), "\"tls_ca_path\": \"/etc/vcv/tls/ca\"")
	assert.Contains(t, string(fileBytes), "\"tls_server_name\": \"vault.service.consul\"")
}

func TestAdminRoutes_VaultRemove_PersistsToSettings(t *testing.T) {
	settingsPath := filepath.Join(t.TempDir(), "settings.json")
	initial := `{"app":{"env":"dev","port":52000,"logging":{"level":"debug","format":"json","output":"stdout","file_path":""}},"cors":{"allowed_origins":[],"allow_credentials":true},"certificates":{"expiration_thresholds":{"critical":1,"warning":2}},"vaults":[{"id":"v1","address":"https://vault.example.com","token":"tok","pki_mount":"pki","pki_mounts":["pki"],"display_name":"v1","tls_insecure":false,"enabled":true}]}`
	require.NoError(t, os.WriteFile(settingsPath, []byte(initial), 0o644))

	router := chi.NewRouter()
	t.Setenv("VCV_ADMIN_PASSWORD", mustBcryptPasswordHash(t, "secret"))
	RegisterAdminRoutes(router, newAdminWebFS(), settingsPath, config.EnvDev)

	loginReq := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader("username=admin&password=secret"))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)
	cookies := loginRec.Result().Cookies()
	require.NotEmpty(t, cookies)

	removeReq := httptest.NewRequest(http.MethodPost, "/admin/vault/remove", strings.NewReader("vaultId=v1"))
	removeReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	removeReq.AddCookie(cookies[0])
	removeRec := httptest.NewRecorder()
	router.ServeHTTP(removeRec, removeReq)
	assert.Equal(t, http.StatusOK, removeRec.Code)

	content, err := os.ReadFile(settingsPath)
	require.NoError(t, err)
	assert.NotContains(t, string(content), "\"id\": \"v1\"")
}
