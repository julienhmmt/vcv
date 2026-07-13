package vault

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"vcv/internal/cache"
	"vcv/internal/certs"
	"vcv/internal/config"
	"vcv/internal/logger"

	"github.com/hashicorp/vault/api"
)

const cacheVersion = "v2"

type realClient struct {
	client   *api.Client
	mounts   []string
	addr     string
	cache    *cache.Cache
	stopChan chan struct{}
}

func decodeBase64String(value string) ([]byte, error) {
	normalized := strings.Join(strings.Fields(value), "")
	decoded, err := base64.StdEncoding.DecodeString(normalized)
	if err == nil {
		return decoded, nil
	}
	decoded, rawErr := base64.RawStdEncoding.DecodeString(normalized)
	if rawErr == nil {
		return decoded, nil
	}
	return nil, err
}

func NewClientFromConfig(cfg config.VaultConfig) (Client, error) {
	if cfg.Addr == "" && cfg.ReadToken == "" {
		logger.Get().Debug().Msg("creating disabled vault client - no address and token provided")
		return &disabledClient{}, nil
	}
	if cfg.Addr == "" {
		return nil, fmt.Errorf("vault address is empty")
	}
	if cfg.ReadToken == "" {
		return nil, fmt.Errorf("vault read token is empty")
	}

	logger.Get().Debug().
		Str("vault_addr", cfg.Addr).
		Strs("vault_mounts", cfg.PKIMounts).
		Msg("creating new vault client")

	clientConfig := api.DefaultConfig()
	if clientConfig == nil {
		return nil, fmt.Errorf("failed to create default Vault config")
	}

	clientConfig.Address = cfg.Addr
	tlsConfig := &api.TLSConfig{CACert: cfg.TLSCACert, CAPath: cfg.TLSCAPath, TLSServerName: cfg.TLSServerName, Insecure: cfg.TLSInsecure}
	if strings.TrimSpace(cfg.TLSCACertBase64) != "" {
		decoded, err := decodeBase64String(strings.TrimSpace(cfg.TLSCACertBase64))
		if err != nil {
			return nil, fmt.Errorf("invalid vault tls ca cert base64: %w", err)
		}
		tlsConfig.CACertBytes = decoded
		tlsConfig.CACert = ""
		tlsConfig.CAPath = ""
	}
	if err := clientConfig.ConfigureTLS(tlsConfig); err != nil {
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
		cache:    cache.New(15 * time.Minute),
		stopChan: make(chan struct{}),
	}

	// Clear cache on startup to invalidate old schema versions
	c.cache.Clear()

	logger.Get().Info().
		Str("vault_addr", cfg.Addr).
		Int("mount_count", len(cfg.PKIMounts)).
		Msg("vault client created successfully")

	if cfg.TLSInsecure {
		logger.Get().Warn().
			Str("vault_addr", cfg.Addr).
			Msg("Vault TLS certificate verification is disabled (tls_insecure=true)")
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
	logger.Get().Debug().
		Str("vault_addr", c.addr).
		Msg("checking vault connection")

	health, err := c.client.Sys().HealthWithContext(ctx)
	if err != nil {
		logger.Get().Error().
			Str("vault_addr", c.addr).
			Err(err).
			Msg("vault health check failed")
		return fmt.Errorf("vault health check failed: %w", err)
	}
	if health == nil {
		return fmt.Errorf("vault health response is nil")
	}
	if !health.Initialized {
		logger.Get().Error().
			Str("vault_addr", c.addr).
			Msg("vault is not initialized")
		return fmt.Errorf("vault is not initialized")
	}
	if health.Sealed {
		logger.Get().Error().
			Str("vault_addr", c.addr).
			Msg("vault is sealed")
		return fmt.Errorf("vault is sealed")
	}

	logger.Get().Debug().
		Str("vault_addr", c.addr).
		Str("version", health.Version).
		Msg("vault liveness check successful")

	// Check if token is usable
	_, err = c.client.Auth().Token().LookupSelfWithContext(ctx)
	if err != nil {
		logger.Get().Error().
			Str("vault_addr", c.addr).
			Err(err).
			Msg("vault token check failed")
		return fmt.Errorf("vault token check failed: %w", err)
	}

	logger.Get().Debug().
		Str("vault_addr", c.addr).
		Msg("vault connection successful")

	return nil
}

// Shutdown stops background goroutines.
func (c *realClient) Shutdown() {
	logger.Get().Debug().
		Str("vault_addr", c.addr).
		Msg("shutting down vault client")
	close(c.stopChan)
}

func (c *realClient) ListCertificates(ctx context.Context) ([]certs.Certificate, error) {
	// Try cache first
	if cached, found := c.cache.Get(cacheVersion + ":certificates"); found {
		if certificates, ok := cached.([]certs.Certificate); ok {
			logger.Get().Debug().
				Str("vault_addr", c.addr).
				Int("cached_certificates", len(certificates)).
				Msg("serving certificates from cache")
			return certificates, nil
		}
	}

	logger.Get().Debug().
		Str("vault_addr", c.addr).
		Strs("mounts", c.mounts).
		Msg("listing certificates from vault mounts")

	if len(c.mounts) == 0 {
		return []certs.Certificate{}, ErrVaultNotConfigured
	}
	var allCertificates []certs.Certificate
	listedMounts := 0
	var lastError error

	// Collect certificates from all mounts
	for _, mount := range c.mounts {
		logger.Get().Debug().
			Str("vault_addr", c.addr).
			Str("mount", mount).
			Msg("listing certificates from mount")

		mountCerts, mountRevoked, err := c.listCertificatesFromMount(ctx, mount)
		if err != nil {
			logger.Get().Error().
				Str("vault_addr", c.addr).
				Str("mount", mount).
				Err(err).
				Msg("failed to list certificates from mount")
			// Log error but continue with other mounts
			lastError = err
			continue
		}

		logger.Get().Debug().
			Str("vault_addr", c.addr).
			Str("mount", mount).
			Int("certificate_count", len(mountCerts)).
			Int("revoked_count", len(mountRevoked)).
			Msg("successfully listed certificates from mount")

		listedMounts += 1
		allCertificates = append(allCertificates, mountCerts...)
	}
	if listedMounts == 0 {
		if lastError != nil {
			return []certs.Certificate{}, lastError
		}
		return []certs.Certificate{}, fmt.Errorf("failed to list certificates from mounts")
	}

	sort.Slice(allCertificates, func(leftIndex, rightIndex int) bool {
		return allCertificates[leftIndex].CommonName < allCertificates[rightIndex].CommonName
	})

	// Cache the result
	c.cache.Set(cacheVersion+":certificates", allCertificates)

	logger.Get().Debug().
		Str("vault_addr", c.addr).
		Int("total_certificates", len(allCertificates)).
		Int("successful_mounts", listedMounts).
		Msg("completed certificate listing and cached result")

	return allCertificates, nil
}

func (c *realClient) listCertificatesFromMount(ctx context.Context, mount string) ([]certs.Certificate, map[string]bool, error) {
	listPath := fmt.Sprintf("%s/certs", mount)
	secret, err := c.client.Logical().ListWithContext(ctx, listPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list certificates from mount %s: %w", mount, err)
	}
	if secret == nil || secret.Data == nil {
		return []certs.Certificate{}, make(map[string]bool), nil
	}

	rawKeys, ok := secret.Data["keys"].([]any)
	if !ok {
		return nil, nil, fmt.Errorf("unexpected list response from Vault for mount %s: missing keys array", mount)
	}

	revokedSet, err := c.fetchRevokedSerialsFromMount(ctx, mount)
	if err != nil {
		return nil, nil, err
	}

	result := make([]certs.Certificate, 0, len(rawKeys))
	for _, value := range rawKeys {
		serial, ok := value.(string)
		if !ok {
			continue
		}
		certificate, err := c.readCertificateFromMount(ctx, mount, serial)
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

func (c *realClient) fetchRevokedSerialsFromMount(ctx context.Context, mount string) (map[string]bool, error) {
	path := fmt.Sprintf("%s/certs/revoked", mount)
	secret, err := c.client.Logical().ListWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list revoked certificates from mount %s: %w", mount, err)
	}

	serials := make(map[string]bool)
	if secret == nil || secret.Data == nil {
		return serials, nil
	}

	rawKeys, ok := secret.Data["keys"].([]any)
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

func (c *realClient) readCertificateFromMount(ctx context.Context, mount, serial string) (certs.Certificate, error) {
	if serial == "" {
		return certs.Certificate{}, fmt.Errorf("serial number cannot be empty")
	}

	path := fmt.Sprintf("%s/cert/%s", mount, serial)
	secret, err := c.client.Logical().ReadWithContext(ctx, path)
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

	subjectAlternativeNames := buildSANs(x509Certificate)

	algo, keySize := certs.KeyAlgoAndSize(x509Certificate)
	// Prefix ID with mount to avoid collisions across mounts
	return certs.Certificate{
		ID:           fmt.Sprintf("%s:%s", mount, serial),
		SerialNumber: serial,
		CommonName:   x509Certificate.Subject.CommonName,
		Sans:         subjectAlternativeNames,
		CertType:     certs.InferCertType(x509Certificate),
		CreatedAt:    x509Certificate.NotBefore.UTC(),
		ExpiresAt:    x509Certificate.NotAfter.UTC(),
		Revoked:      false,
		IssuerCN:     x509Certificate.Issuer.CommonName,
		KeyAlgorithm: algo,
		KeySize:      keySize,
	}, nil
}

func (c *realClient) GetCertificateDetails(ctx context.Context, serialNumber string) (certs.DetailedCertificate, error) {
	// Parse mount and serial from the prefixed ID
	mount, serial, err := c.parseMountAndSerial(serialNumber)
	if err != nil {
		return certs.DetailedCertificate{}, err
	}
	if serial == "" {
		return certs.DetailedCertificate{}, fmt.Errorf("serial number cannot be empty")
	}

	logger.Get().Debug().
		Str("vault_addr", c.addr).
		Str("mount", mount).
		Str("serial", serial).
		Msg("getting certificate details")

	// Try cache first
	cacheKey := fmt.Sprintf("%s:details_%s", cacheVersion, serialNumber)
	if cached, found := c.cache.Get(cacheKey); found {
		if details, ok := cached.(certs.DetailedCertificate); ok {
			logger.Get().Debug().
				Str("vault_addr", c.addr).
				Str("serial", serial).
				Msg("serving certificate details from cache")
			return details, nil
		}
	}

	path := fmt.Sprintf("%s/cert/%s", mount, serial)
	secret, err := c.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		logger.Get().Error().
			Str("vault_addr", c.addr).
			Str("mount", mount).
			Str("serial", serial).
			Err(err).
			Msg("failed to read certificate from vault")
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

	subjectAlternativeNames := buildSANs(x509Certificate)

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
	revokedSet, err := c.fetchRevokedSerialsFromMount(ctx, mount)
	if err != nil {
		return certs.DetailedCertificate{}, err
	}

	algo, keySize := certs.KeyAlgoAndSize(x509Certificate)
	details := certs.DetailedCertificate{
		Certificate: certs.Certificate{
			ID:           serialNumber, // Keep the prefixed ID
			SerialNumber: serial,       // Store only the serial part
			CommonName:   x509Certificate.Subject.CommonName,
			Sans:         subjectAlternativeNames,
			CertType:     certs.InferCertType(x509Certificate),
			CreatedAt:    x509Certificate.NotBefore.UTC(),
			ExpiresAt:    x509Certificate.NotAfter.UTC(),
			Revoked:      revokedSet[serial],
			IssuerCN:     x509Certificate.Issuer.CommonName,
			KeyAlgorithm: algo,
			KeySize:      keySize,
		},
		Issuer:            x509Certificate.Issuer.String(),
		Subject:           x509Certificate.Subject.String(),
		KeyAlgorithm:      algo,
		KeySize:           keySize,
		FingerprintSHA1:   hex.EncodeToString(sha1Fingerprint[:]),
		FingerprintSHA256: hex.EncodeToString(sha256Fingerprint[:]),
		Usage:             usage,
		PEM:               certificatePEM,
	}

	// Cache the full detailed certificate
	c.cache.Set(cacheKey, details)

	logger.Get().Debug().
		Str("vault_addr", c.addr).
		Str("serial", serial).
		Str("common_name", x509Certificate.Subject.CommonName).
		Msg("successfully retrieved and cached certificate details")

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
		if slices.Contains(c.mounts, mount) {
			return mount, serial, nil
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
	logger.Get().Debug().
		Str("vault_addr", c.addr).
		Str("serial_number", serialNumber).
		Msg("getting certificate PEM")

	details, err := c.GetCertificateDetails(ctx, serialNumber)
	if err != nil {
		return certs.PEMResponse{}, err
	}

	logger.Get().Debug().
		Str("vault_addr", c.addr).
		Str("serial_number", serialNumber).
		Msg("successfully retrieved certificate PEM")

	return certs.PEMResponse{
		SerialNumber: details.SerialNumber, // Return only the serial part
		PEM:          details.PEM,
	}, nil
}

func (c *realClient) GetIntermediateCA(ctx context.Context, mount string) (certs.DetailedCertificate, error) {
	logger.Get().Debug().
		Str("vault_addr", c.addr).
		Str("mount", mount).
		Msg("getting intermediate CA certificate")

	if mount == "" {
		return certs.DetailedCertificate{}, fmt.Errorf("mount cannot be empty")
	}

	// Try cache first
	cacheKey := fmt.Sprintf("%s:ca_%s", cacheVersion, mount)
	if cached, found := c.cache.Get(cacheKey); found {
		if details, ok := cached.(certs.DetailedCertificate); ok {
			logger.Get().Debug().
				Str("vault_addr", c.addr).
				Str("mount", mount).
				Msg("serving CA from cache")
			return details, nil
		}
	}

	// Vault PKI exposes the issuing CA cert as JSON at <mount>/cert/ca.
	// The <mount>/ca endpoint returns raw DER and cannot be read via Logical().
	path := fmt.Sprintf("%s/cert/ca", mount)
	secret, err := c.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		logger.Get().Error().
			Str("vault_addr", c.addr).
			Str("mount", mount).
			Str("path", path).
			Err(err).
			Msg("failed to read CA certificate from vault")
		return certs.DetailedCertificate{}, fmt.Errorf("failed to read CA for mount %s: %w", mount, err)
	}
	if secret == nil || secret.Data == nil {
		logger.Get().Warn().
			Str("vault_addr", c.addr).
			Str("mount", mount).
			Str("path", path).
			Msg("CA endpoint returned nil data")
		return certs.DetailedCertificate{}, fmt.Errorf("CA not found in mount %s", mount)
	}

	caPEM, _ := secret.Data["certificate"].(string)
	if caPEM == "" {
		logger.Get().Warn().
			Str("vault_addr", c.addr).
			Str("mount", mount).
			Str("path", path).
			Interface("data_keys", getMapKeys(secret.Data)).
			Msg("certificate field missing in CA response")
		return certs.DetailedCertificate{}, fmt.Errorf("certificate field missing in CA response (keys: %v)", getMapKeys(secret.Data))
	}

	block, _ := pem.Decode([]byte(caPEM))
	if block == nil {
		return certs.DetailedCertificate{}, fmt.Errorf("failed to decode PEM for CA in mount %s", mount)
	}

	x509Certificate, parseError := x509.ParseCertificate(block.Bytes)
	if parseError != nil {
		return certs.DetailedCertificate{}, fmt.Errorf("failed to parse CA in mount %s: %w", mount, parseError)
	}

	caType := "intermediate"
	if x509Certificate.Subject.String() == x509Certificate.Issuer.String() {
		caType = "root"
	}

	// Calculate fingerprints
	sha1Fingerprint := sha1.Sum(x509Certificate.Raw)
	sha256Fingerprint := sha256.Sum256(x509Certificate.Raw)

	algo, keySize := certs.KeyAlgoAndSize(x509Certificate)
	details := certs.DetailedCertificate{
		Certificate: certs.Certificate{
			ID:           fmt.Sprintf("%s:ca", mount),
			SerialNumber: x509Certificate.SerialNumber.String(),
			CommonName:   x509Certificate.Subject.CommonName,
			Sans:         append([]string(nil), x509Certificate.DNSNames...),
			CertType:     certs.InferCertType(x509Certificate),
			CreatedAt:    x509Certificate.NotBefore.UTC(),
			ExpiresAt:    x509Certificate.NotAfter.UTC(),
			Revoked:      false,
			IssuerCN:     x509Certificate.Issuer.CommonName,
			KeyAlgorithm: algo,
			KeySize:      keySize,
		},
		Issuer:            x509Certificate.Issuer.String(),
		Subject:           x509Certificate.Subject.String(),
		KeyAlgorithm:      algo,
		KeySize:           keySize,
		FingerprintSHA1:   hex.EncodeToString(sha1Fingerprint[:]),
		FingerprintSHA256: hex.EncodeToString(sha256Fingerprint[:]),
		PEM:               caPEM,
		CAType:            caType,
	}

	// Cache the result
	c.cache.Set(cacheKey, details)

	logger.Get().Debug().
		Str("vault_addr", c.addr).
		Str("mount", mount).
		Str("common_name", x509Certificate.Subject.CommonName).
		Msg("successfully retrieved and cached intermediate CA")

	return details, nil
}

func (c *realClient) InvalidateCache() {
	logger.Get().Debug().
		Str("vault_addr", c.addr).
		Msg("invalidating vault client cache")
	c.cache.Clear()
	logger.Get().Debug().
		Str("vault_addr", c.addr).
		Msg("cache invalidated successfully")
}

func (c *realClient) CacheSize() int {
	if c.cache == nil {
		return 0
	}
	return c.cache.Size()
}

func getMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// buildSANs collects the subject alternative names (DNS, IP, email) from a certificate.
func buildSANs(cert *x509.Certificate) []string {
	sans := make([]string, 0, len(cert.DNSNames)+len(cert.IPAddresses)+len(cert.EmailAddresses))
	sans = append(sans, cert.DNSNames...)
	for _, address := range cert.IPAddresses {
		sans = append(sans, address.String())
	}
	sans = append(sans, cert.EmailAddresses...)
	return sans
}
