<script lang="ts">
  import { Skeleton } from '$lib/components/ui/skeleton'
  import { getI18n } from '$lib/stores/i18n.svelte'
  import {
    certStatus,
    daysUntilExpiry,
    parseCertID,
    statusBadgeClass,
    rowClassForStatus,
  } from '$lib/utils/cert-status'
  import { formatDate, formatTime } from '$lib/utils/cert-filter'
  import { certDisplayName } from '$lib/utils/cert-label'
  import type { Certificate, CertStatus, ExpirationThresholds } from '$lib/types'

  interface Props {
    certs: Certificate[]
    loading: boolean
    initialLoad: boolean
    hasInventory: boolean
    hasActiveFilters: boolean
    showVaultMount: boolean
    statusMeta: Record<CertStatus, { label: string; desc: string }>
    thresholds: ExpirationThresholds
    onSelect: (cert: Certificate) => void
    onClearFilters: () => void
  }

  const {
    certs,
    loading,
    initialLoad,
    hasInventory,
    hasActiveFilters,
    showVaultMount,
    statusMeta,
    thresholds,
    onSelect,
    onClearFilters,
  }: Props = $props()

  const i18n = getI18n()

  /** Localized expiry label for the table: compact "{n}d" ahead, descriptive when due/past. */
  function expiryLabel(cert: Certificate): string {
    const days = daysUntilExpiry(cert)
    if (days > 0) return i18n.t('daysRemainingShort', '{days}d', { days })
    if (days === 0) return i18n.t('expiringToday', 'Expires today')
    const ago = Math.abs(days)
    return i18n.t(ago === 1 ? 'expiredDaysSingular' : 'expiredDays', 'Expired {days} days ago', { days: ago })
  }
</script>

<div class="vcv-table-wrapper">
  {#if initialLoad && !hasInventory}
    <div class="vcv-table-skeleton">
      {#each Array(8) as _, i (i)}
        <div class="vcv-skeleton-row">
          <Skeleton class="h-5 flex-1" />
          <Skeleton class="h-5 w-24" />
          <Skeleton class="h-5 w-20" />
        </div>
      {/each}
    </div>
  {:else}
    <table class="vcv-table">
      <colgroup>
        <col class="vcv-col-cert" />
        <col class="vcv-col-expiry" />
      </colgroup>
      <thead>
        <tr>
          <th scope="col">{i18n.t('columnCommonName', 'Common Name')}</th>
          <th scope="col" class="vcv-col-expiry-head">{i18n.t('columnExpiresAt', 'Expires')}</th>
        </tr>
      </thead>
      <tbody>
        {#if certs.length === 0}
          <tr>
            <td colspan="2" class="vcv-empty-row">
              {#if loading}
                {i18n.t('labelLoading', 'Loading…')}
              {:else if hasActiveFilters}
                <div class="vcv-empty-state">
                  <p class="vcv-empty-title">{i18n.t('tableNoMatch', 'No certificates match the current filters.')}</p>
                  <button type="button" class="vcv-button vcv-button-small vcv-button-ghost vcv-button-pill" onclick={onClearFilters}>
                    {i18n.t('filterChipReset', 'Clear all')}
                  </button>
                </div>
              {:else}
                <div class="vcv-empty-state">
                  <p class="vcv-empty-title">{i18n.t('tableEmpty', 'No certificates found.')}</p>
                  <p class="vcv-empty-hint">{i18n.t('tableEmptyHint', 'No PKI mount returned any certificates yet.')}</p>
                </div>
              {/if}
            </td>
          </tr>
        {:else}
          {#each certs as cert (cert.id)}
            {@const s = certStatus(cert, thresholds)}
            {@const parts = parseCertID(cert.id)}
            <!-- Whole-row button: role="button" + aria-label gives SR users a single
                 focusable target named by the common name; Enter activates the detail
                 modal. Kept over a nested cell button to preserve one tab stop per row. -->
            <tr
              class="{rowClassForStatus(s)} vcv-row-clickable"
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
              <td class="vcv-col-cert">
                <div class="vcv-cert-header">
                  <span class="vcv-cn-name">{cert.commonName || '—'}</span>
                  <div class="vcv-cert-meta-row">
                    {#if showVaultMount}
                      <span class="vcv-cert-meta-item">{parts.vault || '—'}</span>
                      <span class="vcv-cert-meta-item">{parts.mount || '—'}</span>
                    {/if}
                    <span class="vcv-cert-status-inline {statusBadgeClass(s)}">{statusMeta[s].label}</span>
                  </div>
                </div>
                {#if cert.sans.length > 0}
                  <div class="vcv-san-row">
                    <span class="vcv-san-tag" title={cert.sans.join(', ')}>{cert.sans.join(', ')}</span>
                  </div>
                {/if}
              </td>
              <td class="vcv-col-expiry">
                <div class="vcv-expiry-cell">
                  <div class="vcv-expiry-main">
                    <div class="vcv-expiry-count vcv-days-{s}">{expiryLabel(cert)}</div>
                    <div class="vcv-expiry-datetime">
                      <span class="vcv-expiry-date">{formatDate(cert.expiresAt)}</span>
                      <span class="vcv-date-secondary">· {formatTime(cert.expiresAt)} UTC</span>
                    </div>
                  </div>
                  <span class="vcv-row-action" aria-hidden="true">{i18n.t('buttonDetails', 'Details')}</span>
                </div>
              </td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>
  {/if}
</div>
