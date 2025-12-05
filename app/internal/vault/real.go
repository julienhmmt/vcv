package vault

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"sort"
	"time"

	"vcv/config"
	"vcv/internal/cache"
	"vcv/internal/certs"

	"github.com/hashicorp/vault/api"
)

type realClient struct {
	client   *api.Client
	mount    string
	addr     string
	cache    *cache.Cache
	stopChan chan struct{}
}

func NewClientFromConfig(cfg config.VaultConfig) (Client, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("vault address is empty")
	}
	if cfg.ReadToken == "" {
		return nil, fmt.Errorf("vault read token is empty")
	}

	clientConfig := api.DefaultConfig()
	if clientConfig == nil {
		return nil, fmt.Errorf("failed to create default Vault config")
	}

	clientConfig.Address = cfg.Addr
	if err := clientConfig.ConfigureTLS(&api.TLSConfig{
		Insecure: cfg.TLSInsecure,
	}); err != nil {
		return nil, fmt.Errorf("failed to configure Vault TLS: %w", err)
	}

	apiClient, err := api.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	apiClient.SetToken(cfg.ReadToken)

	c := &realClient{
		client:   apiClient,
		mount:    cfg.PKIMount,
		addr:     cfg.Addr,
		cache:    cache.New(5 * time.Minute),
		stopChan: make(chan struct{}),
	}

	// Start periodic cache cleanup
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.cache.Cleanup()
			case <-c.stopChan:
				return
			}
		}
	}()

	return c, nil
}

// CheckConnection verifies Vault availability and seal status.
func (c *realClient) CheckConnection(ctx context.Context) error {
	health, err := c.client.Sys().HealthWithContext(ctx)
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}
	if health == nil {
		return fmt.Errorf("vault health response is nil")
	}
	if !health.Initialized {
		return fmt.Errorf("vault is not initialized")
	}
	if health.Sealed {
		return fmt.Errorf("vault is sealed")
	}
	return nil
}

// Shutdown stops background goroutines.
func (c *realClient) Shutdown() {
	close(c.stopChan)
}

func (c *realClient) ListCertificates(ctx context.Context) ([]certs.Certificate, error) {
	// Try cache first
	if cached, found := c.cache.Get("certificates"); found {
		if certificates, ok := cached.([]certs.Certificate); ok {
			return certificates, nil
		}
	}

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

	result := make([]certs.Certificate, 0, len(rawKeys)+len(revokedSet))
	seenSerials := make(map[string]bool, len(rawKeys))
	for _, value := range rawKeys {
		serial, ok := value.(string)
		if !ok {
			continue
		}
		seenSerials[serial] = true
		certificate, err := c.readCertificate(serial)
		if err != nil {
			continue
		}
		if revokedSet[serial] {
			certificate.Revoked = true
		}
		result = append(result, certificate)
	}

	for serial := range revokedSet {
		if seenSerials[serial] {
			continue
		}
		certificate, err := c.readCertificate(serial)
		if err != nil {
			continue
		}
		certificate.Revoked = true
		result = append(result, certificate)
	}

	sort.Slice(result, func(leftIndex, rightIndex int) bool {
		return result[leftIndex].CommonName < result[rightIndex].CommonName
	})

	// Cache the result
	c.cache.Set("certificates", result)

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
	subjectAlternativeNames = append(subjectAlternativeNames, x509Certificate.DNSNames...)
	for _, address := range x509Certificate.IPAddresses {
		subjectAlternativeNames = append(subjectAlternativeNames, address.String())
	}
	subjectAlternativeNames = append(subjectAlternativeNames, x509Certificate.EmailAddresses...)

	return certs.Certificate{
		ID:         serial,
		CommonName: x509Certificate.Subject.CommonName,
		Sans:       subjectAlternativeNames,
		CreatedAt:  x509Certificate.NotBefore.UTC(),
		ExpiresAt:  x509Certificate.NotAfter.UTC(),
		Revoked:    false,
	}, nil
}

func (c *realClient) GetCertificateDetails(ctx context.Context, serialNumber string) (certs.DetailedCertificate, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("details_%s", serialNumber)
	if cached, found := c.cache.Get(cacheKey); found {
		if details, ok := cached.(certs.DetailedCertificate); ok {
			return details, nil
		}
	}

	path := fmt.Sprintf("%s/cert/%s", c.mount, serialNumber)
	secret, err := c.client.Logical().Read(path)
	if err != nil {
		return certs.DetailedCertificate{}, fmt.Errorf("failed to read certificate %s from Vault: %w", serialNumber, err)
	}
	if secret == nil || secret.Data == nil {
		return certs.DetailedCertificate{}, fmt.Errorf("certificate %s not found in Vault", serialNumber)
	}

	certificatePEM, ok := secret.Data["certificate"].(string)
	if !ok || certificatePEM == "" {
		return certs.DetailedCertificate{}, fmt.Errorf("certificate field missing for %s", serialNumber)
	}

	block, _ := pem.Decode([]byte(certificatePEM))
	if block == nil {
		return certs.DetailedCertificate{}, fmt.Errorf("failed to decode PEM for certificate %s", serialNumber)
	}

	x509Certificate, parseError := x509.ParseCertificate(block.Bytes)
	if parseError != nil {
		return certs.DetailedCertificate{}, fmt.Errorf("failed to parse certificate %s: %w", serialNumber, parseError)
	}

	// Calculate fingerprints
	sha1Fingerprint := sha1.Sum(x509Certificate.Raw)
	sha256Fingerprint := sha256.Sum256(x509Certificate.Raw)

	subjectAlternativeNames := make([]string, 0, len(x509Certificate.DNSNames)+len(x509Certificate.IPAddresses)+len(x509Certificate.EmailAddresses))
	subjectAlternativeNames = append(subjectAlternativeNames, x509Certificate.DNSNames...)
	for _, address := range x509Certificate.IPAddresses {
		subjectAlternativeNames = append(subjectAlternativeNames, address.String())
	}
	subjectAlternativeNames = append(subjectAlternativeNames, x509Certificate.EmailAddresses...)

	// Extract key usage
	var usage []string
	if len(x509Certificate.ExtKeyUsage) > 0 {
		for _, extUsage := range x509Certificate.ExtKeyUsage {
			switch extUsage {
			case x509.ExtKeyUsageServerAuth:
				usage = append(usage, "Server Auth")
			case x509.ExtKeyUsageClientAuth:
				usage = append(usage, "Client Auth")
			case x509.ExtKeyUsageCodeSigning:
				usage = append(usage, "Code Signing")
			case x509.ExtKeyUsageEmailProtection:
				usage = append(usage, "Email Protection")
			}
		}
	}

	// Get revoked status
	revokedSet, err := c.fetchRevokedSerials()
	if err != nil {
		return certs.DetailedCertificate{}, err
	}

	details := certs.DetailedCertificate{
		Certificate: certs.Certificate{
			ID:         serialNumber,
			CommonName: x509Certificate.Subject.CommonName,
			Sans:       subjectAlternativeNames,
			CreatedAt:  x509Certificate.NotBefore.UTC(),
			ExpiresAt:  x509Certificate.NotAfter.UTC(),
			Revoked:    revokedSet[serialNumber],
		},
		SerialNumber:      serialNumber,
		Issuer:            x509Certificate.Issuer.String(),
		Subject:           x509Certificate.Subject.String(),
		KeyAlgorithm:      x509Certificate.SignatureAlgorithm.String(),
		KeySize:           0, // Would need more complex parsing for RSA/EC key sizes
		FingerprintSHA1:   hex.EncodeToString(sha1Fingerprint[:]),
		FingerprintSHA256: hex.EncodeToString(sha256Fingerprint[:]),
		Usage:             usage,
		PEM:               certificatePEM,
	}

	// Cache the full detailed certificate
	c.cache.Set(cacheKey, details)

	return details, nil
}

func (c *realClient) GetCertificatePEM(ctx context.Context, serialNumber string) (certs.PEMResponse, error) {
	details, err := c.GetCertificateDetails(ctx, serialNumber)
	if err != nil {
		return certs.PEMResponse{}, err
	}

	return certs.PEMResponse{
		SerialNumber: serialNumber,
		PEM:          details.PEM,
	}, nil
}

func (c *realClient) InvalidateCache() {
	c.cache.Clear()
}

func (c *realClient) RotateCRL(ctx context.Context) error {
	path := fmt.Sprintf("%s/crl/rotate", c.mount)
	_, err := c.client.Logical().Read(path)
	if err != nil {
		return fmt.Errorf("failed to rotate CRL: %w", err)
	}
	// Clear cached data to reflect new CRL
	c.cache.Clear()
	return nil
}

func (c *realClient) GetCRL(ctx context.Context) ([]byte, error) {
	path := fmt.Sprintf("%s/crl/pem", c.mount)
	secret, err := c.client.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get CRL from Vault: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("CRL not found in Vault")
	}
	crlData, ok := secret.Data["certificate"].(string)
	if !ok || crlData == "" {
		return nil, fmt.Errorf("CRL data missing from Vault response")
	}
	c.InvalidateCache()
	return []byte(crlData), nil
}
