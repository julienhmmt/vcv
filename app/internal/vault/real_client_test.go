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
	"strings"
	"testing"
	"time"

	"vcv/config"
	"vcv/internal/cache"
	"vcv/internal/certs"
	"vcv/internal/logger"

	"github.com/hashicorp/vault/api"
)

type vaultTestServerState struct {
	certificatePEM string
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
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}
	block := pem.Block{Type: "CERTIFICATE", Bytes: derBytes}
	return string(pem.EncodeToMemory(&block))
}

func newVaultTestServer(state vaultTestServerState) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/sys/health" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"initialized": true, "sealed": false})
			return
		}
		if (r.Method == "LIST" || (r.Method == http.MethodGet && r.URL.Query().Get("list") == "true")) && r.URL.Path == "/v1/pki/certs" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"keys": []string{"aa", "bb"}}})
			return
		}
		if (r.Method == "LIST" || (r.Method == http.MethodGet && r.URL.Query().Get("list") == "true")) && r.URL.Path == "/v1/pki/certs/revoked" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"keys": []string{"bb"}}})
			return
		}
		if r.Method == http.MethodGet && (r.URL.Path == "/v1/pki/cert/aa" || r.URL.Path == "/v1/pki/cert/bb") {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"certificate": state.certificatePEM}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	return httptest.NewServer(handler)
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
			client, err := NewClientFromConfig(tt.cfg)
			if tt.expectError {
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
		})
	}
}

func TestNewClientFromConfig_TLSInsecure_AllowsTLSWithoutCA(t *testing.T) {
	certificatePEM := newVaultTestCertificatePEM(t)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/sys/health" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"initialized": true, "sealed": false})
			return
		}
		if (r.Method == "LIST" || (r.Method == http.MethodGet && r.URL.Query().Get("list") == "true")) && r.URL.Path == "/v1/pki/certs" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"keys": []string{"aa", "bb"}}})
			return
		}
		if r.Method == http.MethodGet && (r.URL.Path == "/v1/pki/cert/aa" || r.URL.Path == "/v1/pki/cert/bb") {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"certificate": certificatePEM}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
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
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/sys/health" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"initialized": true, "sealed": false})
			return
		}
		if (r.Method == "LIST" || (r.Method == http.MethodGet && r.URL.Query().Get("list") == "true")) && r.URL.Path == "/v1/pki/certs" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"keys": []string{"aa"}}})
			return
		}
		if r.Method == http.MethodGet && r.URL.Path == "/v1/pki/cert/aa" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"certificate": certificatePEM}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
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
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/sys/health" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"initialized": true, "sealed": false})
			return
		}
		if (r.Method == "LIST" || (r.Method == http.MethodGet && r.URL.Query().Get("list") == "true")) && r.URL.Path == "/v1/pki/certs" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"keys": []string{"aa"}}})
			return
		}
		if r.Method == http.MethodGet && r.URL.Path == "/v1/pki/cert/aa" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"certificate": certificatePEM}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
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
			mount, serial, err := client.parseMountAndSerial(tt.value)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if mount != tt.expectedMnt {
				t.Fatalf("expected mount %q, got %q", tt.expectedMnt, mount)
			}
			if serial != tt.expectedSer {
				t.Fatalf("expected serial %q, got %q", tt.expectedSer, serial)
			}
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
	details, err := client.GetCertificateDetails(ctx, "pki:aa")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if details.SerialNumber != "aa" {
		t.Fatalf("expected serial %q, got %q", "aa", details.SerialNumber)
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

func TestCheckConnection_NotInitialized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/sys/health" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"initialized": false, "sealed": false})
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
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"initialized": true, "sealed": true})
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
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"keys": "nope"}})
			return
		}
		if r.URL.Path == "/v1/sys/health" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"initialized": true, "sealed": false})
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
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	client := newRealClientForTest(t, server.URL, []string{"pki"})
	_, err := client.readCertificateFromMount("pki", "aa")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestReadCertificateFromMount_InvalidPEM(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet && r.URL.Path == "/v1/pki/cert/aa" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"certificate": "not a pem"}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	client := newRealClientForTest(t, server.URL, []string{"pki"})
	_, err := client.readCertificateFromMount("pki", "aa")
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
	cacheKey := "details_pki:aa"
	client.cache.Set(cacheKey, certs.DetailedCertificate{SerialNumber: "aa"})
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

// TestRealClient_Logging tests that logging works correctly for vault operations
func TestRealClient_Logging(t *testing.T) {
	// Setup logger to capture output
	logger.Init("debug")
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/sys/health":
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"initialized": true,
				"sealed":      false,
				"version":     "1.12.0",
			}); err != nil {
				t.Fatalf("failed to encode health response: %v", err)
			}
		case "/v1/pki/certs":
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string][]string{"keys": {"01"}},
			}); err != nil {
				t.Fatalf("failed to encode certs response: %v", err)
			}
		case "/v1/pki/cert/01":
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]string{
					"certificate": newVaultTestCertificatePEM(t),
				},
			}); err != nil {
				t.Fatalf("failed to encode cert response: %v", err)
			}
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

	output := buf.String()
	if !strings.Contains(output, "checking vault connection") {
		t.Errorf("Expected connection check log, got: %s", output)
	}
	if !strings.Contains(output, "vault connection successful") {
		t.Errorf("Expected successful connection log, got: %s", output)
	}

	// Test ListCertificates logging
	buf.Reset()
	_, err = client.ListCertificates(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output = buf.String()
	if !strings.Contains(output, "listing certificates from vault mounts") {
		t.Errorf("Expected listing logs, got: %s", output)
	}
	if !strings.Contains(output, "listing certificates from mount") {
		t.Errorf("Expected mount listing logs, got: %s", output)
	}
	if !strings.Contains(output, "successfully listed certificates from mount") {
		t.Errorf("Expected success logs, got: %s", output)
	}
	if !strings.Contains(output, "completed certificate listing and cached result") {
		t.Errorf("Expected completion logs, got: %s", output)
	}

	// Test GetCertificateDetails logging
	buf.Reset()
	_, err = client.GetCertificateDetails(context.Background(), "01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output = buf.String()
	if !strings.Contains(output, "getting certificate details") {
		t.Errorf("Expected details log, got: %s", output)
	}
	if !strings.Contains(output, "successfully retrieved and cached certificate details") {
		t.Errorf("Expected success details log, got: %s", output)
	}

	// Test InvalidateCache logging
	buf.Reset()
	client.InvalidateCache()

	output = buf.String()
	if !strings.Contains(output, "invalidating vault client cache") {
		t.Errorf("Expected cache invalidation start log, got: %s", output)
	}
	if !strings.Contains(output, "cache invalidated successfully") {
		t.Errorf("Expected cache invalidation success log, got: %s", output)
	}
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
