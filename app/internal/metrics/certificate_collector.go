package metrics

import (
	"context"
	"errors"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"vcv/config"
	"vcv/internal/certs"
	"vcv/internal/vault"
)

const allLabelValue string = "__all__"

var (
	cacheSizeDesc              = prometheus.NewDesc("vcv_cache_size", "Number of items currently cached", nil, nil)
	certificatesLastFetchDesc  = prometheus.NewDesc("vcv_certificates_last_fetch_timestamp_seconds", "Timestamp of last successful certificates fetch", nil, nil)
	certificatesTotalDesc      = prometheus.NewDesc("vcv_certificates_total", "Total certificates grouped by status", []string{"vault_id", "pki", "status"}, nil)
	expiredCountDesc           = prometheus.NewDesc("vcv_certificates_expired_count", "Number of expired certificates", nil, nil)
	expiringSoonCountDesc      = prometheus.NewDesc("vcv_certificates_expiring_soon_count", "Number of certificates expiring soon within threshold window", []string{"vault_id", "pki", "level"}, nil)
	expiryTimestampDesc        = prometheus.NewDesc("vcv_certificate_expiry_timestamp_seconds", "Certificate expiration timestamp in seconds since epoch", []string{"certificate_id", "common_name", "status", "vault_id", "pki"}, nil)
	daysUntilExpiryDesc        = prometheus.NewDesc("vcv_certificate_days_until_expiry", "Days remaining until certificate expiration (negative if expired)", []string{"certificate_id", "common_name", "status", "vault_id", "pki"}, nil)
	expiryBucketDesc           = prometheus.NewDesc("vcv_certificates_expiry_bucket", "Number of certificates expiring in time bucket", []string{"vault_id", "pki", "bucket"}, nil)
	thresholdCriticalDesc      = prometheus.NewDesc("vcv_expiration_threshold_critical_days", "Configured critical expiration threshold in days", nil, nil)
	thresholdWarningDesc       = prometheus.NewDesc("vcv_expiration_threshold_warning_days", "Configured warning expiration threshold in days", nil, nil)
	lastScrapeDurationDesc     = prometheus.NewDesc("vcv_certificate_exporter_last_scrape_duration_seconds", "Duration of the last certificate scrape in seconds", nil, nil)
	lastScrapeSuccessDesc      = prometheus.NewDesc("vcv_certificate_exporter_last_scrape_success", "Whether the last scrape succeeded (1) or failed (0)", nil, nil)
	vaultConnectedDesc         = prometheus.NewDesc("vcv_vault_connected", "Vault connection status (1=connected,0=disconnected)", []string{"vault_id"}, nil)
	vaultListCertsSuccessDesc  = prometheus.NewDesc("vcv_vault_list_certificates_success", "Whether the last Vault certificate listing succeeded (1) or failed (0)", []string{"vault_id"}, nil)
	vaultListCertsDurationDesc = prometheus.NewDesc("vcv_vault_list_certificates_duration_seconds", "Duration of the last Vault certificate listing in seconds", []string{"vault_id"}, nil)
	vaultListCertsErrorDesc    = prometheus.NewDesc("vcv_vault_list_certificates_error", "Whether the last Vault certificate listing errored (1) or not (0)", []string{"vault_id"}, nil)
	partialScrapeDesc          = prometheus.NewDesc("vcv_certificates_partial_scrape", "Whether the last scrape was partial (1) due to per-vault errors", []string{"vault_id"}, nil)
	configuredVaultsDesc       = prometheus.NewDesc("vcv_vaults_configured", "Number of Vault instances configured", nil, nil)
	configuredMountsDesc       = prometheus.NewDesc("vcv_pki_mounts_configured", "Number of PKI mounts configured for a vault", []string{"vault_id"}, nil)
)

type certificateCollector struct {
	vaultClient      vault.Client
	statusClients    map[string]vault.Client
	thresholds       config.ExpirationThresholds
	perCertificate   bool
	enhancedMetrics  bool
	configuredVaults []config.VaultInstance
	now              func() time.Time
}

// NewCertificateCollector returns a Prometheus collector exposing certificate inventory and expiry status.

func NewCertificateCollector(vaultClient vault.Client, statusClients map[string]vault.Client, thresholds config.ExpirationThresholds) prometheus.Collector {
	critical := thresholds.Critical
	if critical <= 0 {
		critical = 7
	}
	warning := thresholds.Warning
	if warning <= 0 {
		warning = 30
	}
	perCertificate := parseBoolEnv("VCV_METRICS_PER_CERTIFICATE", false)
	enhancedMetrics := parseBoolEnv("VCV_METRICS_ENHANCED", true)
	clients := statusClients
	if clients == nil {
		clients = map[string]vault.Client{}
	}
	return &certificateCollector{
		vaultClient:      vaultClient,
		statusClients:    clients,
		thresholds:       config.ExpirationThresholds{Critical: critical, Warning: warning},
		perCertificate:   perCertificate,
		enhancedMetrics:  enhancedMetrics,
		configuredVaults: []config.VaultInstance{},
		now:              time.Now,
	}
}

func NewCertificateCollectorWithVaults(vaultClient vault.Client, statusClients map[string]vault.Client, thresholds config.ExpirationThresholds, vaults []config.VaultInstance) prometheus.Collector {
	collector := NewCertificateCollector(vaultClient, statusClients, thresholds)
	typed, ok := collector.(*certificateCollector)
	if !ok {
		return collector
	}
	typed.configuredVaults = append([]config.VaultInstance{}, vaults...)
	return typed
}

func (collector *certificateCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cacheSizeDesc
	ch <- certificatesLastFetchDesc
	ch <- certificatesTotalDesc
	ch <- expiredCountDesc
	ch <- expiringSoonCountDesc
	ch <- expiryTimestampDesc
	ch <- daysUntilExpiryDesc
	ch <- expiryBucketDesc
	ch <- thresholdCriticalDesc
	ch <- thresholdWarningDesc
	ch <- lastScrapeDurationDesc
	ch <- lastScrapeSuccessDesc
	ch <- vaultConnectedDesc
	ch <- vaultListCertsSuccessDesc
	ch <- vaultListCertsDurationDesc
	ch <- vaultListCertsErrorDesc
	ch <- partialScrapeDesc
	ch <- configuredVaultsDesc
	ch <- configuredMountsDesc
}

func (collector *certificateCollector) Collect(ch chan<- prometheus.Metric) {
	scrapeStart := time.Now()
	certificates, listResults, err := collector.listCertificatesWithVaultResults()
	scrapeDuration := time.Since(scrapeStart).Seconds()
	ch <- prometheus.MustNewConstMetric(lastScrapeDurationDesc, prometheus.GaugeValue, scrapeDuration)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(lastScrapeSuccessDesc, prometheus.GaugeValue, 0)
		collector.emitConfigurationMetrics(ch)
		collector.emitVaultConnectionMetrics(ch)
		if len(listResults) == 0 {
			collector.emitVaultListingFailureMetrics(ch, scrapeDuration)
			return
		}
		collector.emitVaultListingMetrics(ch, listResults, scrapeDuration)
		return
	}
	ch <- prometheus.MustNewConstMetric(lastScrapeSuccessDesc, prometheus.GaugeValue, 1)
	now := collector.now()

	collector.emitConfigurationMetrics(ch)

	collector.emitVaultConnectionMetrics(ch)
	collector.emitVaultListingMetrics(ch, listResults, scrapeDuration)

	validCount, revokedCount, expiredCount := collector.countStatuses(certificates, now)
	warningSoonCount, criticalSoonCount := collector.countExpiringSoon(certificates, now)
	cacheSize := collector.getCacheSize()

	ch <- prometheus.MustNewConstMetric(cacheSizeDesc, prometheus.GaugeValue, float64(cacheSize))
	ch <- prometheus.MustNewConstMetric(certificatesLastFetchDesc, prometheus.GaugeValue, float64(now.Unix()))
	ch <- prometheus.MustNewConstMetric(certificatesTotalDesc, prometheus.GaugeValue, float64(validCount), allLabelValue, allLabelValue, "valid")
	ch <- prometheus.MustNewConstMetric(certificatesTotalDesc, prometheus.GaugeValue, float64(revokedCount), allLabelValue, allLabelValue, "revoked")
	ch <- prometheus.MustNewConstMetric(certificatesTotalDesc, prometheus.GaugeValue, float64(expiredCount), allLabelValue, allLabelValue, "expired")
	ch <- prometheus.MustNewConstMetric(expiredCountDesc, prometheus.GaugeValue, float64(expiredCount))
	ch <- prometheus.MustNewConstMetric(expiringSoonCountDesc, prometheus.GaugeValue, float64(warningSoonCount), allLabelValue, allLabelValue, "warning")
	ch <- prometheus.MustNewConstMetric(expiringSoonCountDesc, prometheus.GaugeValue, float64(criticalSoonCount), allLabelValue, allLabelValue, "critical")
	ch <- prometheus.MustNewConstMetric(thresholdCriticalDesc, prometheus.GaugeValue, float64(collector.thresholds.Critical))
	ch <- prometheus.MustNewConstMetric(thresholdWarningDesc, prometheus.GaugeValue, float64(collector.thresholds.Warning))
	collector.emitCertificateAggregationMetrics(ch, certificates, now)
	collector.emitPerCertificateMetrics(ch, certificates, now)
	if collector.enhancedMetrics {
		collector.emitEnhancedMetrics(ch, certificates, now)
	}
}

func (collector *certificateCollector) listCertificates() ([]certs.Certificate, error) {
	return collector.vaultClient.ListCertificates(context.Background())
}

func (collector *certificateCollector) listCertificatesWithVaultResults() ([]certs.Certificate, []vault.ListCertificatesByVaultResult, error) {
	if lister, ok := collector.vaultClient.(vault.CertificatesByVaultLister); ok {
		results := lister.ListCertificatesByVault(context.Background())
		combined := make([]certs.Certificate, 0)
		anySuccess := false
		for _, result := range results {
			if result.ListError != nil {
				continue
			}
			anySuccess = true
			combined = append(combined, result.Certificates...)
		}
		if len(results) > 0 && !anySuccess {
			return []certs.Certificate{}, results, errors.New("all vault listings failed")
		}
		return combined, results, nil
	}
	certificates, err := collector.listCertificates()
	return certificates, nil, err
}

func (collector *certificateCollector) emitVaultConnectionMetrics(ch chan<- prometheus.Metric) {
	connectedOverall := 1.0
	if err := collector.vaultClient.CheckConnection(context.Background()); err != nil {
		connectedOverall = 0.0
	}
	ch <- prometheus.MustNewConstMetric(vaultConnectedDesc, prometheus.GaugeValue, connectedOverall, allLabelValue)
	if len(collector.statusClients) == 0 {
		return
	}
	vaultIDs := make([]string, 0, len(collector.statusClients))
	for vaultID := range collector.statusClients {
		trimmed := strings.TrimSpace(vaultID)
		if trimmed == "" {
			continue
		}
		vaultIDs = append(vaultIDs, trimmed)
	}
	sort.Strings(vaultIDs)
	for _, vaultID := range vaultIDs {
		client := collector.statusClients[vaultID]
		if client == nil {
			ch <- prometheus.MustNewConstMetric(vaultConnectedDesc, prometheus.GaugeValue, 0, vaultID)
			continue
		}
		connected := 1.0
		if err := client.CheckConnection(context.Background()); err != nil {
			connected = 0.0
		}
		ch <- prometheus.MustNewConstMetric(vaultConnectedDesc, prometheus.GaugeValue, connected, vaultID)
	}
}

func (collector *certificateCollector) countStatuses(certificates []certs.Certificate, now time.Time) (int, int, int) {
	validCount := 0
	revokedCount := 0
	expiredCount := 0
	for _, certificate := range certificates {
		status := collector.statusLabel(certificate, now)
		switch status {
		case "revoked":
			revokedCount++
		case "expired":
			expiredCount++
		default:
			validCount++
		}
	}
	return validCount, revokedCount, expiredCount
}

func (collector *certificateCollector) countExpiringSoon(certificates []certs.Certificate, now time.Time) (int, int) {
	warningCount := 0
	criticalCount := 0
	for _, certificate := range certificates {
		if collector.statusLabel(certificate, now) != "valid" {
			continue
		}
		if certificate.ExpiresAt.IsZero() {
			continue
		}
		daysRemaining := daysUntil(certificate.ExpiresAt.UTC(), now.UTC())
		if daysRemaining < 0 {
			continue
		}
		if collector.thresholds.Warning > 0 && daysRemaining <= collector.thresholds.Warning {
			warningCount++
		}
		if collector.thresholds.Critical > 0 && daysRemaining <= collector.thresholds.Critical {
			criticalCount++
		}
	}
	return warningCount, criticalCount
}

func (collector *certificateCollector) getCacheSize() int {
	if sizer, ok := collector.vaultClient.(vault.CacheSizer); ok {
		return sizer.CacheSize()
	}
	return 0
}

func (collector *certificateCollector) emitConfigurationMetrics(ch chan<- prometheus.Metric) {
	configuredVaults := collector.configuredVaults
	if len(configuredVaults) == 0 {
		ch <- prometheus.MustNewConstMetric(configuredVaultsDesc, prometheus.GaugeValue, 0)
		ch <- prometheus.MustNewConstMetric(configuredMountsDesc, prometheus.GaugeValue, 0, allLabelValue)
		return
	}
	ch <- prometheus.MustNewConstMetric(configuredVaultsDesc, prometheus.GaugeValue, float64(len(configuredVaults)))
	totalMounts := 0
	for _, instance := range configuredVaults {
		vaultID := strings.TrimSpace(instance.ID)
		if vaultID == "" {
			continue
		}
		mountCount := 0
		for _, mount := range instance.PKIMounts {
			if strings.TrimSpace(mount) == "" {
				continue
			}
			mountCount++
		}
		totalMounts += mountCount
		ch <- prometheus.MustNewConstMetric(configuredMountsDesc, prometheus.GaugeValue, float64(mountCount), vaultID)
	}
	ch <- prometheus.MustNewConstMetric(configuredMountsDesc, prometheus.GaugeValue, float64(totalMounts), allLabelValue)
}

func (collector *certificateCollector) emitVaultListingMetrics(ch chan<- prometheus.Metric, listResults []vault.ListCertificatesByVaultResult, scrapeDuration float64) {
	ch <- prometheus.MustNewConstMetric(vaultListCertsDurationDesc, prometheus.GaugeValue, scrapeDuration, allLabelValue)
	if len(listResults) == 0 {
		ch <- prometheus.MustNewConstMetric(vaultListCertsSuccessDesc, prometheus.GaugeValue, 1, allLabelValue)
		ch <- prometheus.MustNewConstMetric(vaultListCertsErrorDesc, prometheus.GaugeValue, 0, allLabelValue)
		ch <- prometheus.MustNewConstMetric(partialScrapeDesc, prometheus.GaugeValue, 0, allLabelValue)
		return
	}
	sort.Slice(listResults, func(left int, right int) bool {
		return listResults[left].VaultID < listResults[right].VaultID
	})
	partial := 0.0
	for _, result := range listResults {
		vaultID := strings.TrimSpace(result.VaultID)
		if vaultID == "" {
			continue
		}
		duration := result.Duration.Seconds()
		success := 1.0
		errorValue := 0.0
		if result.ListError != nil {
			success = 0.0
			errorValue = 1.0
			partial = 1.0
		}
		ch <- prometheus.MustNewConstMetric(vaultListCertsSuccessDesc, prometheus.GaugeValue, success, vaultID)
		ch <- prometheus.MustNewConstMetric(vaultListCertsErrorDesc, prometheus.GaugeValue, errorValue, vaultID)
		ch <- prometheus.MustNewConstMetric(vaultListCertsDurationDesc, prometheus.GaugeValue, duration, vaultID)
	}
	allSuccess := 1.0
	allError := 0.0
	if partial > 0 {
		allSuccess = 0.0
		allError = 1.0
	}
	ch <- prometheus.MustNewConstMetric(vaultListCertsSuccessDesc, prometheus.GaugeValue, allSuccess, allLabelValue)
	ch <- prometheus.MustNewConstMetric(vaultListCertsErrorDesc, prometheus.GaugeValue, allError, allLabelValue)
	ch <- prometheus.MustNewConstMetric(partialScrapeDesc, prometheus.GaugeValue, partial, allLabelValue)
}

func (collector *certificateCollector) emitVaultListingFailureMetrics(ch chan<- prometheus.Metric, scrapeDuration float64) {
	ch <- prometheus.MustNewConstMetric(vaultListCertsDurationDesc, prometheus.GaugeValue, scrapeDuration, allLabelValue)
	ch <- prometheus.MustNewConstMetric(vaultListCertsSuccessDesc, prometheus.GaugeValue, 0, allLabelValue)
	ch <- prometheus.MustNewConstMetric(vaultListCertsErrorDesc, prometheus.GaugeValue, 1, allLabelValue)
	ch <- prometheus.MustNewConstMetric(partialScrapeDesc, prometheus.GaugeValue, 1, allLabelValue)
}

func (collector *certificateCollector) emitCertificateAggregationMetrics(ch chan<- prometheus.Metric, certificates []certs.Certificate, now time.Time) {
	totals := make(map[string]map[string]int)
	soon := make(map[string]map[string]int)
	for _, certificate := range certificates {
		vaultID, pki := extractVaultIDAndPKI(certificate.ID)
		key := buildAggregationKey(vaultID, pki)
		status := collector.statusLabel(certificate, now)
		if _, ok := totals[key]; !ok {
			totals[key] = map[string]int{"valid": 0, "revoked": 0, "expired": 0}
		}
		totals[key][status] = totals[key][status] + 1
		if status != "valid" {
			continue
		}
		if certificate.ExpiresAt.IsZero() {
			continue
		}
		daysRemaining := daysUntil(certificate.ExpiresAt.UTC(), now.UTC())
		if daysRemaining < 0 {
			continue
		}
		if _, ok := soon[key]; !ok {
			soon[key] = map[string]int{"warning": 0, "critical": 0}
		}
		if collector.thresholds.Warning > 0 && daysRemaining <= collector.thresholds.Warning {
			soon[key]["warning"] = soon[key]["warning"] + 1
		}
		if collector.thresholds.Critical > 0 && daysRemaining <= collector.thresholds.Critical {
			soon[key]["critical"] = soon[key]["critical"] + 1
		}
	}

	keys := make([]string, 0, len(totals))
	for key := range totals {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		vaultID, pki := splitAggregationKey(key)
		for _, status := range []string{"valid", "revoked", "expired"} {
			ch <- prometheus.MustNewConstMetric(certificatesTotalDesc, prometheus.GaugeValue, float64(totals[key][status]), vaultID, pki, status)
		}
		counts, ok := soon[key]
		if !ok {
			counts = map[string]int{"warning": 0, "critical": 0}
		}
		ch <- prometheus.MustNewConstMetric(expiringSoonCountDesc, prometheus.GaugeValue, float64(counts["warning"]), vaultID, pki, "warning")
		ch <- prometheus.MustNewConstMetric(expiringSoonCountDesc, prometheus.GaugeValue, float64(counts["critical"]), vaultID, pki, "critical")
	}
}

func (collector *certificateCollector) emitPerCertificateMetrics(ch chan<- prometheus.Metric, certificates []certs.Certificate, now time.Time) {
	if !collector.perCertificate {
		return
	}
	for _, certificate := range certificates {
		if certificate.ExpiresAt.IsZero() {
			continue
		}
		vaultID, pki := extractVaultIDAndPKI(certificate.ID)
		status := collector.statusLabel(certificate, now)
		expiryTimestamp := float64(certificate.ExpiresAt.Unix())
		ch <- prometheus.MustNewConstMetric(expiryTimestampDesc, prometheus.GaugeValue, expiryTimestamp, certificate.ID, certificate.CommonName, status, vaultID, pki)
		if collector.enhancedMetrics {
			daysRemaining := float64(daysUntil(certificate.ExpiresAt.UTC(), now.UTC()))
			ch <- prometheus.MustNewConstMetric(daysUntilExpiryDesc, prometheus.GaugeValue, daysRemaining, certificate.ID, certificate.CommonName, status, vaultID, pki)
		}
	}
}

func (collector *certificateCollector) statusLabel(certificate certs.Certificate, now time.Time) string {
	if certificate.Revoked {
		return "revoked"
	}
	if !certificate.ExpiresAt.IsZero() && certificate.ExpiresAt.Before(now) {
		return "expired"
	}
	return "valid"
}

func buildAggregationKey(vaultID string, pki string) string {
	return vaultID + "|" + pki
}

func splitAggregationKey(key string) (string, string) {
	parts := strings.SplitN(key, "|", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

func extractVaultIDAndPKI(certificateID string) (string, string) {
	trimmed := strings.TrimSpace(certificateID)
	if trimmed == "" {
		return "", ""
	}
	vaultID := allLabelValue
	mountSerial := trimmed
	if parts := strings.SplitN(trimmed, "|", 2); len(parts) == 2 {
		candidate := strings.TrimSpace(parts[0])
		if candidate != "" {
			vaultID = candidate
		}
		mountSerial = strings.TrimSpace(parts[1])
	}
	parts := strings.SplitN(mountSerial, ":", 2)
	if len(parts) < 2 {
		return vaultID, ""
	}
	return vaultID, strings.TrimSpace(parts[0])
}

func daysUntil(expiresAt time.Time, now time.Time) int {
	if expiresAt.IsZero() {
		return -1
	}
	diff := expiresAt.Sub(now)
	return int(math.Ceil(diff.Hours() / 24))
}

func (collector *certificateCollector) emitEnhancedMetrics(ch chan<- prometheus.Metric, certificates []certs.Certificate, now time.Time) {
	buckets := make(map[string]map[string]map[string]int)
	for _, certificate := range certificates {
		vaultID, pki := extractVaultIDAndPKI(certificate.ID)
		if _, ok := buckets[vaultID]; !ok {
			buckets[vaultID] = make(map[string]map[string]int)
		}
		if _, ok := buckets[vaultID][pki]; !ok {
			buckets[vaultID][pki] = map[string]int{"0-7d": 0, "7-30d": 0, "30-90d": 0, "90d+": 0, "expired": 0, "revoked": 0}
		}
		if certificate.Revoked {
			buckets[vaultID][pki]["revoked"]++
			continue
		}
		if certificate.ExpiresAt.IsZero() {
			continue
		}
		daysRemaining := daysUntil(certificate.ExpiresAt.UTC(), now.UTC())
		if daysRemaining < 0 {
			buckets[vaultID][pki]["expired"]++
		} else if daysRemaining <= 7 {
			buckets[vaultID][pki]["0-7d"]++
		} else if daysRemaining <= 30 {
			buckets[vaultID][pki]["7-30d"]++
		} else if daysRemaining <= 90 {
			buckets[vaultID][pki]["30-90d"]++
		} else {
			buckets[vaultID][pki]["90d+"]++
		}
	}
	vaultIDs := make([]string, 0, len(buckets))
	for vaultID := range buckets {
		vaultIDs = append(vaultIDs, vaultID)
	}
	sort.Strings(vaultIDs)
	for _, vaultID := range vaultIDs {
		pkis := make([]string, 0, len(buckets[vaultID]))
		for pki := range buckets[vaultID] {
			pkis = append(pkis, pki)
		}
		sort.Strings(pkis)
		for _, pki := range pkis {
			for _, bucket := range []string{"0-7d", "7-30d", "30-90d", "90d+", "expired", "revoked"} {
				ch <- prometheus.MustNewConstMetric(expiryBucketDesc, prometheus.GaugeValue, float64(buckets[vaultID][pki][bucket]), vaultID, pki, bucket)
			}
		}
	}
	allBuckets := map[string]int{"0-7d": 0, "7-30d": 0, "30-90d": 0, "90d+": 0, "expired": 0, "revoked": 0}
	for vaultID := range buckets {
		for pki := range buckets[vaultID] {
			for bucket, count := range buckets[vaultID][pki] {
				allBuckets[bucket] += count
			}
		}
	}
	for _, bucket := range []string{"0-7d", "7-30d", "30-90d", "90d+", "expired", "revoked"} {
		ch <- prometheus.MustNewConstMetric(expiryBucketDesc, prometheus.GaugeValue, float64(allBuckets[bucket]), allLabelValue, allLabelValue, bucket)
	}
}

func parseBoolEnv(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}
