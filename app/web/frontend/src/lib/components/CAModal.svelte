<script lang="ts">
  import Check from '@lucide/svelte/icons/check'
  import Copy from '@lucide/svelte/icons/copy'
  import Download from '@lucide/svelte/icons/download'
  import Landmark from '@lucide/svelte/icons/landmark'
  import { toast } from 'svelte-sonner'
  import * as Dialog from '$lib/components/ui/dialog'
  import { ScrollArea } from '$lib/components/ui/scroll-area'
  import { api, ApiError } from '$lib/api'
  import { formatDate, formatTime } from '$lib/utils/cert-filter'
  import { getI18n } from '$lib/stores/i18n.svelte'
  import type { DetailedCertificate } from '$lib/types'

  interface Props {
    certId: string | null
    open: boolean
    onOpenChange: (open: boolean) => void
  }

  const { certId, open, onOpenChange }: Props = $props()
  const i18n = getI18n()

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
        error = err instanceof ApiError ? err.message : i18n.t('loadDetailsNetworkError', 'Failed to load CA')
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

  function downloadPem(): void {
    if (!ca?.pem) return
    const blob = new Blob([ca.pem], { type: 'application/x-pem-file' })
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = `${ca.commonName || 'ca'}.pem`
    document.body.appendChild(anchor)
    anchor.click()
    anchor.remove()
    URL.revokeObjectURL(url)
    toast.success(i18n.t('downloadPEMSuccess', 'Certificate PEM downloaded successfully'))
  }

  const caLabel = $derived(
    ca?.caType === 'root'
      ? i18n.t('labelRootCA', 'Root CA')
      : i18n.t('labelIntermediateCA', 'Intermediate CA'),
  )
</script>

<Dialog.Root {open} {onOpenChange}>
  <Dialog.Content class="max-w-4xl p-0 overflow-hidden">
    {#if loading && !ca}
      <div class="px-8 py-12 text-sm text-muted-foreground">{i18n.t('labelLoading', 'Loading…')}</div>
    {:else if error}
      <div class="px-8 py-12 text-sm text-destructive">{error}</div>
    {:else if ca}
      <ScrollArea class="max-h-[85vh]">
        <div class="vcv-cd-passport vcv-ca-passport">
          <aside class="vcv-cd-passport-sidebar vcv-ca-sidebar">
            <div class="vcv-cd-emblem vcv-ca-emblem">
              <Landmark class="h-8 w-8" />
            </div>

            <div class="vcv-cd-sidebar-status">
              <div class="vcv-ca-type-badge">
                <span class="vcv-ca-type-label">{caLabel}</span>
              </div>
              <strong class="vcv-cd-countdown vcv-ca-cn">{ca.commonName || '—'}</strong>
            </div>

            <div class="vcv-cd-date-stack">
              <div>
                <span>{i18n.t('columnExpiresAt', 'Expires')}</span>
                <strong>{formatDate(ca.expiresAt)}</strong>
                <small>{formatTime(ca.expiresAt)} UTC</small>
              </div>
              <div>
                <span>{i18n.t('columnCreatedAt', 'Created')}</span>
                <strong>{formatDate(ca.createdAt)}</strong>
                <small>{formatTime(ca.createdAt)} UTC</small>
              </div>
            </div>
          </aside>

          <main class="vcv-cd-passport-main">
            <header class="vcv-cd-passport-header">
              <div>
                <h3 class="vcv-cd-cn">{ca.subject || ca.commonName || '—'}</h3>
                {#if ca.issuer && ca.issuer !== ca.subject}
                  <p class="vcv-cd-hero-subject">{ca.issuer}</p>
                {/if}
              </div>
              {#if ca.keyAlgorithm}
                <div class="vcv-cd-hero-meta">
                  <span class="vcv-cd-hero-meta-label">{i18n.t('labelKeyAlgorithm', 'Key')}</span>
                  <span class="vcv-cd-hero-meta-value">
                    {ca.keyAlgorithm}{ca.keySize ? ` (${ca.keySize})` : ''}
                  </span>
                </div>
              {/if}
            </header>

            <section class="vcv-cd-detail-list">
              <div class="vcv-cd-detail-row">
                <span>{i18n.t('labelIssuer', 'Issuer')}</span>
                <strong title={ca.issuer}>{ca.issuer || '—'}</strong>
              </div>

              <div class="vcv-cd-detail-row">
                <span>{i18n.t('labelSerialNumber', 'Serial')}</span>
                <div class="vcv-cd-copy-row">
                  <code class="vcv-cd-serial">{ca.serialNumber}</code>
                  <button
                    type="button"
                    class="vcv-cd-copy-btn"
                    class:vcv-cd-copy-done={copied === 'serial'}
                    onclick={() => copy('serial', ca!.serialNumber)}
                    aria-label={i18n.t('labelCopy', 'Copy')}
                  >
                    {#if copied === 'serial'}<Check class="h-3.5 w-3.5" />{:else}<Copy class="h-3.5 w-3.5" />{/if}
                  </button>
                </div>
              </div>

              {#if ca.usage?.length}
                <div class="vcv-cd-detail-row vcv-cd-detail-row-stack">
                  <span>{i18n.t('labelUsage', 'Usage')}</span>
                  <div class="vcv-cd-san-list">
                    {#each ca.usage as use}
                      <span class="vcv-cd-san-chip"><code>{use}</code></span>
                    {/each}
                  </div>
                </div>
              {/if}

              <div class="vcv-cd-detail-row vcv-cd-detail-row-stack">
                <span>{i18n.t('labelFingerprintSHA256', 'SHA-256')}</span>
                <div class="vcv-cd-copy-row">
                  <code class="vcv-cd-fingerprint">{ca.fingerprintSHA256}</code>
                  <button
                    type="button"
                    class="vcv-cd-copy-btn"
                    class:vcv-cd-copy-done={copied === 'sha256'}
                    onclick={() => copy('sha256', ca!.fingerprintSHA256)}
                    aria-label={i18n.t('labelCopy', 'Copy')}
                  >
                    {#if copied === 'sha256'}<Check class="h-3.5 w-3.5" />{:else}<Copy class="h-3.5 w-3.5" />{/if}
                  </button>
                </div>
              </div>

              {#if ca.fingerprintSHA1}
                <div class="vcv-cd-detail-row vcv-cd-detail-row-stack">
                  <span>{i18n.t('labelFingerprintSHA1', 'SHA-1')}</span>
                  <div class="vcv-cd-copy-row">
                    <code class="vcv-cd-fingerprint">{ca.fingerprintSHA1}</code>
                    <button
                      type="button"
                      class="vcv-cd-copy-btn"
                      class:vcv-cd-copy-done={copied === 'sha1'}
                      onclick={() => copy('sha1', ca!.fingerprintSHA1)}
                      aria-label={i18n.t('labelCopy', 'Copy')}
                    >
                      {#if copied === 'sha1'}<Check class="h-3.5 w-3.5" />{:else}<Copy class="h-3.5 w-3.5" />{/if}
                    </button>
                  </div>
                </div>
              {/if}
            </section>

            {#if ca.pem}
              <div class="vcv-ca-pem-section">
                <div class="vcv-ca-pem-header">
                  <span class="vcv-ca-pem-label">{i18n.t('labelPem', 'PEM')}</span>
                  <div class="vcv-ca-pem-actions">
                    <button type="button" class="vcv-button vcv-button-secondary vcv-ca-pem-btn" onclick={downloadPem}>
                      <Download class="h-3.5 w-3.5" />
                      {i18n.t('buttonDownloadPEM', 'Download PEM')}
                    </button>
                    <button
                      type="button"
                      class="vcv-button vcv-button-secondary vcv-ca-pem-btn"
                      class:vcv-cd-copy-pem-done={copied === 'pem'}
                      onclick={() => copy('pem', ca!.pem)}
                    >
                      {#if copied === 'pem'}
                        <Check class="h-3.5 w-3.5" />
                        {i18n.t('labelCopied', 'Copied!')}
                      {:else}
                        <Copy class="h-3.5 w-3.5" />
                        {i18n.t('labelCopy', 'Copy')}
                      {/if}
                    </button>
                  </div>
                </div>
                <pre class="vcv-pem">{ca.pem}</pre>
              </div>
            {/if}

            <div class="vcv-cd-actions">
              {#if ca.pem}
                <button
                  type="button"
                  class="vcv-button vcv-button-primary"
                  class:vcv-cd-copy-pem-done={copied === 'pem'}
                  onclick={() => copy('pem', ca!.pem)}
                >
                  {#if copied === 'pem'}
                    <Check class="h-4 w-4" />
                    {i18n.t('labelCopied', 'Copied!')}
                  {:else}
                    <Copy class="h-4 w-4" />
                    {i18n.t('labelCopyPem', 'Copy PEM')}
                  {/if}
                </button>
              {/if}
            </div>
          </main>
        </div>
      </ScrollArea>
    {/if}
  </Dialog.Content>
</Dialog.Root>
