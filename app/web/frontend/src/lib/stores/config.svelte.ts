import { api, ApiError } from '$lib/api'
import type { ExpirationThresholds, PublicConfigResponse } from '$lib/types'
import { DEFAULT_THRESHOLDS } from '$lib/utils/cert-status'
import { thresholdsFromConfig } from '$lib/utils/config-thresholds'

type VaultEntry = NonNullable<PublicConfigResponse['vaults']>[number]

export interface ConfigStore {
  readonly thresholds: ExpirationThresholds
  /** Configured "vault|mount" keys, sorted; matches the mount key of each cert ID. */
  readonly mountKeys: string[]
  readonly loading: boolean
  readonly error: string | null
  refresh(): Promise<void>
}

export function createConfigStore(): ConfigStore {
  let thresholds = $state<ExpirationThresholds>({ ...DEFAULT_THRESHOLDS })
  let vaults = $state<VaultEntry[]>([])
  let loading = $state(false)
  let error = $state<string | null>(null)
  /** Ignores out-of-order config responses when refreshes overlap. */
  let refreshGen = 0

  async function refresh(): Promise<void> {
    const gen = ++refreshGen
    loading = true
    error = null
    try {
      const response = await api.config()
      if (gen !== refreshGen) return
      thresholds = thresholdsFromConfig(response.expirationThresholds)
      vaults = response.vaults ?? []
    } catch (err: unknown) {
      if (gen !== refreshGen) return
      error = err instanceof ApiError ? err.message : 'Failed to fetch config'
    } finally {
      if (gen === refreshGen) loading = false
    }
  }

  return {
    get thresholds() {
      return thresholds
    },
    get mountKeys() {
      return vaults.flatMap((v) => v.pkiMounts.map((m) => `${v.id}|${m}`)).sort()
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
