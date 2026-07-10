import { describe, it, expect } from 'vitest'
import { certDisplayName } from '$lib/utils/cert-label'
import type { Certificate } from '$lib/types'

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
