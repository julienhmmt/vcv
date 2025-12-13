package handlers_test

import (
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
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
	handlers.RegisterUIRoutes(router, mockVault, webFS, config.ExpirationThresholds{Critical: 7, Warning: 30})
	return router
}

func TestGetCertificateDetailsUI(t *testing.T) {
	webFS := fstest.MapFS{
		"templates/cert-details.html":       &fstest.MapFile{Data: []byte("<div id=\"cert-id\">{{.CertificateID}}</div>")},
		"templates/footer-status.html":      &fstest.MapFile{Data: []byte("<div>{{.VersionText}}</div>")},
		"templates/certs-fragment.html":     &fstest.MapFile{Data: []byte("{{template \"certs-rows\" .}}{{template \"dashboard-fragment\" .}}{{template \"certs-state\" .}}{{template \"certs-pagination\" .}}{{template \"certs-sort\" .}}")},
		"templates/certs-rows.html":         &fstest.MapFile{Data: []byte("{{define \"certs-rows\"}}{{range .Rows}}<div class=\"row\">{{.CommonName}}</div>{{end}}{{end}}")},
		"templates/dashboard-fragment.html": &fstest.MapFile{Data: []byte("{{define \"dashboard-fragment\"}}{{end}}")},
		"templates/certs-state.html":        &fstest.MapFile{Data: []byte("{{define \"certs-state\"}}<input id=\"vcv-page\" value=\"{{.PageIndex}}\" hx-swap-oob=\"true\" /><input id=\"vcv-sort-key\" value=\"{{.SortKey}}\" hx-swap-oob=\"true\" /><input id=\"vcv-sort-dir\" value=\"{{.SortDirection}}\" hx-swap-oob=\"true\" />{{end}}")},
		"templates/certs-pagination.html":   &fstest.MapFile{Data: []byte("{{define \"certs-pagination\"}}{{end}}")},
		"templates/certs-sort.html":         &fstest.MapFile{Data: []byte("{{define \"certs-sort\"}}{{end}}")},
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
		"templates/cert-details.html":       &fstest.MapFile{Data: []byte("<div id=\"cert-id\">{{.CertificateID}}</div>")},
		"templates/footer-status.html":      &fstest.MapFile{Data: []byte("<div>{{.VersionText}}</div>")},
		"templates/certs-fragment.html":     &fstest.MapFile{Data: []byte("{{template \"certs-rows\" .}}{{template \"certs-state\" .}}{{template \"certs-pagination\" .}}{{template \"certs-sort\" .}}{{template \"dashboard-fragment\" .}}")},
		"templates/certs-rows.html":         &fstest.MapFile{Data: []byte("{{define \"certs-rows\"}}{{range .Rows}}<div class=\"row\">{{.CommonName}}</div>{{end}}{{end}}")},
		"templates/dashboard-fragment.html": &fstest.MapFile{Data: []byte("{{define \"dashboard-fragment\"}}{{end}}")},
		"templates/certs-state.html":        &fstest.MapFile{Data: []byte("{{define \"certs-state\"}}<input id=\"vcv-page\" value=\"{{.PageIndex}}\" hx-swap-oob=\"true\" /><input id=\"vcv-sort-key\" value=\"{{.SortKey}}\" hx-swap-oob=\"true\" /><input id=\"vcv-sort-dir\" value=\"{{.SortDirection}}\" hx-swap-oob=\"true\" />{{end}}")},
		"templates/certs-pagination.html":   &fstest.MapFile{Data: []byte("{{define \"certs-pagination\"}}{{end}}")},
		"templates/certs-sort.html":         &fstest.MapFile{Data: []byte("{{define \"certs-sort\"}}{{end}}")},
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
