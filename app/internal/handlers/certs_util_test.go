package handlers

import (
	"testing"

	"vcv/internal/certs"
)

func TestFilterCertificatesByMounts(t *testing.T) {
	testCertificates := []certs.Certificate{
		{ID: "pki:01", CommonName: "test1.example.com"},
		{ID: "pki:02", CommonName: "test2.example.com"},
		{ID: "custom-pki:01", CommonName: "custom.example.com"},
		{ID: "other:01", CommonName: "other.example.com"},
	}

	tests := []struct {
		name           string
		certificates   []certs.Certificate
		selectedMounts []string
		expected       []certs.Certificate
	}{
		{
			name:           "nil selected mounts returns all",
			certificates:   testCertificates,
			selectedMounts: nil,
			expected:       testCertificates,
		},
		{
			name:           "empty selected mounts returns empty",
			certificates:   testCertificates,
			selectedMounts: []string{},
			expected:       []certs.Certificate{},
		},
		{
			name:           "filter by single mount",
			certificates:   testCertificates,
			selectedMounts: []string{"pki"},
			expected: []certs.Certificate{
				{ID: "pki:01", CommonName: "test1.example.com"},
				{ID: "pki:02", CommonName: "test2.example.com"},
			},
		},
		{
			name:           "filter by multiple mounts",
			certificates:   testCertificates,
			selectedMounts: []string{"pki", "custom-pki"},
			expected: []certs.Certificate{
				{ID: "pki:01", CommonName: "test1.example.com"},
				{ID: "pki:02", CommonName: "test2.example.com"},
				{ID: "custom-pki:01", CommonName: "custom.example.com"},
			},
		},
		{
			name:           "filter by non-existent mount",
			certificates:   testCertificates,
			selectedMounts: []string{"nonexistent"},
			expected:       []certs.Certificate{},
		},
		{
			name:           "certificate without colon",
			certificates:   []certs.Certificate{{ID: "invalid", CommonName: "invalid.example.com"}},
			selectedMounts: []string{"pki"},
			expected:       []certs.Certificate{},
		},
		{
			name:           "empty certificate list",
			certificates:   []certs.Certificate{},
			selectedMounts: []string{"pki"},
			expected:       []certs.Certificate{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterCertificatesByMounts(tt.certificates, tt.selectedMounts)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d certificates, got %d", len(tt.expected), len(result))
			}

			for i := range result {
				if result[i].ID != tt.expected[i].ID {
					t.Errorf("expected certificate ID %q at index %d, got %q", tt.expected[i].ID, i, result[i].ID)
				}
			}
		})
	}
}

func TestBuildPEMDownloadFilename(t *testing.T) {
	tests := []struct {
		name     string
		serial   string
		expected string
	}{
		{
			name:     "normal serial number",
			serial:   "01:23:45:67",
			expected: "certificate-01-23-45-67.pem",
		},
		{
			name:     "serial with slashes",
			serial:   "01/23/45",
			expected: "certificate-01-23-45.pem",
		},
		{
			name:     "serial with backslashes",
			serial:   "01\\23\\45",
			expected: "certificate-01-23-45.pem",
		},
		{
			name:     "serial with double dots",
			serial:   "01..45",
			expected: "certificate-01-45.pem",
		},
		{
			name:     "empty serial",
			serial:   "",
			expected: "certificate.pem",
		},
		{
			name:     "only special characters",
			serial:   ":/\\..",
			expected: "certificate-----.pem",
		},
		{
			name:     "mixed special characters",
			serial:   "01:23/45\\67..89",
			expected: "certificate-01-23-45-67-89.pem",
		},
		{
			name:     "single serial number",
			serial:   "01",
			expected: "certificate-01.pem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPEMDownloadFilename(tt.serial)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
