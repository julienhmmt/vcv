package certs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCountExpiring(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	certs := []Certificate{
		{ID: "critical", ExpiresAt: now.Add(3 * 24 * time.Hour)},
		{ID: "warning", ExpiresAt: now.Add(20 * 24 * time.Hour)},
		{ID: "far-future", ExpiresAt: now.Add(365 * 24 * time.Hour)},
		{ID: "expired", ExpiresAt: now.Add(-24 * time.Hour)},
		{ID: "revoked-but-soon", ExpiresAt: now.Add(1 * 24 * time.Hour), Revoked: true},
		{ID: "no-expiry"},
	}

	warning, critical := CountExpiring(certs, 30, 7, now)
	assert.Equal(t, 2, warning) // critical + warning both fall within the 30-day warning window
	assert.Equal(t, 1, critical)
}

func TestCountExpiring_DisabledThresholds(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	certs := []Certificate{{ID: "soon", ExpiresAt: now.Add(24 * time.Hour)}}

	warning, critical := CountExpiring(certs, 0, 0, now)
	assert.Equal(t, 0, warning)
	assert.Equal(t, 0, critical)
}

func TestCountExpiring_Empty(t *testing.T) {
	warning, critical := CountExpiring(nil, 30, 7, time.Now())
	assert.Equal(t, 0, warning)
	assert.Equal(t, 0, critical)
}
