package certs

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"
)

type Certificate struct {
	ID         string    `json:"id"`
	CommonName string    `json:"commonName"`
	Sans       []string  `json:"sans"`
	CreatedAt  time.Time `json:"createdAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Revoked    bool      `json:"revoked"`
}

type DetailedCertificate struct {
	Certificate
	SerialNumber      string   `json:"serialNumber"`
	Issuer            string   `json:"issuer"`
	Subject           string   `json:"subject"`
	KeyAlgorithm      string   `json:"keyAlgorithm"`
	KeySize           int      `json:"keySize"`
	FingerprintSHA1   string   `json:"fingerprintSHA1"`
	FingerprintSHA256 string   `json:"fingerprintSHA256"`
	Usage             []string `json:"usage"`
	PEM               string   `json:"pem"`
}

type PEMResponse struct {
	SerialNumber string `json:"serialNumber"`
	PEM          string `json:"pem"`
}

// IsExpired returns true if the certificate has expired
func (c *Certificate) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// DaysUntilExpiry returns the number of days until the certificate expires
func (c *Certificate) DaysUntilExpiry() int {
	return int(time.Until(c.ExpiresAt).Hours() / 24)
}

// IsValidAt returns true if the certificate is valid at the given time
func (c *Certificate) IsValidAt(t time.Time) bool {
	return !c.Revoked && t.After(c.CreatedAt) && t.Before(c.ExpiresAt)
}

// HasSubject returns true if the certificate matches the given subject
func (c *Certificate) HasSubject(subject string) bool {
	if strings.EqualFold(c.CommonName, subject) {
		return true
	}
	for _, san := range c.Sans {
		if strings.EqualFold(san, subject) {
			return true
		}
	}
	return false
}

// GetStatus returns a human-readable status for the certificate
func (c *Certificate) GetStatus() string {
	if c.Revoked {
		return "revoked"
	}
	if c.IsExpired() {
		return "expired"
	}
	return "valid"
}

// ParsePEM parses a PEM-encoded certificate and returns a DetailedCertificate
func ParsePEM(pemData string) (*DetailedCertificate, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	detailed := &DetailedCertificate{
		Certificate: Certificate{
			CommonName: cert.Subject.CommonName,
			CreatedAt:  cert.NotBefore,
			ExpiresAt:  cert.NotAfter,
			Revoked:    false, // This would need to be determined from external source
		},
		SerialNumber: cert.SerialNumber.String(),
		Issuer:       cert.Issuer.CommonName,
		Subject:      cert.Subject.CommonName,
	}

	detailed.Usage = getUsage(cert)
	detailed.KeyAlgorithm = cert.PublicKeyAlgorithm.String()
	detailed.KeySize = getKeySize(cert)
	detailed.Sans = append(detailed.Sans, cert.DNSNames...)
	detailed.Sans = append(detailed.Sans, cert.EmailAddresses...)
	for _, ip := range cert.IPAddresses {
		detailed.Sans = append(detailed.Sans, ip.String())
	}

	detailed.PEM = pemData

	return detailed, nil
}

// getKeySize extracts the key size from the certificate
func getKeySize(cert *x509.Certificate) int {
	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return pub.Size() * 8 // Size() returns bytes, convert to bits
	case interface{ Bits() int }:
		return pub.Bits()
	default:
		return 0
	}
}

// getUsage extracts key usage from the certificate
func getUsage(cert *x509.Certificate) []string {
	var usage []string

	if cert.KeyUsage&x509.KeyUsageDigitalSignature != 0 {
		usage = append(usage, "Digital Signature")
	}
	if cert.KeyUsage&x509.KeyUsageKeyEncipherment != 0 {
		usage = append(usage, "Key Encipherment")
	}
	if cert.KeyUsage&x509.KeyUsageKeyAgreement != 0 {
		usage = append(usage, "Key Agreement")
	}
	if cert.KeyUsage&x509.KeyUsageCertSign != 0 {
		usage = append(usage, "Certificate Sign")
	}
	if cert.KeyUsage&x509.KeyUsageCRLSign != 0 {
		usage = append(usage, "CRL Sign")
	}
	if cert.KeyUsage&x509.KeyUsageEncipherOnly != 0 {
		usage = append(usage, "Encipher Only")
	}
	if cert.KeyUsage&x509.KeyUsageDecipherOnly != 0 {
		usage = append(usage, "Decipher Only")
	}

	for _, ext := range cert.ExtKeyUsage {
		switch ext {
		case x509.ExtKeyUsageServerAuth:
			usage = append(usage, "Server Auth")
		case x509.ExtKeyUsageClientAuth:
			usage = append(usage, "Client Auth")
		case x509.ExtKeyUsageCodeSigning:
			usage = append(usage, "Code Signing")
		case x509.ExtKeyUsageEmailProtection:
			usage = append(usage, "Email Protection")
		case x509.ExtKeyUsageTimeStamping:
			usage = append(usage, "Time Stamping")
		case x509.ExtKeyUsageOCSPSigning:
			usage = append(usage, "OCSP Signing")
		}
	}

	return usage
}
