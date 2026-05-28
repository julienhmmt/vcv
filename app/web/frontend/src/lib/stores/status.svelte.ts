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

  async function refresh(): Promise<void> {
    loading = true
    error = null
    try {
      status = await api.status()
    } catch (err: unknown) {
      error = err instanceof ApiError ? err.message : 'Failed to fetch status'
      status = null
    } finally {
      loading = false
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
