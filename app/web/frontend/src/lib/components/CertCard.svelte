<script lang="ts">
  import ChevronRight from '@lucide/svelte/icons/chevron-right'
  import { getI18n } from '$lib/stores/i18n.svelte'
  import {
    certStatus,
    daysUntilExpiry,
    parseCertID,
    statusBadgeClass,
    rowClassForStatus,
    DEFAULT_THRESHOLDS,
  } from '$lib/utils/cert-status'
  import { formatDate, formatTime } from '$lib/utils/cert-filter'
  import type { Certificate } from '$lib/types'

  interface Props {
    cert: Certificate
    showVaultMount: boolean
    statusLabel: string
    onSelect: (cert: Certificate) => void
  }

  const { cert, showVaultMount, statusLabel, onSelect }: Props = $props()

  const i18n = getI18n()

  const s = $derived(certStatus(cert, DEFAULT_THRESHOLDS))
  const parts = $derived(parseCertID(cert.id))

  /** Localized expiry label: compact "{n}d" ahead, descriptive when due/past. */
  const expiry = $derived.by(() => {
    const days = daysUntilExpiry(cert)
    if (days > 0) return i18n.t('daysRemainingShort', '{days}d', { days })
    if (days === 0) return i18n.t('expiringToday', 'Expires today')
    const ago = Math.abs(days)
    return i18n.t(ago === 1 ? 'expiredDaysSingular' : 'expiredDays', 'Expired {days} days ago', { days: ago })
  })
</script>

<div
  class="vcv-cert-card {rowClassForStatus(s)} vcv-row-clickable"
  onclick={() => onSelect(cert)}
  onkeydown={(event) => event.key === 'Enter' && onSelect(cert)}
  tabindex="0"
  role="button"
  aria-label={cert.commonName}
>
  <div class="vcv-cert-card-header">
    <div class="vcv-cert-card-title">
      <span class="vcv-cn-name">{cert.commonName || '—'}</span>
      {#if cert.sans.length > 0}
        <div class="vcv-san-row">
          <span class="vcv-san-tag" title={cert.sans.join(', ')}>{cert.sans.join(', ')}</span>
        </div>
      {/if}
    </div>
    <div class="vcv-status-badges">
      <span class={statusBadgeClass(s)}>{statusLabel}</span>
    </div>
  </div>

  <div class="vcv-cert-card-meta">
    <div class="vcv-cert-card-field">
      <span class="vcv-cert-card-label">{i18n.t('columnExpiresAt', 'Expires')}</span>
      <div class="vcv-expiry-count vcv-days-{s}">{expiry}</div>
      <div class="vcv-expiry-datetime">
        <span class="vcv-expiry-date">{formatDate(cert.expiresAt)}</span>
        <span class="vcv-date-secondary">· {formatTime(cert.expiresAt)} UTC</span>
      </div>
    </div>
    {#if showVaultMount}
      <div class="vcv-cert-card-field">
        <span class="vcv-cert-card-label">{i18n.t('labelVault', 'Vault')} / {i18n.t('labelPki', 'PKI')}</span>
        <span class="vcv-cert-card-source">{parts.vault || '—'}</span>
        <span class="vcv-cert-card-source">{parts.mount || '—'}</span>
      </div>
    {/if}
  </div>

  <div class="vcv-cert-card-action">
    {i18n.t('buttonDetails', 'Details')}
    <ChevronRight class="h-4 w-4" aria-hidden="true" />
  </div>
</div>
