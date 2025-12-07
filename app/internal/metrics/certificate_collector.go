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
	certificatesTotalDesc = prometheus.NewDesc("vcv_certificates_total", "Total certificates grouped by status", []string{"status"}, nil)
	expiryTimestampDesc   = prometheus.NewDesc("vcv_certificate_expiry_timestamp_seconds", "Certificate expiration timestamp in seconds since epoch", []string{"serial_number", "common_name", "status"}, nil)
	expiresInDesc         = prometheus.NewDesc("vcv_certificate_expires_in_seconds", "Seconds until certificate expiration (zero when expired)", []string{"serial_number", "common_name", "status"}, nil)
	expiresSoonDesc       = prometheus.NewDesc("vcv_certificate_expires_soon", "Certificate expires soon within threshold window (1=true,0=false)", []string{"serial_number", "common_name"}, nil)
	lastScrapeSuccessDesc = prometheus.NewDesc("vcv_certificate_exporter_last_scrape_success", "Whether the last scrape succeeded (1) or failed (0)", nil, nil)
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
	ch <- certificatesTotalDesc
	ch <- expiryTimestampDesc
	ch <- expiresInDesc
	ch <- expiresSoonDesc
	ch <- lastScrapeSuccessDesc
}

func (collector *certificateCollector) Collect(ch chan<- prometheus.Metric) {
	certificates, err := collector.listCertificates()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(lastScrapeSuccessDesc, prometheus.GaugeValue, 0)
		return
	}
	ch <- prometheus.MustNewConstMetric(lastScrapeSuccessDesc, prometheus.GaugeValue, 1)
	now := collector.now()
	activeCount, revokedCount := collector.countStatuses(certificates)
	ch <- prometheus.MustNewConstMetric(certificatesTotalDesc, prometheus.GaugeValue, float64(activeCount), "active")
	ch <- prometheus.MustNewConstMetric(certificatesTotalDesc, prometheus.GaugeValue, float64(revokedCount), "revoked")
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
