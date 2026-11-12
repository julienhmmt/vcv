<script lang="ts">
  import ArrowLeft from '@lucide/svelte/icons/arrow-left'
  import Check from '@lucide/svelte/icons/check'
  import Copy from '@lucide/svelte/icons/copy'
  import Download from '@lucide/svelte/icons/download'
  import Link from '@lucide/svelte/icons/link'
  import ShieldCheck from '@lucide/svelte/icons/shield-check'
  import Landmark from '@lucide/svelte/icons/landmark'
  import { toast } from 'svelte-sonner'
  import * as Dialog from '$lib/components/ui/dialog'
  import { ScrollArea } from '$lib/components/ui/scroll-area'
  import { Skeleton } from '$lib/components/ui/skeleton'
  import { api, ApiError } from '$lib/api'
  import { certStatus, daysUntilExpiry, statusBadgeClass, parseCertID, DEFAULT_THRESHOLDS } from '$lib/utils/cert-status'
  import { formatDate, formatTime } from '$lib/utils/cert-filter'
  import { copyToClipboard } from '$lib/utils/clipboard'
  import { getI18n } from '$lib/stores/i18n.svelte'
  import type { Certificate, CertStatus, DetailedCertificate, ExpirationThresholds } from '$lib/types'

  interface Props {
    cert: Certificate | null
    open: boolean
    onOpenChange: (open: boolean) => void
    thresholds?: ExpirationThresholds
  }

  const { cert, open, onOpenChange, thresholds = DEFAULT_THRESHOLDS }: Props = $props()
  const i18n = getI18n()

  const statusLabels = $derived<Record<CertStatus, string>>({
    valid: i18n.t('statusLabelValid', 'Valid'),
    warning: i18n.t('statusLabelWarning', 'Warning'),
    critical: i18n.t('statusLabelCritical', 'Critical'),
    expired: i18n.t('statusLabelExpired', 'Expired'),
    revoked: i18n.t('statusLabelRevoked', 'Revoked'),
  })

  /** Loaded issuer levels, leaf-first (index 0 is the certificate's direct issuer). */
  let issuers = $state<DetailedCertificate[]>([])
  /** null = certificate view; otherwise the issuer level being shown. */
  let issuerIndex = $state<number | null>(null)
  const view = $derived(issuerIndex === null ? 'certificate' : 'issuer')
  const currentIssuer = $derived(issuerIndex === null ? null : (issuers[issuerIndex] ?? null))

  let details = $state<DetailedCertificate | null>(null)
  let loading = $state(false)
  let error = $state<string | null>(null)

  let issuerLoading = $state(false)
  let issuerError = $state<string | null>(null)

  let copiedField = $state<string | null>(null)
  let copyTimer: ReturnType<typeof setTimeout> | null = null
  /** Bumped on each details request so stale responses are ignored. */
  let detailsReqId = 0
  /** Bumped on each issuer request so stale responses are ignored. */
  let issuerReqId = 0

  function loadDetails(): void {
    if (!cert) return
    const reqId = ++detailsReqId
    const id = cert.id
    loading = true
    error = null
    api
      .getCertificateDetails(id)
      .then((data) => {
        if (reqId !== detailsReqId) return
        details = data
      })
      .catch((err: unknown) => {
        if (reqId !== detailsReqId) return
        error = err instanceof ApiError ? err.message : i18n.t('loadDetailsNetworkError', 'Failed to load details')
      })
      .finally(() => {
        if (reqId !== detailsReqId) return
        loading = false
      })
  }

  function loadIssuerAt(index: number): void {
    const parentId = index === 0 ? cert?.id : issuers[index - 1]?.id
    if (!parentId) return
    const reqId = ++issuerReqId
    issuerLoading = true
    issuerError = null
    api
      .getCertificateCA(parentId)
      .then((data) => {
        if (reqId !== issuerReqId) return
        issuers = [...issuers.slice(0, index), data]
        issuerIndex = index
      })
      .catch((err: unknown) => {
        if (reqId !== issuerReqId) return
        issuerError = err instanceof ApiError ? err.message : i18n.t('loadDetailsNetworkError', 'Failed to load details')
      })
      .finally(() => {
        if (reqId !== issuerReqId) return
        issuerLoading = false
      })
  }

  // Reset everything and (re)load whenever the dialog opens for a certificate.
  $effect(() => {
    if (!open || !cert) {
      detailsReqId++
      issuerReqId++
      details = null
      issuers = []
      issuerIndex = null
      error = null
      issuerError = null
      loading = false
      issuerLoading = false
      if (copyTimer) {
        clearTimeout(copyTimer)
        copyTimer = null
      }
      copiedField = null
      return
    }
    // Depend on the selected certificate so switching certs reloads.
    void cert.id
    issuers = []
    issuerIndex = null
    issuerError = null
    issuerReqId++
    loadDetails()
    return () => {
      detailsReqId++
      issuerReqId++
    }
  })

  function showIssuer(): void {
    if (issuerIndex === null) {
      if (issuers.length > 0) issuerIndex = 0
      else loadIssuerAt(0)
      return
    }
    const next = issuerIndex + 1
    if (next >= 5) return
    if (issuers[next]) {
      issuerIndex = next
      return
    }
    if (issuers[issuerIndex]?.caType === 'root') return
    loadIssuerAt(next)
  }

  function goToLevel(index: number | null): void {
    issuerIndex = index
  }

  async function copy(field: string, value: string): Promise<void> {
    const ok = await copyToClipboard(value)
    if (!ok) {
      toast.error(i18n.t('copyFailed', 'Copy failed — clipboard unavailable'))
      return
    }
    if (copyTimer) clearTimeout(copyTimer)
    copiedField = field
    copyTimer = setTimeout(() => {
      if (copiedField === field) copiedField = null
      copyTimer = null
    }, 1500)
  }

  async function copyLink(): Promise<void> {
    if (!cert) return
    const url = new URL(globalThis.location.href)
    url.searchParams.set('cert', cert.id)
    const ok = await copyToClipboard(url.toString())
    if (!ok) {
      toast.error(i18n.t('copyFailed', 'Copy failed — clipboard unavailable'))
      return
    }
    toast.success(i18n.t('labelCopied', 'Copied!'))
  }

  function downloadPem(pem: string, name: string): void {
    if (!pem) return
    const blob = new Blob([pem], { type: 'application/x-pem-file' })
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = `${name || 'certificate'}.pem`
    document.body.appendChild(anchor)
    anchor.click()
    anchor.remove()
    URL.revokeObjectURL(url)
    toast.success(i18n.t('downloadPEMSuccess', 'Certificate PEM downloaded successfully'))
  }

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

  const status = $derived(cert ? certStatus(cert, thresholds) : 'valid')
  const days = $derived(cert ? daysUntilExpiry(cert) : 0)
  const expiryTone = $derived(cert ? tone(status) : 'ok')
  function validityPercent(createdAt: string, expiresAt: string, now = Date.now()): number {
    const start = new Date(createdAt).getTime()
    const end = new Date(expiresAt).getTime()
    if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start) return 0
    return Math.max(0, Math.min(100, Math.round(((now - start) / (end - start)) * 100)))
  }

  const percent = $derived(cert ? validityPercent(cert.createdAt, cert.expiresAt) : 0)
  const source = $derived(cert ? parseCertID(cert.id) : null)
  const issuerLabel = $derived(
    currentIssuer?.caType === 'root'
      ? i18n.t('labelRootCA', 'Root CA')
      : i18n.t('labelIntermediateCA', 'Intermediate CA'),
  )
</script>

<Dialog.Root {open} {onOpenChange}>
  <Dialog.Content class="max-w-4xl p-0 overflow-hidden">
    <Dialog.Title class="sr-only">
      {view === 'issuer'
        ? i18n.t('caIssuerCertificate', 'Issuer certificate')
        : i18n.t('certificateInformationTitle', 'Certificate information')}
    </Dialog.Title>
    {#if loading && !details}
      <div class="vcv-cd-skeleton">
        <Skeleton class="h-8 w-2/3" />
        <Skeleton class="h-5 w-1/3" />
        <Skeleton class="h-24 w-full" />
        <Skeleton class="h-5 w-1/2" />
        <Skeleton class="h-5 w-3/4" />
      </div>
    {:else if error}
      <div class="vcv-cd-error">
        <p class="vcv-cd-error-text">{error}</p>
        <button type="button" class="vcv-button vcv-button-secondary" onclick={loadDetails}>
          {i18n.t('buttonRetry', 'Retry')}
        </button>
      </div>
    {:else if cert && details && view === 'certificate'}
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
            <div
              class="vcv-cd-validity"
              role="progressbar"
              aria-valuemin="0"
              aria-valuemax="100"
              aria-valuenow={percent}
              aria-label={i18n.t('labelValidity', 'Validity')}
            >
              <span class="vcv-cd-validity-fill vcv-cd-validity-{expiryTone}" style="width: {percent}%"></span>
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
              </div>
              <div class="vcv-cd-hero-meta">
                <span class="vcv-cd-hero-meta-label">{i18n.t('labelCertificateType', 'Type')}</span>
                <span class="vcv-cd-hero-meta-value">{details.certType || '—'}</span>
              </div>
            </header>

            <section class="vcv-cd-detail-list">
              {#if source}
                <div class="vcv-cd-detail-row">
                  <span>{i18n.t('labelSource', 'Source')}</span>
                  <strong>{source.vault || '—'} · {source.mount || '—'}</strong>
                </div>
              {/if}

              <div class="vcv-cd-detail-row">
                <span>{i18n.t('labelIssuer', 'Issuer')}</span>
                <strong title={details.issuer}>{details.issuer || '—'}</strong>
              </div>

              <div class="vcv-cd-detail-row">
                <span>{i18n.t('labelUsage', 'Usage')}</span>
                <strong>{details.usage?.join(', ') || '—'}</strong>
              </div>

              {#if details.sans?.length}
                <div class="vcv-cd-detail-row vcv-cd-detail-row-stack">
                  <span>{i18n.t('columnSan', 'SANs')}</span>
                  <div class="vcv-cd-san-list">
                    {#each details.sans as san (san)}
                      <span class="vcv-cd-san-chip"><code>{san}</code></span>
                    {/each}
                  </div>
                </div>
              {/if}
            </section>

            <details class="vcv-cd-technical">
              <summary class="vcv-cd-technical-summary">{i18n.t('technicalDetails', 'Technical details')}</summary>
              <section class="vcv-cd-detail-list vcv-cd-technical-list">
                {#if details.subject}
                  <div class="vcv-cd-detail-row">
                    <span>{i18n.t('labelSubject', 'Subject')}</span>
                    <div class="vcv-cd-copy-row">
                      <strong title={details.subject}>{details.subject}</strong>
                      <button
                        type="button"
                        class="vcv-cd-copy-btn"
                        class:vcv-cd-copy-done={copiedField === 'subject'}
                        onclick={() => copy('subject', details!.subject)}
                        aria-label={i18n.t('labelCopy', 'Copy')}
                      >
                        {#if copiedField === 'subject'}<Check class="h-3.5 w-3.5" />{:else}<Copy class="h-3.5 w-3.5" />{/if}
                      </button>
                    </div>
                  </div>
                {/if}
                <div class="vcv-cd-detail-row">
                  <span>{i18n.t('labelKeyAlgorithm', 'Key')}</span>
                  <strong>{details.keyAlgorithm} {details.keySize ? `(${details.keySize})` : ''}</strong>
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
                      {#if copiedField === 'serial'}<Check class="h-3.5 w-3.5" />{:else}<Copy class="h-3.5 w-3.5" />{/if}
                    </button>
                  </div>
                </div>
                <div class="vcv-cd-detail-row vcv-cd-detail-row-stack">
                  <span>{i18n.t('labelFingerprintSHA256', 'SHA-256')}</span>
                  <div class="vcv-cd-copy-row">
                    <code class="vcv-cd-fingerprint">{details.fingerprintSHA256}</code>
                    <button
                      type="button"
                      class="vcv-cd-copy-btn"
                      class:vcv-cd-copy-done={copiedField === 'sha256'}
                      onclick={() => copy('sha256', details!.fingerprintSHA256)}
                      aria-label={i18n.t('labelCopy', 'Copy')}
                    >
                      {#if copiedField === 'sha256'}<Check class="h-3.5 w-3.5" />{:else}<Copy class="h-3.5 w-3.5" />{/if}
                    </button>
                  </div>
                </div>
                <div class="vcv-cd-detail-row vcv-cd-detail-row-stack">
                  <span>{i18n.t('labelFingerprintSHA1', 'SHA-1')}</span>
                  <div class="vcv-cd-copy-row">
                    <code class="vcv-cd-fingerprint">{details.fingerprintSHA1}</code>
                    <button
                      type="button"
                      class="vcv-cd-copy-btn"
                      class:vcv-cd-copy-done={copiedField === 'sha1'}
                      onclick={() => copy('sha1', details!.fingerprintSHA1)}
                      aria-label={i18n.t('labelCopy', 'Copy')}
                    >
                      {#if copiedField === 'sha1'}<Check class="h-3.5 w-3.5" />{:else}<Copy class="h-3.5 w-3.5" />{/if}
                    </button>
                  </div>
                </div>
              </section>
            </details>

            <div class="vcv-cd-actions">
              <button type="button" class="vcv-button vcv-button-secondary" onclick={showIssuer}>
                <Landmark class="h-4 w-4" />
                {i18n.t('buttonViewCA', 'View issuer CA')}
              </button>
              <button type="button" class="vcv-button vcv-button-secondary" onclick={copyLink}>
                <Link class="h-4 w-4" />
                {i18n.t('buttonCopyLink', 'Copy link')}
              </button>
              {#if details.pem}
                <button
                  type="button"
                  class="vcv-button vcv-button-secondary"
                  onclick={() => downloadPem(details!.pem, cert.commonName)}
                >
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
    {:else if cert && view === 'issuer'}
      <ScrollArea class="max-h-[85vh]">
        <div class="vcv-cd-passport vcv-ca-passport">
          <aside class="vcv-cd-passport-sidebar">
            <div class="vcv-cd-emblem vcv-ca-emblem">
              <Landmark class="h-8 w-8" />
            </div>
            {#if currentIssuer}
              <div class="vcv-ca-type-badge"><span class="vcv-ca-type-label">{issuerLabel}</span></div>
              <div class="vcv-cd-date-stack">
                <div>
                  <span>{i18n.t('columnExpiresAt', 'Expires')}</span>
                  <strong>{formatDate(currentIssuer.expiresAt)}</strong>
                  <small>{formatTime(currentIssuer.expiresAt)} UTC</small>
                </div>
                <div>
                  <span>{i18n.t('columnCreatedAt', 'Created')}</span>
                  <strong>{formatDate(currentIssuer.createdAt)}</strong>
                  <small>{formatTime(currentIssuer.createdAt)} UTC</small>
                </div>
              </div>
            {/if}
          </aside>

          <main class="vcv-cd-passport-main">
            <nav class="vcv-cd-chain" aria-label={i18n.t('caChainLabel', 'Certificate chain')}>
              <button
                type="button"
                class="vcv-cd-chain-step"
                class:vcv-cd-chain-step-active={issuerIndex === null}
                onclick={() => goToLevel(null)}
              >
                {i18n.t('labelCertificate', 'Certificate')}
              </button>
              {#each issuers as entry, index (entry.id)}
                <span class="vcv-cd-chain-sep" aria-hidden="true">/</span>
                <button
                  type="button"
                  class="vcv-cd-chain-step"
                  class:vcv-cd-chain-step-active={issuerIndex === index}
                  onclick={() => goToLevel(index)}
                >
                  {entry.commonName || entry.subject || '—'}
                </button>
              {/each}
            </nav>
            <button type="button" class="vcv-button vcv-button-secondary vcv-cd-back" onclick={() => goToLevel(null)}>
              <ArrowLeft class="h-4 w-4" />
              {i18n.t('buttonBackToCertificate', 'Back to certificate')}
            </button>

            {#if issuerLoading && !currentIssuer}
              <div class="vcv-cd-skeleton">
                <Skeleton class="h-8 w-2/3" />
                <Skeleton class="h-5 w-1/3" />
                <Skeleton class="h-5 w-1/2" />
              </div>
            {:else if issuerError}
              <div class="vcv-cd-error">
                <p class="vcv-cd-error-text">{issuerError}</p>
                <button type="button" class="vcv-button vcv-button-secondary" onclick={() => loadIssuerAt(issuerIndex ?? 0)}>
                  {i18n.t('buttonRetry', 'Retry')}
                </button>
              </div>
            {:else if currentIssuer}
              <header class="vcv-cd-passport-header">
                <h3 class="vcv-cd-cn">{currentIssuer.subject || currentIssuer.commonName || '—'}</h3>
              </header>

              <section class="vcv-cd-detail-list">
                <div class="vcv-cd-detail-row">
                  <span>{i18n.t('labelIssuer', 'Issuer')}</span>
                  <strong title={currentIssuer.issuer}>{currentIssuer.issuer || '—'}</strong>
                </div>
                {#if currentIssuer.keyAlgorithm}
                  <div class="vcv-cd-detail-row">
                    <span>{i18n.t('labelKeyAlgorithm', 'Key')}</span>
                    <strong>{currentIssuer.keyAlgorithm}{currentIssuer.keySize ? ` (${currentIssuer.keySize})` : ''}</strong>
                  </div>
                {/if}
                {#if currentIssuer.usage?.length}
                  <div class="vcv-cd-detail-row">
                    <span>{i18n.t('labelUsage', 'Usage')}</span>
                    <strong>{currentIssuer.usage.join(', ')}</strong>
                  </div>
                {/if}
              </section>

              <details class="vcv-cd-technical">
                <summary class="vcv-cd-technical-summary">{i18n.t('technicalDetails', 'Technical details')}</summary>
                <section class="vcv-cd-detail-list vcv-cd-technical-list">
                  <div class="vcv-cd-detail-row">
                    <span>{i18n.t('labelSerialNumber', 'Serial')}</span>
                    <div class="vcv-cd-copy-row">
                      <code class="vcv-cd-serial">{currentIssuer.serialNumber}</code>
                      <button
                        type="button"
                        class="vcv-cd-copy-btn"
                        class:vcv-cd-copy-done={copiedField === 'ca-serial'}
                        onclick={() => copy('ca-serial', currentIssuer!.serialNumber)}
                        aria-label={i18n.t('labelCopy', 'Copy')}
                      >
                        {#if copiedField === 'ca-serial'}<Check class="h-3.5 w-3.5" />{:else}<Copy class="h-3.5 w-3.5" />{/if}
                      </button>
                    </div>
                  </div>
                  <div class="vcv-cd-detail-row vcv-cd-detail-row-stack">
                    <span>{i18n.t('labelFingerprintSHA256', 'SHA-256')}</span>
                    <div class="vcv-cd-copy-row">
                      <code class="vcv-cd-fingerprint">{currentIssuer.fingerprintSHA256}</code>
                      <button
                        type="button"
                        class="vcv-cd-copy-btn"
                        class:vcv-cd-copy-done={copiedField === 'ca-sha256'}
                        onclick={() => copy('ca-sha256', currentIssuer!.fingerprintSHA256)}
                        aria-label={i18n.t('labelCopy', 'Copy')}
                      >
                        {#if copiedField === 'ca-sha256'}<Check class="h-3.5 w-3.5" />{:else}<Copy class="h-3.5 w-3.5" />{/if}
                      </button>
                    </div>
                  </div>
                </section>
              </details>

              <div class="vcv-cd-actions">
                {#if currentIssuer.caType === 'intermediate'}
                  <button type="button" class="vcv-button vcv-button-secondary" onclick={showIssuer}>
                    <Landmark class="h-4 w-4" />
                    {i18n.t('buttonViewCA', 'View issuer CA')}
                  </button>
                {/if}
              {#if currentIssuer.pem}
                  <button
                    type="button"
                    class="vcv-button vcv-button-secondary"
                    onclick={() => downloadPem(currentIssuer!.pem, currentIssuer!.commonName || 'ca')}
                  >
                    <Download class="h-4 w-4" />
                    {i18n.t('buttonDownloadPEM', 'Download PEM')}
                  </button>
                  <button
                    type="button"
                    class="vcv-button vcv-button-primary"
                    class:vcv-cd-copy-pem-done={copiedField === 'ca-pem'}
                    onclick={() => copy('ca-pem', currentIssuer!.pem)}
                  >
                    {#if copiedField === 'ca-pem'}
                      <Check class="h-4 w-4" />
                      {i18n.t('labelCopied', 'Copied!')}
                    {:else}
                      <Copy class="h-4 w-4" />
                      {i18n.t('labelCopyPem', 'Copy PEM')}
                    {/if}
                  </button>
              {/if}
              </div>
            {/if}
          </main>
        </div>
      </ScrollArea>
    {/if}
  </Dialog.Content>
</Dialog.Root>
