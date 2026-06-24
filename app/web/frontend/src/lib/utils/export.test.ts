import { describe, it, expect } from 'vitest'
import { buildExport } from '$lib/utils/export'
import type { Certificate } from '$lib/types'

function cert(overrides: Partial<Certificate> = {}): Certificate {
  return {
    id: 'vault1|pki-int:0a:1b',
    serialNumber: '0a:1b',
    commonName: 'example.com',
    sans: ['example.com', 'www.example.com'],
    certType: 'machine',
    createdAt: '2024-01-01T00:00:00Z',
    expiresAt: '2999-01-01T00:00:00Z', // far future → valid
    revoked: false,
    ...overrides,
  }
}

describe('buildExport CSV', () => {
  it('emits a header and one row per certificate with derived fields', () => {
    const { content, mime, extension } = buildExport([cert()], 'csv')
    expect(mime).toContain('text/csv')
    expect(extension).toBe('csv')
    const [header, row] = content.split('\r\n')
    expect(header).toBe('commonName,sans,vault,mount,certType,status,expiresAt,serialNumber')
    expect(row).toBe('example.com,example.com www.example.com,vault1,pki-int,machine,valid,2999-01-01T00:00:00Z,0a:1b')
  })

  it('derives revoked and expired status', () => {
    const rows = buildExport(
      [
        cert({ id: 'v|m:1', serialNumber: '1', commonName: 'r', revoked: true }),
        cert({ id: 'v|m:2', serialNumber: '2', commonName: 'e', expiresAt: '2000-01-01T00:00:00Z' }),
      ],
      'csv',
    ).content.split('\r\n')
    expect(rows[1]).toContain(',revoked,')
    expect(rows[2]).toContain(',expired,')
  })

  it('escapes commas and quotes per RFC 4180', () => {
    const { content } = buildExport([cert({ commonName: 'a,b "c"' })], 'csv')
    const row = content.split('\r\n')[1]
    expect(row.startsWith('"a,b ""c"""')).toBe(true)
  })

  it('neutralizes cells that start with a formula trigger character', () => {
    const dangerous = ['=HYPERLINK("http://evil","x")', '+1+2', '-2+3|cmd', '@SUM(A1:A2)']
    for (const cn of dangerous) {
      const row = buildExport([cert({ commonName: cn })], 'csv').content.split('\r\n')[1]
      expect(row.startsWith("'")).toBe(true)
    }
  })

  it('neutralizes a formula hidden behind leading whitespace', () => {
    const row = buildExport([cert({ commonName: ' =cmd' })], 'csv').content.split('\r\n')[1]
    expect(row.startsWith("' =cmd") || row.startsWith('"\' =cmd')).toBe(true)
  })

  it('does not alter safe values', () => {
    const row = buildExport([cert({ commonName: 'example.com' })], 'csv').content.split('\r\n')[1]
    expect(row.startsWith("'")).toBe(false)
    expect(row).toBe('example.com,example.com www.example.com,vault1,pki-int,machine,valid,2999-01-01T00:00:00Z,0a:1b')
  })

  it('sanitizes SAN values too, not just commonName', () => {
    const row = buildExport([cert({ sans: ['=cmd|"/c calc"!A1', 'safe.com'] })], 'csv').content.split('\r\n')[1]
    // SANs join with a space into the second cell; the leading = triggers the guard,
    // then RFC 4180 quoting wraps the cell because it contains a comma/quote.
    expect(row.split(',')[1].startsWith('"\'=') || row.split(',')[1].startsWith("'=")).toBe(true)
  })
})

describe('buildExport JSON', () => {
  it('emits a JSON array of flattened rows', () => {
    const { content, mime, extension } = buildExport([cert()], 'json')
    expect(mime).toBe('application/json')
    expect(extension).toBe('json')
    const parsed = JSON.parse(content)
    expect(parsed).toHaveLength(1)
    expect(parsed[0]).toMatchObject({
      commonName: 'example.com',
      sans: 'example.com www.example.com',
      vault: 'vault1',
      mount: 'pki-int',
      certType: 'machine',
      status: 'valid',
      serialNumber: '0a:1b',
    })
  })

  it('produces an empty array for no certificates', () => {
    expect(JSON.parse(buildExport([], 'json').content)).toEqual([])
  })
})
