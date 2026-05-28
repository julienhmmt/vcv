const STORAGE_KEY = 'vcv-theme'
const DARK_CLASS = 'dark'

export type Theme = 'light' | 'dark'

export interface ThemeStore {
  readonly theme: Theme
  toggle(): void
  set(theme: Theme): void
}

function detectInitial(): Theme {
  if (typeof window === 'undefined') return 'light'
  const stored = window.localStorage.getItem(STORAGE_KEY)
  if (stored === 'light' || stored === 'dark') return stored
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function applyToDocument(theme: Theme): void {
  if (typeof document === 'undefined') return
  document.documentElement.classList.toggle(DARK_CLASS, theme === 'dark')
}

export function createThemeStore(): ThemeStore {
  let theme = $state<Theme>(detectInitial())
  applyToDocument(theme)

  function set(next: Theme): void {
    theme = next
    applyToDocument(next)
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(STORAGE_KEY, next)
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
