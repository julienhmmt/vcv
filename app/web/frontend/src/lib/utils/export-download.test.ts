// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { downloadExport } from '$lib/utils/export'
import type { Certificate } from '$lib/types'

const sample: Certificate = {
  id: 'v|m:1',
  serialNumber: '1',
  commonName: 'c',
  sans: [],
  certType: 'machine',
  createdAt: '2024-01-01T00:00:00Z',
  expiresAt: '2999-01-01T00:00:00Z',
  revoked: false,
}

describe('downloadExport', () => {
  let createObjectURL: ReturnType<typeof vi.fn>
  let revokeObjectURL: ReturnType<typeof vi.fn>
  let clickSpy: ReturnType<typeof vi.spyOn>
  const captured: { href?: string; download?: string } = {}

  beforeEach(() => {
    createObjectURL = vi.fn(() => 'blob:mock-url')
    revokeObjectURL = vi.fn()
    // jsdom does not implement the object URL store.
    URL.createObjectURL = createObjectURL as unknown as typeof URL.createObjectURL
    URL.revokeObjectURL = revokeObjectURL as unknown as typeof URL.revokeObjectURL
    // Avoid jsdom "navigation not implemented" by capturing instead of navigating.
    clickSpy = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(function (this: HTMLAnchorElement) {
      captured.href = this.href
      captured.download = this.download
    })
  })

  afterEach(() => {
    clickSpy.mockRestore()
    delete captured.href
    delete captured.download
  })

  it('creates a blob URL, clicks a dated download anchor, and revokes the URL', () => {
    downloadExport([sample], 'csv')

    expect(createObjectURL).toHaveBeenCalledOnce()
    const blob = createObjectURL.mock.calls[0][0] as Blob
    expect(blob).toBeInstanceOf(Blob)
    expect(blob.type).toContain('text/csv')

    expect(clickSpy).toHaveBeenCalledOnce()
    expect(captured.href).toBe('blob:mock-url')
    expect(captured.download).toMatch(/^vcv-certificates-\d{4}-\d{2}-\d{2}\.csv$/)

    expect(revokeObjectURL).toHaveBeenCalledWith('blob:mock-url')
    // The temporary anchor must not linger in the document.
    expect(document.querySelector('a[download]')).toBeNull()
  })

  it('uses a .json extension for JSON exports', () => {
    downloadExport([sample], 'json')
    expect(captured.download).toMatch(/\.json$/)
  })
})
