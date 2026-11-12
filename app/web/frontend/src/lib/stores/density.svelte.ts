const STORAGE_KEY = 'vcv-density'
const COMPACT_CLASS = 'compact'

export type Density = 'comfortable' | 'compact'

export interface DensityStore {
  readonly compact: boolean
  toggle(): void
}

/** localStorage can throw (private browsing, quota); fall back to comfortable. */
function readStoredDensity(): Density | null {
  try {
    const stored = window.localStorage.getItem(STORAGE_KEY)
    return stored === 'compact' || stored === 'comfortable' ? stored : null
  } catch {
    return null
  }
}

function persistDensity(density: Density): void {
  try {
    window.localStorage.setItem(STORAGE_KEY, density)
  } catch {
    // Ignore: persistence is best-effort.
  }
}

export function createDensityStore(): DensityStore {
  let density = $state<Density>(readStoredDensity() ?? 'comfortable')

  function toggle(): void {
    density = density === 'compact' ? 'comfortable' : 'compact'
    persistDensity(density)
  }

  return {
    get compact() {
      return density === COMPACT_CLASS
    },
    toggle,
  }
}
