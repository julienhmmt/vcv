package vault

import (
	"context"
	"fmt"

	"vcv/internal/certs"
)

type mockClient struct {
	certificates []certs.Certificate
}

func NewMockClient() Client {
	return &mockClient{
		certificates: certs.MockCertificates(),
	}
}

func (c *mockClient) ListCertificates(ctx context.Context) ([]certs.Certificate, error) {
	result := make([]certs.Certificate, len(c.certificates))
	copy(result, c.certificates)
	return result, nil
}

func (c *mockClient) RevokeCertificate(ctx context.Context, serialNumber string, writeToken string) error {
	for index := range c.certificates {
		certificate := &c.certificates[index]
		if certificate.ID == serialNumber {
			certificate.Revoked = true
			return nil
		}
	}
	return fmt.Errorf("certificate with id %s not found", serialNumber)
}
