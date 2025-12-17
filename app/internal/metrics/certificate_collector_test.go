package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"vcv/config"
	"vcv/internal/certs"
	"vcv/internal/vault"
)

func TestCollector_ErrorStopsCollection(t *testing.T) {
	t.Setenv("VCV_METRICS_PER_CERTIFICATE", "false")
	mockVault := new(vault.MockClient)
	mockVault.On("ListCertificates", mock.Anything).Return([]certs.Certificate{}, assert.AnError)
	mockVault.On("CheckConnection", mock.Anything).Return(assert.AnError)

	registry := prometheus.NewRegistry()
	collector := NewCertificateCollector(mockVault, map[string]vault.Client{}, config.ExpirationThresholds{Critical: 7, Warning: 30})
	require.NoError(t, registry.Register(collector))

	metricsCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricsCount, 0)

	// Only last_scrape_success should be emitted with value 0
	value, err := gatherGauge(registry, "vcv_certificate_exporter_last_scrape_success", nil)
	require.NoError(t, err)
	assert.Equal(t, 0.0, value)

	mockVault.AssertExpectations(t)
}

func TestCollector_SuccessMetrics(t *testing.T) {
	t.Setenv("VCV_METRICS_PER_CERTIFICATE", "true")
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	certsList := []certs.Certificate{
		{ID: "pki:active-soon", CommonName: "soon", ExpiresAt: now.Add(10 * 24 * time.Hour), Revoked: false},
		{ID: "pki:active-later", CommonName: "later", ExpiresAt: now.Add(90 * 24 * time.Hour), Revoked: false},
		{ID: "pki:revoked", CommonName: "rev", ExpiresAt: now.Add(20 * 24 * time.Hour), Revoked: true},
		{ID: "pki:expired", CommonName: "old", ExpiresAt: now.Add(-24 * time.Hour), Revoked: false},
	}

	mockVault := new(vault.MockClient)
	mockVault.On("ListCertificates", mock.Anything).Return(certsList, nil)
	mockVault.On("CheckConnection", mock.Anything).Return(nil)
	vaultInstances := []config.VaultInstance{{ID: "vault-a", PKIMounts: []string{"pki"}}}
	clientsByVault := map[string]vault.Client{"vault-a": mockVault}
	multiClient := vault.NewMultiClient(vaultInstances, clientsByVault)
	statusClients := map[string]vault.Client{"vault-a": mockVault}

	registry := prometheus.NewRegistry()
	rawCollector := NewCertificateCollectorWithVaults(multiClient, statusClients, config.ExpirationThresholds{Critical: 7, Warning: 30}, vaultInstances)
	collector, ok := rawCollector.(*certificateCollector)
	require.True(t, ok)
	collector.now = func() time.Time { return now }
	require.NoError(t, registry.Register(collector))

	totalMetrics := testutil.CollectAndCount(collector)
	assert.GreaterOrEqual(t, totalMetrics, 5)

	assertGauge(t, registry, "vcv_certificate_exporter_last_scrape_success", nil, 1.0)
	assertGauge(t, registry, "vcv_vault_connected", map[string]string{"vault_id": "__all__"}, 1.0)
	assertGauge(t, registry, "vcv_vault_connected", map[string]string{"vault_id": "vault-a"}, 1.0)
	assertGauge(t, registry, "vcv_vault_list_certificates_success", map[string]string{"vault_id": "__all__"}, 1.0)
	assertGauge(t, registry, "vcv_vault_list_certificates_success", map[string]string{"vault_id": "vault-a"}, 1.0)
	assertGauge(t, registry, "vcv_vault_list_certificates_error", map[string]string{"vault_id": "__all__"}, 0.0)
	assertGauge(t, registry, "vcv_vault_list_certificates_error", map[string]string{"vault_id": "vault-a"}, 0.0)
	assertGauge(t, registry, "vcv_certificates_partial_scrape", map[string]string{"vault_id": "__all__"}, 0.0)
	assertGauge(t, registry, "vcv_vaults_configured", nil, 1.0)
	assertGauge(t, registry, "vcv_pki_mounts_configured", map[string]string{"vault_id": "vault-a"}, 1.0)
	assertGauge(t, registry, "vcv_pki_mounts_configured", map[string]string{"vault_id": "__all__"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expired_count", nil, 1.0)
	assertGauge(t, registry, "vcv_certificates_total", map[string]string{"vault_id": "__all__", "pki": "__all__", "status": "valid"}, 2.0)
	assertGauge(t, registry, "vcv_certificates_total", map[string]string{"vault_id": "__all__", "pki": "__all__", "status": "expired"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_total", map[string]string{"vault_id": "__all__", "pki": "__all__", "status": "revoked"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_total", map[string]string{"vault_id": "vault-a", "pki": "pki", "status": "valid"}, 2.0)
	assertGauge(t, registry, "vcv_certificates_total", map[string]string{"vault_id": "vault-a", "pki": "pki", "status": "expired"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_total", map[string]string{"vault_id": "vault-a", "pki": "pki", "status": "revoked"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiring_soon_count", map[string]string{"vault_id": "__all__", "pki": "__all__", "level": "warning"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiring_soon_count", map[string]string{"vault_id": "__all__", "pki": "__all__", "level": "critical"}, 0.0)
	assertGauge(t, registry, "vcv_certificates_expiring_soon_count", map[string]string{"vault_id": "vault-a", "pki": "pki", "level": "warning"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiring_soon_count", map[string]string{"vault_id": "vault-a", "pki": "pki", "level": "critical"}, 0.0)

	// Per-certificate expiry timestamp for the "soon" cert
	assertGauge(t, registry, "vcv_certificate_expiry_timestamp_seconds", map[string]string{
		"certificate_id": "vault-a|pki:active-soon",
		"common_name":    "soon",
		"status":         "valid",
		"vault_id":       "vault-a",
		"pki":            "pki",
	}, float64(now.Add(10*24*time.Hour).Unix()))

	mockVault.AssertExpectations(t)
}

func assertGauge(t *testing.T, registry *prometheus.Registry, name string, labels map[string]string, expected float64) {
	t.Helper()
	value, err := gatherGauge(registry, name, labels)
	require.NoError(t, err)
	assert.InDelta(t, expected, value, 0.0001)
}

func gatherGauge(registry *prometheus.Registry, name string, labels map[string]string) (float64, error) {
	families, err := registry.Gather()
	if err != nil {
		return 0, err
	}
	for _, mf := range families {
		if mf.GetName() != name {
			continue
		}
		for _, m := range mf.Metric {
			if !matchLabels(m, labels) {
				continue
			}
			return m.GetGauge().GetValue(), nil
		}
	}
	return 0, nil
}

func matchLabels(metric *dto.Metric, labels map[string]string) bool {
	if len(labels) == 0 {
		return true
	}
	for _, lp := range metric.Label {
		key := lp.GetName()
		val := lp.GetValue()
		if expected, ok := labels[key]; ok {
			if expected != val {
				return false
			}
		}
	}
	// Ensure no expected label is missing
	for expectedKey := range labels {
		found := false
		for _, lp := range metric.Label {
			if lp.GetName() == expectedKey {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
