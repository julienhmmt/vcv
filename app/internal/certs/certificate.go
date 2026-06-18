package certs

import (
	"crypto/x509"
	"strings"
	"time"
)

type Certificate struct {
	ID           string    `json:"id"`
	SerialNumber string    `json:"serialNumber"`
	CommonName   string    `json:"commonName"`
	Sans         []string  `json:"sans"`
	CertType     string    `json:"certType"`
	CreatedAt    time.Time `json:"createdAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
	Revoked      bool      `json:"revoked"`
}

type DetailedCertificate struct {
	Certificate
	Issuer            string   `json:"issuer"`
	Subject           string   `json:"subject"`
	KeyAlgorithm      string   `json:"keyAlgorithm"`
	KeySize           int      `json:"keySize"`
	FingerprintSHA1   string   `json:"fingerprintSHA1"`
	FingerprintSHA256 string   `json:"fingerprintSHA256"`
	Usage             []string `json:"usage"`
	PEM               string   `json:"pem"`
	CAType            string   `json:"caType"` // "intermediate" or "root"
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

func InferCertType(cert *x509.Certificate) string {
	if cert == nil {
		return "unknown"
	}
	hasServerAuth := false
	hasClientAuth := false
	for _, extUsage := range cert.ExtKeyUsage {
		switch extUsage {
		case x509.ExtKeyUsageServerAuth:
			hasServerAuth = true
		case x509.ExtKeyUsageClientAuth:
			hasClientAuth = true
		}
	}
	switch {
	case hasServerAuth && hasClientAuth:
		return "both"
	case hasServerAuth:
		return "machine"
	case hasClientAuth:
		return "user"
	default:
		return "unknown"
	}
}
