package certs

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"strconv"
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
	// IssuerCN is the issuer Common Name (list-time parse for metrics/UI).
	IssuerCN string `json:"issuerCN,omitempty"`
	// KeyAlgorithm is the public key algorithm (RSA, ECDSA, Ed25519, ...).
	KeyAlgorithm string `json:"keyAlgorithm,omitempty"`
	// KeySize is the public key size in bits (0 when not applicable).
	KeySize int `json:"keySize,omitempty"`
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

// KeyAlgoAndSize returns the public key algorithm name and bit size for an x509 certificate.
func KeyAlgoAndSize(certificate *x509.Certificate) (string, int) {
	if certificate == nil {
		return "", 0
	}
	switch pub := certificate.PublicKey.(type) {
	case *rsa.PublicKey:
		return "RSA", pub.N.BitLen()
	case *ecdsa.PublicKey:
		return "ECDSA", pub.Curve.Params().BitSize
	case ed25519.PublicKey:
		return "Ed25519", len(pub) * 8
	default:
		return certificate.PublicKeyAlgorithm.String(), 0
	}
}

// KeySizeLabel formats a key size for metric labels.
func KeySizeLabel(size int) string {
	if size <= 0 {
		return "0"
	}
	return strconv.Itoa(size)
}
