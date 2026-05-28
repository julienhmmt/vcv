import { api, ApiError } from '$lib/api'
import type { Certificate } from '$lib/types'

export interface CertsStore {
  readonly certificates: Certificate[]
  readonly loading: boolean
  readonly error: string | null
  readonly lastFetched: Date | null
  refresh(mounts?: string[]): Promise<void>
}

export function createCertsStore(): CertsStore {
  let certificates = $state<Certificate[]>([])
  let loading = $state(false)
  let error = $state<string | null>(null)
  let lastFetched = $state<Date | null>(null)

  async function refresh(mounts?: string[]): Promise<void> {
    loading = true
    error = null
    try {
      certificates = await api.listCertificates(mounts)
      lastFetched = new Date()
    } catch (err: unknown) {
      const message = err instanceof ApiError ? err.message : 'Failed to load certificates'
      error = message
      certificates = []
    } finally {
      loading = false
    }
  }

  return {
    get certificates() {
      return certificates
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
