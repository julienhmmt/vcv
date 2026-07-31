const STORAGE_KEY = 'vcv-theme'
const DARK_CLASS = 'dark'
const THEME_ATTR = 'data-theme'

export type Theme = 'light' | 'dark'

export interface ThemeStore {
  readonly theme: Theme
  toggle(): void
  set(theme: Theme): void
}

/** localStorage can throw (private browsing, quota); fall back to defaults. */
function readStoredTheme(): Theme | null {
  try {
    const stored = window.localStorage.getItem(STORAGE_KEY)
    return stored === 'light' || stored === 'dark' ? stored : null
  } catch {
    return null
  }
}

function persistTheme(theme: Theme): void {
  try {
    window.localStorage.setItem(STORAGE_KEY, theme)
  } catch {
    // Ignore: persistence is best-effort.
  }
}

function detectInitial(): Theme {
  if (typeof window === 'undefined') return 'light'
  const stored = readStoredTheme()
  if (stored) return stored
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function applyToDocument(theme: Theme): void {
  if (typeof document === 'undefined') return
  document.documentElement.classList.toggle(DARK_CLASS, theme === 'dark')
  document.documentElement.setAttribute(THEME_ATTR, theme)
}

export function createThemeStore(): ThemeStore {
  const initial = detectInitial()
  let theme = $state<Theme>(initial)
  applyToDocument(initial)

  function set(next: Theme): void {
    theme = next
    applyToDocument(next)
    if (typeof window !== 'undefined') {
      persistTheme(next)
    }
  }

  function toggle(): void {
    set(theme === 'dark' ? 'light' : 'dark')
  }

  return {
    get theme() {
      return theme
    },
    set,
    toggle,
  }
}
