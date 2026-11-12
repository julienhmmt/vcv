<script lang="ts">
  import { toast } from 'svelte-sonner'
  import Copy from '@lucide/svelte/icons/copy'
  import { getI18n } from '$lib/stores/i18n.svelte'
  import { copyToClipboard } from '$lib/utils/clipboard'
  import { parseCertID, statusBadgeClass, rowClassForStatus } from '$lib/utils/cert-status'
  import { formatDate, formatTime } from '$lib/utils/cert-filter'
  import { certDisplayName, certExpiryLabel } from '$lib/utils/cert-label'
  import type { Certificate, CertStatus } from '$lib/types'

  interface Props {
    cert: Certificate
    showVaultMount: boolean
    status: CertStatus
    statusLabel: string
    onSelect: (cert: Certificate) => void
  }

  const { cert, showVaultMount, status: s, statusLabel, onSelect }: Props = $props()

  const i18n = getI18n()

  const certTypeLabels = $derived<Record<string, string>>({
    machine: i18n.t('certTypeFilterMachine', 'Machine'),
    user: i18n.t('certTypeFilterUser', 'User'),
    both: i18n.t('certTypeFilterBoth', 'Both'),
    unknown: i18n.t('certTypeFilterUnknown', 'Unknown'),
  })
  const typeLabel = $derived(certTypeLabels[cert.certType] ?? cert.certType)

  const parts = $derived(parseCertID(cert.id))
  const expiry = $derived(certExpiryLabel(cert, i18n.t))

  async function copyValue(value: string): Promise<void> {
    if (!value) return
    if (!(await copyToClipboard(value))) {
      toast.error(i18n.t('copyFailed', 'Copy failed — clipboard unavailable'))
      return
    }
    toast.success(i18n.t('labelCopied', 'Copied!'))
  }
</script>

<div class="vcv-cert-card {rowClassForStatus(s)} vcv-row-clickable">
  <div class="vcv-cert-card-header">
    <div class="vcv-cert-card-title">
      <div class="vcv-cert-name-row">
        <button
          type="button"
          class="vcv-cn-name vcv-cn-button"
          aria-label={`${certDisplayName(cert, i18n.t('certUnnamed', 'Unnamed certificate'))}: ${i18n.t('buttonDetails', 'Details')}`}
          onclick={(event) => {
            event.stopPropagation()
            onSelect(cert)
          }}
        >
          {cert.commonName || '—'}
        </button>
        <span class="vcv-row-quick-actions">
          <button
            type="button"
            class="vcv-icon-action"
            aria-label={i18n.t('labelCopyCommonName', 'Copy common name')}
            onclick={(event) => {
              event.stopPropagation()
              void copyValue(cert.commonName)
            }}
          >
            <Copy class="h-3.5 w-3.5" />
          </button>
          <button
            type="button"
            class="vcv-icon-action"
            aria-label={i18n.t('labelCopySerial', 'Copy serial number')}
            onclick={(event) => {
              event.stopPropagation()
              void copyValue(cert.serialNumber)
            }}
          >
            <Copy class="h-3.5 w-3.5" />
          </button>
        </span>
      </div>
      <span class="vcv-cert-meta-item vcv-cert-type">{typeLabel}</span>
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
