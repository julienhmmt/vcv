package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"vcv/internal/certs"
	"vcv/internal/vault"
)

type certsQueryEnvelope struct {
	Certificates []certs.Certificate `json:"certificates"`
	Errors       []vault.VaultError  `json:"errors"`
	Total        int                 `json:"total"`
	Page         int                 `json:"page"`
	PageSize     int                 `json:"page_size"`
	TotalPages   int                 `json:"total_pages"`
	Counts       map[string]int      `json:"counts"`
}

func queryFixtureCerts(now time.Time) []certs.Certificate {
	return []certs.Certificate{
		{ID: "vault-a|pki:1", SerialNumber: "1", CommonName: "alpha.example.com", Sans: []string{"www.alpha.example.com"}, CertType: "machine", ExpiresAt: now.Add(60 * 24 * time.Hour)},
		{ID: "vault-a|pki:2", SerialNumber: "2", CommonName: "beta.example.com", Sans: nil, CertType: "user", ExpiresAt: now.Add(3 * 24 * time.Hour)},
		{ID: "vault-b|pki2:3", SerialNumber: "3", CommonName: "gamma.example.com", Sans: []string{"Gamma-Alias.EXAMPLE.com"}, CertType: "machine", ExpiresAt: now.Add(-24 * time.Hour)},
		{ID: "vault-b|pki2:4", SerialNumber: "4", CommonName: "delta.example.com", CertType: "both", ExpiresAt: now.Add(10 * 24 * time.Hour)},
		{ID: "vault-a|pki:5", SerialNumber: "5", CommonName: "epsilon.example.com", CertType: "machine", ExpiresAt: now.Add(90 * 24 * time.Hour), Revoked: true},
	}
}

func serveCertsQuery(t *testing.T, target string) (int, certsQueryEnvelope, []byte) {
	t.Helper()
	mockVault := new(vault.MockClient)
	mockVault.On("ListCertificates", mock.Anything).Return(queryFixtureCerts(time.Now()), nil)
	router := setupRouter(mockVault)

	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var got certsQueryEnvelope
	if rec.Code == http.StatusOK {
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	}
	mockVault.AssertExpectations(t)
	return rec.Code, got, rec.Body.Bytes()
}

func certIDs(list []certs.Certificate) []string {
	ids := make([]string, 0, len(list))
	for _, certificate := range list {
		ids = append(ids, certificate.ID)
	}
	return ids
}

func TestListCertsQuery_Defaults(t *testing.T) {
	code, got, _ := serveCertsQuery(t, "/api/certs")

	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, got.Certificates, 5)
	assert.Equal(t, 5, got.Total)
	assert.Equal(t, 1, got.Page)
	assert.Equal(t, 5, got.PageSize)
	assert.Equal(t, 1, got.TotalPages)
	assert.Equal(t, map[string]int{
		"valid": 1, "warning": 1, "critical": 1, "expired": 1, "revoked": 1, "total": 5,
	}, got.Counts)
}

func TestListCertsQuery_Search(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected []string
	}{
		{name: "common name", query: "alpha", expected: []string{"vault-a|pki:1"}},
		{name: "case insensitive", query: "ALPHA", expected: []string{"vault-a|pki:1"}},
		{name: "serial", query: "serial", expected: []string{}},
		{name: "serial number", query: "2", expected: []string{"vault-a|pki:2"}},
		{name: "san", query: "www.alpha", expected: []string{"vault-a|pki:1"}},
		{name: "san case insensitive", query: "gamma-alias", expected: []string{"vault-b|pki2:3"}},
		{name: "no match returns empty array", query: "does-not-exist", expected: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, got, _ := serveCertsQuery(t, "/api/certs?search="+tt.query)
			assert.Equal(t, http.StatusOK, code)
			assert.Equal(t, tt.expected, certIDs(got.Certificates))
			assert.Equal(t, len(tt.expected), got.Total)
			assert.NotNil(t, got.Certificates)
		})
	}
}

func TestListCertsQuery_StatusFilter(t *testing.T) {
	code, got, _ := serveCertsQuery(t, "/api/certs?status=expired,revoked")

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, []string{"vault-b|pki2:3", "vault-a|pki:5"}, certIDs(got.Certificates))
	assert.Equal(t, 2, got.Total)
	// Counts ignore the status filter itself (facet semantics).
	assert.Equal(t, 5, got.Counts["total"])
	assert.Equal(t, 1, got.Counts["expired"])
	assert.Equal(t, 1, got.Counts["revoked"])
}

func TestListCertsQuery_CertType(t *testing.T) {
	code, got, _ := serveCertsQuery(t, "/api/certs?cert_type=machine")

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, []string{"vault-a|pki:1", "vault-b|pki2:3", "vault-a|pki:5"}, certIDs(got.Certificates))
	assert.Equal(t, 3, got.Counts["total"])
}

func TestListCertsQuery_Sort(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			name:     "expires asc",
			query:    "sort=expiresAt&order=asc",
			expected: []string{"vault-b|pki2:3", "vault-a|pki:2", "vault-b|pki2:4", "vault-a|pki:1", "vault-a|pki:5"},
		},
		{
			name:     "expires desc",
			query:    "sort=expiresAt&order=desc",
			expected: []string{"vault-a|pki:5", "vault-a|pki:1", "vault-b|pki2:4", "vault-a|pki:2", "vault-b|pki2:3"},
		},
		{
			name:     "common name desc",
			query:    "sort=commonName&order=desc",
			expected: []string{"vault-b|pki2:3", "vault-a|pki:5", "vault-b|pki2:4", "vault-a|pki:2", "vault-a|pki:1"},
		},
		{
			name:     "vault asc",
			query:    "sort=vault",
			expected: []string{"vault-a|pki:1", "vault-a|pki:2", "vault-a|pki:5", "vault-b|pki2:3", "vault-b|pki2:4"},
		},
		{
			name:     "pki asc",
			query:    "sort=pki",
			expected: []string{"vault-a|pki:1", "vault-a|pki:2", "vault-a|pki:5", "vault-b|pki2:3", "vault-b|pki2:4"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, got, _ := serveCertsQuery(t, "/api/certs?"+tt.query)
			assert.Equal(t, http.StatusOK, code)
			assert.Equal(t, tt.expected, certIDs(got.Certificates))
		})
	}
}

func TestListCertsQuery_Pagination(t *testing.T) {
	code, first, _ := serveCertsQuery(t, "/api/certs?page=1&page_size=2&sort=commonName")
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, 5, first.Total)
	assert.Equal(t, 1, first.Page)
	assert.Equal(t, 2, first.PageSize)
	assert.Equal(t, 3, first.TotalPages)
	assert.Len(t, first.Certificates, 2)

	code, second, _ := serveCertsQuery(t, "/api/certs?page=2&page_size=2&sort=commonName")
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, second.Certificates, 2)
	assert.Empty(t, intersectIDs(first.Certificates, second.Certificates))

	code, third, _ := serveCertsQuery(t, "/api/certs?page=3&page_size=2&sort=commonName")
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, third.Certificates, 1)

	code, beyond, raw := serveCertsQuery(t, "/api/certs?page=9&page_size=2")
	assert.Equal(t, http.StatusOK, code)
	assert.Empty(t, beyond.Certificates)
	assert.NotNil(t, beyond.Certificates)
	assert.Equal(t, 3, beyond.TotalPages)
	assert.Contains(t, string(raw), `"certificates":[]`)
}

func TestListCertsQuery_PageSizeAll(t *testing.T) {
	code, got, _ := serveCertsQuery(t, "/api/certs?page=1&page_size=all")

	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, got.Certificates, 5)
	assert.Equal(t, 5, got.PageSize)
	assert.Equal(t, 1, got.TotalPages)
}

func TestListCertsQuery_Invalid(t *testing.T) {
	for _, target := range []string{
		"/api/certs?status=bogus",
		"/api/certs?cert_type=bogus",
		"/api/certs?sort=bogus",
		"/api/certs?order=sideways",
		"/api/certs?page=0",
		"/api/certs?page=abc",
		"/api/certs?page_size=0",
		"/api/certs?page_size=abc",
	} {
		t.Run(target, func(t *testing.T) {
			mockVault := new(vault.MockClient)
			mockVault.On("ListCertificates", mock.Anything).Return(queryFixtureCerts(time.Now()), nil)
			router := setupRouter(mockVault)

			req := httptest.NewRequest(http.MethodGet, target, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			var body map[string]string
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.NotEmpty(t, body["error"])
			mockVault.AssertExpectations(t)
		})
	}
}

func intersectIDs(a, b []certs.Certificate) []string {
	inB := make(map[string]struct{}, len(b))
	for _, certificate := range b {
		inB[certificate.ID] = struct{}{}
	}
	var overlap []string
	for _, certificate := range a {
		if _, ok := inB[certificate.ID]; ok {
			overlap = append(overlap, certificate.ID)
		}
	}
	return overlap
}
