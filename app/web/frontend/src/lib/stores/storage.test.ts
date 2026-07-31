// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

const { i18nFn, ApiError } = vi.hoisted(() => {
  class ApiError extends Error {
    status: number
    constructor(status: number, message: string) {
      super(message)
      this.status = status
      this.name = 'ApiError'
    }
  }
  return { i18nFn: vi.fn(), ApiError }
})

vi.mock('$lib/api', () => ({
  api: { i18n: i18nFn },
  ApiError,
}))

import { createThemeStore } from '$lib/stores/theme.svelte'
import { createI18nStore } from '$lib/stores/i18n.svelte'

/** Simulates private browsing / disabled storage, where every access throws. */
function breakStorage(): void {
  vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
    throw new Error('SecurityError: storage disabled')
  })
  vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
    throw new Error('QuotaExceededError')
  })
}

beforeEach(() => {
  i18nFn.mockResolvedValue({ language: 'en', messages: {} })
  window.localStorage.clear()
})

afterEach(() => {
  vi.restoreAllMocks()
})

describe('createThemeStore with unusable localStorage', () => {
  it('reads the stored theme when storage works', () => {
    window.localStorage.setItem('vcv-theme', 'dark')
    expect(createThemeStore().theme).toBe('dark')
  })

  it('falls back to the OS preference when reading throws', () => {
    window.localStorage.setItem('vcv-theme', 'dark')
    breakStorage()
    // The matchMedia stub in vitest.setup.ts reports matches: false.
    expect(createThemeStore().theme).toBe('light')
  })

  it('still switches theme when persisting throws', () => {
    breakStorage()
    const store = createThemeStore()
    expect(() => store.toggle()).not.toThrow()
    expect(store.theme).toBe('dark')
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark')
  })
})

describe('createI18nStore with unusable localStorage', () => {
  it('reads the stored language when storage works', () => {
    window.localStorage.setItem('vcv-lang', 'de')
    expect(createI18nStore().lang).toBe('de')
  })

  it('falls back to the default language when reading throws', () => {
    window.localStorage.setItem('vcv-lang', 'de')
    breakStorage()
    expect(createI18nStore().lang).toBe('en')
  })

  it('still switches language when persisting throws', async () => {
    breakStorage()
    const store = createI18nStore()
    await expect(store.setLang('it')).resolves.toBeUndefined()
    expect(store.lang).toBe('it')
    expect(document.documentElement.getAttribute('lang')).toBe('it')
  })
})
