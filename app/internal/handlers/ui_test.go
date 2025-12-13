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

	"vcv/internal/certs"
	"vcv/internal/handlers"
	"vcv/internal/vault"
	"vcv/middleware"
)

func setupUIRouter(mockVault *vault.MockClient, webFS fs.FS) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	handlers.RegisterUIRoutes(router, mockVault, webFS)
	return router
}

func TestGetCertificateDetailsUI(t *testing.T) {
	webFS := fstest.MapFS{
		"templates/cert-details.html":  &fstest.MapFile{Data: []byte("<div id=\"cert-id\">{{.CertificateID}}</div>")},
		"templates/footer-status.html": &fstest.MapFile{Data: []byte("<div>{{.VersionText}}</div>")},
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
