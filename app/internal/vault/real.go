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
	"strings"
	"time"

	"vcv/config"
	"vcv/internal/cache"
	"vcv/internal/certs"

	"github.com/hashicorp/vault/api"
)

type realClient struct {
	client   *api.Client
	mounts   []string
	addr     string
	cache    *cache.Cache
	stopChan chan struct{}
}

func NewClientFromConfig(cfg config.VaultConfig) (Client, error) {
	if cfg.Addr == "" && cfg.ReadToken == "" {
		return &disabledClient{}, nil
	}
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
		mounts:   cfg.PKIMounts,
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

	var allCertificates []certs.Certificate
	revokedSet := make(map[string]bool)

	// Collect certificates from all mounts
	for _, mount := range c.mounts {
		mountCerts, mountRevoked, err := c.listCertificatesFromMount(ctx, mount)
		if err != nil {
			// Log error but continue with other mounts
			continue
		}
		allCertificates = append(allCertificates, mountCerts...)
		for serial := range mountRevoked {
			revokedSet[serial] = true
		}
	}

	sort.Slice(allCertificates, func(leftIndex, rightIndex int) bool {
		return allCertificates[leftIndex].CommonName < allCertificates[rightIndex].CommonName
	})

	// Cache the result
	c.cache.Set("certificates", allCertificates)

	return allCertificates, nil
}

func (c *realClient) listCertificatesFromMount(_ context.Context, mount string) ([]certs.Certificate, map[string]bool, error) {
	listPath := fmt.Sprintf("%s/certs", mount)
	secret, err := c.client.Logical().List(listPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list certificates from mount %s: %w", mount, err)
	}
	if secret == nil || secret.Data == nil {
		return []certs.Certificate{}, make(map[string]bool), nil
	}

	rawKeys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("unexpected list response from Vault for mount %s: missing keys array", mount)
	}

	revokedSet, err := c.fetchRevokedSerialsFromMount(mount)
	if err != nil {
		return nil, nil, err
	}

	result := make([]certs.Certificate, 0, len(rawKeys))
	for _, value := range rawKeys {
		serial, ok := value.(string)
		if !ok {
			continue
		}
		certificate, err := c.readCertificateFromMount(mount, serial)
		if err != nil {
			continue
		}
		if revokedSet[serial] {
			certificate.Revoked = true
		}
		result = append(result, certificate)
	}

	return result, revokedSet, nil
}

func (c *realClient) fetchRevokedSerialsFromMount(mount string) (map[string]bool, error) {
	path := fmt.Sprintf("%s/certs/revoked", mount)
	secret, err := c.client.Logical().List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list revoked certificates from mount %s: %w", mount, err)
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

func (c *realClient) readCertificateFromMount(mount, serial string) (certs.Certificate, error) {
	path := fmt.Sprintf("%s/cert/%s", mount, serial)
	secret, err := c.client.Logical().Read(path)
	if err != nil {
		return certs.Certificate{}, fmt.Errorf("failed to read certificate %s from mount %s: %w", serial, mount, err)
	}
	if secret == nil || secret.Data == nil {
		return certs.Certificate{}, fmt.Errorf("certificate %s not found in mount %s", serial, mount)
	}

	certificatePEM, ok := secret.Data["certificate"].(string)
	if !ok || certificatePEM == "" {
		return certs.Certificate{}, fmt.Errorf("certificate field missing for %s in mount %s", serial, mount)
	}

	block, _ := pem.Decode([]byte(certificatePEM))
	if block == nil {
		return certs.Certificate{}, fmt.Errorf("failed to decode PEM for certificate %s in mount %s", serial, mount)
	}

	x509Certificate, parseError := x509.ParseCertificate(block.Bytes)
	if parseError != nil {
		return certs.Certificate{}, fmt.Errorf("failed to parse certificate %s in mount %s: %w", serial, mount, parseError)
	}

	subjectAlternativeNames := make([]string, 0, len(x509Certificate.DNSNames)+len(x509Certificate.IPAddresses)+len(x509Certificate.EmailAddresses))
	subjectAlternativeNames = append(subjectAlternativeNames, x509Certificate.DNSNames...)
	for _, address := range x509Certificate.IPAddresses {
		subjectAlternativeNames = append(subjectAlternativeNames, address.String())
	}
	subjectAlternativeNames = append(subjectAlternativeNames, x509Certificate.EmailAddresses...)

	// Prefix ID with mount to avoid collisions across mounts
	return certs.Certificate{
		ID:         fmt.Sprintf("%s:%s", mount, serial),
		CommonName: x509Certificate.Subject.CommonName,
		Sans:       subjectAlternativeNames,
		CreatedAt:  x509Certificate.NotBefore.UTC(),
		ExpiresAt:  x509Certificate.NotAfter.UTC(),
		Revoked:    false,
	}, nil
}

func (c *realClient) GetCertificateDetails(ctx context.Context, serialNumber string) (certs.DetailedCertificate, error) {
	// Parse mount and serial from the prefixed ID
	mount, serial, err := c.parseMountAndSerial(serialNumber)
	if err != nil {
		return certs.DetailedCertificate{}, err
	}

	// Try cache first
	cacheKey := fmt.Sprintf("details_%s", serialNumber)
	if cached, found := c.cache.Get(cacheKey); found {
		if details, ok := cached.(certs.DetailedCertificate); ok {
			return details, nil
		}
	}

	path := fmt.Sprintf("%s/cert/%s", mount, serial)
	secret, err := c.client.Logical().Read(path)
	if err != nil {
		return certs.DetailedCertificate{}, fmt.Errorf("failed to read certificate %s from mount %s: %w", serial, mount, err)
	}
	if secret == nil || secret.Data == nil {
		return certs.DetailedCertificate{}, fmt.Errorf("certificate %s not found in mount %s", serial, mount)
	}

	certificatePEM, ok := secret.Data["certificate"].(string)
	if !ok || certificatePEM == "" {
		return certs.DetailedCertificate{}, fmt.Errorf("certificate field missing for %s in mount %s", serial, mount)
	}

	block, _ := pem.Decode([]byte(certificatePEM))
	if block == nil {
		return certs.DetailedCertificate{}, fmt.Errorf("failed to decode PEM for certificate %s in mount %s", serial, mount)
	}

	x509Certificate, parseError := x509.ParseCertificate(block.Bytes)
	if parseError != nil {
		return certs.DetailedCertificate{}, fmt.Errorf("failed to parse certificate %s in mount %s: %w", serial, mount, parseError)
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
	revokedSet, err := c.fetchRevokedSerialsFromMount(mount)
	if err != nil {
		return certs.DetailedCertificate{}, err
	}

	details := certs.DetailedCertificate{
		Certificate: certs.Certificate{
			ID:         serialNumber, // Keep the prefixed ID
			CommonName: x509Certificate.Subject.CommonName,
			Sans:       subjectAlternativeNames,
			CreatedAt:  x509Certificate.NotBefore.UTC(),
			ExpiresAt:  x509Certificate.NotAfter.UTC(),
			Revoked:    revokedSet[serial],
		},
		SerialNumber:      serial, // Store only the serial part
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

func (c *realClient) parseMountAndSerial(serialNumber string) (string, string, error) {
	// Parse mount and serial from the prefixed ID
	// Format: "mount:serial" (e.g., "pki:1234-5678", "pki_dev:abcd-efgh")
	// This prevents ID collisions across multiple PKI mounts
	parts := strings.SplitN(serialNumber, ":", 2)
	if len(parts) == 2 {
		mount := parts[0]
		serial := parts[1]

		// Validate that the mount is configured
		for _, configuredMount := range c.mounts {
			if configuredMount == mount {
				return mount, serial, nil
			}
		}
		return "", "", fmt.Errorf("mount %s is not configured", mount)
	}

	// Legacy behavior: if no prefix, use the first configured mount
	if len(c.mounts) == 0 {
		return "", "", fmt.Errorf("no mounts configured")
	}
	return c.mounts[0], serialNumber, nil
}

func (c *realClient) GetCertificatePEM(ctx context.Context, serialNumber string) (certs.PEMResponse, error) {
	details, err := c.GetCertificateDetails(ctx, serialNumber)
	if err != nil {
		return certs.PEMResponse{}, err
	}

	return certs.PEMResponse{
		SerialNumber: details.SerialNumber, // Return only the serial part
		PEM:          details.PEM,
	}, nil
}

func (c *realClient) InvalidateCache() {
	c.cache.Clear()
}
