import { api, ApiError } from '$lib/api'

const STORAGE_KEY = 'vcv-lang'
const FALLBACK = 'en'

export interface I18nStore {
  readonly messages: Record<string, string>
  readonly lang: string
  readonly loading: boolean
  readonly error: string | null
  readonly ready: Promise<void>
  t(key: string, fallback?: string): string
  setLang(lang: string): Promise<void>
}

function detectInitial(): string {
  if (typeof window === 'undefined') return FALLBACK
  const stored = window.localStorage.getItem(STORAGE_KEY)
  if (stored) return stored
  const browserLang = window.navigator.language.split('-')[0]
  return browserLang || FALLBACK
}

export function createI18nStore(): I18nStore {
  const initial = detectInitial()
  let lang = $state(initial)
  let messages = $state<Record<string, string>>({})
  let loading = $state(false)
  let error = $state<string | null>(null)

  async function load(target: string): Promise<void> {
    loading = true
    error = null
    try {
      messages = await api.i18n(target)
    } catch (err: unknown) {
      error = err instanceof ApiError ? err.message : 'Failed to load translations'
      messages = {}
    } finally {
      loading = false
    }
  }

  const ready = load(initial)

  async function setLang(next: string): Promise<void> {
    lang = next
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(STORAGE_KEY, next)
    }
    await load(next)
  }

  function t(key: string, fallback?: string): string {
    return messages[key] ?? fallback ?? key
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
