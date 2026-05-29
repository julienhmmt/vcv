package vault

import (
	"context"
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

// VaultError reports a per-vault failure during a partial-success listing.
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
