package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"vcv/internal/certs"
	"vcv/internal/vault"
)

const expirySoonWindowDays int = 30

var (
	cacheSizeDesc             = prometheus.NewDesc("vcv_cache_size", "Number of items currently cached", nil, nil)
	certificatesLastFetchDesc = prometheus.NewDesc("vcv_certificates_last_fetch_timestamp_seconds", "Timestamp of last successful certificates fetch", nil, nil)
	certificatesTotalDesc     = prometheus.NewDesc("vcv_certificates_total", "Total certificates grouped by status", []string{"status"}, nil)
	expiredCountDesc          = prometheus.NewDesc("vcv_certificates_expired_count", "Number of expired certificates", nil, nil)
	expiresInDesc             = prometheus.NewDesc("vcv_certificate_expires_in_seconds", "Seconds until certificate expiration (zero when expired)", []string{"serial_number", "common_name", "status"}, nil)
	expiresSoonCountDesc      = prometheus.NewDesc("vcv_certificates_expires_soon_count", "Number of certificates expiring soon within threshold window", nil, nil)
	expiresSoonDesc           = prometheus.NewDesc("vcv_certificate_expires_soon", "Certificate expires soon within threshold window (1=true,0=false)", []string{"serial_number", "common_name"}, nil)
	expiryTimestampDesc       = prometheus.NewDesc("vcv_certificate_expiry_timestamp_seconds", "Certificate expiration timestamp in seconds since epoch", []string{"serial_number", "common_name", "status"}, nil)
	lastScrapeSuccessDesc     = prometheus.NewDesc("vcv_certificate_exporter_last_scrape_success", "Whether the last scrape succeeded (1) or failed (0)", nil, nil)
	vaultConnectedDesc        = prometheus.NewDesc("vcv_vault_connected", "Vault connection status (1=connected,0=disconnected)", nil, nil)
)

type certificateCollector struct {
	vaultClient      vault.Client
	expirySoonWindow time.Duration
	now              func() time.Time
}

// NewCertificateCollector returns a Prometheus collector exposing certificate inventory and expiry status.
func NewCertificateCollector(vaultClient vault.Client) prometheus.Collector {
	return &certificateCollector{
		vaultClient:      vaultClient,
		expirySoonWindow: time.Duration(expirySoonWindowDays) * 24 * time.Hour,
		now:              time.Now,
	}
}

func (collector *certificateCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cacheSizeDesc
	ch <- certificatesLastFetchDesc
	ch <- certificatesTotalDesc
	ch <- expiredCountDesc
	ch <- expiresInDesc
	ch <- expiresSoonCountDesc
	ch <- expiresSoonDesc
	ch <- expiryTimestampDesc
	ch <- lastScrapeSuccessDesc
	ch <- vaultConnectedDesc
}

func (collector *certificateCollector) Collect(ch chan<- prometheus.Metric) {
	certificates, err := collector.listCertificates()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(lastScrapeSuccessDesc, prometheus.GaugeValue, 0)
		return
	}
	ch <- prometheus.MustNewConstMetric(lastScrapeSuccessDesc, prometheus.GaugeValue, 1)
	now := collector.now()

	// Check Vault connection
	vaultConnected := 1.0
	if err := collector.vaultClient.CheckConnection(context.Background()); err != nil {
		vaultConnected = 0.0
	}

	activeCount, revokedCount := collector.countStatuses(certificates)
	expiredCount := collector.countExpired(certificates, now)
	expiresSoonCount := collector.countExpiresSoon(certificates, now)
	cacheSize := collector.getCacheSize()

	ch <- prometheus.MustNewConstMetric(cacheSizeDesc, prometheus.GaugeValue, float64(cacheSize))
	ch <- prometheus.MustNewConstMetric(certificatesLastFetchDesc, prometheus.GaugeValue, float64(now.Unix()))
	ch <- prometheus.MustNewConstMetric(certificatesTotalDesc, prometheus.GaugeValue, float64(activeCount), "active")
	ch <- prometheus.MustNewConstMetric(certificatesTotalDesc, prometheus.GaugeValue, float64(revokedCount), "revoked")
	ch <- prometheus.MustNewConstMetric(expiredCountDesc, prometheus.GaugeValue, float64(expiredCount))
	ch <- prometheus.MustNewConstMetric(expiresSoonCountDesc, prometheus.GaugeValue, float64(expiresSoonCount))
	ch <- prometheus.MustNewConstMetric(vaultConnectedDesc, prometheus.GaugeValue, vaultConnected)
	collector.emitCertificateMetrics(ch, certificates, now)
}

func (collector *certificateCollector) listCertificates() ([]certs.Certificate, error) {
	return collector.vaultClient.ListCertificates(context.Background())
}

func (collector *certificateCollector) countStatuses(certificates []certs.Certificate) (int, int) {
	activeCount := 0
	revokedCount := 0
	for _, certificate := range certificates {
		if certificate.Revoked {
			revokedCount++
			continue
		}
		activeCount++
	}
	return activeCount, revokedCount
}

func (collector *certificateCollector) countExpired(certificates []certs.Certificate, now time.Time) int {
	count := 0
	for _, certificate := range certificates {
		if !certificate.Revoked && certificate.ExpiresAt.Before(now) {
			count++
		}
	}
	return count
}

func (collector *certificateCollector) countExpiresSoon(certificates []certs.Certificate, now time.Time) int {
	count := 0
	for _, certificate := range certificates {
		if collector.expiresSoonValue(certificate, now) == 1 {
			count++
		}
	}
	return count
}

func (collector *certificateCollector) getCacheSize() int {
	// Try to access cache via reflection or interface if available
	// For now, return 0 as cache size is not exposed by vault.Client interface
	return 0
}

func (collector *certificateCollector) emitCertificateMetrics(ch chan<- prometheus.Metric, certificates []certs.Certificate, now time.Time) {
	for _, certificate := range certificates {
		status := collector.statusLabel(certificate.Revoked)
		expiryTimestamp := float64(certificate.ExpiresAt.Unix())
		secondsToExpiry := certificate.ExpiresAt.Sub(now).Seconds()
		if secondsToExpiry < 0 {
			secondsToExpiry = 0
		}
		expiresSoon := collector.expiresSoonValue(certificate, now)
		ch <- prometheus.MustNewConstMetric(expiryTimestampDesc, prometheus.GaugeValue, expiryTimestamp, certificate.ID, certificate.CommonName, status)
		ch <- prometheus.MustNewConstMetric(expiresInDesc, prometheus.GaugeValue, secondsToExpiry, certificate.ID, certificate.CommonName, status)
		ch <- prometheus.MustNewConstMetric(expiresSoonDesc, prometheus.GaugeValue, expiresSoon, certificate.ID, certificate.CommonName)
	}
}

func (collector *certificateCollector) expiresSoonValue(certificate certs.Certificate, now time.Time) float64 {
	if certificate.Revoked {
		return 0
	}
	if certificate.ExpiresAt.Before(now) {
		return 0
	}
	if certificate.ExpiresAt.Sub(now) <= collector.expirySoonWindow {
		return 1
	}
	return 0
}

func (collector *certificateCollector) statusLabel(revoked bool) string {
	if revoked {
		return "revoked"
	}
	return "active"
}
