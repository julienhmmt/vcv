package handlers_test

import (
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"vcv/config"
	"vcv/internal/certs"
	"vcv/internal/handlers"
	"vcv/internal/vault"
	"vcv/middleware"
)

func setupUIRouter(mockVault *vault.MockClient, webFS fs.FS) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	handlers.RegisterUIRoutes(router, mockVault, []config.VaultInstance{}, map[string]vault.Client{}, webFS, config.ExpirationThresholds{Critical: 7, Warning: 30})
	return router
}

func TestStatusFragment_MultiVaultAddsSummaryPill(t *testing.T) {
	webFS := fstest.MapFS{
		"templates/status-indicator.html":      &fstest.MapFile{Data: []byte("{{if .Summary}}<button class=\"{{.Summary.Class}}\">{{.Summary.Text}}</button>{{end}}")},
		"templates/theme-toggle-fragment.html": &fstest.MapFile{Data: []byte("<div></div>")},
		"templates/cert-details.html":          &fstest.MapFile{Data: []byte("<div></div>")},
		"templates/certs-fragment.html":        &fstest.MapFile{Data: []byte("{{define \"certs-fragment\"}}{{end}}")},
		"templates/certs-rows.html":            &fstest.MapFile{Data: []byte("{{define \"certs-rows\"}}{{end}}")},
		"templates/certs-state.html":           &fstest.MapFile{Data: []byte("{{define \"certs-state\"}}{{end}}")},
		"templates/certs-pagination.html":      &fstest.MapFile{Data: []byte("{{define \"certs-pagination\"}}{{end}}")},
		"templates/certs-sort.html":            &fstest.MapFile{Data: []byte("{{define \"certs-sort\"}}{{end}}")},
		"templates/dashboard-fragment.html":    &fstest.MapFile{Data: []byte("{{define \"dashboard-fragment\"}}{{end}}")},
	}
	vaultInstances := []config.VaultInstance{{ID: "vault-1", DisplayName: "Vault 1"}, {ID: "vault-2", DisplayName: "Vault 2"}}
	statusClient1 := &vault.MockClient{}
	statusClient1.On("CheckConnection", mock.Anything).Return(nil)
	statusClient2 := &vault.MockClient{}
	statusClient2.On("CheckConnection", mock.Anything).Return(errors.New("boom"))
	vaultStatusClients := map[string]vault.Client{"vault-1": statusClient1, "vault-2": statusClient2}
	primaryClient := &vault.MockClient{}
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	handlers.RegisterUIRoutes(router, primaryClient, vaultInstances, vaultStatusClients, webFS, config.ExpirationThresholds{Critical: 7, Warning: 30})
	req := httptest.NewRequest(http.MethodGet, "/ui/status?lang=en", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "Vaults: 1/2 up")
	assert.Contains(t, body, "vcv-status-state-error")
	statusClient1.AssertExpectations(t)
	statusClient2.AssertExpectations(t)
}

func TestToggleThemeFragment(t *testing.T) {
	webFS := fstest.MapFS{
		"templates/cert-details.html":          &fstest.MapFile{Data: []byte("<div></div>")},
		"templates/status-indicator.html":      &fstest.MapFile{Data: []byte("<div></div>")},
		"templates/certs-fragment.html":        &fstest.MapFile{Data: []byte("{{define \"certs-fragment\"}}{{end}}")},
		"templates/certs-rows.html":            &fstest.MapFile{Data: []byte("{{define \"certs-rows\"}}{{end}}")},
		"templates/certs-state.html":           &fstest.MapFile{Data: []byte("{{define \"certs-state\"}}{{end}}")},
		"templates/certs-pagination.html":      &fstest.MapFile{Data: []byte("{{define \"certs-pagination\"}}{{end}}")},
		"templates/certs-sort.html":            &fstest.MapFile{Data: []byte("{{define \"certs-sort\"}}{{end}}")},
		"templates/dashboard-fragment.html":    &fstest.MapFile{Data: []byte("{{define \"dashboard-fragment\"}}{{end}}")},
		"templates/theme-toggle-fragment.html": &fstest.MapFile{Data: []byte("<span id=\"theme-icon\" hx-swap-oob=\"true\">{{.Icon}}</span><input id=\"vcv-theme-value\" hx-swap-oob=\"true\" value=\"{{.Theme}}\" />")},
	}
	mockVault := &vault.MockClient{}
	router := setupUIRouter(mockVault, webFS)
	req := httptest.NewRequest(http.MethodPost, "/ui/theme/toggle", strings.NewReader("theme=light"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "id=\"theme-icon\"")
	assert.Contains(t, body, "value=\"dark\"")
}

func TestGetCertificateDetailsUI(t *testing.T) {
	webFS := fstest.MapFS{
		"templates/cert-details.html":          &fstest.MapFile{Data: []byte("<div id=\"cert-id\">{{.CertificateID}}</div>")},
		"templates/status-indicator.html":      &fstest.MapFile{Data: []byte("<div>{{.VersionText}}</div>")},
		"templates/certs-fragment.html":        &fstest.MapFile{Data: []byte("{{template \"certs-rows\" .}}{{template \"dashboard-fragment\" .}}{{template \"certs-state\" .}}{{template \"certs-pagination\" .}}{{template \"certs-sort\" .}}")},
		"templates/certs-rows.html":            &fstest.MapFile{Data: []byte("{{define \"certs-rows\"}}{{range .Rows}}<div class=\"row\">{{.CommonName}}</div>{{end}}{{end}}")},
		"templates/dashboard-fragment.html":    &fstest.MapFile{Data: []byte("{{define \"dashboard-fragment\"}}{{end}}")},
		"templates/theme-toggle-fragment.html": &fstest.MapFile{Data: []byte("<span id=\"theme-icon\" hx-swap-oob=\"true\">{{.Icon}}</span><input id=\"vcv-theme-value\" hx-swap-oob=\"true\" value=\"{{.Theme}}\" />")},
		"templates/certs-state.html":           &fstest.MapFile{Data: []byte("{{define \"certs-state\"}}<input id=\"vcv-page\" value=\"{{.PageIndex}}\" hx-swap-oob=\"true\" /><input id=\"vcv-sort-key\" value=\"{{.SortKey}}\" hx-swap-oob=\"true\" /><input id=\"vcv-sort-dir\" value=\"{{.SortDirection}}\" hx-swap-oob=\"true\" />{{end}}")},
		"templates/certs-pagination.html":      &fstest.MapFile{Data: []byte("{{define \"certs-pagination\"}}{{end}}")},
		"templates/certs-sort.html":            &fstest.MapFile{Data: []byte("{{define \"certs-sort\"}}{{end}}")},
	}
	tests := []struct {
		name                 string
		path                 string
		setupMock            func(mockVault *vault.MockClient)
		expectedStatus       int
		expectedBodyContains string
	}{
		{
			name: "success unescapes id",
			path: "/ui/certs/pki_dev%3A33%3Aaa/details",
			setupMock: func(mockVault *vault.MockClient) {
				mockVault.On("GetCertificateDetails", mock.Anything, "pki_dev:33:aa").Return(certs.DetailedCertificate{Certificate: certs.Certificate{ID: "pki_dev:33:aa", CommonName: "cn", ExpiresAt: time.Now()}, SerialNumber: "33:aa"}, nil)
			},
			expectedStatus:       http.StatusOK,
			expectedBodyContains: "pki_dev:33:aa",
		},
		{
			name:                 "bad request when id is missing",
			path:                 "/ui/certs//details",
			setupMock:            func(mockVault *vault.MockClient) {},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "",
		},
		{
			name: "internal server error when vault fails",
			path: "/ui/certs/pki_dev%3A33%3Aaa/details",
			setupMock: func(mockVault *vault.MockClient) {
				mockVault.On("GetCertificateDetails", mock.Anything, "pki_dev:33:aa").Return(certs.DetailedCertificate{}, errors.New("boom"))
			},
			expectedStatus:       http.StatusInternalServerError,
			expectedBodyContains: "",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockVault := new(vault.MockClient)
			tt.setupMock(mockVault)
			router := setupUIRouter(mockVault, webFS)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBodyContains != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBodyContains)
			}
			mockVault.AssertExpectations(t)
		})
	}
}

func TestGetCertificatesFragment(t *testing.T) {
	webFS := fstest.MapFS{
		"templates/cert-details.html":          &fstest.MapFile{Data: []byte("<div id=\"cert-id\">{{.CertificateID}}</div>")},
		"templates/status-indicator.html":      &fstest.MapFile{Data: []byte("<div>{{.VersionText}}</div>")},
		"templates/certs-fragment.html":        &fstest.MapFile{Data: []byte("{{template \"certs-rows\" .}}{{template \"certs-state\" .}}{{template \"certs-pagination\" .}}{{template \"certs-sort\" .}}{{template \"dashboard-fragment\" .}}")},
		"templates/certs-rows.html":            &fstest.MapFile{Data: []byte("{{define \"certs-rows\"}}{{range .Rows}}<div class=\"row\">{{.CommonName}}</div>{{end}}{{end}}")},
		"templates/dashboard-fragment.html":    &fstest.MapFile{Data: []byte("{{define \"dashboard-fragment\"}}{{end}}")},
		"templates/theme-toggle-fragment.html": &fstest.MapFile{Data: []byte("<span id=\"theme-icon\" hx-swap-oob=\"true\">{{.Icon}}</span><input id=\"vcv-theme-value\" hx-swap-oob=\"true\" value=\"{{.Theme}}\" />")},
		"templates/certs-state.html":           &fstest.MapFile{Data: []byte("{{define \"certs-state\"}}<input id=\"vcv-page\" value=\"{{.PageIndex}}\" hx-swap-oob=\"true\" /><input id=\"vcv-sort-key\" value=\"{{.SortKey}}\" hx-swap-oob=\"true\" /><input id=\"vcv-sort-dir\" value=\"{{.SortDirection}}\" hx-swap-oob=\"true\" />{{end}}")},
		"templates/certs-pagination.html":      &fstest.MapFile{Data: []byte("{{define \"certs-pagination\"}}{{end}}")},
		"templates/certs-sort.html":            &fstest.MapFile{Data: []byte("{{define \"certs-sort\"}}{{end}}")},
	}
	certificates := []certs.Certificate{
		{ID: "pki:a", CommonName: "alpha.example", Sans: []string{"alpha"}, CreatedAt: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC), ExpiresAt: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)},
		{ID: "pki:b", CommonName: "beta.example", Sans: []string{"beta"}, CreatedAt: time.Date(2025, 2, 1, 10, 0, 0, 0, time.UTC), ExpiresAt: time.Date(2027, 1, 1, 10, 0, 0, 0, time.UTC)},
		{ID: "pki:c", CommonName: "gamma.example", Sans: []string{"gamma"}, CreatedAt: time.Date(2025, 3, 1, 10, 0, 0, 0, time.UTC), ExpiresAt: time.Date(2028, 1, 1, 10, 0, 0, 0, time.UTC)},
	}
	tests := []struct {
		name          string
		path          string
		headerTrigger string
		assertBody    func(t *testing.T, body string)
	}{
		{
			name: "success renders rows",
			path: "/ui/certs",
			assertBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "alpha.example")
				assert.Contains(t, body, "beta.example")
				assert.Contains(t, body, "gamma.example")
			},
		},
		{
			name: "search filters",
			path: "/ui/certs?search=beta",
			assertBody: func(t *testing.T, body string) {
				assert.NotContains(t, body, "alpha.example")
				assert.Contains(t, body, "beta.example")
				assert.NotContains(t, body, "gamma.example")
			},
		},
		{
			name: "pagination next advances page",
			path: "/ui/certs?pageSize=1&page=0&pageAction=next",
			assertBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "beta.example")
				assert.Contains(t, body, "id=\"vcv-page\" value=\"1\"")
			},
		},
		{
			name: "sort toggle changes direction",
			path: "/ui/certs?sortKey=commonName&sortDir=asc&sort=commonName",
			assertBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "id=\"vcv-sort-dir\" value=\"desc\"")
			},
		},
		{
			name: "mounts empty returns no rows",
			path: "/ui/certs?mounts=",
			assertBody: func(t *testing.T, body string) {
				assert.NotContains(t, body, "alpha.example")
				assert.NotContains(t, body, "beta.example")
				assert.NotContains(t, body, "gamma.example")
			},
		},
		{
			name: "mounts sentinel returns all rows",
			path: "/ui/certs?mounts=__all__",
			assertBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "alpha.example")
				assert.Contains(t, body, "beta.example")
				assert.Contains(t, body, "gamma.example")
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			mockVault := &vault.MockClient{}
			mockVault.On("ListCertificates", mock.Anything).Return(certificates, nil)
			router := setupUIRouter(mockVault, webFS)
			req := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			if testCase.headerTrigger != "" {
				req.Header.Set("HX-Trigger", testCase.headerTrigger)
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			assert.Equal(t, http.StatusOK, recorder.Code)
			body := recorder.Body.String()
			testCase.assertBody(t, body)
			mockVault.AssertExpectations(t)
		})
	}
}
