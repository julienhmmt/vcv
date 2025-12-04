package certs

import "time"

func MockCertificates() []Certificate {
	now := time.Now().UTC()
	return []Certificate{
		{
			ID:         "serial-001",
			CommonName: "example.internal",
			Sans:       []string{"example.internal"},
			CreatedAt:  now.Add(-24 * time.Hour),
			ExpiresAt:  now.Add(24 * time.Hour),
			Revoked:    false,
		},
		{
			ID:         "serial-002",
			CommonName: "api.internal",
			Sans:       []string{"api.internal", "api-alt.internal"},
			CreatedAt:  now.Add(-7 * 24 * time.Hour),
			ExpiresAt:  now.Add(7 * 24 * time.Hour),
			Revoked:    false,
		},
		{
			ID:         "serial-003",
			CommonName: "old.internal",
			Sans:       []string{"old.internal"},
			CreatedAt:  now.Add(-60 * 24 * time.Hour),
			ExpiresAt:  now.Add(-1 * 24 * time.Hour),
			Revoked:    false,
		},
		{
			ID:         "serial-004",
			CommonName: "revoked.internal",
			Sans:       []string{"revoked.internal"},
			CreatedAt:  now.Add(-60 * 24 * time.Hour),
			ExpiresAt:  now.Add(-10 * 24 * time.Hour),
			Revoked:    true,
		},
	}
}
