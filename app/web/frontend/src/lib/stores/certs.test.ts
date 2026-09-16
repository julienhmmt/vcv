import { describe, it, expect, vi, beforeEach } from 'vitest'
import type { Certificate, CertificatesEnvelope } from '$lib/types'

const { listCertificates, ApiError } = vi.hoisted(() => {
  class ApiError extends Error {
    status: number
    constructor(status: number, message: string) {
      super(message)
      this.status = status
      this.name = 'ApiError'
    }
  }
  return {
    listCertificates: vi.fn(),
    ApiError,
  }
})

vi.mock('$lib/api', () => ({
  api: { listCertificates },
  ApiError,
}))

import { createCertsStore } from '$lib/stores/certs.svelte'
import type { I18nStore } from '$lib/stores/i18n.svelte'

const i18n = {
  t: (_key: string, fallback?: string) => fallback ?? _key,
} as unknown as I18nStore

function sampleCert(id: string): Certificate {
  return {
    id,
    serialNumber: '1',
    commonName: id,
    sans: [],
    certType: 'machine',
    createdAt: '2024-01-01T00:00:00Z',
    expiresAt: '2030-01-01T00:00:00Z',
    revoked: false,
  }
}

function envelopeFor(ids: string[]): CertificatesEnvelope {
  return {
    certificates: ids.map(sampleCert),
    errors: [],
    total: ids.length,
    page: 1,
    page_size: ids.length,
    total_pages: 1,
    counts: { valid: ids.length, warning: 0, critical: 0, expired: 0, revoked: 0, total: ids.length },
  }
}

function deferred<T>(): {
  promise: Promise<T>
  resolve: (value: T) => void
  reject: (err: unknown) => void
} {
  let resolve!: (value: T) => void
  let reject!: (err: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

beforeEach(() => {
  listCertificates.mockReset()
})

describe('createCertsStore', () => {
  it('loads the table page and total on refreshTable', async () => {
    const envelope = envelopeFor(['a'])
    listCertificates.mockResolvedValueOnce(envelope)
    const store = createCertsStore(i18n)
    await store.refreshTable({ page: 1, pageSize: 25 })
    expect(listCertificates).toHaveBeenCalledWith({ page: 1, pageSize: 25 })
    expect(store.certificates).toEqual(envelope.certificates)
    expect(store.total).toBe(1)
    expect(store.totalPages).toBe(1)
    expect(store.error).toBeNull()
    expect(store.loading).toBe(false)
  })

  it('loads the full inventory on refreshInventory', async () => {
    const envelope = envelopeFor(['a', 'b'])
    listCertificates.mockResolvedValueOnce(envelope)
    const store = createCertsStore(i18n)
    await store.refreshInventory(null)
    expect(listCertificates).toHaveBeenCalledWith({ pageSize: 'all' })
    expect(store.inventory).toEqual(envelope.certificates)
    expect(store.loading).toBe(false)
  })

  it('passes mount scope through to refreshInventory', async () => {
    listCertificates.mockResolvedValueOnce(envelopeFor(['a']))
    const store = createCertsStore(i18n)
    await store.refreshInventory(['vault-a|pki'])
    expect(listCertificates).toHaveBeenCalledWith({ mounts: ['vault-a|pki'], pageSize: 'all' })
  })

  it('ignores a stale slower response after a newer refreshTable wins', async () => {
    const first = deferred<CertificatesEnvelope>()
    const second = deferred<CertificatesEnvelope>()
    listCertificates.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise)

    const store = createCertsStore(i18n)
    const p1 = store.refreshTable({})
    const p2 = store.refreshTable({})

    second.resolve(envelopeFor(['newer']))
    await p2
    expect(store.certificates.map((c) => c.id)).toEqual(['newer'])
    expect(store.loading).toBe(false)

    first.resolve(envelopeFor(['stale']))
    await p1
    expect(store.certificates.map((c) => c.id)).toEqual(['newer'])
    expect(store.loading).toBe(false)
  })

  it('does not apply a stale error after a successful newer refreshTable', async () => {
    const first = deferred<CertificatesEnvelope>()
    const second = deferred<CertificatesEnvelope>()
    listCertificates.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise)

    const store = createCertsStore(i18n)
    const p1 = store.refreshTable({})
    const p2 = store.refreshTable({})

    second.resolve(envelopeFor(['ok']))
    await p2
    expect(store.certificates.map((c) => c.id)).toEqual(['ok'])
    expect(store.error).toBeNull()

    first.reject(new ApiError(500, 'stale failure'))
    await p1
    expect(store.certificates.map((c) => c.id)).toEqual(['ok'])
    expect(store.error).toBeNull()
    expect(store.loading).toBe(false)
  })

  it('keeps loading true until the latest in-flight refreshTable finishes', async () => {
    const first = deferred<CertificatesEnvelope>()
    const second = deferred<CertificatesEnvelope>()
    listCertificates.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise)

    const store = createCertsStore(i18n)
    const p1 = store.refreshTable({})
    const p2 = store.refreshTable({})
    expect(store.loading).toBe(true)

    first.resolve(envelopeFor(['stale']))
    await p1
    expect(store.loading).toBe(true)

    second.resolve(envelopeFor(['latest']))
    await p2
    expect(store.loading).toBe(false)
    expect(store.certificates.map((c) => c.id)).toEqual(['latest'])
  })
})
