package vault

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestCertReadPath(t *testing.T) {
	tests := []struct {
		name    string
		mount   string
		serial  string
		wantErr bool
		want    string
	}{
		{name: "prefixed colon serial", mount: "pki", serial: "1a-2b-3c", want: "pki/cert/1a-2b-3c"},
		{name: "dev mount colon octets", mount: "pki_dev", serial: "0A:FF", want: "pki_dev/cert/0A:FF"},
		{name: "legacy bare hex", mount: "pki", serial: "1a2b", want: "pki/cert/1a2b"},
		{name: "slash in serial", mount: "pki", serial: "a/b", wantErr: true},
		{name: "dot dot segment", mount: "pki", serial: "..", wantErr: true},
		{name: "traversal to other path", mount: "pki", serial: "../x", wantErr: true},
		{name: "empty serial", mount: "pki", serial: "", wantErr: true},
		{name: "non hex chars", mount: "pki", serial: "zz", wantErr: true},
		{name: "trailing separator", mount: "pki", serial: "1a:", wantErr: true},
		{name: "octet too long", mount: "pki", serial: strings.Repeat("a", 65), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := certReadPath(tt.mount, tt.serial)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidSerial) {
					t.Fatalf("expected ErrInvalidSerial, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected path %q, got %q", tt.want, got)
			}
		})
	}
}

// recordingVaultServer returns a fake Vault and the list of request paths it saw.
func recordingVaultServer(t *testing.T, listKeys []string) (*httptest.Server, func() []string) {
	t.Helper()
	inner := newVaultHTTPHandler(newVaultTestCertificatePEM(t), listKeys, []string{"aa", "bb"})
	var mu sync.Mutex
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()
		inner(w, r)
	}))
	return server, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), paths...)
	}
}

func TestInvalidSerialNeverReachesVault(t *testing.T) {
	server, requests := recordingVaultServer(t, []string{"aa"})
	defer server.Close()
	client := newRealClientForTest(t, server.URL, []string{"pki"})
	ctx := context.Background()

	if _, err := client.readCertificateFromMount(ctx, "pki", "../x"); !errors.Is(err, ErrInvalidSerial) {
		t.Fatalf("readCertificateFromMount: expected ErrInvalidSerial, got %v", err)
	}
	if _, err := client.GetCertificatePEM(ctx, "pki:zz"); !errors.Is(err, ErrInvalidSerial) {
		t.Fatalf("GetCertificatePEM: expected ErrInvalidSerial, got %v", err)
	}
	if _, err := client.GetCertificateDetails(ctx, "pki:a/b"); !errors.Is(err, ErrInvalidSerial) {
		t.Fatalf("GetCertificateDetails: expected ErrInvalidSerial, got %v", err)
	}
	for _, p := range requests() {
		if strings.HasPrefix(p, "/v1/pki/cert/") {
			t.Fatalf("invalid serial reached Vault path %q", p)
		}
	}
}

func TestListCertificates_SkipsInvalidSerial(t *testing.T) {
	server, requests := recordingVaultServer(t, []string{"aa", "../x"})
	defer server.Close()
	client := newRealClientForTest(t, server.URL, []string{"pki"})

	certificates, err := client.ListCertificates(context.Background())
	if err != nil {
		t.Fatalf("expected listing to succeed, got %v", err)
	}
	if len(certificates) != 1 || certificates[0].SerialNumber != "aa" {
		t.Fatalf("expected only serial aa, got %+v", certificates)
	}
	for _, p := range requests() {
		if strings.HasPrefix(p, "/v1/pki/cert/") && p != "/v1/pki/cert/aa" {
			t.Fatalf("unexpected cert read %q", p)
		}
	}
}
