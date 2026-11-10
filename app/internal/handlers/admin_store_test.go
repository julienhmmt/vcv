package handlers

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vcv/internal/config"
	vcverrors "vcv/internal/errors"
)

func TestAdminSettingsStore_SaveFallsBackToDirectWrite(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("read-only directory cannot block root")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	require.NoError(t, os.WriteFile(path, []byte("{}"), 0o600))
	require.NoError(t, os.Chmod(dir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	store := newAdminSettingsStore(path, config.EnvDev)
	err := store.save(config.SettingsFile{App: config.AppSettings{Env: "dev", Port: 52001}})
	require.NoError(t, err)

	data, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "52001")
	info, statErr := os.Stat(path)
	require.NoError(t, statErr)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestAdminSettingsStore_SaveRenameFailureWithoutFallbackReturnsError(t *testing.T) {
	dir := t.TempDir()
	// A directory at the target path makes rename fail with a non-fallback error.
	path := filepath.Join(dir, "settings.json")
	require.NoError(t, os.Mkdir(path, 0o755))

	store := newAdminSettingsStore(path, config.EnvDev)
	err := store.save(config.SettingsFile{App: config.AppSettings{Env: "dev", Port: 52001}})
	assert.Error(t, err)
}

func TestAdminSettingsStore_SaveValidationErrorLeavesFileUnchanged(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	original := []byte(`{"app":{"port":52000}}`)
	require.NoError(t, os.WriteFile(path, original, 0o600))

	store := newAdminSettingsStore(path, config.EnvDev)
	invalid := config.SettingsFile{}
	invalid.Notifications.WebhookURL = "ftp://example.com/hook"

	err := store.save(invalid)
	require.ErrorIs(t, err, vcverrors.ErrInvalidWebhookURL)

	data, readErr := os.ReadFile(path)
	require.NoError(t, readErr)
	assert.Equal(t, original, data)
}

func TestFallbackWriteSettings_WriteFailureReturnsOriginalError(t *testing.T) {
	missingDir := filepath.Join(t.TempDir(), "missing", "settings.json")
	err := fallbackWriteSettings(missingDir, []byte(`{}`), syscall.EACCES)
	assert.ErrorIs(t, err, syscall.EACCES)
}

func TestValidationErrorStatus(t *testing.T) {
	sentinels := []error{
		vcverrors.ErrInvalidAddress,
		vcverrors.ErrInvalidToken,
		vcverrors.ErrInvalidThreshold,
		vcverrors.ErrInvalidWebhookURL,
		vcverrors.ErrInvalidWebhookTarget,
		vcverrors.ErrVaultIDEmpty,
		vcverrors.ErrDuplicateVaultID,
	}
	for _, err := range sentinels {
		t.Run(err.Error(), func(t *testing.T) {
			assert.Equal(t, 400, validationErrorStatus(err))
		})
	}
	t.Run("unknown error maps to 500", func(t *testing.T) {
		assert.Equal(t, 500, validationErrorStatus(errors.New("x")))
	})
}

func TestAdminSettingsStore_SaveMkdirFailureReturnsError(t *testing.T) {
	// A regular file where the parent directory should be makes MkdirAll fail.
	blocker := filepath.Join(t.TempDir(), "blocker")
	require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o600))

	store := newAdminSettingsStore(filepath.Join(blocker, "settings.json"), config.EnvDev)
	assert.Error(t, store.save(config.SettingsFile{}))
}
