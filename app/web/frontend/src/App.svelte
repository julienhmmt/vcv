<script lang="ts">
  import { onMount } from 'svelte'
  import { Button } from '$lib/components/ui/button'
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card'
  import CertTable from '$lib/components/CertTable.svelte'
  import CertDetailDialog from '$lib/components/CertDetailDialog.svelte'
  import StatusHeader from '$lib/components/StatusHeader.svelte'
  import { createCertsStore } from '$lib/stores/certs.svelte'
  import { createStatusStore } from '$lib/stores/status.svelte'
  import type { Certificate } from '$lib/types'

  const certs = createCertsStore()
  const status = createStatusStore()

  let selected = $state<Certificate | null>(null)
  let dialogOpen = $state(false)

  onMount(() => {
    void certs.refresh()
    void status.refresh()
  })

  function handleSelect(cert: Certificate): void {
    selected = cert
    dialogOpen = true
  }
</script>

<main class="min-h-svh bg-background text-foreground">
  <div class="mx-auto max-w-6xl space-y-6 p-6">
    <header class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-semibold tracking-tight">VaultCertsViewer</h1>
        <p class="text-sm text-muted-foreground">
          Inspect certificates across Vault / OpenBao PKI mounts.
        </p>
      </div>
      <Button
        variant="outline"
        onclick={() => {
          void certs.refresh()
          void status.refresh()
        }}
        disabled={certs.loading}
      >
        {certs.loading ? 'Refreshing…' : 'Refresh'}
      </Button>
    </header>

    <StatusHeader status={status.status} loading={status.loading} error={status.error} />

    <Card>
      <CardHeader>
        <CardTitle>Certificates</CardTitle>
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
    />
  </div>
</main>
