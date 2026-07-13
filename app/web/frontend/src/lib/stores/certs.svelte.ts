import { api, ApiError } from '$lib/api'
import type { I18nStore } from '$lib/stores/i18n.svelte'
import type { Certificate, VaultListError } from '$lib/types'

export interface CertsStore {
  readonly certificates: Certificate[]
  readonly vaultErrors: VaultListError[]
  readonly loading: boolean
  readonly error: string | null
  readonly lastFetched: Date | null
  refresh(mounts?: string[]): Promise<void>
}

export function createCertsStore(i18n: I18nStore): CertsStore {
  let certificates = $state<Certificate[]>([])
  let vaultErrors = $state<VaultListError[]>([])
  let loading = $state(false)
  let error = $state<string | null>(null)
  let lastFetched = $state<Date | null>(null)
  /** Ignores out-of-order list responses when refreshes overlap. */
  let refreshGen = 0

  async function refresh(mounts?: string[]): Promise<void> {
    const gen = ++refreshGen
    loading = true
    error = null
    try {
      const envelope = await api.listCertificates(mounts)
      if (gen !== refreshGen) return
      certificates = envelope.certificates ?? []
      vaultErrors = envelope.errors ?? []
      lastFetched = new Date()
    } catch (err: unknown) {
      if (gen !== refreshGen) return
      error =
        err instanceof ApiError
          ? err.message
          : i18n.t('loadNetworkError', 'Network error loading certificates. Please try again.')
      certificates = []
      vaultErrors = []
    } finally {
      // Only the latest in-flight request may clear loading.
      if (gen === refreshGen) loading = false
    }
  }

  return {
    get certificates() {
      return certificates
    },
    get vaultErrors() {
      return vaultErrors
    },
    get loading() {
      return loading
    },
    get error() {
      return error
    },
    get lastFetched() {
      return lastFetched
    },
    refresh,
  }
}
