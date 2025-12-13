package vault

import (
	"context"

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
