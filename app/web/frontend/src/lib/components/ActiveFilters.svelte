<script lang="ts">
  import X from '@lucide/svelte/icons/x'
  import { getI18n } from '$lib/stores/i18n.svelte'
  import type { CertTypeFilter, StatusFilter } from '$lib/utils/cert-filter'

  interface Props {
    search: string
    statusFilter: StatusFilter
    certTypeFilter: CertTypeFilter
    mountFilter: string[] | null
    allMountsCount: number
    onClearSearch: () => void
    onClearStatus: () => void
    onClearCertType: () => void
    onClearMounts: () => void
    onClearAll: () => void
  }

  const {
    search,
    statusFilter,
    certTypeFilter,
    mountFilter,
    allMountsCount,
    onClearSearch,
    onClearStatus,
    onClearCertType,
    onClearMounts,
    onClearAll,
  }: Props = $props()

  const i18n = getI18n()

  const statusLabels = $derived<Record<string, string>>({
    valid: i18n.t('statusLabelValid', 'Valid'),
    warning: i18n.t('statusLabelWarning', 'Warning'),
    critical: i18n.t('statusLabelCritical', 'Critical'),
    expired: i18n.t('statusLabelExpired', 'Expired'),
    revoked: i18n.t('statusLabelRevoked', 'Revoked'),
  })
  const certTypeLabels = $derived<Record<string, string>>({
    machine: i18n.t('certTypeFilterMachine', 'Machine'),
    user: i18n.t('certTypeFilterUser', 'User'),
    both: i18n.t('certTypeFilterBoth', 'Both'),
    unknown: i18n.t('certTypeFilterUnknown', 'Unknown'),
  })

  const hasMountFilter = $derived(mountFilter !== null && mountFilter.length !== allMountsCount)
  const hasAny = $derived(
    !!search ||
      statusFilter !== 'all' ||
      certTypeFilter !== 'all' ||
      hasMountFilter,
  )
</script>

{#if hasAny}
  <div class="vcv-active-filters" aria-live="polite">
    {#if search}
      <button type="button" class="vcv-filter-chip" onclick={onClearSearch}>
        {i18n.t('filterChipSearch', 'Search')}: <strong>{search}</strong>
        <X class="h-3 w-3" />
      </button>
    {/if}
    {#if statusFilter !== 'all'}
      <button type="button" class="vcv-filter-chip vcv-filter-chip-{statusFilter}" onclick={onClearStatus}>
        {i18n.t('filterChipStatus', 'Status')}: <strong>{statusLabels[statusFilter] ?? statusFilter}</strong>
        <X class="h-3 w-3" />
      </button>
    {/if}
    {#if certTypeFilter !== 'all'}
      <button type="button" class="vcv-filter-chip" onclick={onClearCertType}>
        {i18n.t('filterChipCertType', 'Type')}: <strong>{certTypeLabels[certTypeFilter] ?? certTypeFilter}</strong>
        <X class="h-3 w-3" />
      </button>
    {/if}
    {#if hasMountFilter}
      <button type="button" class="vcv-filter-chip" onclick={onClearMounts}>
        {i18n.t('filterChipSources', 'Sources')}: <strong>{mountFilter?.length ?? 0} / {allMountsCount}</strong>
        <X class="h-3 w-3" />
      </button>
    {/if}
    <button type="button" class="vcv-button vcv-button-small vcv-button-ghost" onclick={onClearAll}>
      {i18n.t('filterChipReset', 'Clear all')}
    </button>
  </div>
{/if}
