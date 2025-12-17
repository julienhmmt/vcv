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
