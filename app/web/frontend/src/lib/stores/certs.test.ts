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
  it('loads certificates on refresh', async () => {
    const envelope: CertificatesEnvelope = {
      certificates: [sampleCert('a')],
      errors: [],
    }
    listCertificates.mockResolvedValueOnce(envelope)
    const store = createCertsStore(i18n)
    await store.refresh()
    expect(store.certificates).toEqual(envelope.certificates)
    expect(store.error).toBeNull()
    expect(store.loading).toBe(false)
  })

  it('ignores a stale slower response after a newer refresh wins', async () => {
    const first = deferred<CertificatesEnvelope>()
    const second = deferred<CertificatesEnvelope>()
    listCertificates.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise)

    const store = createCertsStore(i18n)
    const p1 = store.refresh()
    const p2 = store.refresh()

    second.resolve({
      certificates: [sampleCert('newer')],
      errors: [],
    })
    await p2
    expect(store.certificates.map((c) => c.id)).toEqual(['newer'])
    expect(store.loading).toBe(false)

    first.resolve({
      certificates: [sampleCert('stale')],
      errors: [],
    })
    await p1
    expect(store.certificates.map((c) => c.id)).toEqual(['newer'])
    expect(store.loading).toBe(false)
  })

  it('does not apply a stale error after a successful newer refresh', async () => {
    const first = deferred<CertificatesEnvelope>()
    const second = deferred<CertificatesEnvelope>()
    listCertificates.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise)

    const store = createCertsStore(i18n)
    const p1 = store.refresh()
    const p2 = store.refresh()

    second.resolve({
      certificates: [sampleCert('ok')],
      errors: [],
    })
    await p2
    expect(store.certificates.map((c) => c.id)).toEqual(['ok'])
    expect(store.error).toBeNull()

    first.reject(new ApiError(500, 'stale failure'))
    await p1
    expect(store.certificates.map((c) => c.id)).toEqual(['ok'])
    expect(store.error).toBeNull()
    expect(store.loading).toBe(false)
  })

  it('keeps loading true until the latest in-flight refresh finishes', async () => {
    const first = deferred<CertificatesEnvelope>()
    const second = deferred<CertificatesEnvelope>()
    listCertificates.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise)

    const store = createCertsStore(i18n)
    const p1 = store.refresh()
    const p2 = store.refresh()
    expect(store.loading).toBe(true)

    first.resolve({ certificates: [sampleCert('stale')], errors: [] })
    await p1
    expect(store.loading).toBe(true)

    second.resolve({ certificates: [sampleCert('latest')], errors: [] })
    await p2
    expect(store.loading).toBe(false)
    expect(store.certificates.map((c) => c.id)).toEqual(['latest'])
  })
})
