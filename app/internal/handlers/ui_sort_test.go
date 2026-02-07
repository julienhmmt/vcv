package handlers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"vcv/config"
	"vcv/internal/certs"
)

func TestSortCertificates_ByCommonName_Asc(t *testing.T) {
	items := []certs.Certificate{
		{CommonName: "Bravo"},
		{CommonName: "Alpha"},
		{CommonName: "Charlie"},
	}
	sorted := sortCertificates(items, "commonName", "asc")
	assert.Equal(t, "Alpha", sorted[0].CommonName)
	assert.Equal(t, "Bravo", sorted[1].CommonName)
	assert.Equal(t, "Charlie", sorted[2].CommonName)
}

func TestSortCertificates_ByCommonName_Desc(t *testing.T) {
	items := []certs.Certificate{
		{CommonName: "Alpha"},
		{CommonName: "Charlie"},
		{CommonName: "Bravo"},
	}
	sorted := sortCertificates(items, "commonName", "desc")
	assert.Equal(t, "Charlie", sorted[0].CommonName)
	assert.Equal(t, "Bravo", sorted[1].CommonName)
	assert.Equal(t, "Alpha", sorted[2].CommonName)
}

func TestSortCertificates_ByCreatedAt_Asc(t *testing.T) {
	now := time.Now()
	items := []certs.Certificate{
		{CommonName: "B", CreatedAt: now.Add(2 * time.Hour)},
		{CommonName: "A", CreatedAt: now},
		{CommonName: "C", CreatedAt: now.Add(1 * time.Hour)},
	}
	sorted := sortCertificates(items, "createdAt", "asc")
	assert.Equal(t, "A", sorted[0].CommonName)
	assert.Equal(t, "C", sorted[1].CommonName)
	assert.Equal(t, "B", sorted[2].CommonName)
}

func TestSortCertificates_ByCreatedAt_Desc(t *testing.T) {
	now := time.Now()
	items := []certs.Certificate{
		{CommonName: "A", CreatedAt: now},
		{CommonName: "B", CreatedAt: now.Add(2 * time.Hour)},
	}
	sorted := sortCertificates(items, "createdAt", "desc")
	assert.Equal(t, "B", sorted[0].CommonName)
	assert.Equal(t, "A", sorted[1].CommonName)
}

func TestSortCertificates_ByExpiresAt_Asc(t *testing.T) {
	now := time.Now()
	items := []certs.Certificate{
		{CommonName: "B", ExpiresAt: now.Add(2 * time.Hour)},
		{CommonName: "A", ExpiresAt: now},
	}
	sorted := sortCertificates(items, "expiresAt", "asc")
	assert.Equal(t, "A", sorted[0].CommonName)
	assert.Equal(t, "B", sorted[1].CommonName)
}

func TestSortCertificates_ByExpiresAt_Desc(t *testing.T) {
	now := time.Now()
	items := []certs.Certificate{
		{CommonName: "A", ExpiresAt: now},
		{CommonName: "B", ExpiresAt: now.Add(2 * time.Hour)},
	}
	sorted := sortCertificates(items, "expiresAt", "desc")
	assert.Equal(t, "B", sorted[0].CommonName)
	assert.Equal(t, "A", sorted[1].CommonName)
}

func TestSortCertificates_ByVault_Asc(t *testing.T) {
	items := []certs.Certificate{
		{ID: "vault-b|pki:cert1", CommonName: "C1"},
		{ID: "vault-a|pki:cert2", CommonName: "C2"},
	}
	sorted := sortCertificates(items, "vault", "asc")
	assert.Equal(t, "C2", sorted[0].CommonName)
	assert.Equal(t, "C1", sorted[1].CommonName)
}

func TestSortCertificates_ByVault_Desc(t *testing.T) {
	items := []certs.Certificate{
		{ID: "vault-a|pki:cert1", CommonName: "C1"},
		{ID: "vault-b|pki:cert2", CommonName: "C2"},
	}
	sorted := sortCertificates(items, "vault", "desc")
	assert.Equal(t, "C2", sorted[0].CommonName)
	assert.Equal(t, "C1", sorted[1].CommonName)
}

func TestSortCertificates_ByVault_SameVault_SortsByMount(t *testing.T) {
	items := []certs.Certificate{
		{ID: "vault-a|pki-b:cert1", CommonName: "C1"},
		{ID: "vault-a|pki-a:cert2", CommonName: "C2"},
	}
	sorted := sortCertificates(items, "vault", "asc")
	assert.Equal(t, "C2", sorted[0].CommonName)
	assert.Equal(t, "C1", sorted[1].CommonName)
}

func TestSortCertificates_ByPKI_Asc(t *testing.T) {
	items := []certs.Certificate{
		{ID: "vault-a|pki-b:cert1", CommonName: "C1"},
		{ID: "vault-a|pki-a:cert2", CommonName: "C2"},
	}
	sorted := sortCertificates(items, "pki", "asc")
	assert.Equal(t, "C2", sorted[0].CommonName)
	assert.Equal(t, "C1", sorted[1].CommonName)
}

func TestSortCertificates_ByPKI_Desc(t *testing.T) {
	items := []certs.Certificate{
		{ID: "vault-a|pki-a:cert1", CommonName: "C1"},
		{ID: "vault-a|pki-b:cert2", CommonName: "C2"},
	}
	sorted := sortCertificates(items, "pki", "desc")
	assert.Equal(t, "C2", sorted[0].CommonName)
	assert.Equal(t, "C1", sorted[1].CommonName)
}

func TestSortCertificates_ByPKI_SameMount_SortsByVault(t *testing.T) {
	items := []certs.Certificate{
		{ID: "vault-b|pki:cert1", CommonName: "C1"},
		{ID: "vault-a|pki:cert2", CommonName: "C2"},
	}
	sorted := sortCertificates(items, "pki", "asc")
	assert.Equal(t, "C2", sorted[0].CommonName)
	assert.Equal(t, "C1", sorted[1].CommonName)
}

func TestCountUniqueMounts(t *testing.T) {
	tests := []struct {
		name      string
		instances []config.VaultInstance
		expected  int
	}{
		{name: "empty", instances: nil, expected: 0},
		{name: "single", instances: []config.VaultInstance{{PKIMounts: []string{"pki"}}}, expected: 1},
		{name: "duplicates", instances: []config.VaultInstance{
			{PKIMounts: []string{"pki", "pki"}},
		}, expected: 1},
		{name: "multiple_vaults", instances: []config.VaultInstance{
			{PKIMounts: []string{"pki", "pki_dev"}},
			{PKIMounts: []string{"pki", "pki_stg"}},
		}, expected: 3},
		{name: "empty_mount_trimmed", instances: []config.VaultInstance{
			{PKIMounts: []string{"  ", "pki"}},
		}, expected: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countUniqueMounts(tt.instances)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestShouldShowVaultMount(t *testing.T) {
	assert.True(t, shouldShowVaultMount([]config.VaultInstance{{ID: "a"}, {ID: "b"}}))
	assert.True(t, shouldShowVaultMount([]config.VaultInstance{{PKIMounts: []string{"pki", "pki2"}}}))
	assert.False(t, shouldShowVaultMount([]config.VaultInstance{{PKIMounts: []string{"pki"}}}))
}
