<script lang="ts">
  import { getI18n } from '$lib/stores/i18n.svelte'
  import { certStatus, parseCertID, statusBadgeClass, rowClassForStatus, DEFAULT_THRESHOLDS } from '$lib/utils/cert-status'
  import { formatDate, formatTime } from '$lib/utils/cert-filter'
  import { certDisplayName, certExpiryLabel } from '$lib/utils/cert-label'
  import type { Certificate, ExpirationThresholds } from '$lib/types'

  interface Props {
    cert: Certificate
    showVaultMount: boolean
    statusLabel: string
    thresholds?: ExpirationThresholds
    onSelect: (cert: Certificate) => void
  }

  const { cert, showVaultMount, statusLabel, thresholds = DEFAULT_THRESHOLDS, onSelect }: Props = $props()

  const i18n = getI18n()

  const s = $derived(certStatus(cert, thresholds))
  const parts = $derived(parseCertID(cert.id))
  const expiry = $derived(certExpiryLabel(cert, i18n.t))
</script>

<div
  class="vcv-cert-card {rowClassForStatus(s)} vcv-row-clickable"
  onclick={() => onSelect(cert)}
  onkeydown={(event) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      onSelect(cert)
    }
  }}
  tabindex="0"
  role="button"
  aria-label={`${certDisplayName(cert, i18n.t('certUnnamed', 'Unnamed certificate'))}: ${i18n.t('buttonDetails', 'Details')}`}
>
  <div class="vcv-cert-card-header">
    <div class="vcv-cert-card-title">
      <span class="vcv-cn-name">{cert.commonName || '—'}</span>
      <span class="vcv-cert-status-inline {statusBadgeClass(s)}">{statusLabel}</span>
      {#if cert.sans.length > 0}
        <div class="vcv-san-row">
          <span class="vcv-san-tag" title={cert.sans.join(', ')}>{cert.sans.join(', ')}</span>
        </div>
      {/if}
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
  </div>
</div>
