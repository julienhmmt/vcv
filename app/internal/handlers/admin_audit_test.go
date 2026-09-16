package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"vcv/internal/config"
	"vcv/internal/handlers"
	"vcv/internal/logger"
)

func auditTestRouter(t *testing.T, password string) (*chi.Mux, string) {
	t.Helper()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	settingsPath := t.TempDir() + "/settings.json"
	settings := config.SettingsFile{
		App:   config.AppSettings{Env: "dev", Port: 52000},
		Admin: config.AdminSettings{Password: string(hashedPassword)},
	}
	data, err := json.Marshal(settings)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(settingsPath, data, 0644))

	r := chi.NewRouter()
	handlers.RegisterAdminRoutes(r, settingsPath, config.EnvDev, nil, nil, nil, false)
	return r, settingsPath
}

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	t.Cleanup(func() {
		logger.SetOutput(os.Stderr)
	})
	return &buf
}

// containsSuccess matches the success flag in both zerolog console
// (success=false) and JSON ("success":false) output.
func containsSuccess(t *testing.T, output string, ok bool) {
	t.Helper()
	want := "success=false"
	if ok {
		want = "success=true"
	}
	jsonWant := `"success":false`
	if ok {
		jsonWant = `"success":true`
	}
	assert.True(t, strings.Contains(output, want) || strings.Contains(output, jsonWant),
		"expected audit success=%v in %q", ok, output)
}

func adminLogin(t *testing.T, r *chi.Mux, username, password string) (int, []*http.Cookie) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code, w.Result().Cookies()
}

func TestAdminAudit_LoginFailure(t *testing.T) {
	r, _ := auditTestRouter(t, "correct-horse")
	buf := captureLogs(t)

	code, _ := adminLogin(t, r, "admin", "wrong-password")

	assert.Equal(t, http.StatusUnauthorized, code)
	output := buf.String()
	assert.Contains(t, output, "admin.login")
	assert.Contains(t, output, "audit")
	assert.Contains(t, output, "admin")
	containsSuccess(t, output, false)
	assert.NotContains(t, output, "wrong-password")
}

func TestAdminAudit_LoginSuccessLogoutAndSettingsPut(t *testing.T) {
	r, _ := auditTestRouter(t, "correct-horse")
	buf := captureLogs(t)

	code, cookies := adminLogin(t, r, "admin", "correct-horse")
	require.Equal(t, http.StatusOK, code)
	assert.Contains(t, buf.String(), "admin.login")
	containsSuccess(t, buf.String(), true)
	assert.NotContains(t, buf.String(), "correct-horse")

	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "vcv_admin_session" {
			sessionCookie = c
		}
	}
	require.NotNil(t, sessionCookie)

	buf.Reset()
	settingsBody, _ := json.Marshal(config.SettingsFile{
		App:   config.AppSettings{Env: "dev", Port: 52000},
		Admin: config.AdminSettings{Password: "x"},
	})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings", bytes.NewReader(settingsBody))
	req.AddCookie(sessionCookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	// Password "x" is not a bcrypt hash; the save path rejects it, but the
	// audit line fires either way.
	assert.Contains(t, buf.String(), "admin.settings_put")

	buf.Reset()
	req = httptest.NewRequest(http.MethodPost, "/api/admin/logout", nil)
	req.AddCookie(sessionCookie)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Contains(t, buf.String(), "admin.logout")
}
