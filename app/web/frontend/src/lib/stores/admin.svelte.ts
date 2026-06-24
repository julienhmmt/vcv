import { api, ApiError } from '$lib/api'
import type { I18nStore } from '$lib/stores/i18n.svelte'
import type { AdminVaultStatus, SettingsFile } from '$lib/types'

export interface AdminStore {
  readonly authenticated: boolean
  readonly settings: SettingsFile | null
  readonly vaultStatuses: AdminVaultStatus[]
  readonly loading: boolean
  readonly error: string | null
  readonly successMessage: string | null
  checkSession(): Promise<void>
  login(username: string, password: string): Promise<boolean>
  logout(): Promise<void>
  loadSettings(): Promise<void>
  saveSettings(settings: SettingsFile): Promise<boolean>
  addVault(): Promise<void>
  removeVault(id: string): Promise<void>
  invalidateCache(): Promise<void>
  clearMessages(): void
}

export function createAdminStore(i18n: I18nStore): AdminStore {
  let authenticated = $state(false)
  let settings = $state<SettingsFile | null>(null)
  let vaultStatuses = $state<AdminVaultStatus[]>([])
  let loading = $state(false)
  let error = $state<string | null>(null)
  let successMessage = $state<string | null>(null)

  function applyError(err: unknown, fallback: string): void {
    error = err instanceof ApiError ? err.message : fallback
  }

  async function checkSession(): Promise<void> {
    try {
      const response = await api.adminSession()
      authenticated = response.authenticated
    } catch {
      authenticated = false
    }
  }

  async function login(username: string, password: string): Promise<boolean> {
    loading = true
    error = null
    try {
      const response = await api.adminLogin(username, password)
      authenticated = response.authenticated
      return authenticated
    } catch (err: unknown) {
      authenticated = false
      applyError(err, 'Login failed')
      return false
    } finally {
      loading = false
    }
  }

  async function logout(): Promise<void> {
    await api.adminLogout()
    authenticated = false
    settings = null
    vaultStatuses = []
  }

  async function loadSettings(): Promise<void> {
    loading = true
    error = null
    try {
      const response = await api.adminGetSettings()
      settings = {
        ...response.settings,
        vaults: response.settings.vaults.map((v) => ({ ...v, original_id: v.id })),
      }
      vaultStatuses = response.vault_statuses
    } catch (err: unknown) {
      applyError(err, 'Failed to load settings')
    } finally {
      loading = false
    }
  }

  async function saveSettings(next: SettingsFile): Promise<boolean> {
    loading = true
    error = null
    successMessage = null
    try {
      // Never re-send stored tokens. The backend preserves the existing token
      // when the incoming token is empty (mergeVaultTokens), so only forward a
      // token the admin actually typed into the field.
      const toSend: SettingsFile = {
        ...next,
        vaults: next.vaults.map((v) => ({
          ...v,
          token: v.token && v.token.trim() !== '' ? v.token : '',
        })),
      }
      const response = await api.adminPutSettings(toSend)
      settings = {
        ...response.settings,
        vaults: response.settings.vaults.map((v) => ({ ...v, original_id: v.id })),
      }
      vaultStatuses = response.vault_statuses
      successMessage = i18n.t('adminSettingsSaved', 'Settings saved')
      return true
    } catch (err: unknown) {
      applyError(err, 'Failed to save settings')
      return false
    } finally {
      loading = false
    }
  }

  async function addVault(): Promise<void> {
    if (!settings) return
    try {
      const response = await api.adminAddVault()
      settings = {
        ...settings,
        vaults: [...settings.vaults, response.vault],
      }
    } catch (err: unknown) {
      applyError(err, 'Failed to add vault')
    }
  }

  async function removeVault(id: string): Promise<void> {
    try {
      await api.adminDeleteVault(id)
      await loadSettings()
    } catch (err: unknown) {
      applyError(err, 'Failed to remove vault')
    }
  }

  async function invalidateCache(): Promise<void> {
    try {
      await api.adminInvalidateCache()
      successMessage = i18n.t('cacheInvalidated', 'Cache invalidated')
    } catch (err: unknown) {
      applyError(err, i18n.t('cacheInvalidateFailed', 'Failed to invalidate cache'))
    }
  }

  function clearMessages(): void {
    error = null
    successMessage = null
  }

  return {
    get authenticated() {
      return authenticated
    },
    get settings() {
      return settings
    },
    get vaultStatuses() {
      return vaultStatuses
    },
    get loading() {
      return loading
    },
    get error() {
      return error
    },
    get successMessage() {
      return successMessage
    },
    checkSession,
    login,
    logout,
    loadSettings,
    saveSettings,
    addVault,
    removeVault,
    invalidateCache,
    clearMessages,
  }
}
