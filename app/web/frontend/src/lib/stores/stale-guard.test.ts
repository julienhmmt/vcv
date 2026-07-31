import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest'
import type { I18nResponse, PublicConfigResponse, StatusResponse } from '$lib/types'

const { configFn, statusFn, i18nFn, ApiError } = vi.hoisted(() => {
  class ApiError extends Error {
    status: number
    constructor(status: number, message: string) {
      super(message)
      this.status = status
      this.name = 'ApiError'
    }
  }
  return {
    configFn: vi.fn(),
    statusFn: vi.fn(),
    i18nFn: vi.fn(),
    ApiError,
  }
})

vi.mock('$lib/api', () => ({
  api: { config: configFn, status: statusFn, i18n: i18nFn },
  ApiError,
}))

import { createConfigStore } from '$lib/stores/config.svelte'
import { createStatusStore } from '$lib/stores/status.svelte'
import { createI18nStore } from '$lib/stores/i18n.svelte'

interface AsyncStore {
  readonly loading: boolean
  readonly error: string | null
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

/**
 * The contract every generation-guarded store shares: a slower earlier request
 * must never overwrite the value, the error, or the loading flag owned by a
 * newer one. Mirrors the cases certs.test.ts already covers for the certs store.
 */
function staleGuardSuite<S extends AsyncStore, R>(opts: {
  label: string
  mock: Mock
  /** Builds the store and a zero-arg trigger for one fetch. */
  setup: () => { store: S; fetch: () => Promise<void> }
  newer: R
  stale: R
  read: (store: S) => unknown
  newerValue: unknown
}): void {
  describe(opts.label, () => {
    it('ignores a stale slower response after a newer request wins', async () => {
      const { store, fetch } = opts.setup()
      const first = deferred<R>()
      const second = deferred<R>()
      opts.mock.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise)

      const p1 = fetch()
      const p2 = fetch()

      second.resolve(opts.newer)
      await p2
      expect(opts.read(store)).toEqual(opts.newerValue)

      first.resolve(opts.stale)
      await p1
      expect(opts.read(store)).toEqual(opts.newerValue)
      expect(store.loading).toBe(false)
    })

    it('does not apply a stale error after a successful newer request', async () => {
      const { store, fetch } = opts.setup()
      const first = deferred<R>()
      const second = deferred<R>()
      opts.mock.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise)

      const p1 = fetch()
      const p2 = fetch()

      second.resolve(opts.newer)
      await p2
      expect(store.error).toBeNull()

      first.reject(new ApiError(500, 'stale failure'))
      await p1
      expect(opts.read(store)).toEqual(opts.newerValue)
      expect(store.error).toBeNull()
      expect(store.loading).toBe(false)
    })

    it('keeps loading true until the latest in-flight request finishes', async () => {
      const { store, fetch } = opts.setup()
      const first = deferred<R>()
      const second = deferred<R>()
      opts.mock.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise)

      const p1 = fetch()
      const p2 = fetch()
      expect(store.loading).toBe(true)

      first.resolve(opts.stale)
      await p1
      expect(store.loading).toBe(true)

      second.resolve(opts.newer)
      await p2
      expect(store.loading).toBe(false)
      expect(opts.read(store)).toEqual(opts.newerValue)
    })
  })
}

beforeEach(() => {
  configFn.mockReset()
  statusFn.mockReset()
  i18nFn.mockReset()
})

staleGuardSuite<ReturnType<typeof createConfigStore>, PublicConfigResponse>({
  label: 'createConfigStore',
  mock: configFn,
  setup: () => {
    const store = createConfigStore()
    return { store, fetch: () => store.refresh() }
  },
  newer: { expirationThresholds: { critical: 3, warning: 14 } },
  stale: { expirationThresholds: { critical: 9, warning: 99 } },
  read: (store) => store.thresholds,
  newerValue: { critical: 3, warning: 14 },
})

staleGuardSuite<ReturnType<typeof createStatusStore>, StatusResponse>({
  label: 'createStatusStore',
  mock: statusFn,
  setup: () => {
    const store = createStatusStore()
    return { store, fetch: () => store.refresh() }
  },
  newer: { version: 'newer', vault_connected: true, vaults: [] },
  stale: { version: 'stale', vault_connected: false, vaults: [] },
  read: (store) => store.status?.version,
  newerValue: 'newer',
})

staleGuardSuite<ReturnType<typeof createI18nStore>, I18nResponse>({
  label: 'createI18nStore',
  mock: i18nFn,
  setup: () => {
    // The constructor kicks off its own load; give it a settled response so the
    // deferreds the suite queues belong to the two overlapping setLang calls.
    i18nFn.mockResolvedValueOnce({ language: 'en', messages: {} })
    const store = createI18nStore()
    const langs = ['fr', 'de']
    let call = 0
    return { store, fetch: () => store.setLang(langs[call++]) }
  },
  newer: { language: 'de', messages: { greeting: 'Hallo' } },
  stale: { language: 'fr', messages: { greeting: 'Bonjour' } },
  read: (store) => store.messages.greeting,
  newerValue: 'Hallo',
})
