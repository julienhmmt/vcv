<script lang="ts">
  import { Copy, Check, Landmark } from '@lucide/svelte'
  import * as Dialog from '$lib/components/ui/dialog'
  import { ScrollArea } from '$lib/components/ui/scroll-area'
  import { api, ApiError } from '$lib/api'
  import { formatDate, formatTime } from '$lib/utils/cert-filter'
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
  let copied = $state<string | null>(null)

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

  async function copy(field: string, value: string): Promise<void> {
    await navigator.clipboard.writeText(value)
    copied = field
    setTimeout(() => {
      if (copied === field) copied = null
    }, 1500)
  }
</script>

<Dialog.Root {open} {onOpenChange}>
  <Dialog.Content class="max-w-2xl p-0 overflow-hidden">
    <Dialog.Header class="px-6 pt-6">
      <Dialog.Title class="flex items-center gap-2">
        <Landmark class="h-5 w-5 text-primary" />
        {ca?.caType === 'root' ? 'Root CA' : 'Intermediate CA'}
      </Dialog.Title>
      <Dialog.Description>{ca?.commonName ?? 'Issuer certificate'}</Dialog.Description>
    </Dialog.Header>

    {#if loading && !ca}
      <div class="px-6 py-8 text-sm text-muted-foreground">Loading…</div>
    {:else if error}
      <div class="px-6 py-8 text-sm text-destructive">{error}</div>
    {:else if ca}
      <ScrollArea class="max-h-[70vh]">
        <div class="px-6 pb-6 space-y-4">
          <div class="grid gap-2 text-sm">
            <div class="grid grid-cols-[100px_1fr] gap-3">
              <span class="text-xs uppercase tracking-wide text-muted-foreground">Subject</span>
              <span>{ca.subject}</span>
            </div>
            <div class="grid grid-cols-[100px_1fr] gap-3">
              <span class="text-xs uppercase tracking-wide text-muted-foreground">Issuer</span>
              <span>{ca.issuer}</span>
            </div>
            <div class="grid grid-cols-[100px_1fr] gap-3">
              <span class="text-xs uppercase tracking-wide text-muted-foreground">Serial</span>
              <span class="font-mono text-xs break-all">{ca.serialNumber}</span>
            </div>
            <div class="grid grid-cols-[100px_1fr] gap-3">
              <span class="text-xs uppercase tracking-wide text-muted-foreground">Validity</span>
              <span>{formatDate(ca.createdAt)} — {formatDate(ca.expiresAt)} ({formatTime(ca.expiresAt)} UTC)</span>
            </div>
          </div>

          <div>
            <div class="text-xs uppercase tracking-wide text-muted-foreground mb-1">SHA-256</div>
            <div class="flex items-center gap-2">
              <code class="font-mono text-xs break-all flex-1">{ca.fingerprintSHA256}</code>
              <button
                type="button"
                class="vcv-cd-copy-btn"
                class:vcv-cd-copy-done={copied === 'sha256'}
                onclick={() => copy('sha256', ca!.fingerprintSHA256)}
                aria-label="Copy SHA-256"
              >
                {#if copied === 'sha256'}<Check class="h-3.5 w-3.5" />{:else}<Copy class="h-3.5 w-3.5" />{/if}
              </button>
            </div>
          </div>

          {#if ca.usage?.length}
            <div>
              <div class="text-xs uppercase tracking-wide text-muted-foreground mb-1">Usage</div>
              <div class="flex flex-wrap gap-1">
                {#each ca.usage as use}
                  <span class="vcv-cd-san-chip"><code>{use}</code></span>
                {/each}
              </div>
            </div>
          {/if}

          {#if ca.pem}
            <div>
              <div class="flex items-center justify-between mb-1">
                <span class="text-xs uppercase tracking-wide text-muted-foreground">PEM</span>
                <button
                  type="button"
                  class="vcv-button vcv-button-small"
                  onclick={() => copy('pem', ca!.pem)}
                >
                  {copied === 'pem' ? 'Copied!' : 'Copy'}
                </button>
              </div>
              <pre class="vcv-pem">{ca.pem}</pre>
            </div>
          {/if}
        </div>
      </ScrollArea>
    {/if}
  </Dialog.Content>
</Dialog.Root>
