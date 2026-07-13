import { describe, it, expect, vi, beforeEach } from 'vitest'
import type { SettingsFile, VaultInstance } from '$lib/types'

const {
  adminSession,
  adminLogin,
  adminLogout,
  adminGetSettings,
  adminPutSettings,
  adminAddVault,
  adminDeleteVault,
  adminInvalidateCache,
  ApiError,
} = vi.hoisted(() => {
  class ApiError extends Error {
    status: number
    constructor(status: number, message: string) {
      super(message)
      this.status = status
      this.name = 'ApiError'
    }
  }
  return {
    adminSession: vi.fn(),
    adminLogin: vi.fn(),
    adminLogout: vi.fn(),
    adminGetSettings: vi.fn(),
    adminPutSettings: vi.fn(),
    adminAddVault: vi.fn(),
    adminDeleteVault: vi.fn(),
    adminInvalidateCache: vi.fn(),
    ApiError,
  }
})

vi.mock('$lib/api', () => ({
  api: {
    adminSession,
    adminLogin,
    adminLogout,
    adminGetSettings,
    adminPutSettings,
    adminAddVault,
    adminDeleteVault,
    adminInvalidateCache,
  },
  ApiError,
}))

import { createAdminStore, mapVaultsForPut } from '$lib/stores/admin.svelte'
import type { I18nStore } from '$lib/stores/i18n.svelte'

const i18n = {
  t: (_key: string, fallback?: string) => fallback ?? _key,
} as unknown as I18nStore

function sampleSettings(vaults: VaultInstance[]): SettingsFile {
  return {
    app: { env: 'dev', port: 52000 },
    certificates: { expiration_thresholds: { critical: 7, warning: 30 } },
    metrics: {},
    cors: {},
    vaults,
  }
}

function sampleVault(overrides: Partial<VaultInstance> = {}): VaultInstance {
  return {
    id: 'vault1',
    address: 'https://vault.example',
    pki_mounts: ['pki'],
    ...overrides,
  }
}

beforeEach(() => {
  adminSession.mockReset()
  adminLogin.mockReset()
  adminLogout.mockReset()
  adminGetSettings.mockReset()
  adminPutSettings.mockReset()
  adminAddVault.mockReset()
  adminDeleteVault.mockReset()
  adminInvalidateCache.mockReset()
})

describe('mapVaultsForPut', () => {
  it('blanks undefined, empty, and whitespace-only tokens', () => {
    const vaults = [
      sampleVault({ token: undefined }),
      sampleVault({ id: 'v2', token: '' }),
      sampleVault({ id: 'v3', token: '   ' }),
    ]
    expect(mapVaultsForPut(vaults).map((v) => v.token)).toEqual(['', '', ''])
  })

  it('forwards a non-empty typed token', () => {
    const vaults = [sampleVault({ token: 'test-token-value' })]
    expect(mapVaultsForPut(vaults)[0]?.token).toBe('test-token-value')
  })
})

describe('createAdminStore', () => {
  it('checkSession sets authenticated from API true/false', async () => {
    const store = createAdminStore(i18n)
    adminSession.mockResolvedValueOnce({ authenticated: true })
    await store.checkSession()
    expect(store.authenticated).toBe(true)
    adminSession.mockResolvedValueOnce({ authenticated: false })
    await store.checkSession()
    expect(store.authenticated).toBe(false)
  })

  it('checkSession treats errors as unauthenticated', async () => {
    const store = createAdminStore(i18n)
    adminSession.mockRejectedValueOnce(new Error('network'))
    await store.checkSession()
    expect(store.authenticated).toBe(false)
  })

  it('login success authenticates; failure clears auth and sets error', async () => {
    const store = createAdminStore(i18n)
    adminLogin.mockResolvedValueOnce({ authenticated: true })
    expect(await store.login('admin', 'secret')).toBe(true)
    expect(store.authenticated).toBe(true)
    adminLogin.mockRejectedValueOnce(new ApiError(401, 'bad credentials'))
    expect(await store.login('admin', 'wrong')).toBe(false)
    expect(store.authenticated).toBe(false)
    expect(store.error).toBe('bad credentials')
  })

  it('loadSettings stamps original_id on vaults', async () => {
    const store = createAdminStore(i18n)
    adminGetSettings.mockResolvedValueOnce({
      settings: sampleSettings([sampleVault({ id: 'primary' })]),
      vault_statuses: [{ id: 'primary', enabled: true, connected: true }],
    })
    await store.loadSettings()
    expect(store.settings?.vaults[0]).toMatchObject({ id: 'primary', original_id: 'primary' })
    expect(store.vaultStatuses).toEqual([{ id: 'primary', enabled: true, connected: true }])
  })

  it('saveSettings blanks empty tokens and forwards typed tokens', async () => {
    const store = createAdminStore(i18n)
    const next = sampleSettings([
      sampleVault({ id: 'a', token: undefined }),
      sampleVault({ id: 'b', token: '   ' }),
      sampleVault({ id: 'c', token: 'test-token-value' }),
    ])
    adminPutSettings.mockResolvedValueOnce({
      settings: sampleSettings([sampleVault({ id: 'a' }), sampleVault({ id: 'b' }), sampleVault({ id: 'c' })]),
      vault_statuses: [],
    })
    expect(await store.saveSettings(next)).toBe(true)
    expect(adminPutSettings).toHaveBeenCalledTimes(1)
    const sent = adminPutSettings.mock.calls[0]?.[0] as SettingsFile
    expect(sent.vaults.map((v) => v.token)).toEqual(['', '', 'test-token-value'])
    expect(store.successMessage).toBe('Settings saved')
  })

  it('logout clears authenticated state and settings', async () => {
    const store = createAdminStore(i18n)
    adminLogin.mockResolvedValueOnce({ authenticated: true })
    await store.login('admin', 'secret')
    adminGetSettings.mockResolvedValueOnce({
      settings: sampleSettings([sampleVault()]),
      vault_statuses: [],
    })
    await store.loadSettings()
    adminLogout.mockResolvedValueOnce(undefined)
    await store.logout()
    expect(store.authenticated).toBe(false)
    expect(store.settings).toBeNull()
    expect(store.vaultStatuses).toEqual([])
  })

  it('invalidateCache sets successMessage on success', async () => {
    const store = createAdminStore(i18n)
    adminInvalidateCache.mockResolvedValueOnce(undefined)
    await store.invalidateCache()
    expect(store.successMessage).toBe('Cache invalidated')
    expect(store.error).toBeNull()
  })
})
