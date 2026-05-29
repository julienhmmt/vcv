import { api, ApiError } from '$lib/api'
import type { Certificate, VaultListError } from '$lib/types'

export interface CertsStore {
  readonly certificates: Certificate[]
  readonly vaultErrors: VaultListError[]
  readonly loading: boolean
  readonly error: string | null
  readonly lastFetched: Date | null
  refresh(mounts?: string[]): Promise<void>
}

export function createCertsStore(): CertsStore {
  let certificates = $state<Certificate[]>([])
  let vaultErrors = $state<VaultListError[]>([])
  let loading = $state(false)
  let error = $state<string | null>(null)
  let lastFetched = $state<Date | null>(null)

  async function refresh(mounts?: string[]): Promise<void> {
    loading = true
    error = null
    try {
      const envelope = await api.listCertificates(mounts)
      certificates = envelope.certificates ?? []
      vaultErrors = envelope.errors ?? []
      lastFetched = new Date()
    } catch (err: unknown) {
      const message = err instanceof ApiError ? err.message : 'Failed to load certificates'
      error = message
      certificates = []
      vaultErrors = []
    } finally {
      loading = false
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
