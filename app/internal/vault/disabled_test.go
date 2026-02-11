package vault

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"vcv/internal/certs"
)

func TestDisabledClient_Functions(t *testing.T) {
	client := &disabledClient{}

	// Test CheckConnection
	err := client.CheckConnection(context.Background())
	assert.Error(t, err)
	assert.Equal(t, ErrVaultNotConfigured, err)

	// Test GetCertificateDetails
	details, err := client.GetCertificateDetails(context.Background(), "test-cert")
	assert.Error(t, err)
	assert.Equal(t, ErrVaultNotConfigured, err)
	assert.Equal(t, certs.DetailedCertificate{}, details)

	// Test GetCertificatePEM
	pemResp, err := client.GetCertificatePEM(context.Background(), "test-cert")
	assert.Error(t, err)
	assert.Equal(t, ErrVaultNotConfigured, err)
	assert.Equal(t, certs.PEMResponse{}, pemResp)

	// Test InvalidateCache (should not panic)
	assert.NotPanics(t, func() {
		client.InvalidateCache()
	})

	// Test ListCertificates
	certList, err := client.ListCertificates(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, certList)
	assert.Equal(t, []certs.Certificate{}, certList)

	// Test Shutdown (should not panic)
	assert.NotPanics(t, func() {
		client.Shutdown()
	})
}

func TestDisabledClient_ErrorConsistency(t *testing.T) {
	client := &disabledClient{}

	// All methods except ListCertificates should return the same error
	expectedErr := ErrVaultNotConfigured

	assert.Equal(t, expectedErr, client.CheckConnection(context.Background()))
	_, err := client.GetCertificateDetails(context.Background(), "test")
	assert.Equal(t, expectedErr, err)
	_, err = client.GetCertificatePEM(context.Background(), "test")
	assert.Equal(t, expectedErr, err)

	// ListCertificates should return nil error
	_, err = client.ListCertificates(context.Background())
	assert.NoError(t, err)
}

func TestErrVaultNotConfigured(t *testing.T) {
	// Test the error variable
	assert.NotNil(t, ErrVaultNotConfigured)
	assert.Equal(t, "vault is not configured", ErrVaultNotConfigured.Error())

	// Test it's a proper error type
	err := ErrVaultNotConfigured
	assert.Error(t, err)
	assert.Equal(t, "vault is not configured", err.Error())
}
