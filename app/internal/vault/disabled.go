package vault

import (
	"context"
	"errors"

	"vcv/internal/certs"
)

var ErrVaultNotConfigured = errors.New("vault is not configured")

type disabledClient struct{}

func (c *disabledClient) CheckConnection(_ context.Context) error {
	return ErrVaultNotConfigured
}

func (c *disabledClient) GetCertificateDetails(_ context.Context, _ string) (certs.DetailedCertificate, error) {
	return certs.DetailedCertificate{}, ErrVaultNotConfigured
}

func (c *disabledClient) GetCertificatePEM(_ context.Context, _ string) (certs.PEMResponse, error) {
	return certs.PEMResponse{}, ErrVaultNotConfigured
}

func (c *disabledClient) InvalidateCache() {
}

func (c *disabledClient) ListCertificates(_ context.Context) ([]certs.Certificate, error) {
	return []certs.Certificate{}, nil
}

func (c *disabledClient) Shutdown() {
}
