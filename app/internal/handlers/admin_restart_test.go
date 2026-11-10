package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vcv/internal/config"
)

func baseRestartSettings() config.SettingsFile {
	return config.SettingsFile{
		Certificates: config.CertificateSettings{
			ExpirationThresholds: config.ExpirationThresholds{Critical: 7, Warning: 30},
		},
		Vaults: []config.VaultInstance{
			{ID: "v1", Address: "https://vault.example.test:8200", Token: "tok", PKIMounts: []string{"pki"}},
			{ID: "v2", Address: "https://vault2.example.test:8200", Token: "tok2", PKIMounts: []string{"pki"}},
		},
	}
}

func TestRestartRequired(t *testing.T) {
	disabled := false
	tests := []struct {
		name   string
		mutate func(s *config.SettingsFile)
		want   bool
	}{
		{name: "no change", mutate: func(*config.SettingsFile) {}, want: false},
		{name: "enabled toggle only", mutate: func(s *config.SettingsFile) { s.Vaults[0].Enabled = &disabled }, want: false},
		{name: "display name only", mutate: func(s *config.SettingsFile) { s.Vaults[0].DisplayName = "Prod" }, want: false},
		{name: "vault added", mutate: func(s *config.SettingsFile) {
			s.Vaults = append(s.Vaults, config.VaultInstance{ID: "v3", Address: "https://v3.example.test:8200", Token: "t"})
		}, want: true},
		{name: "vault removed", mutate: func(s *config.SettingsFile) { s.Vaults = s.Vaults[:1] }, want: true},
		{name: "vault renamed", mutate: func(s *config.SettingsFile) { s.Vaults[0].ID = "renamed" }, want: true},
		{name: "address changed", mutate: func(s *config.SettingsFile) { s.Vaults[1].Address = "https://other.example.test:8200" }, want: true},
		{name: "token changed", mutate: func(s *config.SettingsFile) { s.Vaults[0].Token = "new" }, want: true},
		{name: "pki mounts changed", mutate: func(s *config.SettingsFile) { s.Vaults[0].PKIMounts = []string{"pki", "pki2"} }, want: true},
		{name: "legacy pki mount equivalent", mutate: func(s *config.SettingsFile) {
			s.Vaults[0].PKIMounts = nil
			s.Vaults[0].PKIMount = "pki"
		}, want: false},
		{name: "tls insecure changed", mutate: func(s *config.SettingsFile) { s.Vaults[0].TLSInsecure = true }, want: true},
		{name: "tls ca cert changed", mutate: func(s *config.SettingsFile) { s.Vaults[0].TLSCACert = "pem" }, want: true},
		{name: "tls ca path changed", mutate: func(s *config.SettingsFile) { s.Vaults[0].TLSCAPath = "/ca.pem" }, want: true},
		{name: "tls server name changed", mutate: func(s *config.SettingsFile) { s.Vaults[0].TLSServerName = "vault" }, want: true},
		{name: "thresholds changed", mutate: func(s *config.SettingsFile) { s.Certificates.ExpirationThresholds.Warning = 60 }, want: true},
		{name: "webhook change is live", mutate: func(s *config.SettingsFile) { s.Notifications.WebhookURL = "https://hook.example.test" }, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := baseRestartSettings()
			after := baseRestartSettings()
			tt.mutate(&after)
			assert.Equal(t, tt.want, restartRequired(before, after))
		})
	}
}

func putSettings(t *testing.T, r http.Handler, cookie *http.Cookie, settings config.SettingsFile) adminSettingsResponse {
	t.Helper()
	body, err := json.Marshal(settings)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings", bytes.NewReader(body))
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var resp adminSettingsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	return resp
}

func TestRegisterAdminAPIRoutes_SettingsPut_RestartRequired(t *testing.T) {
	r, _, _, _ := setupAdminAPIRouter(t)
	cookie := loginAdmin(t, r)

	thresholds := config.ExpirationThresholds{Critical: 7, Warning: 30}
	baseline := config.SettingsFile{
		Certificates: config.CertificateSettings{ExpirationThresholds: thresholds},
		Vaults: []config.VaultInstance{
			{ID: "v1", Address: "http://localhost:8200", Token: "", PKIMounts: []string{"pki"}},
		},
	}
	// Establish a known baseline; its own flag is not under test.
	putSettings(t, r, cookie, baseline)

	disabled := false
	toggled := baseline
	toggled.Vaults = []config.VaultInstance{
		{ID: "v1", Address: "http://localhost:8200", Token: "", PKIMounts: []string{"pki"}, Enabled: &disabled},
	}
	assert.False(t, putSettings(t, r, cookie, toggled).RestartRequired, "enabled toggle must not require restart")

	addressChanged := toggled
	addressChanged.Vaults = []config.VaultInstance{
		{ID: "v1", Address: "http://localhost:9999", Token: "", PKIMounts: []string{"pki"}, Enabled: &disabled},
	}
	assert.True(t, putSettings(t, r, cookie, addressChanged).RestartRequired, "address change must require restart")
}
