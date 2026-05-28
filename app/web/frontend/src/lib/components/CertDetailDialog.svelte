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
  import type { Certificate, DetailedCertificate } from '$lib/types'

  interface Props {
    cert: Certificate | null
    open: boolean
    onOpenChange: (open: boolean) => void
    onShowCA?: (certId: string) => void
  }

  const { cert, open, onOpenChange, onShowCA }: Props = $props()

  let details = $state<DetailedCertificate | null>(null)
  let loading = $state(false)
  let error = $state<string | null>(null)

  $effect(() => {
    if (!open || !cert) {
      details = null
      error = null
      return
    }
    loading = true
    error = null
    api
      .getCertificateDetails(cert.id)
      .then((data) => {
        details = data
      })
      .catch((err: unknown) => {
        error = err instanceof ApiError ? err.message : 'Failed to load details'
      })
      .finally(() => {
        loading = false
      })
  })

  async function copyPem(): Promise<void> {
    if (!details?.pem) return
    await navigator.clipboard.writeText(details.pem)
  }
</script>

<Dialog {open} {onOpenChange}>
  <DialogContent class="max-w-3xl">
    <DialogHeader>
      <DialogTitle>{cert?.commonName || 'Certificate'}</DialogTitle>
      <DialogDescription>
        Serial: <span class="font-mono">{cert?.serialNumber}</span>
      </DialogDescription>
    </DialogHeader>

    {#if loading}
      <p class="text-sm text-muted-foreground">Loading…</p>
    {:else if error}
      <p class="text-sm text-destructive">{error}</p>
    {:else if details}
      {#snippet Field(label: string, value: string, mono = false)}
        <div>
          <p class="text-xs uppercase tracking-wide text-muted-foreground">{label}</p>
          <p class="break-all {mono ? 'font-mono text-xs' : ''}">{value || '—'}</p>
        </div>
      {/snippet}
      <div class="grid gap-3 text-sm">
        {@render Field('Issuer', details.issuer)}
        {@render Field('Subject', details.subject)}
        {@render Field('Key', `${details.keyAlgorithm} ${details.keySize ? `(${details.keySize} bits)` : ''}`)}
        {@render Field('SHA1', details.fingerprintSHA1, true)}
        {@render Field('SHA256', details.fingerprintSHA256, true)}

        {#if details.sans?.length}
          <div>
            <p class="text-xs uppercase tracking-wide text-muted-foreground">SANs</p>
            <div class="mt-1 flex flex-wrap gap-1">
              {#each details.sans as san}
                <Badge variant="outline">{san}</Badge>
              {/each}
            </div>
          </div>
        {/if}

        {#if details.usage?.length}
          <div>
            <p class="text-xs uppercase tracking-wide text-muted-foreground">Usage</p>
            <div class="mt-1 flex flex-wrap gap-1">
              {#each details.usage as use}
                <Badge variant="secondary">{use}</Badge>
              {/each}
            </div>
          </div>
        {/if}

        {#if cert && onShowCA}
          <div>
            <Button
              size="sm"
              variant="outline"
              onclick={() => {
                onOpenChange(false)
                onShowCA(cert.id)
              }}
            >
              View issuer CA
            </Button>
          </div>
        {/if}

        {#if details.pem}
          <div>
            <div class="flex items-center justify-between">
              <p class="text-xs uppercase tracking-wide text-muted-foreground">PEM</p>
              <Button size="sm" variant="outline" onclick={copyPem}>Copy</Button>
            </div>
            <pre
              class="mt-1 max-h-64 overflow-auto rounded-md bg-muted p-3 text-xs font-mono">{details.pem}</pre>
          </div>
        {/if}
      </div>
    {/if}
  </DialogContent>
</Dialog>

