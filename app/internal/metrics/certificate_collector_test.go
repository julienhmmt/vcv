package metrics

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"vcv/internal/certs"
	"vcv/internal/config"
	"vcv/internal/vault"
)

func TestCollector_ErrorStopsCollection(t *testing.T) {
	t.Setenv("VCV_METRICS_PER_CERTIFICATE", "false")
	mockVault := new(vault.MockClient)
	mockVault.On("ListCertificates", mock.Anything).Return([]certs.Certificate{}, assert.AnError)
	mockVault.On("CheckConnection", mock.Anything).Return(assert.AnError)

	registry := prometheus.NewRegistry()
	metricsConfig := config.MetricsConfig{PerCertificate: false, EnhancedMetrics: false}
	collector := NewCertificateCollectorWithVaults(mockVault, map[string]vault.Client{}, config.ExpirationThresholds{Critical: 7, Warning: 30}, metricsConfig, []config.VaultInstance{})
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
	t.Setenv("VCV_METRICS_ENHANCED", "true")
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	certsList := []certs.Certificate{
		{ID: "pki:active-soon", CommonName: "soon", ExpiresAt: now.Add(10 * 24 * time.Hour), CreatedAt: now.Add(-20 * 24 * time.Hour), Revoked: false},
		{ID: "pki:active-later", CommonName: "later", ExpiresAt: now.Add(90 * 24 * time.Hour), CreatedAt: now.Add(-10 * 24 * time.Hour), Revoked: false},
		{ID: "pki:revoked", CommonName: "rev", ExpiresAt: now.Add(20 * 24 * time.Hour), CreatedAt: now.Add(-30 * 24 * time.Hour), Revoked: true},
		{ID: "pki:expired", CommonName: "old", ExpiresAt: now.Add(-24 * time.Hour), CreatedAt: now.Add(-100 * 24 * time.Hour), Revoked: false},
	}

	mockVault := new(vault.MockClient)
	mockVault.On("ListCertificates", mock.Anything).Return(certsList, nil)
	mockVault.On("CheckConnection", mock.Anything).Return(nil)
	vaultInstances := []config.VaultInstance{{ID: "vault-a", PKIMounts: []string{"pki"}}}
	clientsByVault := map[string]vault.Client{"vault-a": mockVault}
	multiClient := vault.NewMultiClient(vaultInstances, clientsByVault, nil)
	statusClients := map[string]vault.Client{"vault-a": mockVault}

	registry := prometheus.NewRegistry()
	metricsConfig := config.MetricsConfig{PerCertificate: true, EnhancedMetrics: true}
	rawCollector := NewCertificateCollectorWithVaults(multiClient, statusClients, config.ExpirationThresholds{Critical: 7, Warning: 30}, metricsConfig, vaultInstances)
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
	assertGauge(t, registry, "vcv_expiration_threshold_critical_days", nil, 7.0)
	assertGauge(t, registry, "vcv_expiration_threshold_warning_days", nil, 30.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "0-7d"}, 0.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "7-30d"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "30-90d"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "90d+"}, 0.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "expired"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "revoked"}, 1.0)

	// Per-certificate expiry timestamp for the "soon" cert
	assertGauge(t, registry, "vcv_certificate_expiry_timestamp_seconds", map[string]string{
		"certificate_id": "vault-a|pki:active-soon",
		"common_name":    "soon",
		"status":         "valid",
		"vault_id":       "vault-a",
		"pki":            "pki",
	}, float64(now.Add(10*24*time.Hour).Unix()))
	assertGauge(t, registry, "vcv_certificate_days_until_expiry", map[string]string{
		"certificate_id": "vault-a|pki:active-soon",
		"common_name":    "soon",
		"status":         "valid",
		"vault_id":       "vault-a",
		"pki":            "pki",
	}, 10.0)

	mockVault.AssertExpectations(t)
}

func TestCollector_EnhancedMetrics(t *testing.T) {
	t.Setenv("VCV_METRICS_PER_CERTIFICATE", "false")
	t.Setenv("VCV_METRICS_ENHANCED", "true")
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	certsList := []certs.Certificate{
		{ID: "pki:cert1", CommonName: "cert1", ExpiresAt: now.Add(5 * 24 * time.Hour), CreatedAt: now.Add(-25 * 24 * time.Hour), Revoked: false},
		{ID: "pki:cert2", CommonName: "cert2", ExpiresAt: now.Add(15 * 24 * time.Hour), CreatedAt: now.Add(-15 * 24 * time.Hour), Revoked: false},
		{ID: "pki:cert3", CommonName: "cert3", ExpiresAt: now.Add(45 * 24 * time.Hour), CreatedAt: now.Add(-45 * 24 * time.Hour), Revoked: false},
		{ID: "pki:cert4", CommonName: "cert4", ExpiresAt: now.Add(120 * 24 * time.Hour), CreatedAt: now.Add(-10 * 24 * time.Hour), Revoked: false},
		{ID: "pki:cert5", CommonName: "cert5", ExpiresAt: now.Add(-10 * 24 * time.Hour), CreatedAt: now.Add(-100 * 24 * time.Hour), Revoked: false},
		{ID: "pki:cert6", CommonName: "cert6", ExpiresAt: now.Add(30 * 24 * time.Hour), CreatedAt: now.Add(-30 * 24 * time.Hour), Revoked: true},
	}

	mockVault := new(vault.MockClient)
	mockVault.On("ListCertificates", mock.Anything).Return(certsList, nil)
	mockVault.On("CheckConnection", mock.Anything).Return(nil)
	vaultInstances := []config.VaultInstance{{ID: "vault-1", PKIMounts: []string{"pki"}}}
	clientsByVault := map[string]vault.Client{"vault-1": mockVault}
	multiClient := vault.NewMultiClient(vaultInstances, clientsByVault, nil)
	statusClients := map[string]vault.Client{"vault-1": mockVault}

	registry := prometheus.NewRegistry()
	metricsConfig := config.MetricsConfig{PerCertificate: false, EnhancedMetrics: true}
	rawCollector := NewCertificateCollectorWithVaults(multiClient, statusClients, config.ExpirationThresholds{Critical: 7, Warning: 30}, metricsConfig, vaultInstances)
	collector, ok := rawCollector.(*certificateCollector)
	require.True(t, ok)
	collector.now = func() time.Time { return now }
	require.NoError(t, registry.Register(collector))

	totalMetrics := testutil.CollectAndCount(collector)
	assert.GreaterOrEqual(t, totalMetrics, 10)

	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "vault-1", "pki": "pki", "bucket": "0-7d"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "vault-1", "pki": "pki", "bucket": "7-30d"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "vault-1", "pki": "pki", "bucket": "30-90d"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "vault-1", "pki": "pki", "bucket": "90d+"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "vault-1", "pki": "pki", "bucket": "expired"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "vault-1", "pki": "pki", "bucket": "revoked"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "0-7d"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "7-30d"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "30-90d"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "90d+"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "expired"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "revoked"}, 1.0)

	mockVault.AssertExpectations(t)
}

func TestCollector_ThresholdMetrics(t *testing.T) {
	t.Setenv("VCV_METRICS_PER_CERTIFICATE", "false")
	t.Setenv("VCV_METRICS_ENHANCED", "false")
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	certsList := []certs.Certificate{
		{ID: "pki:cert1", CommonName: "cert1", ExpiresAt: now.Add(1 * 24 * time.Hour), Revoked: false},
		{ID: "pki:cert2", CommonName: "cert2", ExpiresAt: now.Add(5 * 24 * time.Hour), Revoked: false},
		{ID: "pki:cert3", CommonName: "cert3", ExpiresAt: now.Add(8 * 24 * time.Hour), Revoked: false},
	}

	mockVault := new(vault.MockClient)
	mockVault.On("ListCertificates", mock.Anything).Return(certsList, nil)
	mockVault.On("CheckConnection", mock.Anything).Return(nil)

	registry := prometheus.NewRegistry()
	metricsConfig := config.MetricsConfig{PerCertificate: false, EnhancedMetrics: false}
	collector := NewCertificateCollector(mockVault, map[string]vault.Client{}, config.ExpirationThresholds{Critical: 2, Warning: 10}, metricsConfig)
	typed, ok := collector.(*certificateCollector)
	require.True(t, ok)
	typed.now = func() time.Time { return now }
	require.NoError(t, registry.Register(collector))

	testutil.CollectAndCount(collector)

	assertGauge(t, registry, "vcv_expiration_threshold_critical_days", nil, 2.0)
	assertGauge(t, registry, "vcv_expiration_threshold_warning_days", nil, 10.0)
	assertGauge(t, registry, "vcv_certificates_expiring_soon_count", map[string]string{"vault_id": "__all__", "pki": "__all__", "level": "critical"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiring_soon_count", map[string]string{"vault_id": "__all__", "pki": "__all__", "level": "warning"}, 3.0)

	mockVault.AssertExpectations(t)
}

func TestCollector_ZeroExpiresAtExcludedFromBuckets(t *testing.T) {
	t.Setenv("VCV_METRICS_PER_CERTIFICATE", "false")
	t.Setenv("VCV_METRICS_ENHANCED", "true")
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	certsList := []certs.Certificate{
		{ID: "pki:cert1", CommonName: "cert1", ExpiresAt: now.Add(5 * 24 * time.Hour), Revoked: false},
		{ID: "pki:cert2", CommonName: "cert2", ExpiresAt: time.Time{}, Revoked: false},
		{ID: "pki:cert3", CommonName: "cert3", ExpiresAt: time.Time{}, Revoked: false},
	}

	mockVault := new(vault.MockClient)
	mockVault.On("ListCertificates", mock.Anything).Return(certsList, nil)
	mockVault.On("CheckConnection", mock.Anything).Return(nil)

	registry := prometheus.NewRegistry()
	metricsConfig := config.MetricsConfig{PerCertificate: false, EnhancedMetrics: true}
	collector := NewCertificateCollector(mockVault, map[string]vault.Client{}, config.ExpirationThresholds{Critical: 7, Warning: 30}, metricsConfig)
	typed, ok := collector.(*certificateCollector)
	require.True(t, ok)
	typed.now = func() time.Time { return now }
	require.NoError(t, registry.Register(collector))

	testutil.CollectAndCount(collector)

	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "0-7d"}, 1.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "7-30d"}, 0.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "30-90d"}, 0.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "90d+"}, 0.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "expired"}, 0.0)
	assertGauge(t, registry, "vcv_certificates_expiry_bucket", map[string]string{"vault_id": "__all__", "pki": "__all__", "bucket": "revoked"}, 0.0)
	assertGauge(t, registry, "vcv_certificates_total", map[string]string{"vault_id": "__all__", "pki": "__all__", "status": "valid"}, 3.0)

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

	for _, f := range families {
		if f.GetName() != name {
			continue
		}
		// Convert family to text format
		var buf bytes.Buffer
		encoder := expfmt.NewEncoder(&buf, expfmt.NewFormat(expfmt.TypeTextPlain))
		if err := encoder.Encode(f); err != nil {
			return 0, err
		}
		// Parse the text format to find matching labels
		value, err := parseGaugeValue(buf.String(), labels)
		if err == nil {
			return value, nil
		}
	}
	return 0, fmt.Errorf("metric %s not found", name)
}

// parseGaugeValue extracts gauge value from exposition format text for matching labels.
// Example format: vcv_certificates_total{vault_id="__all__",pki="__all__",status="valid"} 2
func parseGaugeValue(output string, labels map[string]string) (float64, error) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Check if all labels match
		if !matchLabelsInLine(line, labels) {
			continue
		}
		// Extract value after the last quote/brace
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			var value float64
			if _, err := fmt.Sscanf(parts[len(parts)-1], "%f", &value); err == nil {
				return value, nil
			}
		}
	}
	return 0, fmt.Errorf("metric not found")
}

// matchLabelsInLine checks if a Prometheus exposition line matches all expected labels.
func matchLabelsInLine(line string, labels map[string]string) bool {
	if len(labels) == 0 {
		return true
	}
	// Extract label section from braces if present
	start := strings.Index(line, "{")
	end := strings.Index(line, "}")
	if start == -1 || end == -1 {
		// No labels in line, match if we expect no labels
		return len(labels) == 0
	}
	labelSection := line[start+1 : end]

	// Parse each label=value pair
	found := make(map[string]string)
	pairs := strings.Split(labelSection, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.Trim(kv[1], `"`)
		found[key] = val
	}

	// Check all expected labels match
	for key, expectedVal := range labels {
		if found[key] != expectedVal {
			return false
		}
	}
	return true
}
