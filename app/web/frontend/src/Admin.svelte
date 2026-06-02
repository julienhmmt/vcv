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

<main class="min-h-svh bg-background text-foreground">
  <div class="mx-auto max-w-4xl p-6">
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
      <p class="text-sm text-muted-foreground">{i18n.t('labelLoading', 'Loading…')}</p>
    {/if}
  </div>
</main>
