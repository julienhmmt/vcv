import type { Certificate, CertStatus, ExpirationThresholds } from '$lib/types'
import { certStatus, parseCertID, DEFAULT_THRESHOLDS } from './cert-status'

export type SortKey = 'commonName' | 'expiresAt' | 'vault' | 'pki'
export type SortDirection = 'asc' | 'desc'
export type CertTypeFilter = 'all' | 'machine' | 'user' | 'both' | 'unknown'

export interface FilterState {
  search: string
  statuses: CertStatus[] // empty = all
  certType: CertTypeFilter
  mounts: string[] | null // null = all
}

export function matchesFilters(
  cert: Certificate,
  state: FilterState,
  thresholds: ExpirationThresholds = DEFAULT_THRESHOLDS,
  now: Date = new Date(),
): boolean {
  const status = certStatus(cert, thresholds, now)
  if (state.statuses.length > 0 && !state.statuses.includes(status)) return false
  if (state.certType !== 'all' && cert.certType !== state.certType) return false

  if (state.search.trim()) {
    const q = state.search.toLowerCase()
    const inCN = cert.commonName.toLowerCase().includes(q)
    const inSerial = cert.serialNumber.toLowerCase().includes(q)
    const inSans = cert.sans.some((s) => s.toLowerCase().includes(q))
    if (!inCN && !inSerial && !inSans) return false
  }

  if (state.mounts !== null) {
    if (state.mounts.length === 0) return false
    const { mountKey, mount } = parseCertID(cert.id)
    if (!state.mounts.includes(mountKey) && !state.mounts.includes(mount)) return false
  }

  return true
}

export function sortCerts(
  items: Certificate[],
  key: SortKey,
  direction: SortDirection,
): Certificate[] {
  const dir = direction === 'asc' ? 1 : -1
  const copy = [...items]
  copy.sort((a, b) => {
    let av: string | number = ''
    let bv: string | number = ''
    switch (key) {
      case 'commonName':
        av = a.commonName.toLowerCase()
        bv = b.commonName.toLowerCase()
        break
      case 'expiresAt':
        av = new Date(a.expiresAt).getTime()
        bv = new Date(b.expiresAt).getTime()
        break
      case 'vault':
        av = parseCertID(a.id).vault.toLowerCase()
        bv = parseCertID(b.id).vault.toLowerCase()
        break
      case 'pki':
        av = parseCertID(a.id).mount.toLowerCase()
        bv = parseCertID(b.id).mount.toLowerCase()
        break
    }
    if (av < bv) return -1 * dir
    if (av > bv) return 1 * dir
    return 0
  })
  return copy
}

export function paginate<T>(items: T[], pageIndex: number, pageSize: number | 'all'): T[] {
  if (pageSize === 'all') return items
  const start = pageIndex * pageSize
  return items.slice(start, start + pageSize)
}

export interface DashboardCounts {
  valid: number
  warning: number
  critical: number
  expired: number
  revoked: number
  total: number
}

export function dashboardCounts(
  certs: Certificate[],
  thresholds: ExpirationThresholds = DEFAULT_THRESHOLDS,
  now: Date = new Date(),
): DashboardCounts {
  const counts: DashboardCounts = { valid: 0, warning: 0, critical: 0, expired: 0, revoked: 0, total: 0 }
  for (const cert of certs) {
    const s = certStatus(cert, thresholds, now)
    counts[s] += 1
    counts.total += 1
  }
  return counts
}

function isValidDate(date: Date): boolean {
  return !Number.isNaN(date.getTime())
}

/** Format an ISO date string as YYYY-MM-DD; returns '—' for invalid input. */
export function formatDate(iso: string): string {
  const date = new Date(iso)
  if (!isValidDate(date)) return '—'
  return date.toISOString().split('T')[0]
}

/** Format an ISO date string as HH:MM (UTC); returns '—' for invalid input. */
export function formatTime(iso: string): string {
  const date = new Date(iso)
  if (!isValidDate(date)) return '—'
  return date.toISOString().split('T')[1].slice(0, 5)
}
