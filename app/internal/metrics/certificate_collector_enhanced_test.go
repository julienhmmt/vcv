package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"vcv/internal/certs"
)

func TestExtractIssuerCN_UsesIssuerField(t *testing.T) {
	got := extractIssuerCN(certs.Certificate{CommonName: "app.example.com", IssuerCN: "Internal Intermediate CA"})
	assert.Equal(t, "Internal Intermediate CA", got)
	assert.Equal(t, "unknown", extractIssuerCN(certs.Certificate{CommonName: "app.example.com"}))
}

func TestExtractKeyInfo_UsesCertificateFields(t *testing.T) {
	algo, size := extractKeyInfo(certs.Certificate{KeyAlgorithm: "RSA", KeySize: 1024})
	assert.Equal(t, "RSA", algo)
	assert.Equal(t, "1024", size)
	assert.True(t, isWeakKey(algo, size))
	algo, size = extractKeyInfo(certs.Certificate{})
	assert.Equal(t, "unknown", algo)
	assert.Equal(t, "0", size)
}

func TestEmitKeyTypeMetrics_WeakRSA1024(t *testing.T) {
	collector := &certificateCollector{enhancedMetrics: true}
	ch := make(chan prometheus.Metric, 16)
	certificates := []certs.Certificate{{
		ID:           "vault-a|pki:aa:bb",
		CommonName:   "weak.example.com",
		KeyAlgorithm: "RSA",
		KeySize:      1024,
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour),
	}}
	collector.emitKeyTypeMetrics(ch, certificates)
	close(ch)
	metricCount := 0
	for range ch {
		metricCount++
	}
	assert.GreaterOrEqual(t, metricCount, 1)
	assert.True(t, isWeakKey("RSA", "1024"))
}

func TestEmitIssuerMetrics_UsesIssuerCN(t *testing.T) {
	collector := &certificateCollector{enhancedMetrics: true}
	ch := make(chan prometheus.Metric, 8)
	certificates := []certs.Certificate{{
		ID:         "vault-a|pki:aa:bb",
		CommonName: "app.example.com",
		IssuerCN:   "My Issuer CA",
		ExpiresAt:  time.Now().Add(30 * 24 * time.Hour),
	}}
	collector.emitIssuerMetrics(ch, certificates)
	close(ch)
	count := 0
	for range ch {
		count++
	}
	assert.GreaterOrEqual(t, count, 1)
}
