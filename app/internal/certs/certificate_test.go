package certs

import (
	"crypto/x509"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCertificate_JSONSerialization(t *testing.T) {
	tests := []struct {
		name        string
		certificate Certificate
		expectError bool
	}{
		{
			name: "valid certificate",
			certificate: Certificate{
				ID:         "test-vault-test-mount-1234567890",
				CommonName: "test.example.com",
				Sans:       []string{"test.example.com", "www.test.example.com"},
				CertType:   "machine",
				CreatedAt:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				ExpiresAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Revoked:    false,
			},
			expectError: false,
		},
		{
			name: "certificate with empty fields",
			certificate: Certificate{
				ID:         "",
				CommonName: "",
				Sans:       []string{},
				CertType:   "unknown",
				CreatedAt:  time.Time{},
				ExpiresAt:  time.Time{},
				Revoked:    false,
			},
			expectError: false,
		},
		{
			name: "revoked certificate",
			certificate: Certificate{
				ID:         "vault-mount-serial",
				CommonName: "revoked.example.com",
				Sans:       []string{"revoked.example.com"},
				CertType:   "user",
				CreatedAt:  time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
				ExpiresAt:  time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
				Revoked:    true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			data, err := json.Marshal(tt.certificate)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, data)

			// Test JSON unmarshaling
			var unmarshaled Certificate
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tt.certificate, unmarshaled)
		})
	}
}

func TestDetailedCertificate_JSONSerialization(t *testing.T) {
	tests := []struct {
		name                string
		detailedCertificate DetailedCertificate
		expectError         bool
	}{
		{
			name: "valid detailed certificate",
			detailedCertificate: DetailedCertificate{
				Certificate: Certificate{
					ID:           "vault-mount-serial",
					SerialNumber: "1234567890ABCDEF",
					CommonName:   "detailed.example.com",
					Sans:         []string{"detailed.example.com", "www.detailed.example.com"},
					CertType:     "machine",
					CreatedAt:    time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					ExpiresAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					Revoked:      false,
				},
				Issuer:            "CN=Test CA",
				Subject:           "CN=detailed.example.com",
				KeyAlgorithm:      "RSA",
				KeySize:           2048,
				FingerprintSHA1:   "AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD",
				FingerprintSHA256: "AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD",
				Usage:             []string{"Digital Signature", "Key Encipherment"},
				PEM:               "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
			},
			expectError: false,
		},
		{
			name: "detailed certificate with minimal fields",
			detailedCertificate: DetailedCertificate{
				Certificate: Certificate{
					ID:           "minimal",
					SerialNumber: "",
					CommonName:   "minimal.example.com",
					Sans:         []string{},
					CertType:     "unknown",
					CreatedAt:    time.Time{},
					ExpiresAt:    time.Time{},
					Revoked:      false,
				},
				Issuer:            "",
				Subject:           "",
				KeyAlgorithm:      "",
				KeySize:           0,
				FingerprintSHA1:   "",
				FingerprintSHA256: "",
				Usage:             []string{},
				PEM:               "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			data, err := json.Marshal(tt.detailedCertificate)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, data)

			// Test JSON unmarshaling
			var unmarshaled DetailedCertificate
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tt.detailedCertificate, unmarshaled)
		})
	}
}

func TestDetailedCertificate_Inheritance(t *testing.T) {
	baseCert := Certificate{
		ID:         "inherit-test",
		CommonName: "inherit.example.com",
		Sans:       []string{"inherit.example.com"},
		CreatedAt:  time.Date(2023, 3, 15, 10, 30, 0, 0, time.UTC),
		ExpiresAt:  time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
		Revoked:    false,
	}

	baseCert.SerialNumber = "1111222233334444"
	detailedCert := DetailedCertificate{
		Certificate: baseCert,
		Issuer:      "CN=Test Issuer",
		Subject:     "CN=inherit.example.com",
	}

	// Verify that the Certificate fields are accessible through DetailedCertificate
	assert.Equal(t, baseCert.ID, detailedCert.ID)
	assert.Equal(t, baseCert.CommonName, detailedCert.CommonName)
	assert.Equal(t, baseCert.Sans, detailedCert.Sans)
	assert.Equal(t, baseCert.CreatedAt, detailedCert.CreatedAt)
	assert.Equal(t, baseCert.ExpiresAt, detailedCert.ExpiresAt)
	assert.Equal(t, baseCert.Revoked, detailedCert.Revoked)

	// Verify DetailedCertificate specific fields
	assert.Equal(t, "1111222233334444", detailedCert.SerialNumber)
	assert.Equal(t, "CN=Test Issuer", detailedCert.Issuer)
	assert.Equal(t, "CN=inherit.example.com", detailedCert.Subject)
}

func TestPEMResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name        string
		pemResponse PEMResponse
		expectError bool
	}{
		{
			name: "valid PEM response",
			pemResponse: PEMResponse{
				SerialNumber: "1234567890ABCDEF",
				PEM:          "-----BEGIN CERTIFICATE-----\nMIIFazCCBFOgAwIBAgISA2Qx2V2...\n-----END CERTIFICATE-----",
			},
			expectError: false,
		},
		{
			name: "empty PEM response",
			pemResponse: PEMResponse{
				SerialNumber: "",
				PEM:          "",
			},
			expectError: false,
		},
		{
			name: "PEM response with only serial",
			pemResponse: PEMResponse{
				SerialNumber: "ABCDEF1234567890",
				PEM:          "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			data, err := json.Marshal(tt.pemResponse)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, data)

			// Test JSON unmarshaling
			var unmarshaled PEMResponse
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tt.pemResponse, unmarshaled)
		})
	}
}

func TestCertificateFieldTypes(t *testing.T) {
	cert := Certificate{
		ID:         "test-id",
		CommonName: "test.example.com",
		Sans:       []string{"test.example.com", "www.test.example.com"},
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(365 * 24 * time.Hour),
		Revoked:    false,
	}

	// Verify field types
	assert.IsType(t, "", cert.ID)
	assert.IsType(t, "", cert.CommonName)
	assert.IsType(t, []string{}, cert.Sans)
	assert.IsType(t, time.Time{}, cert.CreatedAt)
	assert.IsType(t, time.Time{}, cert.ExpiresAt)
	assert.IsType(t, false, cert.Revoked)
}

func TestDetailedCertificateFieldTypes(t *testing.T) {
	detailedCert := DetailedCertificate{
		Certificate: Certificate{
			ID:           "test-id",
			SerialNumber: "1234567890",
			CommonName:   "test.example.com",
			Sans:         []string{"test.example.com"},
			CreatedAt:    time.Now(),
			ExpiresAt:    time.Now().Add(365 * 24 * time.Hour),
			Revoked:      false,
		},
		Issuer:            "CN=Test CA",
		Subject:           "CN=test.example.com",
		KeyAlgorithm:      "RSA",
		KeySize:           2048,
		FingerprintSHA1:   "AA:BB:CC:DD:EE:FF",
		FingerprintSHA256: "AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD",
		Usage:             []string{"Digital Signature", "Key Encipherment"},
		PEM:               "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
	}

	// Verify field types
	assert.IsType(t, "", detailedCert.SerialNumber)
	assert.IsType(t, "", detailedCert.Issuer)
	assert.IsType(t, "", detailedCert.Subject)
	assert.IsType(t, "", detailedCert.KeyAlgorithm)
	assert.IsType(t, 0, detailedCert.KeySize)
	assert.IsType(t, "", detailedCert.FingerprintSHA1)
	assert.IsType(t, "", detailedCert.FingerprintSHA256)
	assert.IsType(t, []string{}, detailedCert.Usage)
	assert.IsType(t, "", detailedCert.PEM)
}

func TestPEMResponseFieldTypes(t *testing.T) {
	pemResp := PEMResponse{
		SerialNumber: "1234567890ABCDEF",
		PEM:          "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
	}

	// Verify field types
	assert.IsType(t, "", pemResp.SerialNumber)
	assert.IsType(t, "", pemResp.PEM)
}

func TestCertificate_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		cert     Certificate
		expected bool
	}{
		{
			name: "expired certificate",
			cert: Certificate{
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			},
			expected: true,
		},
		{
			name: "valid certificate",
			cert: Certificate{
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			expected: false,
		},
		{
			name: "certificate expiring now",
			cert: Certificate{
				ExpiresAt: time.Now().Add(1 * time.Second), // Add 1 second to account for timing
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.cert.IsExpired())
		})
	}
}

func TestCertificate_DaysUntilExpiry(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		cert     Certificate
		expected int
	}{
		{
			name: "expired certificate",
			cert: Certificate{
				ExpiresAt: now.Add(-24 * time.Hour),
			},
			expected: -1,
		},
		{
			name: "expires tomorrow",
			cert: Certificate{
				ExpiresAt: now.Add(24 * time.Hour),
			},
			expected: 1,
		},
		{
			name: "expires in 10 days",
			cert: Certificate{
				ExpiresAt: now.Add(10 * 24 * time.Hour),
			},
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cert.DaysUntilExpiry()
			// Allow for 1 day difference due to timing
			assert.InDelta(t, tt.expected, result, 1)
		})
	}
}

func TestCertificate_IsValidAt(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	tests := []struct {
		name     string
		cert     Certificate
		testTime time.Time
		expected bool
	}{
		{
			name: "valid certificate at current time",
			cert: Certificate{
				CreatedAt: yesterday,
				ExpiresAt: tomorrow,
				Revoked:   false,
			},
			testTime: now,
			expected: true,
		},
		{
			name: "revoked certificate",
			cert: Certificate{
				CreatedAt: yesterday,
				ExpiresAt: tomorrow,
				Revoked:   true,
			},
			testTime: now,
			expected: false,
		},
		{
			name: "certificate not yet valid",
			cert: Certificate{
				CreatedAt: tomorrow,
				ExpiresAt: tomorrow.Add(365 * 24 * time.Hour),
				Revoked:   false,
			},
			testTime: now,
			expected: false,
		},
		{
			name: "expired certificate",
			cert: Certificate{
				CreatedAt: yesterday.Add(-365 * 24 * time.Hour),
				ExpiresAt: yesterday,
				Revoked:   false,
			},
			testTime: now,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.cert.IsValidAt(tt.testTime))
		})
	}
}

func TestCertificate_HasSubject(t *testing.T) {
	tests := []struct {
		name     string
		cert     Certificate
		subject  string
		expected bool
	}{
		{
			name: "matches common name exactly",
			cert: Certificate{
				CommonName: "test.example.com",
				Sans:       []string{},
			},
			subject:  "test.example.com",
			expected: true,
		},
		{
			name: "matches common name case insensitive",
			cert: Certificate{
				CommonName: "Test.Example.COM",
				Sans:       []string{},
			},
			subject:  "test.example.com",
			expected: true,
		},
		{
			name: "matches SAN",
			cert: Certificate{
				CommonName: "primary.example.com",
				Sans:       []string{"san.example.com", "another.example.com"},
			},
			subject:  "san.example.com",
			expected: true,
		},
		{
			name: "matches SAN case insensitive",
			cert: Certificate{
				CommonName: "primary.example.com",
				Sans:       []string{"SAN.Example.COM"},
			},
			subject:  "san.example.com",
			expected: true,
		},
		{
			name: "no match",
			cert: Certificate{
				CommonName: "primary.example.com",
				Sans:       []string{"san.example.com"},
			},
			subject:  "nomatch.example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.cert.HasSubject(tt.subject))
		})
	}
}

func TestCertificate_GetStatus(t *testing.T) {
	tests := []struct {
		name     string
		cert     Certificate
		expected string
	}{
		{
			name: "revoked certificate",
			cert: Certificate{
				Revoked:   true,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			expected: "revoked",
		},
		{
			name: "expired certificate",
			cert: Certificate{
				Revoked:   false,
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			},
			expected: "expired",
		},
		{
			name: "valid certificate",
			cert: Certificate{
				Revoked:   false,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			expected: "valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.cert.GetStatus())
		})
	}
}

func TestInferCertType(t *testing.T) {
	tests := []struct {
		name     string
		usages   []x509.ExtKeyUsage
		expected string
	}{
		{name: "server auth is machine", usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, expected: "machine"},
		{name: "client auth is user", usages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, expected: "user"},
		{name: "server and client auth is both", usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}, expected: "both"},
		{name: "missing usage is unknown", usages: nil, expected: "unknown"},
		{name: "other usage is unknown", usages: []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning}, expected: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			certificate := &x509.Certificate{ExtKeyUsage: tt.usages}
			assert.Equal(t, tt.expected, InferCertType(certificate))
		})
	}
	assert.Equal(t, "unknown", InferCertType(nil))
}
