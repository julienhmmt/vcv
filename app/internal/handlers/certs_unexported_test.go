package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"vcv/internal/certs"
)

func TestDecodeCertificateIDParam_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		urlParam     string
		expectedID   string
		expectedCode int
	}{
		{
			name:         "valid id",
			urlParam:     "serial123",
			expectedID:   "serial123",
			expectedCode: http.StatusOK,
		},
		{
			name:         "url encoded id",
			urlParam:     "vault1%7Cpki%3Aserial",
			expectedID:   "vault1|pki:serial",
			expectedCode: http.StatusOK,
		},
		{
			name:         "empty id",
			urlParam:     "",
			expectedID:   "",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "whitespace only",
			urlParam:     "%20%20",
			expectedID:   "",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/certs/"+tt.urlParam+"/details", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.urlParam)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			id, code, _ := decodeCertificateIDParam(req)
			assert.Equal(t, tt.expectedCode, code)
			if tt.expectedCode == http.StatusOK {
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestFilterCertificatesByMounts_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		certificates   []certs.Certificate
		selectedMounts []string
		expectedLen    int
	}{
		{
			name:           "nil mounts returns all",
			certificates:   []certs.Certificate{{ID: "pki:a"}, {ID: "pki:b"}},
			selectedMounts: nil,
			expectedLen:    2,
		},
		{
			name:           "empty mounts returns empty",
			certificates:   []certs.Certificate{{ID: "pki:a"}, {ID: "pki:b"}},
			selectedMounts: []string{},
			expectedLen:    0,
		},
		{
			name:           "no matching mounts",
			certificates:   []certs.Certificate{{ID: "pki:a"}, {ID: "pki:b"}},
			selectedMounts: []string{"other"},
			expectedLen:    0,
		},
		{
			name:           "vault prefix mount filter",
			certificates:   []certs.Certificate{{ID: "v1|pki:a"}, {ID: "v2|pki:b"}},
			selectedMounts: []string{"v1|pki"},
			expectedLen:    1,
		},
		{
			name:           "certificate with empty id is skipped",
			certificates:   []certs.Certificate{{ID: ""}, {ID: "pki:a"}},
			selectedMounts: []string{"pki"},
			expectedLen:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterCertificatesByMounts(tt.certificates, tt.selectedMounts)
			assert.Len(t, result, tt.expectedLen)
		})
	}
}

func TestExtractVaultMountFromCertificateID(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		expectedKey  string
		expectedName string
	}{
		{name: "empty", value: "", expectedKey: "", expectedName: ""},
		{name: "legacy mount only", value: "pki:serial", expectedKey: "pki", expectedName: "pki"},
		{name: "vault prefix", value: "v1|pki:serial", expectedKey: "v1|pki", expectedName: "pki"},
		{name: "no colon separator", value: "invalid", expectedKey: "", expectedName: ""},
		{name: "empty mount", value: " :serial", expectedKey: "", expectedName: ""},
		{name: "whitespace only", value: "   ", expectedKey: "", expectedName: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, name := extractVaultMountFromCertificateID(tt.value)
			assert.Equal(t, tt.expectedKey, key)
			assert.Equal(t, tt.expectedName, name)
		})
	}
}
