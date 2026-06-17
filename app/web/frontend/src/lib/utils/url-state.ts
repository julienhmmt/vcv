import type { CertStatus } from '$lib/types'
import type { CertTypeFilter, SortDirection, SortKey } from '$lib/utils/cert-filter'

/** Serializable view state synced to the URL query string. */
export interface UrlState {
  search: string
  statusFilters: CertStatus[]
  certTypeFilter: CertTypeFilter
  mountFilter: string[] | null
  sortKey: SortKey
  sortDir: SortDirection
  pageSize: number | 'all'
  pageIndex: number
}

const VALID_STATUS: readonly CertStatus[] = ['valid', 'warning', 'critical', 'expired', 'revoked']
const VALID_CERT_TYPE: readonly CertTypeFilter[] = ['all', 'machine', 'user', 'both', 'unknown']
const VALID_SORT_KEY: readonly SortKey[] = ['commonName', 'expiresAt', 'vault', 'pki']
const VALID_SORT_DIR: readonly SortDirection[] = ['asc', 'desc']
const VALID_PAGE_SIZE: readonly number[] = [25, 50, 100]

/** Read view state from the current URL, falling back to `defaults` for absent/invalid params. */
export function parseUrlState(defaults: UrlState, searchString = window.location.search): UrlState {
  const params = new URLSearchParams(searchString)

  const statusRaw = (params.get('status') ?? '').split(',').filter(Boolean)
  const statusFilters = statusRaw.filter((s): s is CertStatus => VALID_STATUS.includes(s as CertStatus))

  const certType = params.get('type')
  const certTypeFilter = VALID_CERT_TYPE.includes(certType as CertTypeFilter)
    ? (certType as CertTypeFilter)
    : defaults.certTypeFilter

  const mountsRaw = params.get('mounts')
  const mountFilter =
    mountsRaw === null ? defaults.mountFilter : mountsRaw.split(',').map(decodeURIComponent).filter(Boolean)

  const sortKey = params.get('sort')
  const sortDir = params.get('dir')

  const sizeRaw = params.get('size')
  let pageSize: number | 'all' = defaults.pageSize
  if (sizeRaw === 'all') pageSize = 'all'
  else if (sizeRaw !== null && VALID_PAGE_SIZE.includes(Number(sizeRaw))) pageSize = Number(sizeRaw)

  const pageRaw = Number(params.get('page'))
  const pageIndex = Number.isInteger(pageRaw) && pageRaw > 0 ? pageRaw - 1 : defaults.pageIndex

  return {
    search: params.get('q') ?? defaults.search,
    statusFilters: statusFilters.length > 0 ? statusFilters : defaults.statusFilters,
    certTypeFilter,
    mountFilter,
    sortKey: VALID_SORT_KEY.includes(sortKey as SortKey) ? (sortKey as SortKey) : defaults.sortKey,
    sortDir: VALID_SORT_DIR.includes(sortDir as SortDirection) ? (sortDir as SortDirection) : defaults.sortDir,
    pageSize,
    pageIndex,
  }
}

/** Build a query string from view state, omitting params left at their default. */
export function serializeUrlState(state: UrlState, defaults: UrlState): string {
  const params = new URLSearchParams()

  if (state.search) params.set('q', state.search)
  if (state.statusFilters.length > 0) params.set('status', state.statusFilters.join(','))
  if (state.certTypeFilter !== defaults.certTypeFilter) params.set('type', state.certTypeFilter)
  if (state.mountFilter !== null) params.set('mounts', state.mountFilter.map(encodeURIComponent).join(','))
  if (state.sortKey !== defaults.sortKey) params.set('sort', state.sortKey)
  if (state.sortDir !== defaults.sortDir) params.set('dir', state.sortDir)
  if (state.pageSize !== defaults.pageSize) params.set('size', String(state.pageSize))
  if (state.pageIndex > 0) params.set('page', String(state.pageIndex + 1))

  return params.toString()
}

/** Replace the URL query string in place without adding a history entry. */
export function writeUrlState(state: UrlState, defaults: UrlState): void {
  const query = serializeUrlState(state, defaults)
  const url = query ? `${window.location.pathname}?${query}` : window.location.pathname
  window.history.replaceState(window.history.state, '', url)
}
