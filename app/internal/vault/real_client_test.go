package vault

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"vcv/config"
	"vcv/internal/cache"

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
	if clientConfig == nil {
		t.Fatalf("expected default config")
	}
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
		name string
		cfg  config.VaultConfig
	}{
		{name: "empty address", cfg: config.VaultConfig{Addr: "", ReadToken: "token"}},
		{name: "empty token", cfg: config.VaultConfig{Addr: "http://localhost:8200", ReadToken: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClientFromConfig(tt.cfg)
			if err == nil {
				t.Fatalf("expected error")
			}
			if client != nil {
				t.Fatalf("expected nil client")
			}
		})
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
