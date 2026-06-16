<script lang="ts">
  import { onMount } from 'svelte'
  import AdminLogin from '$lib/components/admin/AdminLogin.svelte'
  import AdminPanel from '$lib/components/admin/AdminPanel.svelte'
  import { createAdminStore } from '$lib/stores/admin.svelte'
  import { createThemeStore } from '$lib/stores/theme.svelte'
  import { createI18nStore, setI18nContext } from '$lib/stores/i18n.svelte'

  const i18n = setI18nContext(createI18nStore())
  const admin = createAdminStore(i18n)
  createThemeStore()

  onMount(async () => {
    await admin.checkSession()
    if (admin.authenticated) {
      await admin.loadSettings()
    }
  })

  async function handleLogin(username: string, password: string): Promise<void> {
    const ok = await admin.login(username, password)
    if (ok) {
      await admin.loadSettings()
    }
  }
</script>

{#if !admin.authenticated}
  <AdminLogin loading={admin.loading} error={admin.error} onSubmit={handleLogin} />
{:else if admin.settings}
  <AdminPanel
    settings={admin.settings}
    statuses={admin.vaultStatuses}
    loading={admin.loading}
    error={admin.error}
    successMessage={admin.successMessage}
    onSave={(next) => void admin.saveSettings(next)}
    onAddVault={() => void admin.addVault()}
    onRemoveVault={(id) => void admin.removeVault(id)}
    onInvalidateCache={() => void admin.invalidateCache()}
    onLogout={() => void admin.logout()}
  />
{:else}
  <div class="adm-boot-loading">
    <span>{i18n.t('labelLoading', 'Loading…')}</span>
  </div>
{/if}

<style>
  :global(html) {
    scroll-behavior: smooth;
  }

  .adm-boot-loading {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100svh;
    font-size: 0.875rem;
    color: var(--vcv-color-muted);
    background: var(--vcv-color-bg);
  }
</style>
