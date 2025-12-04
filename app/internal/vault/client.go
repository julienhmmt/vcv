package vault

import (
	"context"

	"vcv/internal/certs"
)

type Client interface {
	ListCertificates(ctx context.Context) ([]certs.Certificate, error)
	RevokeCertificate(ctx context.Context, serialNumber string, writeToken string) error
}
