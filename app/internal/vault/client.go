package vault

import (
	"context"
	"errors"
	"time"

	"vcv/internal/certs"
)

// Client defines the interface for interacting with Vault PKI.
type Client interface {
	CheckConnection(ctx context.Context) error
	GetCertificateDetails(ctx context.Context, serialNumber string) (certs.DetailedCertificate, error)
	GetCertificatePEM(ctx context.Context, serialNumber string) (certs.PEMResponse, error)
	GetIntermediateCA(ctx context.Context, mount string) (certs.DetailedCertificate, error)
	InvalidateCache()
	ListCertificates(ctx context.Context) ([]certs.Certificate, error)
	Shutdown()
}

type CacheSizer interface {
	CacheSize() int
}

type ListCertificatesByVaultResult struct {
	VaultID      string
	Certificates []certs.Certificate
	Duration     time.Duration
	ListError    error
}

type CertificatesByVaultLister interface {
	ListCertificatesByVault(ctx context.Context) []ListCertificatesByVaultResult
}

// PublicErrorMessage maps a vault error to a stable, non-sensitive string for
// unauthenticated API responses: raw errors carry Vault URLs, paths and API
// messages, which belong in server logs only.
func PublicErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, ErrVaultNotConfigured) {
		return "vault not configured"
	}
	return "vault unavailable"
}

// VaultError reports a per-vault failure during a partial-success listing.
// Message is sanitized (see PublicErrorMessage).
type VaultError struct {
	VaultID string `json:"vaultId"`
	Message string `json:"message"`
}

// CertificatesEnvelopeLister returns successful certificates alongside
// per-vault errors. Allows the handler to surface partial-success state to the
// frontend without failing the whole request when a subset of vaults is down.
type CertificatesEnvelopeLister interface {
	ListCertificatesEnvelope(ctx context.Context) ([]certs.Certificate, []VaultError)
}
