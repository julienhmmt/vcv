import { describe, it, expect } from 'vitest'
import {
  daysUntilExpiry,
  certStatus,
  statusBadgeClass,
  rowClassForStatus,
  parseCertID,
} from '$lib/utils/cert-status'
import type { Certificate } from '$lib/types'

const NOW = new Date('2026-06-17T00:00:00Z')

function cert(expiresAt: string, revoked = false): Certificate {
  return {
    id: 'v|m:1',
    serialNumber: '1',
    commonName: 'c',
    sans: [],
    certType: 'machine',
    createdAt: '2024-01-01T00:00:00Z',
    expiresAt,
    revoked,
  }
}

describe('daysUntilExpiry', () => {
  it('counts whole days ahead and behind', () => {
    expect(daysUntilExpiry(cert('2026-06-27T00:00:00Z'), NOW)).toBe(10)
    expect(daysUntilExpiry(cert('2026-06-16T00:00:00Z'), NOW)).toBe(-1)
  })
})

describe('certStatus', () => {
  it('classifies by threshold with default 7/30', () => {
    expect(certStatus(cert('2026-09-01T00:00:00Z'), undefined, NOW)).toBe('valid') // >30d
    expect(certStatus(cert('2026-07-01T00:00:00Z'), undefined, NOW)).toBe('warning') // 14d
    expect(certStatus(cert('2026-06-20T00:00:00Z'), undefined, NOW)).toBe('critical') // 3d
    expect(certStatus(cert('2026-06-10T00:00:00Z'), undefined, NOW)).toBe('expired') // past
  })

  it('revoked takes precedence over expiry', () => {
    expect(certStatus(cert('2026-09-01T00:00:00Z', true), undefined, NOW)).toBe('revoked')
  })

  it('treats the threshold boundary as inclusive', () => {
    expect(certStatus(cert('2026-06-24T00:00:00Z'), undefined, NOW)).toBe('critical') // exactly 7d
    expect(certStatus(cert('2026-07-17T00:00:00Z'), undefined, NOW)).toBe('warning') // exactly 30d
  })
})

describe('class helpers', () => {
  it('maps status to badge and row classes', () => {
    expect(statusBadgeClass('critical')).toBe('vcv-badge vcv-badge-critical')
    expect(rowClassForStatus('expired')).toBe('vcv-row-expired')
  })
})

describe('parseCertID', () => {
  it('splits vault|mount:serial', () => {
    expect(parseCertID('vault1|pki-int:0a:1b')).toEqual({
      vault: 'vault1',
      mount: 'pki-int',
      mountKey: 'vault1|pki-int',
    })
  })

  it('handles an id without a vault prefix', () => {
    expect(parseCertID('pki:serial')).toEqual({ vault: '', mount: 'pki', mountKey: 'pki' })
  })
})
