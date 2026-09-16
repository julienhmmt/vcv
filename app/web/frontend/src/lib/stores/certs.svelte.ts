import { api, ApiError } from '$lib/api'
import type { CertsQuery } from '$lib/api'
import type { I18nStore } from '$lib/stores/i18n.svelte'
import type { Certificate, CertStatusCounts, VaultListError } from '$lib/types'

export interface CertsStore {
  /** Current table page: server-filtered, server-sorted. */
  readonly certificates: Certificate[]
  /** Full inventory for dashboard, timeline, palette, export, and mount list. */
  readonly inventory: Certificate[]
  /** Filtered total across all pages (drives result counts and pagination). */
  readonly total: number
  readonly totalPages: number
  /** Server status breakdown (informational; dashboard keeps client counts). */
  readonly counts: CertStatusCounts | null
  readonly vaultErrors: VaultListError[]
  readonly loading: boolean
  readonly error: string | null
  readonly lastFetched: Date | null
  refreshTable(query: CertsQuery): Promise<void>
  refreshInventory(mounts?: string[] | null): Promise<void>
}

function loadError(err: unknown, i18n: I18nStore): string {
  return err instanceof ApiError
    ? err.message
    : i18n.t('loadNetworkError', 'Network error loading certificates. Please try again.')
}

export function createCertsStore(i18n: I18nStore): CertsStore {
  let certificates = $state<Certificate[]>([])
  let inventory = $state<Certificate[]>([])
  let total = $state(0)
  let totalPages = $state(1)
  let counts = $state<CertStatusCounts | null>(null)
  let vaultErrors = $state<VaultListError[]>([])
  let tableLoading = $state(false)
  let inventoryLoading = $state(false)
  let error = $state<string | null>(null)
  let lastFetched = $state<Date | null>(null)
  /** Ignores out-of-order responses within each pipeline independently. */
  let tableGen = 0
  let inventoryGen = 0

  async function refreshTable(query: CertsQuery): Promise<void> {
    const gen = ++tableGen
    tableLoading = true
    error = null
    try {
      const envelope = await api.listCertificates(query)
      if (gen !== tableGen) return
      certificates = envelope.certificates ?? []
      total = envelope.total ?? certificates.length
      totalPages = Math.max(1, envelope.total_pages ?? 1)
      counts = envelope.counts ?? null
      vaultErrors = envelope.errors ?? []
      lastFetched = new Date()
    } catch (err: unknown) {
      if (gen !== tableGen) return
      error = loadError(err, i18n)
      certificates = []
      total = 0
      totalPages = 1
      counts = null
      vaultErrors = []
    } finally {
      // Only the latest in-flight request may clear loading.
      if (gen === tableGen) tableLoading = false
    }
  }

  async function refreshInventory(mounts?: string[] | null): Promise<void> {
    const gen = ++inventoryGen
    inventoryLoading = true
    try {
      const envelope = await api.listCertificates({
        ...(mounts ? { mounts } : {}),
        pageSize: 'all',
      })
      if (gen !== inventoryGen) return
      inventory = envelope.certificates ?? []
    } catch (err: unknown) {
      if (gen !== inventoryGen) return
      error = loadError(err, i18n)
      inventory = []
    } finally {
      if (gen === inventoryGen) inventoryLoading = false
    }
  }

  return {
    get certificates() {
      return certificates
    },
    get inventory() {
      return inventory
    },
    get total() {
      return total
    },
    get totalPages() {
      return totalPages
    },
    get counts() {
      return counts
    },
    get vaultErrors() {
      return vaultErrors
    },
    get loading() {
      return tableLoading || inventoryLoading
    },
    get error() {
      return error
    },
    get lastFetched() {
      return lastFetched
    },
    refreshTable,
    refreshInventory,
  }
}
