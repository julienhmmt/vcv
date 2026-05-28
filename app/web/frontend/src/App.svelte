<script lang="ts">
  import { onMount } from 'svelte'
  import { Sun, Moon } from '@lucide/svelte'
  import { Button } from '$lib/components/ui/button'
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card'
  import CertTable from '$lib/components/CertTable.svelte'
  import CertDetailDialog from '$lib/components/CertDetailDialog.svelte'
  import CADetailDialog from '$lib/components/CADetailDialog.svelte'
  import MountFilter from '$lib/components/MountFilter.svelte'
  import StatusHeader from '$lib/components/StatusHeader.svelte'
  import { createCertsStore } from '$lib/stores/certs.svelte'
  import { createStatusStore } from '$lib/stores/status.svelte'
  import { createThemeStore } from '$lib/stores/theme.svelte'
  import { createI18nStore } from '$lib/stores/i18n.svelte'
  import type { Certificate } from '$lib/types'

  const certs = createCertsStore()
  const status = createStatusStore()
  const theme = createThemeStore()
  const i18n = createI18nStore()

  let selected = $state<Certificate | null>(null)
  let dialogOpen = $state(false)
  let caCertId = $state<string | null>(null)
  let caDialogOpen = $state(false)
  let mounts = $state<string[] | null>(null)

  let allCertificates = $state<Certificate[]>([])

  onMount(async () => {
    await certs.refresh()
    allCertificates = certs.certificates
    void status.refresh()
  })

  async function refresh(): Promise<void> {
    await certs.refresh(mounts ?? undefined)
    if (mounts === null) {
      allCertificates = certs.certificates
    }
    void status.refresh()
  }

  function handleSelect(cert: Certificate): void {
    selected = cert
    dialogOpen = true
  }

  async function handleMountChange(next: string[] | null): Promise<void> {
    mounts = next
    await certs.refresh(next ?? undefined)
  }
</script>

<main class="min-h-svh bg-background text-foreground">
  <div class="mx-auto max-w-6xl space-y-6 p-6">
    <header class="flex items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">VaultCertsViewer</h1>
        <p class="text-sm text-muted-foreground">
          {i18n.t('dashboard.subtitle', 'Inspect certificates across Vault / OpenBao PKI mounts.')}
        </p>
      </div>
      <div class="flex items-center gap-2">
        <select
          class="h-9 rounded-md border border-input bg-background px-2 text-sm"
          value={i18n.lang}
          onchange={(event) => void i18n.setLang((event.target as HTMLSelectElement).value)}
        >
          <option value="en">EN</option>
          <option value="fr">FR</option>
        </select>
        <Button variant="ghost" size="icon" onclick={theme.toggle} aria-label="Toggle theme">
          {#if theme.theme === 'dark'}
            <Sun class="h-4 w-4" />
          {:else}
            <Moon class="h-4 w-4" />
          {/if}
        </Button>
        <Button variant="outline" onclick={refresh} disabled={certs.loading}>
          {certs.loading ? i18n.t('common.refreshing', 'Refreshing…') : i18n.t('common.refresh', 'Refresh')}
        </Button>
      </div>
    </header>

    <StatusHeader status={status.status} loading={status.loading} error={status.error} />

    <Card>
      <CardHeader class="flex flex-col gap-3">
        <CardTitle>{i18n.t('certificates.title', 'Certificates')}</CardTitle>
        <MountFilter
          certificates={allCertificates.length ? allCertificates : certs.certificates}
          selected={mounts}
          onChange={handleMountChange}
        />
      </CardHeader>
      <CardContent>
        <CertTable
          certificates={certs.certificates}
          loading={certs.loading}
          error={certs.error}
          onSelect={handleSelect}
        />
      </CardContent>
    </Card>

    <CertDetailDialog
      cert={selected}
      open={dialogOpen}
      onOpenChange={(value) => (dialogOpen = value)}
      onShowCA={(id) => {
        caCertId = id
        caDialogOpen = true
      }}
    />

    <CADetailDialog
      certId={caCertId}
      open={caDialogOpen}
      onOpenChange={(value) => (caDialogOpen = value)}
    />
  </div>
</main>
