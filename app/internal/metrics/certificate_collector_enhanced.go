package metrics

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"vcv/internal/certs"
)

// ageBucketNames lists the certificate age buckets in emission order.
var ageBucketNames = []string{"0-30d", "30-90d", "90-180d", "180-365d", "1y+"}

// sortedStringKeys returns sorted keys from a string-keyed map.
func sortedStringKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// matchesPattern checks if a string matches a pattern (supports wildcards).
func matchesPattern(pattern, value string) bool {
	matched, err := filepath.Match(pattern, value)
	if err != nil {
		return false
	}
	return matched
}

// sanBucketLabel classifies a SAN count into a SAN bucket name.
func sanBucketLabel(sanCount int) string {
	switch {
	case sanCount == 0:
		return "0"
	case sanCount <= 5:
		return "1-5"
	case sanCount <= 10:
		return "6-10"
	default:
		return "11+"
	}
}

// emitIssuerMetrics emits metrics grouped by certificate issuer CN.
func (collector *certificateCollector) emitIssuerMetrics(ch chan<- prometheus.Metric, certificates []certs.Certificate) {
	issuerCounts := make(map[string]map[string]map[string]int)
	for _, certificate := range certificates {
		vaultID, pki := extractVaultIDAndPKI(certificate.ID)
		incCounter3(issuerCounts, vaultID, pki, extractIssuerCN(certificate))
	}
	for _, vaultID := range sortedStringKeys(issuerCounts) {
		for _, pki := range sortedStringKeys(issuerCounts[vaultID]) {
			for _, issuer := range sortedStringKeys(issuerCounts[vaultID][pki]) {
				ch <- prometheus.MustNewConstMetric(certsByIssuerDesc, prometheus.GaugeValue, float64(issuerCounts[vaultID][pki][issuer]), vaultID, pki, issuer)
			}
		}
	}
}

// emitKeyTypeMetrics emits metrics grouped by key algorithm and size, including weak key detection.
func (collector *certificateCollector) emitKeyTypeMetrics(ch chan<- prometheus.Metric, certificates []certs.Certificate) {
	keyTypeCounts := make(map[string]map[string]map[string]int)
	weakKeyCounts := make(map[string]map[string]int)
	for _, certificate := range certificates {
		vaultID, pki := extractVaultIDAndPKI(certificate.ID)
		algorithm, keySize := extractKeyInfo(certificate)
		incCounter3(keyTypeCounts, vaultID, pki, algorithm+"_"+keySize)
		if isWeakKey(algorithm, keySize) {
			incCounter2(weakKeyCounts, vaultID, pki)
		}
	}
	for _, vaultID := range sortedStringKeys(keyTypeCounts) {
		for _, pki := range sortedStringKeys(keyTypeCounts[vaultID]) {
			for _, keyType := range sortedStringKeys(keyTypeCounts[vaultID][pki]) {
				parts := strings.SplitN(keyType, "_", 2)
				algorithm := "unknown"
				keySize := "0"
				if len(parts) == 2 {
					algorithm = parts[0]
					keySize = parts[1]
				}
				ch <- prometheus.MustNewConstMetric(certsByKeyTypeDesc, prometheus.GaugeValue, float64(keyTypeCounts[vaultID][pki][keyType]), vaultID, pki, algorithm, keySize)
			}
			ch <- prometheus.MustNewConstMetric(weakKeysDesc, prometheus.GaugeValue, float64(counter2At(weakKeyCounts, vaultID, pki)), vaultID, pki)
		}
	}
}

// emitSANMetrics emits metrics related to Subject Alternative Names.
func (collector *certificateCollector) emitSANMetrics(ch chan<- prometheus.Metric, certificates []certs.Certificate) {
	sanCounts := make(map[string]map[string]int)
	sanBuckets := make(map[string]map[string]map[string]int)
	for _, certificate := range certificates {
		vaultID, pki := extractVaultIDAndPKI(certificate.ID)
		sanCount := len(certificate.Sans)
		if sanCount > 0 {
			incCounter2(sanCounts, vaultID, pki)
		}
		incCounter3(sanBuckets, vaultID, pki, sanBucketLabel(sanCount))
	}
	for _, vaultID := range sortedStringKeys(sanBuckets) {
		for _, pki := range sortedStringKeys(sanBuckets[vaultID]) {
			ch <- prometheus.MustNewConstMetric(certsWithSansDesc, prometheus.GaugeValue, float64(counter2At(sanCounts, vaultID, pki)), vaultID, pki)
			for _, bucket := range []string{"0", "1-5", "6-10", "11+"} {
				ch <- prometheus.MustNewConstMetric(sanCountBucketDesc, prometheus.GaugeValue, float64(sanBuckets[vaultID][pki][bucket]), vaultID, pki, bucket)
			}
		}
	}
}

// emitAgeMetrics emits metrics about certificate age (time since issuance).
func (collector *certificateCollector) emitAgeMetrics(ch chan<- prometheus.Metric, certificates []certs.Certificate, now time.Time) {
	ageBuckets := make(map[string]map[string]map[string]int)
	for _, certificate := range certificates {
		vaultID, pki := extractVaultIDAndPKI(certificate.ID)
		if certificate.CreatedAt.IsZero() {
			continue
		}
		ageDays := int(now.Sub(certificate.CreatedAt).Hours() / 24)
		if ageDays < 0 {
			continue
		}
		incCounter3(ageBuckets, vaultID, pki, ageBucketLabel(ageDays))
	}
	for _, vaultID := range sortedStringKeys(ageBuckets) {
		for _, pki := range sortedStringKeys(ageBuckets[vaultID]) {
			for _, bucket := range ageBucketNames {
				ch <- prometheus.MustNewConstMetric(ageBucketDesc, prometheus.GaugeValue, float64(ageBuckets[vaultID][pki][bucket]), vaultID, pki, bucket)
			}
		}
	}
}

// ageBucketLabel classifies a certificate age in days into an age bucket name.
func ageBucketLabel(ageDays int) string {
	switch {
	case ageDays <= 30:
		return "0-30d"
	case ageDays <= 90:
		return "30-90d"
	case ageDays <= 180:
		return "90-180d"
	case ageDays <= 365:
		return "180-365d"
	default:
		return "1y+"
	}
}

// emitRenewalMetrics emits metrics about certificate renewal rates.
func (collector *certificateCollector) emitRenewalMetrics(ch chan<- prometheus.Metric, certificates []certs.Certificate, now time.Time) {
	issued24h := make(map[string]map[string]int)
	issued7d := make(map[string]map[string]int)
	issued30d := make(map[string]map[string]int)
	for _, certificate := range certificates {
		vaultID, pki := extractVaultIDAndPKI(certificate.ID)
		if certificate.CreatedAt.IsZero() {
			continue
		}
		ageDuration := now.Sub(certificate.CreatedAt)
		if ageDuration < 0 {
			continue
		}
		if ageDuration <= 24*time.Hour {
			incCounter2(issued24h, vaultID, pki)
		}
		if ageDuration <= 7*24*time.Hour {
			incCounter2(issued7d, vaultID, pki)
		}
		if ageDuration <= 30*24*time.Hour {
			incCounter2(issued30d, vaultID, pki)
		}
	}
	for _, vaultID := range unionKeys(issued24h, issued7d, issued30d) {
		for _, pki := range unionKeys(issued24h[vaultID], issued7d[vaultID], issued30d[vaultID]) {
			ch <- prometheus.MustNewConstMetric(issuedLast24hDesc, prometheus.GaugeValue, float64(counter2At(issued24h, vaultID, pki)), vaultID, pki)
			ch <- prometheus.MustNewConstMetric(issuedLast7dDesc, prometheus.GaugeValue, float64(counter2At(issued7d, vaultID, pki)), vaultID, pki)
			ch <- prometheus.MustNewConstMetric(issuedLast30dDesc, prometheus.GaugeValue, float64(counter2At(issued30d, vaultID, pki)), vaultID, pki)
		}
	}
}

// isPinnedCertificate reports whether the certificate matches any pinned
// pattern by common name, ID, or SAN.
func isPinnedCertificate(pinnedPatterns map[string]bool, certificate certs.Certificate) bool {
	values := []string{
		strings.ToLower(strings.TrimSpace(certificate.CommonName)),
		strings.ToLower(strings.TrimSpace(certificate.ID)),
	}
	for _, san := range certificate.Sans {
		values = append(values, strings.ToLower(strings.TrimSpace(san)))
	}
	for pattern := range pinnedPatterns {
		for _, value := range values {
			if matchesPattern(pattern, value) {
				return true
			}
		}
	}
	return false
}

// emitPinnedCertificateMetrics emits per-certificate metrics only for pinned certificates.
func (collector *certificateCollector) emitPinnedCertificateMetrics(ch chan<- prometheus.Metric, certificates []certs.Certificate, now time.Time) {
	if len(collector.pinnedCertificates) == 0 {
		return
	}
	pinnedMap := make(map[string]bool)
	for _, pinned := range collector.pinnedCertificates {
		pinnedMap[strings.ToLower(strings.TrimSpace(pinned))] = true
	}
	for _, certificate := range certificates {
		if !isPinnedCertificate(pinnedMap, certificate) || certificate.ExpiresAt.IsZero() {
			continue
		}
		vaultID, pki := extractVaultIDAndPKI(certificate.ID)
		status := collector.statusLabel(certificate, now)
		expiryTimestamp := float64(certificate.ExpiresAt.Unix())
		daysRemaining := float64(daysUntil(certificate.ExpiresAt.UTC(), now.UTC()))
		ch <- prometheus.MustNewConstMetric(pinnedCertExpiryDesc, prometheus.GaugeValue, expiryTimestamp, certificate.ID, certificate.CommonName, status, vaultID, pki)
		ch <- prometheus.MustNewConstMetric(pinnedCertDaysDesc, prometheus.GaugeValue, daysRemaining, certificate.ID, certificate.CommonName, status, vaultID, pki)
	}
}

// extractIssuerCN returns the issuer Common Name from list-time certificate fields.
func extractIssuerCN(certificate certs.Certificate) string {
	issuerCN := strings.TrimSpace(certificate.IssuerCN)
	if issuerCN != "" {
		return issuerCN
	}
	return "unknown"
}

// extractKeyInfo returns key algorithm and size labels from list-time certificate fields.
func extractKeyInfo(certificate certs.Certificate) (string, string) {
	algorithm := strings.TrimSpace(certificate.KeyAlgorithm)
	if algorithm == "" {
		algorithm = "unknown"
	}
	return algorithm, certs.KeySizeLabel(certificate.KeySize)
}

// isWeakKey determines if a key is considered weak based on algorithm and size.
func isWeakKey(algorithm string, keySize string) bool {
	if algorithm == "RSA" {
		if keySize == "1024" || keySize == "512" {
			return true
		}
	}
	if algorithm == "DSA" {
		return true
	}
	return false
}
