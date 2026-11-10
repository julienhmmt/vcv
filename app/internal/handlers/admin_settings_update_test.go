package handlers

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vcv/internal/config"
	vcverrors "vcv/internal/errors"
)

func TestAdminSettingsStore_UpdateSerializesConcurrentWriters(t *testing.T) {
	store := newAdminSettingsStore(filepath.Join(t.TempDir(), "settings.json"), config.EnvDev)

	const writers = 20
	var wg sync.WaitGroup
	errs := make(chan error, writers)
	for i := range writers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := store.update(func(s *config.SettingsFile) error {
				s.Metrics.PinnedCertificates = append(s.Metrics.PinnedCertificates, fmt.Sprintf("cert-%d", i))
				return nil
			})
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	settings, err := store.load()
	require.NoError(t, err)
	assert.Len(t, settings.Metrics.PinnedCertificates, writers, "no concurrent update may be lost")
}

func TestAdminSettingsStore_UpdateErrorWritesNothing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	store := newAdminSettingsStore(path, config.EnvDev)

	_, err := store.update(func(*config.SettingsFile) error { return errVaultNotFound })
	require.ErrorIs(t, err, errVaultNotFound)

	_, err = store.update(func(s *config.SettingsFile) error {
		s.Notifications.WebhookURL = "not a url"
		return nil
	})
	require.ErrorIs(t, err, vcverrors.ErrInvalidWebhookURL)

	settings, err := store.load()
	require.NoError(t, err)
	assert.Empty(t, settings.Notifications.WebhookURL)
}

func fullyPopulatedSettings() config.SettingsFile {
	perCert := true
	enhanced := false
	return config.SettingsFile{
		Certificates: config.CertificateSettings{ExpirationThresholds: config.ExpirationThresholds{Critical: 7, Warning: 30}},
		Metrics: config.MetricsSettings{
			PerCertificate:     &perCert,
			EnhancedMetrics:    &enhanced,
			PinnedCertificates: []string{"pinned-a"},
		},
		CORS:          config.CORSSettings{AllowedOrigins: []string{"https://ui.example.com"}},
		Notifications: config.NotificationSettings{WebhookURL: "https://hooks.example.com/services/T/B/S"},
		Vaults:        []config.VaultInstance{{ID: "vault1", Address: "https://vault.example.com", Token: "s.token"}},
	}
}

func TestMergeAdminSettings_EmptyIncomingKeepsCurrent(t *testing.T) {
	current := fullyPopulatedSettings()

	merged := mergeAdminSettings(current, config.SettingsFile{})

	assert.Equal(t, current, merged)
}

func TestMergeAdminSettings_ExplicitEmptyClears(t *testing.T) {
	current := fullyPopulatedSettings()

	merged := mergeAdminSettings(current, config.SettingsFile{
		Vaults:  []config.VaultInstance{},
		Metrics: config.MetricsSettings{PinnedCertificates: []string{}},
		CORS:    config.CORSSettings{AllowedOrigins: []string{}},
	})

	assert.Empty(t, merged.Vaults, "explicit empty vaults list must clear")
	assert.Empty(t, merged.Metrics.PinnedCertificates, "explicit empty pinned list must clear")
	assert.Empty(t, merged.CORS.AllowedOrigins, "explicit empty origins list must clear")
	assert.Equal(t, current.Certificates, merged.Certificates, "absent thresholds keep current")
}
