package certs

import (
	"math"
	"time"
)

// CountExpiring counts non-revoked, unexpired certificates whose remaining days
// until expiry fall within the warning and critical thresholds. A certificate
// inside the critical window is also counted toward warning when its remaining
// days are within the warning threshold too - callers get both counts from one
// pass. warningDays or criticalDays <= 0 disables that threshold.
func CountExpiring(certificates []Certificate, warningDays, criticalDays int, now time.Time) (warning, critical int) {
	for _, certificate := range certificates {
		if certificate.Revoked || certificate.ExpiresAt.IsZero() || certificate.ExpiresAt.Before(now) {
			continue
		}
		daysRemaining := daysUntilExpiry(certificate.ExpiresAt, now)
		if daysRemaining < 0 {
			continue
		}
		if warningDays > 0 && daysRemaining <= warningDays {
			warning++
		}
		if criticalDays > 0 && daysRemaining <= criticalDays {
			critical++
		}
	}
	return warning, critical
}

func daysUntilExpiry(expiresAt, now time.Time) int {
	diff := expiresAt.Sub(now)
	return int(math.Ceil(diff.Hours() / 24))
}
