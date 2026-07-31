import { describe, it, expect } from 'vitest'
import { certDisplayName, certExpiryLabel, type TranslateFn } from '$lib/utils/cert-label'
import type { Certificate } from '$lib/types'

/** English fallback translator that interpolates `{name}` placeholders. */
const en: TranslateFn = (key, fallback = key, params) =>
  Object.entries(params ?? {}).reduce(
    (acc, [name, value]) => acc.replaceAll(`{${name}}`, String(value)),
    fallback,
  )

function expiresInDays(days: number): string {
  return new Date(Date.now() + days * 86_400_000).toISOString()
}

function cert(overrides: Partial<Certificate> = {}): Certificate {
  return {
    id: 'v|m:1',
    serialNumber: 'aa:bb',
    commonName: 'web.example.com',
    sans: ['web.example.com', 'api.example.com'],
    certType: 'machine',
    createdAt: '2024-01-01T00:00:00Z',
    expiresAt: '2999-01-01T00:00:00Z',
    revoked: false,
    ...overrides,
  }
}

describe('certDisplayName', () => {
  it('returns the common name when present', () => {
    expect(certDisplayName(cert(), 'unnamed')).toBe('web.example.com')
  })

  it('falls back to the first SAN when common name is empty', () => {
    expect(certDisplayName(cert({ commonName: '', sans: ['api.example.com'] }), 'unnamed')).toBe('api.example.com')
  })

  it('falls back to the serial number when common name and SANs are empty', () => {
    expect(certDisplayName(cert({ commonName: '', sans: [], serialNumber: 'aa:bb' }), 'unnamed')).toBe('aa:bb')
  })

  it('falls back to the provided string when everything else is empty', () => {
    expect(certDisplayName(cert({ commonName: '', sans: [], serialNumber: '' }), 'unnamed')).toBe('unnamed')
  })

  it('treats a whitespace-only common name as empty', () => {
    expect(certDisplayName(cert({ commonName: '   ', sans: ['api.example.com'] }), 'unnamed')).toBe('api.example.com')
  })
})

describe('certExpiryLabel', () => {
  const cases: { name: string; expiresAt: string; expected: string }[] = [
    { name: 'compact days ahead while valid', expiresAt: expiresInDays(5.5), expected: '5d' },
    { name: 'descriptive when expiring today', expiresAt: expiresInDays(0.5), expected: 'Expires today' },
    { name: 'singular when expired one day ago', expiresAt: expiresInDays(-0.5), expected: 'Expired 1 day ago' },
    { name: 'plural when expired several days ago', expiresAt: expiresInDays(-4.5), expected: 'Expired 5 days ago' },
  ]

  for (const { name, expiresAt, expected } of cases) {
    it(name, () => {
      expect(certExpiryLabel(cert({ expiresAt }), en)).toBe(expected)
    })
  }

  it('uses the translator for localized output', () => {
    const fr: TranslateFn = (key, fallback = key, params) =>
      ({ daysRemainingShort: '{days}j', expiringToday: "Expire aujourd'hui", expiredDays: 'Expiré depuis {days} jours', expiredDaysSingular: 'Expiré depuis {days} jour' })[key]
        ?.replace('{days}', String(params?.days ?? '')) ?? fallback
    expect(certExpiryLabel(cert({ expiresAt: expiresInDays(3.5) }), fr)).toBe('3j')
    expect(certExpiryLabel(cert({ expiresAt: expiresInDays(-0.5) }), fr)).toBe('Expiré depuis 1 jour')
  })
})
