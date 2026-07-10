import { getContext, setContext } from 'svelte'
import { api, ApiError } from '$lib/api'

const STORAGE_KEY = 'vcv-lang'
const FALLBACK = 'en'
const I18N_KEY = Symbol('vcv-i18n')

export type TParams = Record<string, string | number>

export interface I18nStore {
  readonly messages: Record<string, string>
  readonly lang: string
  readonly loading: boolean
  readonly error: string | null
  readonly ready: Promise<void>
  t(key: string, fallback?: string, params?: TParams): string
  setLang(lang: string): Promise<void>
}

/** Supported languages with their native display names, ordered for the picker. */
export const LANGUAGES: { code: string; name: string }[] = [
  { code: 'de', name: 'Deutsch' },
  { code: 'en', name: 'English' },
  { code: 'es', name: 'Español' },
  { code: 'fr', name: 'Français' },
  { code: 'it', name: 'Italiano' },
]

const SUPPORTED = new Set(LANGUAGES.map((l) => l.code))

/** Set <html lang> so screen readers and translation engines use the active language. */
function applyLangToDocument(lang: string): void {
  if (typeof document === 'undefined') return
  document.documentElement.setAttribute('lang', lang)
}

function detectInitial(): string {
  if (typeof window === 'undefined') return FALLBACK
  const stored = window.localStorage.getItem(STORAGE_KEY)
  if (stored && SUPPORTED.has(stored)) return stored
  const browserLang = window.navigator.language.split('-')[0]
  return browserLang && SUPPORTED.has(browserLang) ? browserLang : FALLBACK
}

/**
 * Replace `{{name}}` and `{name}` placeholders with values from `params`.
 * Unknown placeholders are left intact so missing data is visible.
 */
function interpolate(template: string, params?: TParams): string {
  if (!params) return template
  return template
    .replace(/\{\{\s*(\w+)\s*\}\}/g, (match, key) => (key in params ? String(params[key]) : match))
    .replace(/\{\s*(\w+)\s*\}/g, (match, key) => (key in params ? String(params[key]) : match))
}

export function createI18nStore(): I18nStore {
  const initial = detectInitial()
  applyLangToDocument(initial)
  let lang = $state(initial)
  let messages = $state<Record<string, string>>({})
  let loading = $state(false)
  let error = $state<string | null>(null)

  async function load(target: string): Promise<void> {
    loading = true
    error = null
    try {
      const response = await api.i18n(target)
      messages = response.messages ?? {}
    } catch (err: unknown) {
      error = err instanceof ApiError ? err.message : 'Failed to load translations'
      messages = {}
    } finally {
      loading = false
    }
  }

  const ready = load(initial)

  async function setLang(next: string): Promise<void> {
    if (!SUPPORTED.has(next)) return
    lang = next
    applyLangToDocument(next)
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(STORAGE_KEY, next)
    }
    await load(next)
  }

  function t(key: string, fallback?: string, params?: TParams): string {
    return interpolate(messages[key] ?? fallback ?? key, params)
  }

  return {
    get messages() {
      return messages
    },
    get lang() {
      return lang
    },
    get loading() {
      return loading
    },
    get error() {
      return error
    },
    ready,
    t,
    setLang,
  }
}

/** Provide a single i18n store to the component tree. Call once at the root. */
export function setI18nContext(store: I18nStore): I18nStore {
  return setContext(I18N_KEY, store)
}

/** Read the i18n store provided by an ancestor via {@link setI18nContext}. */
export function getI18n(): I18nStore {
  const store = getContext<I18nStore | undefined>(I18N_KEY)
  if (!store) {
    throw new Error('getI18n() called without a provider. Wrap the tree with setI18nContext().')
  }
  return store
}
