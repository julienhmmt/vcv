package vault

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sort"

	"vcv/config"
	"vcv/internal/certs"

	"github.com/hashicorp/vault/api"
)

type realClient struct {
	client *api.Client
	mount  string
	addr   string
}

func NewClientFromConfig(cfg config.VaultConfig) (Client, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("vault address is empty")
	}
	if cfg.ReadToken == "" {
		return nil, fmt.Errorf("vault read token is empty")
	}

	apiConfig := api.DefaultConfig()
	apiConfig.Address = cfg.Addr

	client, err := api.NewClient(apiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}
	client.SetToken(cfg.ReadToken)

	return &realClient{
		client: client,
		mount:  cfg.PKIMount,
		addr:   cfg.Addr,
	}, nil
}

func (c *realClient) ListCertificates(ctx context.Context) ([]certs.Certificate, error) {
	listPath := fmt.Sprintf("%s/certs", c.mount)
	secret, err := c.client.Logical().List(listPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list certificates from Vault: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return []certs.Certificate{}, nil
	}

	rawKeys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected list response from Vault: missing keys array")
	}

	revokedSet, err := c.fetchRevokedSerials()
	if err != nil {
		return nil, err
	}

	result := make([]certs.Certificate, 0, len(rawKeys))
	for _, value := range rawKeys {
		serial, ok := value.(string)
		if !ok {
			continue
		}
		certificate, err := c.readCertificate(serial)
		if err != nil {
			continue
		}
		if revokedSet[serial] {
			certificate.Revoked = true
		}
		result = append(result, certificate)
	}

	sort.Slice(result, func(leftIndex, rightIndex int) bool {
		return result[leftIndex].CommonName < result[rightIndex].CommonName
	})

	return result, nil
}

func (c *realClient) fetchRevokedSerials() (map[string]bool, error) {
	path := fmt.Sprintf("%s/certs/revoked", c.mount)
	secret, err := c.client.Logical().List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list revoked certificates from Vault: %w", err)
	}

	serials := make(map[string]bool)
	if secret == nil || secret.Data == nil {
		return serials, nil
	}

	rawKeys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return serials, nil
	}

	for _, value := range rawKeys {
		serial, ok := value.(string)
		if ok {
			serials[serial] = true
		}
	}

	return serials, nil
}

func (c *realClient) readCertificate(serial string) (certs.Certificate, error) {
	path := fmt.Sprintf("%s/cert/%s", c.mount, serial)
	secret, err := c.client.Logical().Read(path)
	if err != nil {
		return certs.Certificate{}, fmt.Errorf("failed to read certificate %s from Vault: %w", serial, err)
	}
	if secret == nil || secret.Data == nil {
		return certs.Certificate{}, fmt.Errorf("certificate %s not found in Vault", serial)
	}

	certificatePEM, ok := secret.Data["certificate"].(string)
	if !ok || certificatePEM == "" {
		return certs.Certificate{}, fmt.Errorf("certificate field missing for %s", serial)
	}

	block, _ := pem.Decode([]byte(certificatePEM))
	if block == nil {
		return certs.Certificate{}, fmt.Errorf("failed to decode PEM for certificate %s", serial)
	}

	x509Certificate, parseError := x509.ParseCertificate(block.Bytes)
	if parseError != nil {
		return certs.Certificate{}, fmt.Errorf("failed to parse certificate %s: %w", serial, parseError)
	}

	subjectAlternativeNames := make([]string, 0, len(x509Certificate.DNSNames)+len(x509Certificate.IPAddresses)+len(x509Certificate.EmailAddresses))
	for _, name := range x509Certificate.DNSNames {
		subjectAlternativeNames = append(subjectAlternativeNames, name)
	}
	for _, address := range x509Certificate.IPAddresses {
		subjectAlternativeNames = append(subjectAlternativeNames, address.String())
	}
	for _, email := range x509Certificate.EmailAddresses {
		subjectAlternativeNames = append(subjectAlternativeNames, email)
	}

	return certs.Certificate{
		ID:         serial,
		CommonName: x509Certificate.Subject.CommonName,
		Sans:       subjectAlternativeNames,
		CreatedAt:  x509Certificate.NotBefore.UTC(),
		ExpiresAt:  x509Certificate.NotAfter.UTC(),
		Revoked:    false,
	}, nil
}

func (c *realClient) RevokeCertificate(ctx context.Context, serialNumber string, writeToken string) error {
	if writeToken == "" {
		return fmt.Errorf("write token is required to revoke a certificate")
	}

	apiConfig := api.DefaultConfig()
	apiConfig.Address = c.addr

	client, err := api.NewClient(apiConfig)
	if err != nil {
		return fmt.Errorf("failed to create Vault client for revocation: %w", err)
	}
	client.SetToken(writeToken)

	path := fmt.Sprintf("%s/revoke", c.mount)
	_, err = client.Logical().Write(path, map[string]interface{}{
		"serial_number": serialNumber,
	})
	if err != nil {
		return fmt.Errorf("failed to revoke certificate %s in Vault: %w", serialNumber, err)
	}

	return nil
}
