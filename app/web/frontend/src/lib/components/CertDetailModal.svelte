<script lang="ts">
  import { Copy, Check, ShieldCheck } from '@lucide/svelte'
  import * as Dialog from '$lib/components/ui/dialog'
  import { ScrollArea } from '$lib/components/ui/scroll-area'
  import { api, ApiError } from '$lib/api'
  import { certStatus, daysUntilExpiry, statusBadgeClass, DEFAULT_THRESHOLDS } from '$lib/utils/cert-status'
  import { formatDate, formatTime } from '$lib/utils/cert-filter'
  import type { Certificate, CertStatus, DetailedCertificate } from '$lib/types'

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
  let copiedField = $state<string | null>(null)

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

  function tone(s: CertStatus): string {
    switch (s) {
      case 'valid':
        return 'ok'
      case 'warning':
        return 'warning'
      case 'critical':
      case 'expired':
      case 'revoked':
        return 'critical'
    }
  }

  async function copy(field: string, value: string): Promise<void> {
    await navigator.clipboard.writeText(value)
    copiedField = field
    setTimeout(() => {
      if (copiedField === field) copiedField = null
    }, 1500)
  }

  const status = $derived(cert ? certStatus(cert, DEFAULT_THRESHOLDS) : 'valid')
  const days = $derived(cert ? daysUntilExpiry(cert) : 0)
  const expiryTone = $derived(cert ? tone(status) : 'ok')
</script>

<Dialog.Root {open} {onOpenChange}>
  <Dialog.Content class="max-w-3xl p-0 overflow-hidden">
    <Dialog.Header class="px-6 pt-6">
      <Dialog.Title class="flex items-center gap-2">
        <ShieldCheck class="h-5 w-5 text-primary" />
        Certificate
      </Dialog.Title>
    </Dialog.Header>

    {#if loading && !details}
      <div class="px-6 py-8 text-sm text-muted-foreground">Loading…</div>
    {:else if error}
      <div class="px-6 py-8 text-sm text-destructive">{error}</div>
    {:else if cert && details}
      <ScrollArea class="max-h-[70vh]">
        <div class="vcv-cd-passport vcv-cd-passport-{expiryTone}">
          <aside class="vcv-cd-passport-sidebar">
            <div class="vcv-cd-emblem">
              <ShieldCheck class="h-7 w-7" />
            </div>
            <div class="vcv-cd-sidebar-status">
              <div class="vcv-cd-badges">
                <span class={statusBadgeClass(status)}>{status}</span>
              </div>
              {#if days >= 0}
                <strong class="vcv-cd-countdown vcv-cd-expiry-value-{expiryTone}">
                  {days}d left
                </strong>
              {:else}
                <strong class="vcv-cd-countdown vcv-cd-expiry-value-critical">
                  {Math.abs(days)}d ago
                </strong>
              {/if}
            </div>
            <div class="vcv-cd-date-stack">
              <div>
                <span>Expires</span>
                <strong>{formatDate(cert.expiresAt)}</strong>
                <small>{formatTime(cert.expiresAt)} UTC</small>
              </div>
              <div>
                <span>Created</span>
                <strong>{formatDate(cert.createdAt)}</strong>
                <small>{formatTime(cert.createdAt)} UTC</small>
              </div>
            </div>
          </aside>

          <main class="vcv-cd-passport-main">
            <header class="vcv-cd-passport-header">
              <div>
                <h3 class="vcv-cd-cn">{cert.commonName || '—'}</h3>
                {#if details.subject}
                  <p class="vcv-cd-hero-subject">{details.subject}</p>
                {/if}
              </div>
              <div class="vcv-cd-hero-meta">
                <span class="vcv-cd-hero-meta-label">Type</span>
                <span class="vcv-cd-hero-meta-value">{details.certType || '—'}</span>
              </div>
            </header>

            <section class="vcv-cd-info-strip">
              <div class="vcv-cd-strip-item">
                <span>Key</span>
                <strong>{details.keyAlgorithm} {details.keySize ? `(${details.keySize})` : ''}</strong>
              </div>
              <div class="vcv-cd-strip-item">
                <span>Usage</span>
                <strong>{details.usage?.join(', ') || '—'}</strong>
              </div>
            </section>

            <section class="vcv-cd-detail-list">
              <div class="vcv-cd-detail-row">
                <span>Issuer</span>
                <strong title={details.issuer}>{details.issuer || '—'}</strong>
              </div>

              <div class="vcv-cd-detail-row">
                <span>Serial</span>
                <div class="vcv-cd-copy-row">
                  <code class="vcv-cd-serial">{details.serialNumber}</code>
                  <button
                    type="button"
                    class="vcv-cd-copy-btn"
                    class:vcv-cd-copy-done={copiedField === 'serial'}
                    onclick={() => copy('serial', details!.serialNumber)}
                    aria-label="Copy serial"
                  >
                    {#if copiedField === 'serial'}
                      <Check class="h-3.5 w-3.5" />
                    {:else}
                      <Copy class="h-3.5 w-3.5" />
                    {/if}
                  </button>
                </div>
              </div>

              {#if details.sans?.length}
                <div class="vcv-cd-detail-row vcv-cd-detail-row-stack">
                  <span>SANs</span>
                  <div class="vcv-cd-san-list">
                    {#each details.sans as san}
                      <span class="vcv-cd-san-chip"><code>{san}</code></span>
                    {/each}
                  </div>
                </div>
              {/if}

              <div class="vcv-cd-detail-row vcv-cd-detail-row-stack">
                <span>SHA-256</span>
                <div class="vcv-cd-copy-row">
                  <code class="vcv-cd-fingerprint">{details!.fingerprintSHA256}</code>
                  <button
                    type="button"
                    class="vcv-cd-copy-btn"
                    class:vcv-cd-copy-done={copiedField === 'sha256'}
                    onclick={() => copy('sha256', details!.fingerprintSHA256)}
                    aria-label="Copy SHA-256"
                  >
                    {#if copiedField === 'sha256'}
                      <Check class="h-3.5 w-3.5" />
                    {:else}
                      <Copy class="h-3.5 w-3.5" />
                    {/if}
                  </button>
                </div>
              </div>

              <div class="vcv-cd-detail-row vcv-cd-detail-row-stack">
                <span>SHA-1</span>
                <div class="vcv-cd-copy-row">
                  <code class="vcv-cd-fingerprint">{details!.fingerprintSHA1}</code>
                  <button
                    type="button"
                    class="vcv-cd-copy-btn"
                    class:vcv-cd-copy-done={copiedField === 'sha1'}
                    onclick={() => copy('sha1', details!.fingerprintSHA1)}
                    aria-label="Copy SHA-1"
                  >
                    {#if copiedField === 'sha1'}
                      <Check class="h-3.5 w-3.5" />
                    {:else}
                      <Copy class="h-3.5 w-3.5" />
                    {/if}
                  </button>
                </div>
              </div>
            </section>

            <div class="vcv-cd-actions">
              {#if onShowCA}
                <button
                  type="button"
                  class="vcv-button vcv-button-secondary"
                  onclick={() => {
                    onOpenChange(false)
                    onShowCA(cert.id)
                  }}
                >
                  View issuer CA
                </button>
              {/if}
              {#if details.pem}
                <button
                  type="button"
                  class="vcv-button"
                  onclick={() => copy('pem', details!.pem)}
                >
                  {copiedField === 'pem' ? 'Copied!' : 'Copy PEM'}
                </button>
              {/if}
            </div>
          </main>
        </div>
      </ScrollArea>
    {/if}
  </Dialog.Content>
</Dialog.Root>
