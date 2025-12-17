package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"vcv/internal/certs"
	"vcv/internal/handlers"
	"vcv/internal/vault"
	"vcv/middleware"
)

func setupRouter(mockVault *vault.MockClient) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	handlers.RegisterCertRoutes(r, mockVault)
	return r
}

func TestListCertificates_Success(t *testing.T) {
	mockVault := new(vault.MockClient)
	certsList := []certs.Certificate{
		{ID: "1", CommonName: "a", ExpiresAt: time.Now()},
	}
	mockVault.On("ListCertificates", mock.Anything).Return(certsList, nil)
	router := setupRouter(mockVault)

	req := httptest.NewRequest(http.MethodGet, "/api/certs", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var got []certs.Certificate
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Len(t, got, 1)
	mockVault.AssertExpectations(t)
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
			ID:         "serial",
			CommonName: "cn",
			ExpiresAt:  time.Now(),
		},
		SerialNumber: "serial",
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

func TestInvalidateCache(t *testing.T) {
	mockVault := new(vault.MockClient)
	mockVault.On("InvalidateCache").Return()
	router := setupRouter(mockVault)

	req := httptest.NewRequest(http.MethodPost, "/api/cache/invalidate", bytes.NewBuffer(nil))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	mockVault.AssertExpectations(t)
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
	certsList := []certs.Certificate{{ID: "pki:a", CommonName: "a", ExpiresAt: time.Now()}, {ID: "pki:b", CommonName: "b", ExpiresAt: time.Now()}}
	mockVault.On("ListCertificates", mock.Anything).Return(certsList, nil)
	router := setupRouter(mockVault)
	req := httptest.NewRequest(http.MethodGet, "/api/certs?mounts=__all__", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var got []certs.Certificate
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Len(t, got, 2)
	mockVault.AssertExpectations(t)
}

func TestListCertificates_MountsQuery_EmptyReturnsEmpty(t *testing.T) {
	mockVault := new(vault.MockClient)
	certsList := []certs.Certificate{{ID: "pki:a", CommonName: "a", ExpiresAt: time.Now()}}
	mockVault.On("ListCertificates", mock.Anything).Return(certsList, nil)
	router := setupRouter(mockVault)
	req := httptest.NewRequest(http.MethodGet, "/api/certs?mounts=", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var got []certs.Certificate
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Len(t, got, 0)
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
	expected := certs.DetailedCertificate{Certificate: certs.Certificate{ID: "serial", CommonName: "cn", ExpiresAt: time.Now()}, SerialNumber: "serial"}
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

func TestDownloadCertificatePEM_Success(t *testing.T) {
	mockVault := new(vault.MockClient)
	pemResp := certs.PEMResponse{SerialNumber: "serial", PEM: "pem-data"}
	mockVault.On("GetCertificatePEM", mock.Anything, "serial").Return(pemResp, nil)
	router := setupRouter(mockVault)
	req := httptest.NewRequest(http.MethodGet, "/api/certs/serial/pem/download", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/x-pem-file", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Header().Get("Content-Disposition"), "attachment;")
	assert.Equal(t, "pem-data", strings.TrimSpace(rec.Body.String()))
	mockVault.AssertExpectations(t)
}

func TestDownloadCertificatePEM_WriteError_DoesNotPanic(t *testing.T) {
	mockVault := new(vault.MockClient)
	pemResp := certs.PEMResponse{SerialNumber: "serial", PEM: "pem-data"}
	mockVault.On("GetCertificatePEM", mock.Anything, "serial").Return(pemResp, nil)
	router := setupRouter(mockVault)
	req := httptest.NewRequest(http.MethodGet, "/api/certs/serial/pem/download", nil)
	w := &failingResponseWriter{}
	router.ServeHTTP(w, req)
	mockVault.AssertExpectations(t)
}
