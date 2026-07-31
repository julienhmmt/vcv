import { describe, it, expect } from 'vitest'
import { buildExpiryTimeline, TIMELINE_FAR_HORIZON_DAYS, type ExpiryBucket } from '$lib/utils/expiry-timeline'
import { DEFAULT_THRESHOLDS } from '$lib/utils/cert-status'
import type { Certificate, ExpirationThresholds } from '$lib/types'

const NOW = new Date('2026-07-31T12:00:00Z')

function cert(overrides: Partial<Certificate> = {}): Certificate {
  return {
    id: 'v|m:1',
    serialNumber: 'aa:bb',
    commonName: 'web.example.com',
    sans: [],
    certType: 'machine',
    createdAt: '2024-01-01T00:00:00Z',
    expiresAt: '2999-01-01T00:00:00Z',
    revoked: false,
    ...overrides,
  }
}

function expiresInDays(days: number): string {
  return new Date(NOW.getTime() + days * 86_400_000).toISOString()
}

function counts(buckets: ExpiryBucket[]): Record<string, number> {
  return Object.fromEntries(buckets.map((bucket) => [bucket.key, bucket.count]))
}

describe('buildExpiryTimeline', () => {
  const thresholds: ExpirationThresholds = DEFAULT_THRESHOLDS // critical 7, warning 30

  const cases: { name: string; certs: Certificate[]; expected: Record<string, number> }[] = [
    {
      name: 'cert expiring within the critical threshold lands in critical',
      certs: [cert({ expiresAt: expiresInDays(3.5) })],
      expected: { critical: 1, warning: 0, near: 0, later: 0 },
    },
    {
      name: 'cert expiring today counts as critical, not expired',
      certs: [cert({ expiresAt: expiresInDays(0.5) })],
      expected: { critical: 1, warning: 0, near: 0, later: 0 },
    },
    {
      name: 'cert between thresholds lands in warning',
      certs: [cert({ expiresAt: expiresInDays(20.5) })],
      expected: { critical: 0, warning: 1, near: 0, later: 0 },
    },
    {
      name: 'cert past the warning threshold lands in near',
      certs: [cert({ expiresAt: expiresInDays(60.5) })],
      expected: { critical: 0, warning: 0, near: 1, later: 0 },
    },
    {
      name: 'cert past the far horizon lands in later',
      certs: [cert({ expiresAt: expiresInDays(200.5) })],
      expected: { critical: 0, warning: 0, near: 0, later: 1 },
    },
    {
      name: 'already expired certs are excluded',
      certs: [cert({ expiresAt: expiresInDays(-2.5) })],
      expected: { critical: 0, warning: 0, near: 0, later: 0 },
    },
    {
      name: 'revoked certs are excluded',
      certs: [cert({ expiresAt: expiresInDays(3.5), revoked: true })],
      expected: { critical: 0, warning: 0, near: 0, later: 0 },
    },
    {
      name: 'unparseable expiry dates are excluded',
      certs: [cert({ expiresAt: 'not-a-date' })],
      expected: { critical: 0, warning: 0, near: 0, later: 0 },
    },
  ]

  for (const { name, certs, expected } of cases) {
    it(name, () => {
      expect(counts(buildExpiryTimeline(certs, thresholds, NOW))).toEqual(expected)
    })
  }

  it('respects custom thresholds for the critical/warning boundary', () => {
    const custom: ExpirationThresholds = { critical: 14, warning: 45 }
    const buckets = buildExpiryTimeline([cert({ expiresAt: expiresInDays(10.5) })], custom, NOW)
    expect(counts(buckets)).toEqual({ critical: 1, warning: 0, near: 0, later: 0 })
    expect(buckets[0]).toMatchObject({ key: 'critical', from: 0, to: 14 })
    expect(buckets[1]).toMatchObject({ key: 'warning', from: 15, to: 45 })
  })

  it('drops the near bucket when the warning threshold reaches the far horizon', () => {
    const wide: ExpirationThresholds = { critical: 7, warning: TIMELINE_FAR_HORIZON_DAYS + 10 }
    const buckets = buildExpiryTimeline([cert({ expiresAt: expiresInDays(200.5) })], wide, NOW)
    expect(buckets.map((bucket) => bucket.key)).toEqual(['critical', 'warning', 'later'])
    expect(buckets.at(-1)).toMatchObject({ key: 'later', from: TIMELINE_FAR_HORIZON_DAYS + 11, to: null, count: 1 })
  })
})
