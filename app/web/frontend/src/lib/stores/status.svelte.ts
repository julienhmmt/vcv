import { api, ApiError } from '$lib/api'
import type { StatusResponse } from '$lib/types'

export interface StatusStore {
  readonly status: StatusResponse | null
  readonly loading: boolean
  readonly error: string | null
  refresh(): Promise<void>
}

export function createStatusStore(): StatusStore {
  let status = $state<StatusResponse | null>(null)
  let loading = $state(false)
  let error = $state<string | null>(null)
  /** Ignores out-of-order status responses when polls overlap. */
  let refreshGen = 0

  async function refresh(): Promise<void> {
    const gen = ++refreshGen
    loading = true
    error = null
    try {
      const next = await api.status()
      if (gen !== refreshGen) return
      status = next
    } catch (err: unknown) {
      if (gen !== refreshGen) return
      error = err instanceof ApiError ? err.message : 'Failed to fetch status'
      status = null
    } finally {
      if (gen === refreshGen) loading = false
    }
  }

  return {
    get status() {
      return status
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
