package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vcv/internal/certs"
)

// pinnedMetricCertIDs drains ch and returns the certificate_id label value of every
// emitted metric family produced by emitPinnedCertificateMetrics.
func pinnedMetricCertIDs(t *testing.T, ch chan prometheus.Metric) []string {
	t.Helper()
	close(ch)
	var ids []string
	for m := range ch {
		var pb dto.Metric
		require.NoError(t, m.Write(&pb))
		for _, label := range pb.GetLabel() {
			if label.GetName() == "certificate_id" {
				ids = append(ids, label.GetValue())
			}
		}
	}
	return ids
}

func TestEmitPinnedCertificateMetrics_NoPinnedConfigured(t *testing.T) {
	collector := &certificateCollector{}
	ch := make(chan prometheus.Metric, 8)
	certificates := []certs.Certificate{{
		ID:         "vault-a|pki:aa",
		CommonName: "app.example.com",
		ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
	}}

	collector.emitPinnedCertificateMetrics(ch, certificates, time.Now())

	assert.Empty(t, pinnedMetricCertIDs(t, ch))
}

func TestEmitPinnedCertificateMetrics_ExactCommonNameMatch(t *testing.T) {
	collector := &certificateCollector{pinnedCertificates: []string{"root.example.com"}}
	ch := make(chan prometheus.Metric, 8)
	certificates := []certs.Certificate{
		{ID: "vault-a|pki:aa", CommonName: "root.example.com", ExpiresAt: time.Now().Add(30 * 24 * time.Hour)},
		{ID: "vault-a|pki:bb", CommonName: "other.example.com", ExpiresAt: time.Now().Add(30 * 24 * time.Hour)},
	}

	collector.emitPinnedCertificateMetrics(ch, certificates, time.Now())

	ids := pinnedMetricCertIDs(t, ch)
	assert.Contains(t, ids, "vault-a|pki:aa")
	assert.NotContains(t, ids, "vault-a|pki:bb")
}

func TestEmitPinnedCertificateMetrics_WildcardSANMatch(t *testing.T) {
	collector := &certificateCollector{pinnedCertificates: []string{"*.internal.example.com"}}
	ch := make(chan prometheus.Metric, 8)
	certificates := []certs.Certificate{
		{
			ID:         "vault-a|pki:cc",
			CommonName: "service.example.com",
			Sans:       []string{"service.example.com", "svc.internal.example.com"},
			ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
		},
		{
			ID:         "vault-a|pki:dd",
			CommonName: "unrelated.example.com",
			Sans:       []string{"unrelated.example.com"},
			ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
		},
	}

	collector.emitPinnedCertificateMetrics(ch, certificates, time.Now())

	ids := pinnedMetricCertIDs(t, ch)
	assert.Contains(t, ids, "vault-a|pki:cc")
	assert.NotContains(t, ids, "vault-a|pki:dd")
}

func TestEmitPinnedCertificateMetrics_CaseInsensitive(t *testing.T) {
	collector := &certificateCollector{pinnedCertificates: []string{"ROOT.EXAMPLE.COM"}}
	ch := make(chan prometheus.Metric, 8)
	certificates := []certs.Certificate{
		{ID: "vault-a|pki:aa", CommonName: "root.example.com", ExpiresAt: time.Now().Add(30 * 24 * time.Hour)},
	}

	collector.emitPinnedCertificateMetrics(ch, certificates, time.Now())

	assert.Contains(t, pinnedMetricCertIDs(t, ch), "vault-a|pki:aa")
}

func TestEmitPinnedCertificateMetrics_SkipsZeroExpiry(t *testing.T) {
	collector := &certificateCollector{pinnedCertificates: []string{"root.example.com"}}
	ch := make(chan prometheus.Metric, 8)
	certificates := []certs.Certificate{
		{ID: "vault-a|pki:aa", CommonName: "root.example.com"}, // ExpiresAt zero value
	}

	collector.emitPinnedCertificateMetrics(ch, certificates, time.Now())

	assert.Empty(t, pinnedMetricCertIDs(t, ch))
}
