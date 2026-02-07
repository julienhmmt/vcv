package handlers

import (
	"testing"
	"time"

	"vcv/config"
	"vcv/internal/certs"
	"vcv/internal/i18n"

	"github.com/stretchr/testify/assert"
)

func TestBuildCertRows_BadgeLogic(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	thresholds := config.ExpirationThresholds{
		Critical: 7,
		Warning:  30,
	}
	messages := i18n.MessagesForLanguage("en")

	tests := []struct {
		name              string
		cert              certs.Certificate
		expectedClass     string
		expectedLabel     string
		expectedDaysClass string
		expectedRowClass  string
	}{
		{
			name: "Valid and Safe (far future)",
			cert: certs.Certificate{
				ID:        "test",
				ExpiresAt: now.Add(40 * 24 * time.Hour),
			},
			expectedClass:     "vcv-badge vcv-badge-valid",
			expectedLabel:     messages.StatusLabelValid,
			expectedDaysClass: "",
			expectedRowClass:  "vcv-row-valid",
		},
		{
			name: "Valid but Warning (soon expiring)",
			cert: certs.Certificate{
				ID:        "test",
				ExpiresAt: now.Add(20 * 24 * time.Hour),
			},
			expectedClass:     "vcv-badge vcv-badge-warning",
			expectedLabel:     messages.StatusLabelValid,
			expectedDaysClass: "vcv-days-remaining vcv-days-warning",
			expectedRowClass:  "vcv-row-warning",
		},
		{
			name: "Valid but Critical (very soon expiring)",
			cert: certs.Certificate{
				ID:        "test",
				ExpiresAt: now.Add(5 * 24 * time.Hour),
			},
			expectedClass:     "vcv-badge vcv-badge-critical",
			expectedLabel:     messages.StatusLabelValid,
			expectedDaysClass: "vcv-days-remaining vcv-days-critical",
			expectedRowClass:  "vcv-row-critical",
		},
		{
			name: "Expiring today",
			cert: certs.Certificate{
				ID:        "test",
				ExpiresAt: now.Add(2 * time.Hour),
			},
			expectedClass:     "vcv-badge vcv-badge-critical",
			expectedLabel:     messages.StatusLabelValid,
			expectedDaysClass: "vcv-days-remaining vcv-days-critical",
			expectedRowClass:  "vcv-row-critical",
		},
		{
			name: "Expired",
			cert: certs.Certificate{
				ID:        "test",
				ExpiresAt: now.Add(-1 * 24 * time.Hour),
			},
			expectedClass:     "vcv-badge vcv-badge-expired",
			expectedLabel:     messages.StatusLabelExpired,
			expectedDaysClass: "vcv-days-remaining vcv-days-expired",
			expectedRowClass:  "vcv-row-expired",
		},
		{
			name: "Expired today",
			cert: certs.Certificate{
				ID:        "test",
				ExpiresAt: now.Add(-2 * time.Hour),
			},
			expectedClass:     "vcv-badge vcv-badge-expired",
			expectedLabel:     messages.StatusLabelExpired,
			expectedDaysClass: "vcv-days-remaining vcv-days-expired",
			expectedRowClass:  "vcv-row-expired",
		},
		{
			name: "Revoked",
			cert: certs.Certificate{
				ID:        "test",
				Revoked:   true,
				ExpiresAt: now.Add(40 * 24 * time.Hour),
			},
			expectedClass:     "vcv-badge vcv-badge-revoked",
			expectedLabel:     messages.StatusLabelRevoked,
			expectedDaysClass: "",
			expectedRowClass:  "vcv-row-revoked",
		},
		{
			name: "Revoked and expiring soon (badge stays revoked)",
			cert: certs.Certificate{
				ID:        "test",
				Revoked:   true,
				ExpiresAt: now.Add(5 * 24 * time.Hour),
			},
			expectedClass:     "vcv-badge vcv-badge-revoked",
			expectedLabel:     messages.StatusLabelRevoked,
			expectedDaysClass: "vcv-days-remaining vcv-days-critical",
			expectedRowClass:  "vcv-row-revoked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := buildCertRows([]certs.Certificate{tt.cert}, messages, thresholds, nil, false, now)
			assert.NotEmpty(t, rows)
			row := rows[0]
			assert.NotEmpty(t, row.Badges)
			badge := row.Badges[0]
			assert.Equal(t, tt.expectedLabel, badge.Label)
			assert.Equal(t, tt.expectedClass, badge.Class)
			if tt.expectedDaysClass != "" {
				assert.Equal(t, tt.expectedDaysClass, row.DaysRemainingClass)
			}
			assert.Equal(t, tt.expectedRowClass, row.RowClass)
		})
	}
}

func TestBuildCertRows_DaysRemainingText(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	thresholds := config.ExpirationThresholds{Critical: 7, Warning: 30}
	messages := i18n.MessagesForLanguage("en")

	tests := []struct {
		name         string
		cert         certs.Certificate
		expectedText string
	}{
		{
			name:         "Expiring today shows ExpiringToday",
			cert:         certs.Certificate{ID: "t", ExpiresAt: now.Add(6 * time.Hour)},
			expectedText: messages.ExpiringToday,
		},
		{
			name:         "Expired today shows ExpiredDaysSingular",
			cert:         certs.Certificate{ID: "t", ExpiresAt: now.Add(-2 * time.Hour)},
			expectedText: "Expired 1 day ago",
		},
		{
			name:         "1 day remaining uses singular",
			cert:         certs.Certificate{ID: "t", ExpiresAt: now.Add(36 * time.Hour)},
			expectedText: "1 day remaining",
		},
		{
			name:         "Expired 1.5 days ago uses plural (floor rounds to 2)",
			cert:         certs.Certificate{ID: "t", ExpiresAt: now.Add(-36 * time.Hour)},
			expectedText: "Expired 2 days ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := buildCertRows([]certs.Certificate{tt.cert}, messages, thresholds, nil, false, now)
			assert.NotEmpty(t, rows)
			assert.Equal(t, tt.expectedText, rows[0].DaysRemainingText)
		})
	}
}

func TestResolveExpirationLevel(t *testing.T) {
	thresholds := config.ExpirationThresholds{Critical: 7, Warning: 30}

	tests := []struct {
		name     string
		days     int
		expected string
	}{
		{name: "critical", days: 5, expected: "critical"},
		{name: "warning", days: 15, expected: "warning"},
		{name: "ok", days: 40, expected: "ok"},
		{name: "boundary critical", days: 7, expected: "critical"},
		{name: "boundary warning", days: 30, expected: "warning"},
		{name: "zero days is critical", days: 0, expected: "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, resolveExpirationLevel(tt.days, thresholds))
		})
	}
}
