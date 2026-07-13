import { api, ApiError } from '$lib/api'
import type { ExpirationThresholds } from '$lib/types'
import { DEFAULT_THRESHOLDS } from '$lib/utils/cert-status'
import { thresholdsFromConfig } from '$lib/utils/config-thresholds'

export interface ConfigStore {
  readonly thresholds: ExpirationThresholds
  readonly loading: boolean
  readonly error: string | null
  refresh(): Promise<void>
}

export function createConfigStore(): ConfigStore {
  let thresholds = $state<ExpirationThresholds>({ ...DEFAULT_THRESHOLDS })
  let loading = $state(false)
  let error = $state<string | null>(null)

  async function refresh(): Promise<void> {
    loading = true
    error = null
    try {
      const response = await api.config()
      thresholds = thresholdsFromConfig(response.expirationThresholds)
    } catch (err: unknown) {
      error = err instanceof ApiError ? err.message : 'Failed to fetch config'
    } finally {
      loading = false
    }
  }

  return {
    get thresholds() {
      return thresholds
    },
    get loading() {
      return loading
    },
    get error() {
      return error
    },
    refresh,
  }
}
