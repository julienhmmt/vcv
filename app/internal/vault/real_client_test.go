package vault

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"vcv/internal/cache"
	"vcv/internal/certs"
	"vcv/internal/config"
	"vcv/internal/logger"

	"github.com/hashicorp/vault/api"
)

type vaultTestServerState struct {
	certificatePEM string
}

// writeVaultTestJSON writes a JSON response with the given status.
func writeVaultTestJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// isVaultListRequest reports whether the request is a Vault LIST of path.
func isVaultListRequest(r *http.Request, path string) bool {
	if r.URL.Path != path {
		return false
	}
	return r.Method == "LIST" || (r.Method == http.MethodGet && r.URL.Query().Get("list") == "true")
}

// newVaultHTTPHandler returns a minimal Vault PKI handler serving the given
// certificate PEM for the listed serials and a cert list with the given keys.
func newVaultHTTPHandler(certificatePEM string, listKeys []string, serials []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/sys/health":
			writeVaultTestJSON(w, http.StatusOK, map[string]any{"initialized": true, "sealed": false})
		case r.URL.Path == "/v1/auth/token/lookup-self":
			writeVaultTestJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"id": "token"}})
		case isVaultListRequest(r, "/v1/pki/certs"):
			writeVaultTestJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"keys": listKeys}})
		case isVaultListRequest(r, "/v1/pki/certs/revoked"):
			writeVaultTestJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"keys": []string{"bb"}}})
		case r.Method == http.MethodGet && slices.Contains(serials, strings.TrimPrefix(r.URL.Path, "/v1/pki/cert/")):
			writeVaultTestJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"certificate": certificatePEM}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func newVaultTestCertificatePEM(t *testing.T) string {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		t.Fatalf("failed to generate serial: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "test.example.com",
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour),
		DNSNames:  []string{"test.example.com"},
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}
	block := pem.Block{Type: "CERTIFICATE", Bytes: derBytes}
	return string(pem.EncodeToMemory(&block))
}

func newVaultTestServer(state vaultTestServerState) *httptest.Server {
	return httptest.NewServer(newVaultHTTPHandler(state.certificatePEM, []string{"aa", "bb"}, []string{"aa", "bb", "ca"}))
}

func newRealClientForTest(t *testing.T, serverURL string, mounts []string) *realClient {
	clientConfig := api.DefaultConfig()
	clientConfig.Address = serverURL
	apiClient, err := api.NewClient(clientConfig)
	if err != nil {
		t.Fatalf("failed to create api client: %v", err)
	}
	apiClient.SetToken("token")
	return &realClient{client: apiClient, mounts: mounts, addr: serverURL, cache: cache.New(5 * time.Minute), stopChan: make(chan struct{})}
}

// assertClientCreation runs NewClientFromConfig and asserts the expected outcome.
func assertClientCreation(t *testing.T, cfg config.VaultConfig, expectError bool) {
	t.Helper()
	client, err := NewClientFromConfig(cfg)
	if expectError {
		if err == nil {
			t.Fatalf("expected error")
		}
		if client != nil {
			t.Fatalf("expected nil client")
		}
		return
	}
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatalf("expected client")
	}
}

func TestNewClientFromConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.VaultConfig
		expectError bool
	}{
		{name: "no vault configured", cfg: config.VaultConfig{Addr: "", ReadToken: ""}, expectError: false},
		{name: "empty address", cfg: config.VaultConfig{Addr: "", ReadToken: "token"}, expectError: true},
		{name: "empty token", cfg: config.VaultConfig{Addr: "http://localhost:8200", ReadToken: ""}, expectError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertClientCreation(t, tt.cfg, tt.expectError)
		})
	}
}

func TestNewClientFromConfig_TLSInsecure_AllowsTLSWithoutCA(t *testing.T) {
	certificatePEM := newVaultTestCertificatePEM(t)
	server := httptest.NewTLSServer(newVaultHTTPHandler(certificatePEM, []string{"aa", "bb"}, []string{"aa", "bb"}))
	defer server.Close()
	c, err := NewClientFromConfig(config.VaultConfig{Addr: server.URL, ReadToken: "token", PKIMounts: []string{"pki"}, TLSInsecure: true})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := c.CheckConnection(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestNewClientFromConfig_TLSCACert_AllowsTLSWithCA(t *testing.T) {
	certificatePEM := newVaultTestCertificatePEM(t)
	server := httptest.NewTLSServer(newVaultHTTPHandler(certificatePEM, []string{"aa"}, []string{"aa"}))
	defer server.Close()
	serverURL, parseErr := url.Parse(server.URL)
	if parseErr != nil {
		t.Fatalf("failed to parse server url: %v", parseErr)
	}
	hostname := serverURL.Hostname()
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.pem")
	if writeErr := os.WriteFile(caPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.TLS.Certificates[0].Certificate[0]}), 0o600); writeErr != nil {
		t.Fatalf("failed to write ca cert: %v", writeErr)
	}
	c, err := NewClientFromConfig(config.VaultConfig{Addr: server.URL, ReadToken: "token", PKIMounts: []string{"pki"}, TLSCACert: caPath, TLSServerName: hostname})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := c.CheckConnection(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestNewClientFromConfig_TLSCACert_BadPathReturnsError(t *testing.T) {
	_, err := NewClientFromConfig(config.VaultConfig{Addr: "https://vault.example", ReadToken: "token", PKIMounts: []string{"pki"}, TLSCACert: "/path/does/not/exist.pem"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestNewClientFromConfig_TLSCACertBase64_AllowsTLSWithCA(t *testing.T) {
	certificatePEM := newVaultTestCertificatePEM(t)
	server := httptest.NewTLSServer(newVaultHTTPHandler(certificatePEM, []string{"aa"}, []string{"aa"}))
	defer server.Close()
	serverURL, parseErr := url.Parse(server.URL)
	if parseErr != nil {
		t.Fatalf("failed to parse server url: %v", parseErr)
	}
	hostname := serverURL.Hostname()
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.TLS.Certificates[0].Certificate[0]})
	encoded := base64.RawStdEncoding.EncodeToString(pemBytes)
	c, err := NewClientFromConfig(config.VaultConfig{Addr: server.URL, ReadToken: "token", PKIMounts: []string{"pki"}, TLSCACertBase64: encoded, TLSServerName: hostname})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := c.CheckConnection(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestNewClientFromConfig_TLSCACertBase64_InvalidBase64ReturnsError(t *testing.T) {
	_, err := NewClientFromConfig(config.VaultConfig{Addr: "https://vault.example", ReadToken: "token", PKIMounts: []string{"pki"}, TLSCACertBase64: "not base64"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

// assertParseMountAndSerial asserts the parse outcome for one test case.
func assertParseMountAndSerial(t *testing.T, client *realClient, value string, expectedMnt string, expectedSer string, expectErr bool) {
	t.Helper()
	mount, serial, err := client.parseMountAndSerial(value)
	if expectErr {
		if err == nil {
			t.Fatalf("expected error")
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mount != expectedMnt {
		t.Fatalf("expected mount %q, got %q", expectedMnt, mount)
	}
	if serial != expectedSer {
		t.Fatalf("expected serial %q, got %q", expectedSer, serial)
	}
}

func TestParseMountAndSerial(t *testing.T) {
	client := &realClient{mounts: []string{"pki", "pki_dev"}}
	tests := []struct {
		name        string
		value       string
		expectedMnt string
		expectedSer string
		expectErr   bool
	}{
		{name: "prefixed configured mount", value: "pki:aa", expectedMnt: "pki", expectedSer: "aa", expectErr: false},
		{name: "prefixed unconfigured mount", value: "unknown:aa", expectErr: true},
		{name: "legacy no prefix", value: "aa", expectedMnt: "pki", expectedSer: "aa", expectErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertParseMountAndSerial(t, client, tt.value, tt.expectedMnt, tt.expectedSer, tt.expectErr)
		})
	}
	clientNoMounts := &realClient{mounts: []string{}}
	_, _, err := clientNoMounts.parseMountAndSerial("aa")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckConnection(t *testing.T) {
	certificatePEM := newVaultTestCertificatePEM(t)
	server := newVaultTestServer(vaultTestServerState{certificatePEM: certificatePEM})
	defer server.Close()
	client := newRealClientForTest(t, server.URL, []string{"pki"})
	err := client.CheckConnection(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRealClient_ListCertificates_And_Details(t *testing.T) {
	certificatePEM := newVaultTestCertificatePEM(t)
	server := newVaultTestServer(vaultTestServerState{certificatePEM: certificatePEM})
	defer server.Close()
	client := newRealClientForTest(t, server.URL, []string{"pki"})
	ctx := context.Background()
	certificates, err := client.ListCertificates(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(certificates) != 2 {
		t.Fatalf("expected 2 certificates, got %d", len(certificates))
	}
	if certificates[0].CertType != "machine" {
		t.Fatalf("expected machine certificate type, got %q", certificates[0].CertType)
	}
	details, err := client.GetCertificateDetails(ctx, "pki:aa")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if details.SerialNumber != "aa" {
		t.Fatalf("expected serial %q, got %q", "aa", details.SerialNumber)
	}
	if details.CertType != "machine" {
		t.Fatalf("expected machine certificate type, got %q", details.CertType)
	}
	if details.PEM == "" {
		t.Fatalf("expected pem")
	}
	pemResponse, err := client.GetCertificatePEM(ctx, "pki:bb")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pemResponse.SerialNumber != "bb" {
		t.Fatalf("expected serial %q, got %q", "bb", pemResponse.SerialNumber)
	}
	if pemResponse.PEM == "" {
		t.Fatalf("expected pem")
	}
	client.InvalidateCache()
	client.Shutdown()
}

func TestRealClient_ListCertificates_CacheHit(t *testing.T) {
	certificatePEM := newVaultTestCertificatePEM(t)
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		newVaultHTTPHandler(certificatePEM, []string{"aa", "bb"}, []string{"aa", "bb"}).ServeHTTP(w, r)
	}))
	defer server.Close()

	client := newRealClientForTest(t, server.URL, []string{"pki"})
	ctx := context.Background()

	first, err := client.ListCertificates(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(first) != 2 {
		t.Fatalf("expected 2 certificates, got %d", len(first))
	}
	countAfterFirst := requestCount.Load()
	if countAfterFirst == 0 {
		t.Fatalf("expected the first call to reach the vault server")
	}

	second, err := client.ListCertificates(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(second) != len(first) {
		t.Fatalf("expected cached result length %d, got %d", len(first), len(second))
	}
	if got := requestCount.Load(); got != countAfterFirst {
		t.Fatalf("expected second ListCertificates call to be served from cache with no new vault requests, got %d additional requests", got-countAfterFirst)
	}
}

func TestRealClient_GetIntermediateCA(t *testing.T) {
	certificatePEM := newVaultTestCertificatePEM(t)
	server := newVaultTestServer(vaultTestServerState{certificatePEM: certificatePEM})
	defer server.Close()
	client := newRealClientForTest(t, server.URL, []string{"pki"})
	ctx := context.Background()
	caDetails, err := client.GetIntermediateCA(ctx, "pki")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if caDetails.CommonName != "test.example.com" {
		t.Fatalf("expected common name %q, got %q", "test.example.com", caDetails.CommonName)
	}
	if caDetails.PEM == "" {
		t.Fatalf("expected pem")
	}
	// Cache hit
	caDetails2, err := client.GetIntermediateCA(ctx, "pki")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if caDetails2.CommonName != "test.example.com" {
		t.Fatalf("expected cached common name %q, got %q", "test.example.com", caDetails2.CommonName)
	}
}

func TestRealClient_GetIntermediateCA_UnconfiguredMount(t *testing.T) {
	certificatePEM := newVaultTestCertificatePEM(t)
	server := newVaultTestServer(vaultTestServerState{certificatePEM: certificatePEM})
	defer server.Close()
	client := newRealClientForTest(t, server.URL, []string{"pki"})
	_, err := client.GetIntermediateCA(context.Background(), "other")
	if err == nil {
		t.Fatalf("expected error for unconfigured mount")
	}
}

func TestRealClient_GetIntermediateCA_BadResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/v1/pki/ca/pem" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"ca_chain": []string{"not a pem"}}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	client := newRealClientForTest(t, server.URL, []string{"pki"})
	_, err := client.GetIntermediateCA(context.Background(), "pki")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckConnection_NotInitialized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/sys/health" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"initialized": false, "sealed": false})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	client := newRealClientForTest(t, server.URL, []string{"pki"})
	err := client.CheckConnection(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestCheckConnection_Sealed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/sys/health" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"initialized": true, "sealed": true})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	client := newRealClientForTest(t, server.URL, []string{"pki"})
	err := client.CheckConnection(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestListCertificatesFromMount_KeysWrongType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if (r.Method == "LIST" || (r.Method == http.MethodGet && r.URL.Query().Get("list") == "true")) && r.URL.Path == "/v1/pki/certs" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"keys": "nope"}})
			return
		}
		if r.URL.Path == "/v1/sys/health" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"initialized": true, "sealed": false})
			return
		}
		if r.URL.Path == "/v1/auth/token/lookup-self" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"id": "token"}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	client := newRealClientForTest(t, server.URL, []string{"pki"})
	_, _, err := client.listCertificatesFromMount(context.Background(), "pki")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestReadCertificateFromMount_MissingCertificateField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/v1/pki/cert/aa" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	client := newRealClientForTest(t, server.URL, []string{"pki"})
	_, err := client.readCertificateFromMount(context.Background(), "pki", "aa")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestReadCertificateFromMount_InvalidPEM(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/v1/pki/cert/aa" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"certificate": "not a pem"}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	client := newRealClientForTest(t, server.URL, []string{"pki"})
	_, err := client.readCertificateFromMount(context.Background(), "pki", "aa")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestGetCertificateDetails_CacheHit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	client := newRealClientForTest(t, server.URL, []string{"pki"})
	cacheKey := "v2:details_pki:aa"
	client.cache.Set(cacheKey, certs.DetailedCertificate{Certificate: certs.Certificate{SerialNumber: "aa"}})
	result, err := client.GetCertificateDetails(context.Background(), "pki:aa")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.SerialNumber != "aa" {
		t.Fatalf("expected cached details")
	}
}

func TestRealClient_CacheSize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	client := newRealClientForTest(t, server.URL, []string{"pki"})
	if client.CacheSize() != 0 {
		t.Fatalf("expected cache size 0")
	}
	client.cache.Set("k1", "v1")
	client.cache.Set("k2", "v2")
	if client.CacheSize() != 2 {
		t.Fatalf("expected cache size 2")
	}
}

// assertLogContains asserts that the captured log output contains every
// expected substring.
func assertLogContains(t *testing.T, buf *bytes.Buffer, expected []string) {
	t.Helper()
	output := buf.String()
	for _, want := range expected {
		if !strings.Contains(output, want) {
			t.Errorf("Expected %q in log, got: %s", want, output)
		}
	}
}

// TestRealClient_Logging tests that logging works correctly for vault operations
func TestRealClient_Logging(t *testing.T) {
	// Setup logger to capture output
	logger.Init("debug")
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/sys/health":
			writeVaultTestJSON(w, http.StatusOK, map[string]any{"initialized": true, "sealed": false, "version": "1.12.0"})
		case "/v1/auth/token/lookup-self":
			writeVaultTestJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"id": "token"}})
		case "/v1/pki/certs":
			writeVaultTestJSON(w, http.StatusOK, map[string]any{"data": map[string][]string{"keys": {"01"}}})
		case "/v1/pki/cert/01":
			writeVaultTestJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"certificate": newVaultTestCertificatePEM(t)}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := newRealClientForTest(t, server.URL, []string{"pki"})

	// Test CheckConnection logging
	buf.Reset()
	err := client.CheckConnection(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertLogContains(t, &buf, []string{"checking vault connection", "vault connection successful"})

	// Test ListCertificates logging
	buf.Reset()
	_, err = client.ListCertificates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertLogContains(t, &buf, []string{
		"listing certificates from vault mounts",
		"listing certificates from mount",
		"successfully listed certificates from mount",
		"completed certificate listing and cached result",
	})

	// Test GetCertificateDetails logging
	buf.Reset()
	_, err = client.GetCertificateDetails(context.Background(), "01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertLogContains(t, &buf, []string{"getting certificate details", "successfully retrieved and cached certificate details"})

	// Test InvalidateCache logging
	buf.Reset()
	client.InvalidateCache()
	assertLogContains(t, &buf, []string{"invalidating vault client cache", "cache invalidated successfully"})
}

// TestRealClient_LoggingErrors tests that error logging works correctly
func TestRealClient_LoggingErrors(t *testing.T) {
	// Setup logger to capture output
	logger.Init("debug")
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	// Create a server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/sys/health":
			w.WriteHeader(http.StatusInternalServerError)
		case "/v1/pki/certs":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := newRealClientForTest(t, server.URL, []string{"pki"})

	// Test CheckConnection error logging
	buf.Reset()
	err := client.CheckConnection(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	output := buf.String()
	if !strings.Contains(output, "vault health check failed") {
		t.Errorf("Expected health check error log, got: %s", output)
	}

	// Test ListCertificates error logging
	buf.Reset()
	_, err = client.ListCertificates(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	output = buf.String()
	if !strings.Contains(output, "failed to list certificates from mount") {
		t.Errorf("Expected mount listing error log, got: %s", output)
	}
}
