package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"vcv/internal/certs"
	"vcv/internal/handlers"
	"vcv/internal/middleware"
	"vcv/internal/vault"
)

func setupRouter(mockVault *vault.MockClient) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	handlers.RegisterCertRoutes(r, mockVault)
	return r
}

type certsEnvelopeResponse struct {
	Certificates []certs.Certificate `json:"certificates"`
	Errors       []vault.VaultError  `json:"errors"`
}

func TestListCertificates_Success(t *testing.T) {
	mockVault := new(vault.MockClient)
	certsList := []certs.Certificate{
		{ID: "1", SerialNumber: "1", CommonName: "a", ExpiresAt: time.Now()},
	}
	mockVault.On("ListCertificates", mock.Anything).Return(certsList, nil)
	router := setupRouter(mockVault)

	req := httptest.NewRequest(http.MethodGet, "/api/certs", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var got certsEnvelopeResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Len(t, got.Certificates, 1)
	assert.Empty(t, got.Errors)
	mockVault.AssertExpectations(t)
}

type envelopeMockClient struct {
	*vault.MockClient
	certs  []certs.Certificate
	errors []vault.VaultError
}

func (e *envelopeMockClient) ListCertificatesEnvelope(_ context.Context) ([]certs.Certificate, []vault.VaultError) {
	return e.certs, e.errors
}

func TestListCertificates_Envelope_PartialSuccess(t *testing.T) {
	envClient := &envelopeMockClient{
		MockClient: new(vault.MockClient),
		certs:      []certs.Certificate{{ID: "vault-a|pki:1", SerialNumber: "1", CommonName: "a", ExpiresAt: time.Now()}},
		errors:     []vault.VaultError{{VaultID: "vault-b", Message: "dial tcp: lookup vault-b: no such host"}},
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	handlers.RegisterCertRoutes(r, envClient)

	req := httptest.NewRequest(http.MethodGet, "/api/certs", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var got certsEnvelopeResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Len(t, got.Certificates, 1)
	assert.Len(t, got.Errors, 1)
	assert.Equal(t, "vault-b", got.Errors[0].VaultID)
}

func TestListCertificates_Error(t *testing.T) {
	mockVault := new(vault.MockClient)
	mockVault.On("ListCertificates", mock.Anything).Return([]certs.Certificate{}, errors.New("boom"))
	router := setupRouter(mockVault)

	req := httptest.NewRequest(http.MethodGet, "/api/certs", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockVault.AssertExpectations(t)
}

func TestGetCertificateDetails_Success(t *testing.T) {
	mockVault := new(vault.MockClient)
	expected := certs.DetailedCertificate{
		Certificate: certs.Certificate{
			ID:           "serial",
			SerialNumber: "serial",
			CommonName:   "cn",
			ExpiresAt:    time.Now(),
		},
	}
	mockVault.On("GetCertificateDetails", mock.Anything, "serial").Return(expected, nil)
	router := setupRouter(mockVault)

	req := httptest.NewRequest(http.MethodGet, "/api/certs/serial/details", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var got certs.DetailedCertificate
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "serial", got.SerialNumber)
	mockVault.AssertExpectations(t)
}

func TestGetCertificateDetails_BadRequest(t *testing.T) {
	mockVault := new(vault.MockClient)
	router := setupRouter(mockVault)

	req := httptest.NewRequest(http.MethodGet, "/api/certs//details", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockVault.AssertExpectations(t)
}

func TestGetCertificateDetails_Error(t *testing.T) {
	mockVault := new(vault.MockClient)
	mockVault.On("GetCertificateDetails", mock.Anything, "serial").Return(certs.DetailedCertificate{}, errors.New("fail"))
	router := setupRouter(mockVault)

	req := httptest.NewRequest(http.MethodGet, "/api/certs/serial/details", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockVault.AssertExpectations(t)
}

func TestGetCertificatePEM_Success(t *testing.T) {
	mockVault := new(vault.MockClient)
	pemResp := certs.PEMResponse{SerialNumber: "serial", PEM: "pem-data"}
	mockVault.On("GetCertificatePEM", mock.Anything, "serial").Return(pemResp, nil)
	router := setupRouter(mockVault)

	req := httptest.NewRequest(http.MethodGet, "/api/certs/serial/pem", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var got certs.PEMResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "pem-data", got.PEM)
	mockVault.AssertExpectations(t)
}

func TestGetCertificatePEM_Error(t *testing.T) {
	mockVault := new(vault.MockClient)
	mockVault.On("GetCertificatePEM", mock.Anything, "serial").Return(certs.PEMResponse{}, errors.New("fail"))
	router := setupRouter(mockVault)

	req := httptest.NewRequest(http.MethodGet, "/api/certs/serial/pem", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockVault.AssertExpectations(t)
}

func TestInvalidateCache_NotRegisteredOnCertRoutes(t *testing.T) {
	mockVault := new(vault.MockClient)
	router := setupRouter(mockVault)
	req := httptest.NewRequest(http.MethodPost, "/api/cache/invalidate", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	mockVault.AssertNotCalled(t, "InvalidateCache")
}

type failingResponseWriter struct {
	header http.Header
}

func (w *failingResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *failingResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func (w *failingResponseWriter) WriteHeader(statusCode int) {
}

func TestListCertificates_MountsQuery_AllSentinelReturnsAll(t *testing.T) {
	mockVault := new(vault.MockClient)
	certsList := []certs.Certificate{{ID: "pki:a", SerialNumber: "a", CommonName: "a", ExpiresAt: time.Now()}, {ID: "pki:b", SerialNumber: "b", CommonName: "b", ExpiresAt: time.Now()}}
	mockVault.On("ListCertificates", mock.Anything).Return(certsList, nil)
	router := setupRouter(mockVault)
	req := httptest.NewRequest(http.MethodGet, "/api/certs?mounts=__all__", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var got certsEnvelopeResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Len(t, got.Certificates, 2)
	mockVault.AssertExpectations(t)
}

func TestListCertificates_MountsQuery_EmptyReturnsEmpty(t *testing.T) {
	mockVault := new(vault.MockClient)
	certsList := []certs.Certificate{{ID: "pki:a", SerialNumber: "a", CommonName: "a", ExpiresAt: time.Now()}}
	mockVault.On("ListCertificates", mock.Anything).Return(certsList, nil)
	router := setupRouter(mockVault)
	req := httptest.NewRequest(http.MethodGet, "/api/certs?mounts=", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var got certsEnvelopeResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Len(t, got.Certificates, 0)
	mockVault.AssertExpectations(t)
}

func TestListCertificates_EncodingError_DoesNotPanic(t *testing.T) {
	mockVault := new(vault.MockClient)
	mockVault.On("ListCertificates", mock.Anything).Return([]certs.Certificate{}, nil)
	router := setupRouter(mockVault)
	req := httptest.NewRequest(http.MethodGet, "/api/certs", nil)
	w := &failingResponseWriter{}
	router.ServeHTTP(w, req)
	mockVault.AssertExpectations(t)
}

func TestGetCertificateDetails_EncodingError_DoesNotPanic(t *testing.T) {
	mockVault := new(vault.MockClient)
	expected := certs.DetailedCertificate{Certificate: certs.Certificate{ID: "serial", SerialNumber: "serial", CommonName: "cn", ExpiresAt: time.Now()}}
	mockVault.On("GetCertificateDetails", mock.Anything, "serial").Return(expected, nil)
	router := setupRouter(mockVault)
	req := httptest.NewRequest(http.MethodGet, "/api/certs/serial/details", nil)
	w := &failingResponseWriter{}
	router.ServeHTTP(w, req)
	mockVault.AssertExpectations(t)
}

func TestGetCertificatePEM_EncodingError_DoesNotPanic(t *testing.T) {
	mockVault := new(vault.MockClient)
	pemResp := certs.PEMResponse{SerialNumber: "serial", PEM: "pem-data"}
	mockVault.On("GetCertificatePEM", mock.Anything, "serial").Return(pemResp, nil)
	router := setupRouter(mockVault)
	req := httptest.NewRequest(http.MethodGet, "/api/certs/serial/pem", nil)
	w := &failingResponseWriter{}
	router.ServeHTTP(w, req)
	mockVault.AssertExpectations(t)
}


func TestGetIntermediateCA_Success(t *testing.T) {
	mockVault := new(vault.MockClient)
	expected := certs.DetailedCertificate{
		Certificate: certs.Certificate{
			ID:           "pki:ca-id",
			SerialNumber: "ca-serial",
			CommonName:   "ca-cn",
			ExpiresAt:    time.Now(),
		},
	}
	mockVault.On("GetIntermediateCA", mock.Anything, "vault1|pki").Return(expected, nil)
	router := setupRouter(mockVault)

	req := httptest.NewRequest(http.MethodGet, "/api/certs/vault1%7Cpki%3Aserial/ca", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var got certs.DetailedCertificate
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	mockVault.AssertExpectations(t)
}

func TestGetIntermediateCA_BadRequest(t *testing.T) {
	mockVault := new(vault.MockClient)
	router := setupRouter(mockVault)

	req := httptest.NewRequest(http.MethodGet, "/api/certs//ca", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockVault.AssertExpectations(t)
}

func TestGetIntermediateCA_Error(t *testing.T) {
	mockVault := new(vault.MockClient)
	mockVault.On("GetIntermediateCA", mock.Anything, "vault1|pki").Return(certs.DetailedCertificate{}, errors.New("fail"))
	router := setupRouter(mockVault)

	req := httptest.NewRequest(http.MethodGet, "/api/certs/vault1%7Cpki%3Aserial/ca", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockVault.AssertExpectations(t)
}

func TestGetCertificateCA_EncodingError_DoesNotPanic(t *testing.T) {
	mockVault := new(vault.MockClient)
	expected := certs.DetailedCertificate{Certificate: certs.Certificate{ID: "ca-id", SerialNumber: "ca-serial", CommonName: "ca-cn"}}
	mockVault.On("GetIntermediateCA", mock.Anything, "vault1|pki").Return(expected, nil)
	router := setupRouter(mockVault)
	req := httptest.NewRequest(http.MethodGet, "/api/certs/vault1%7Cpki%3Aserial/ca", nil)
	w := &failingResponseWriter{}
	router.ServeHTTP(w, req)
	mockVault.AssertExpectations(t)
}

// Unexported function tests are in certs_unexported_test.go
