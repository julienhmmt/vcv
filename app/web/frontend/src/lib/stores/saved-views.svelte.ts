import type { CertStatus } from '$lib/types'
import type { CertTypeFilter, SortDirection, SortKey } from '$lib/utils/cert-filter'

const STORAGE_KEY = 'vcv-saved-views'
const MAX_VIEWS = 10

/** Filter/sort state captured by a saved view. Page and open certificate are excluded. */
export interface ViewSnapshot {
  search: string
  statusFilters: CertStatus[]
  certTypeFilter: CertTypeFilter
  mountFilter: string[] | null
  sortKey: SortKey
  sortDir: SortDirection
  pageSize: number | 'all'
  minDays: number | null
  maxDays: number | null
}

export interface SavedView {
  id: string
  name: string
  snapshot: ViewSnapshot
}

export interface SavedViewsStore {
  readonly views: SavedView[]
  save(name: string, snapshot: ViewSnapshot): void
  remove(id: string): void
}

function isSnapshot(value: unknown): value is ViewSnapshot {
  if (typeof value !== 'object' || value === null) return false
  const snapshot = value as Partial<ViewSnapshot>
  return (
    typeof snapshot.search === 'string' &&
    Array.isArray(snapshot.statusFilters) &&
    typeof snapshot.certTypeFilter === 'string' &&
    (snapshot.mountFilter === null || Array.isArray(snapshot.mountFilter)) &&
    typeof snapshot.sortKey === 'string' &&
    typeof snapshot.sortDir === 'string' &&
    (typeof snapshot.pageSize === 'number' || snapshot.pageSize === 'all') &&
    (snapshot.minDays === null || typeof snapshot.minDays === 'number') &&
    (snapshot.maxDays === null || typeof snapshot.maxDays === 'number')
  )
}

/** Read saved views, dropping anything malformed. Storage can throw; treat that as empty. */
function readStoredViews(): SavedView[] {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    const parsed: unknown = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed.filter(
      (entry): entry is SavedView =>
        typeof entry === 'object' &&
        entry !== null &&
        typeof (entry as SavedView).id === 'string' &&
        typeof (entry as SavedView).name === 'string' &&
        isSnapshot((entry as SavedView).snapshot),
    )
  } catch {
    return []
  }
}

function persistViews(views: SavedView[]): void {
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(views))
  } catch {
    // Ignore: persistence is best-effort.
  }
}

function newId(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
}

export function createSavedViewsStore(): SavedViewsStore {
  let views = $state<SavedView[]>(readStoredViews())

  function save(name: string, snapshot: ViewSnapshot): void {
    const trimmed = name.trim()
    if (!trimmed) return
    const existing = views.find((view) => view.name === trimmed)
    if (existing) {
      views = views.map((view) => (view.id === existing.id ? { ...view, snapshot } : view))
    } else {
      const next = [...views, { id: newId(), name: trimmed, snapshot }]
      views = next.slice(-MAX_VIEWS)
    }
    persistViews(views)
  }

  function remove(id: string): void {
    views = views.filter((view) => view.id !== id)
    persistViews(views)
  }

  return {
    get views() {
      return views
    },
    save,
    remove,
  }
}
