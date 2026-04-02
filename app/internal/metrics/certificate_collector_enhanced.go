package metrics

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"vcv/internal/certs"
)

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

// emitIssuerMetrics emits metrics grouped by certificate issuer CN.
func (collector *certificateCollector) emitIssuerMetrics(ch chan<- prometheus.Metric, certificates []certs.Certificate) {
	issuerCounts := make(map[string]map[string]map[string]int)
	for _, certificate := range certificates {
		vaultID, pki := extractVaultIDAndPKI(certificate.ID)
		issuerCN := extractIssuerCN(certificate.CommonName)
		if _, ok := issuerCounts[vaultID]; !ok {
			issuerCounts[vaultID] = make(map[string]map[string]int)
		}
		if _, ok := issuerCounts[vaultID][pki]; !ok {
			issuerCounts[vaultID][pki] = make(map[string]int)
		}
		issuerCounts[vaultID][pki][issuerCN]++
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
		algorithm, keySize := extractKeyInfo(certificate.CommonName)
		if _, ok := keyTypeCounts[vaultID]; !ok {
			keyTypeCounts[vaultID] = make(map[string]map[string]int)
		}
		if _, ok := keyTypeCounts[vaultID][pki]; !ok {
			keyTypeCounts[vaultID][pki] = make(map[string]int)
		}
		keyTypeLabel := algorithm + "_" + keySize
		keyTypeCounts[vaultID][pki][keyTypeLabel]++
		if isWeakKey(algorithm, keySize) {
			if _, ok := weakKeyCounts[vaultID]; !ok {
				weakKeyCounts[vaultID] = make(map[string]int)
			}
			weakKeyCounts[vaultID][pki]++
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
			weakCount := 0
			if counts, ok := weakKeyCounts[vaultID]; ok {
				weakCount = counts[pki]
			}
			ch <- prometheus.MustNewConstMetric(weakKeysDesc, prometheus.GaugeValue, float64(weakCount), vaultID, pki)
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
		if _, ok := sanCounts[vaultID]; !ok {
			sanCounts[vaultID] = make(map[string]int)
		}
		if sanCount > 0 {
			sanCounts[vaultID][pki]++
		}
		if _, ok := sanBuckets[vaultID]; !ok {
			sanBuckets[vaultID] = make(map[string]map[string]int)
		}
		if _, ok := sanBuckets[vaultID][pki]; !ok {
			sanBuckets[vaultID][pki] = map[string]int{"0": 0, "1-5": 0, "6-10": 0, "11+": 0}
		}
		if sanCount == 0 {
			sanBuckets[vaultID][pki]["0"]++
		} else if sanCount <= 5 {
			sanBuckets[vaultID][pki]["1-5"]++
		} else if sanCount <= 10 {
			sanBuckets[vaultID][pki]["6-10"]++
		} else {
			sanBuckets[vaultID][pki]["11+"]++
		}
	}
	for _, vaultID := range sortedStringKeys(sanBuckets) {
		for _, pki := range sortedStringKeys(sanBuckets[vaultID]) {
			count := 0
			if counts, ok := sanCounts[vaultID]; ok {
				count = counts[pki]
			}
			ch <- prometheus.MustNewConstMetric(certsWithSansDesc, prometheus.GaugeValue, float64(count), vaultID, pki)
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
		if _, ok := ageBuckets[vaultID]; !ok {
			ageBuckets[vaultID] = make(map[string]map[string]int)
		}
		if _, ok := ageBuckets[vaultID][pki]; !ok {
			ageBuckets[vaultID][pki] = map[string]int{"0-30d": 0, "30-90d": 0, "90-180d": 0, "180-365d": 0, "1y+": 0}
		}
		if ageDays <= 30 {
			ageBuckets[vaultID][pki]["0-30d"]++
		} else if ageDays <= 90 {
			ageBuckets[vaultID][pki]["30-90d"]++
		} else if ageDays <= 180 {
			ageBuckets[vaultID][pki]["90-180d"]++
		} else if ageDays <= 365 {
			ageBuckets[vaultID][pki]["180-365d"]++
		} else {
			ageBuckets[vaultID][pki]["1y+"]++
		}
	}
	for _, vaultID := range sortedStringKeys(ageBuckets) {
		for _, pki := range sortedStringKeys(ageBuckets[vaultID]) {
			for _, bucket := range []string{"0-30d", "30-90d", "90-180d", "180-365d", "1y+"} {
				ch <- prometheus.MustNewConstMetric(ageBucketDesc, prometheus.GaugeValue, float64(ageBuckets[vaultID][pki][bucket]), vaultID, pki, bucket)
			}
		}
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
			if _, ok := issued24h[vaultID]; !ok {
				issued24h[vaultID] = make(map[string]int)
			}
			issued24h[vaultID][pki]++
		}
		if ageDuration <= 7*24*time.Hour {
			if _, ok := issued7d[vaultID]; !ok {
				issued7d[vaultID] = make(map[string]int)
			}
			issued7d[vaultID][pki]++
		}
		if ageDuration <= 30*24*time.Hour {
			if _, ok := issued30d[vaultID]; !ok {
				issued30d[vaultID] = make(map[string]int)
			}
			issued30d[vaultID][pki]++
		}
	}
	allVaultIDs := make(map[string]bool)
	for vaultID := range issued24h {
		allVaultIDs[vaultID] = true
	}
	for vaultID := range issued7d {
		allVaultIDs[vaultID] = true
	}
	for vaultID := range issued30d {
		allVaultIDs[vaultID] = true
	}
	for _, vaultID := range sortedStringKeys(allVaultIDs) {
		allPKIs := make(map[string]bool)
		if pkis, ok := issued24h[vaultID]; ok {
			for pki := range pkis {
				allPKIs[pki] = true
			}
		}
		if pkis, ok := issued7d[vaultID]; ok {
			for pki := range pkis {
				allPKIs[pki] = true
			}
		}
		if pkis, ok := issued30d[vaultID]; ok {
			for pki := range pkis {
				allPKIs[pki] = true
			}
		}
		for _, pki := range sortedStringKeys(allPKIs) {
			count24h := 0
			if counts, ok := issued24h[vaultID]; ok {
				count24h = counts[pki]
			}
			count7d := 0
			if counts, ok := issued7d[vaultID]; ok {
				count7d = counts[pki]
			}
			count30d := 0
			if counts, ok := issued30d[vaultID]; ok {
				count30d = counts[pki]
			}
			ch <- prometheus.MustNewConstMetric(issuedLast24hDesc, prometheus.GaugeValue, float64(count24h), vaultID, pki)
			ch <- prometheus.MustNewConstMetric(issuedLast7dDesc, prometheus.GaugeValue, float64(count7d), vaultID, pki)
			ch <- prometheus.MustNewConstMetric(issuedLast30dDesc, prometheus.GaugeValue, float64(count30d), vaultID, pki)
		}
	}
}

// emitPinnedCertificateMetrics emits per-certificate metrics only for pinned certificates.
func (collector *certificateCollector) emitPinnedCertificateMetrics(ch chan<- prometheus.Metric, certificates []certs.Certificate, now time.Time) {
	if len(collector.pinnedCertificates) == 0 {
		return
	}
	pinnedMap := make(map[string]bool)
	for _, pinned := range collector.pinnedCertificates {
		normalized := strings.ToLower(strings.TrimSpace(pinned))
		pinnedMap[normalized] = true
	}
	for _, certificate := range certificates {
		normalizedCN := strings.ToLower(strings.TrimSpace(certificate.CommonName))
		normalizedID := strings.ToLower(strings.TrimSpace(certificate.ID))
		isPinned := false
		for pattern := range pinnedMap {
			if matchesPattern(pattern, normalizedCN) || matchesPattern(pattern, normalizedID) {
				isPinned = true
				break
			}
		}
		if !isPinned {
			for _, san := range certificate.Sans {
				normalizedSAN := strings.ToLower(strings.TrimSpace(san))
				for pattern := range pinnedMap {
					if matchesPattern(pattern, normalizedSAN) {
						isPinned = true
						break
					}
				}
				if isPinned {
					break
				}
			}
		}
		if !isPinned {
			continue
		}
		if certificate.ExpiresAt.IsZero() {
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

// extractIssuerCN extracts a simplified issuer CN from certificate data.
// PLACEHOLDER: Returns domain-based heuristic until PEM parsing is implemented.
// Real issuer CN requires parsing the x509 certificate issuer field.
func extractIssuerCN(commonName string) string {
	parts := strings.Split(commonName, ".")
	if len(parts) > 1 {
		return parts[len(parts)-2] + "." + parts[len(parts)-1]
	}
	return "self-signed"
}

// extractKeyInfo extracts key algorithm and size information.
// PLACEHOLDER: Returns hardcoded RSA 2048 until PEM parsing is implemented.
// Real key info requires parsing the x509 certificate public key.
func extractKeyInfo(commonName string) (string, string) {
	return "RSA", "2048"
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
