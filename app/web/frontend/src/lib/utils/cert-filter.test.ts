import { describe, it, expect } from 'vitest'
import {
  matchesFilters,
  sortCerts,
  paginate,
  dashboardCounts,
  formatDate,
  formatTime,
  type FilterState,
} from '$lib/utils/cert-filter'
import type { Certificate } from '$lib/types'

const NOW = new Date('2026-06-17T00:00:00Z')

function cert(overrides: Partial<Certificate> = {}): Certificate {
  return {
    id: 'vault1|pki-int:1',
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

const noFilters: FilterState = { search: '', statuses: [], certType: 'all', mounts: null }

describe('matchesFilters', () => {
  it('matches everything with no active filters', () => {
    expect(matchesFilters(cert(), noFilters, undefined, NOW)).toBe(true)
  })

  it('filters by status', () => {
    expect(matchesFilters(cert(), { ...noFilters, statuses: ['expired'] }, undefined, NOW)).toBe(false)
    expect(matchesFilters(cert({ revoked: true }), { ...noFilters, statuses: ['revoked'] }, undefined, NOW)).toBe(true)
  })

  it('filters by cert type', () => {
    expect(matchesFilters(cert(), { ...noFilters, certType: 'user' }, undefined, NOW)).toBe(false)
    expect(matchesFilters(cert(), { ...noFilters, certType: 'machine' }, undefined, NOW)).toBe(true)
  })

  it('searches CN, serial, and SANs case-insensitively', () => {
    expect(matchesFilters(cert(), { ...noFilters, search: 'API' }, undefined, NOW)).toBe(true)
    expect(matchesFilters(cert(), { ...noFilters, search: 'AA:BB' }, undefined, NOW)).toBe(true)
    expect(matchesFilters(cert(), { ...noFilters, search: 'nope' }, undefined, NOW)).toBe(false)
  })

  it('filters by mount key or mount name; empty selection matches nothing', () => {
    expect(matchesFilters(cert(), { ...noFilters, mounts: ['vault1|pki-int'] }, undefined, NOW)).toBe(true)
    expect(matchesFilters(cert(), { ...noFilters, mounts: ['pki-int'] }, undefined, NOW)).toBe(true)
    expect(matchesFilters(cert(), { ...noFilters, mounts: [] }, undefined, NOW)).toBe(false)
  })
})

describe('sortCerts', () => {
  const a = cert({ commonName: 'a', expiresAt: '2030-01-01T00:00:00Z' })
  const b = cert({ commonName: 'b', expiresAt: '2025-01-01T00:00:00Z' })

  it('sorts by common name and does not mutate the input', () => {
    const input = [b, a]
    const out = sortCerts(input, 'commonName', 'asc')
    expect(out.map((c) => c.commonName)).toEqual(['a', 'b'])
    expect(input.map((c) => c.commonName)).toEqual(['b', 'a'])
  })

  it('sorts by expiry descending', () => {
    expect(sortCerts([b, a], 'expiresAt', 'desc').map((c) => c.commonName)).toEqual(['a', 'b'])
  })
})

describe('paginate', () => {
  const items = [1, 2, 3, 4, 5]
  it('slices by page index and size', () => {
    expect(paginate(items, 0, 2)).toEqual([1, 2])
    expect(paginate(items, 2, 2)).toEqual([5])
  })
  it('returns everything for "all"', () => {
    expect(paginate(items, 0, 'all')).toEqual(items)
  })
})

describe('dashboardCounts', () => {
  it('tallies per status and total', () => {
    const counts = dashboardCounts(
      [cert(), cert({ revoked: true }), cert({ expiresAt: '2000-01-01T00:00:00Z' })],
      undefined,
      NOW,
    )
    expect(counts).toMatchObject({ valid: 1, revoked: 1, expired: 1, total: 3 })
  })
})

describe('formatDate / formatTime', () => {
  it('formats ISO into date and HH:MM (UTC)', () => {
    expect(formatDate('2026-06-17T08:30:45Z')).toBe('2026-06-17')
    expect(formatTime('2026-06-17T08:30:45Z')).toBe('08:30')
  })
})
