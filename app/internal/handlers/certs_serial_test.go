package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"vcv/internal/certs"
	"vcv/internal/vault"
)

func TestCertEndpoints_InvalidSerialReturnsBadRequest(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		mockOn func(*vault.MockClient)
	}{
		{
			name: "details",
			path: "/api/certs/pki:zz/details",
			mockOn: func(m *vault.MockClient) {
				m.On("GetCertificateDetails", mock.Anything, "pki:zz").Return(certs.DetailedCertificate{}, fmt.Errorf("wrapped: %w", vault.ErrInvalidSerial))
			},
		},
		{
			name: "pem",
			path: "/api/certs/pki:zz/pem",
			mockOn: func(m *vault.MockClient) {
				m.On("GetCertificatePEM", mock.Anything, "pki:zz").Return(certs.PEMResponse{}, fmt.Errorf("wrapped: %w", vault.ErrInvalidSerial))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockVault := new(vault.MockClient)
			tt.mockOn(mockVault)
			router := setupRouter(mockVault)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			mockVault.AssertExpectations(t)
		})
	}
}
