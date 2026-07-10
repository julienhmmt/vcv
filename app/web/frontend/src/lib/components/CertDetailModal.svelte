<script lang="ts">
  import Check from '@lucide/svelte/icons/check'
  import Copy from '@lucide/svelte/icons/copy'
  import Download from '@lucide/svelte/icons/download'
  import ShieldCheck from '@lucide/svelte/icons/shield-check'
  import Landmark from '@lucide/svelte/icons/landmark'
  import { toast } from 'svelte-sonner'
  import * as Dialog from '$lib/components/ui/dialog'
  import { ScrollArea } from '$lib/components/ui/scroll-area'
  import { api, ApiError } from '$lib/api'
  import { certStatus, daysUntilExpiry, statusBadgeClass, DEFAULT_THRESHOLDS } from '$lib/utils/cert-status'
  import { formatDate, formatTime } from '$lib/utils/cert-filter'
  import { copyToClipboard } from '$lib/utils/clipboard'
  import { getI18n } from '$lib/stores/i18n.svelte'
  import type { Certificate, CertStatus, DetailedCertificate } from '$lib/types'

  interface Props {
    cert: Certificate | null
    open: boolean
    onOpenChange: (open: boolean) => void
    onShowCA?: (certId: string) => void
  }

  const { cert, open, onOpenChange, onShowCA }: Props = $props()
  const i18n = getI18n()

  const statusLabels = $derived<Record<CertStatus, string>>({
    valid: i18n.t('statusLabelValid', 'Valid'),
    warning: i18n.t('statusLabelWarning', 'Warning'),
    critical: i18n.t('statusLabelCritical', 'Critical'),
    expired: i18n.t('statusLabelExpired', 'Expired'),
    revoked: i18n.t('statusLabelRevoked', 'Revoked'),
  })

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
        error = err instanceof ApiError ? err.message : i18n.t('loadDetailsNetworkError', 'Failed to load details')
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
    const ok = await copyToClipboard(value)
    if (!ok) {
      toast.error(i18n.t('copyFailed', 'Copy failed — clipboard unavailable'))
      return
    }
    copiedField = field
    setTimeout(() => {
      if (copiedField === field) copiedField = null
    }, 1500)
  }

  function downloadPem(): void {
    if (!details?.pem) return
    const blob = new Blob([details.pem], { type: 'application/x-pem-file' })
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = `${cert?.commonName || 'certificate'}.pem`
    document.body.appendChild(anchor)
    anchor.click()
    anchor.remove()
    URL.revokeObjectURL(url)
    toast.success(i18n.t('downloadPEMSuccess', 'Certificate PEM downloaded successfully'))
  }

  const status = $derived(cert ? certStatus(cert, DEFAULT_THRESHOLDS) : 'valid')
  const days = $derived(cert ? daysUntilExpiry(cert) : 0)
  const expiryTone = $derived(cert ? tone(status) : 'ok')
</script>

<Dialog.Root {open} {onOpenChange}>
  <Dialog.Content class="max-w-4xl p-0 overflow-hidden">
    {#if loading && !details}
      <div class="px-8 py-12 text-sm text-muted-foreground">{i18n.t('labelLoading', 'Loading…')}</div>
    {:else if error}
      <div class="px-8 py-12 text-sm text-destructive">{error}</div>
    {:else if cert && details}
      <ScrollArea class="max-h-[85vh]">
        <div class="vcv-cd-passport vcv-cd-passport-{expiryTone}">
          <aside class="vcv-cd-passport-sidebar">
            <div class="vcv-cd-emblem">
              <ShieldCheck class="h-8 w-8" />
            </div>
            <div class="vcv-cd-sidebar-status">
              <div class="vcv-cd-badges">
                <span class={statusBadgeClass(status)}>{statusLabels[status]}</span>
              </div>
              {#if days > 0}
                <strong class="vcv-cd-countdown vcv-cd-expiry-value-{expiryTone}">
                  {i18n.t(days === 1 ? 'daysRemainingSingular' : 'daysRemaining', '{days} days remaining', { days })}
                </strong>
              {:else if days === 0}
                <strong class="vcv-cd-countdown vcv-cd-expiry-value-{expiryTone}">
                  {i18n.t('expiringToday', 'Expires today')}
                </strong>
              {:else}
                <strong class="vcv-cd-countdown vcv-cd-expiry-value-critical">
                  {i18n.t(Math.abs(days) === 1 ? 'expiredDaysSingular' : 'expiredDays', 'Expired {days} days ago', {
                    days: Math.abs(days),
                  })}
                </strong>
              {/if}
            </div>
            <div class="vcv-cd-date-stack">
              <div>
                <span>{i18n.t('columnExpiresAt', 'Expires')}</span>
                <strong>{formatDate(cert.expiresAt)}</strong>
                <small>{formatTime(cert.expiresAt)} UTC</small>
              </div>
              <div>
                <span>{i18n.t('columnCreatedAt', 'Created')}</span>
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
                <span class="vcv-cd-hero-meta-label">{i18n.t('labelCertificateType', 'Type')}</span>
                <span class="vcv-cd-hero-meta-value">{details.certType || '—'}</span>
              </div>
            </header>

            <section class="vcv-cd-info-strip">
              <div class="vcv-cd-strip-item">
                <span>{i18n.t('labelKeyAlgorithm', 'Key')}</span>
                <strong>{details.keyAlgorithm} {details.keySize ? `(${details.keySize})` : ''}</strong>
              </div>
              <div class="vcv-cd-strip-item">
                <span>{i18n.t('labelUsage', 'Usage')}</span>
                <strong>{details.usage?.join(', ') || '—'}</strong>
              </div>
            </section>

            <section class="vcv-cd-detail-list">
              <div class="vcv-cd-detail-row">
                <span>{i18n.t('labelIssuer', 'Issuer')}</span>
                <strong title={details.issuer}>{details.issuer || '—'}</strong>
              </div>

              <div class="vcv-cd-detail-row">
                <span>{i18n.t('labelSerialNumber', 'Serial')}</span>
                <div class="vcv-cd-copy-row">
                  <code class="vcv-cd-serial">{details.serialNumber}</code>
                  <button
                    type="button"
                    class="vcv-cd-copy-btn"
                    class:vcv-cd-copy-done={copiedField === 'serial'}
                    onclick={() => copy('serial', details!.serialNumber)}
                    aria-label={i18n.t('labelCopy', 'Copy')}
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
                  <span>{i18n.t('columnSan', 'SANs')}</span>
                  <div class="vcv-cd-san-list">
                    {#each details.sans as san}
                      <span class="vcv-cd-san-chip"><code>{san}</code></span>
                    {/each}
                  </div>
                </div>
              {/if}

              <div class="vcv-cd-detail-row vcv-cd-detail-row-stack">
                <span>{i18n.t('labelFingerprintSHA256', 'SHA-256')}</span>
                <div class="vcv-cd-copy-row">
                  <code class="vcv-cd-fingerprint">{details!.fingerprintSHA256}</code>
                  <button
                    type="button"
                    class="vcv-cd-copy-btn"
                    class:vcv-cd-copy-done={copiedField === 'sha256'}
                    onclick={() => copy('sha256', details!.fingerprintSHA256)}
                    aria-label={i18n.t('labelCopy', 'Copy')}
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
                <span>{i18n.t('labelFingerprintSHA1', 'SHA-1')}</span>
                <div class="vcv-cd-copy-row">
                  <code class="vcv-cd-fingerprint">{details!.fingerprintSHA1}</code>
                  <button
                    type="button"
                    class="vcv-cd-copy-btn"
                    class:vcv-cd-copy-done={copiedField === 'sha1'}
                    onclick={() => copy('sha1', details!.fingerprintSHA1)}
                    aria-label={i18n.t('labelCopy', 'Copy')}
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
                  <Landmark class="h-4 w-4" />
                  {i18n.t('buttonViewCA', 'View issuer CA')}
                </button>
              {/if}
              {#if details.pem}
                <button type="button" class="vcv-button vcv-button-secondary" onclick={downloadPem}>
                  <Download class="h-4 w-4" />
                  {i18n.t('buttonDownloadPEM', 'Download PEM')}
                </button>
                <button
                  type="button"
                  class="vcv-button vcv-button-primary"
                  class:vcv-cd-copy-pem-done={copiedField === 'pem'}
                  onclick={() => copy('pem', details!.pem)}
                >
                  {#if copiedField === 'pem'}
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
