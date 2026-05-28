<script lang="ts">
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
  } from '$lib/components/ui/dialog'
  import { Badge } from '$lib/components/ui/badge'
  import { Button } from '$lib/components/ui/button'
  import { api, ApiError } from '$lib/api'
  import type { DetailedCertificate } from '$lib/types'

  interface Props {
    certId: string | null
    open: boolean
    onOpenChange: (open: boolean) => void
  }

  const { certId, open, onOpenChange }: Props = $props()

  let ca = $state<DetailedCertificate | null>(null)
  let loading = $state(false)
  let error = $state<string | null>(null)

  $effect(() => {
    if (!open || !certId) {
      ca = null
      error = null
      return
    }
    loading = true
    error = null
    api
      .getCertificateCA(certId)
      .then((data) => {
        ca = data
      })
      .catch((err: unknown) => {
        error = err instanceof ApiError ? err.message : 'Failed to load CA'
      })
      .finally(() => {
        loading = false
      })
  })

  async function copyPem(): Promise<void> {
    if (!ca?.pem) return
    await navigator.clipboard.writeText(ca.pem)
  }
</script>

<Dialog {open} {onOpenChange}>
  <DialogContent class="max-w-3xl">
    <DialogHeader>
      <DialogTitle>
        {ca?.caType === 'root' ? 'Root CA' : 'Intermediate CA'}
      </DialogTitle>
      <DialogDescription>
        {ca?.commonName ?? 'Issuer certificate'}
      </DialogDescription>
    </DialogHeader>

    {#if loading}
      <p class="text-sm text-muted-foreground">Loading…</p>
    {:else if error}
      <p class="text-sm text-destructive">{error}</p>
    {:else if ca}
      <div class="grid gap-3 text-sm">
        <div>
          <p class="text-xs uppercase tracking-wide text-muted-foreground">Subject</p>
          <p class="break-all">{ca.subject || '—'}</p>
        </div>
        <div>
          <p class="text-xs uppercase tracking-wide text-muted-foreground">Issuer</p>
          <p class="break-all">{ca.issuer || '—'}</p>
        </div>
        <div>
          <p class="text-xs uppercase tracking-wide text-muted-foreground">Serial</p>
          <p class="font-mono text-xs break-all">{ca.serialNumber}</p>
        </div>
        <div>
          <p class="text-xs uppercase tracking-wide text-muted-foreground">SHA256</p>
          <p class="font-mono text-xs break-all">{ca.fingerprintSHA256}</p>
        </div>
        {#if ca.usage?.length}
          <div>
            <p class="text-xs uppercase tracking-wide text-muted-foreground">Usage</p>
            <div class="mt-1 flex flex-wrap gap-1">
              {#each ca.usage as use}
                <Badge variant="secondary">{use}</Badge>
              {/each}
            </div>
          </div>
        {/if}
        {#if ca.pem}
          <div>
            <div class="flex items-center justify-between">
              <p class="text-xs uppercase tracking-wide text-muted-foreground">PEM</p>
              <Button size="sm" variant="outline" onclick={copyPem}>Copy</Button>
            </div>
            <pre
              class="mt-1 max-h-64 overflow-auto rounded-md bg-muted p-3 text-xs font-mono">{ca.pem}</pre>
          </div>
        {/if}
      </div>
    {/if}
  </DialogContent>
</Dialog>
