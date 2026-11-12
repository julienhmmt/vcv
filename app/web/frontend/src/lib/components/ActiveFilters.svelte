<script lang="ts">
  import X from '@lucide/svelte/icons/x'
  import { getI18n } from '$lib/stores/i18n.svelte'
  import type { CertTypeFilter } from '$lib/utils/cert-filter'
  import type { CertStatus } from '$lib/types'

  interface Props {
    search: string
    statusFilters: CertStatus[]
    certTypeFilter: CertTypeFilter
    mountFilter: string[] | null
    minDays: number | null
    maxDays: number | null
    allMountsCount: number
    onClearSearch: () => void
    onRemoveStatus: (key: CertStatus) => void
    onClearCertType: () => void
    onClearMounts: () => void
    onClearExpiry: () => void
    onClearAll: () => void
  }

  const {
    search,
    statusFilters,
    certTypeFilter,
    mountFilter,
    minDays,
    maxDays,
    allMountsCount,
    onClearSearch,
    onRemoveStatus,
    onClearCertType,
    onClearMounts,
    onClearExpiry,
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
  const hasExpiryRange = $derived(minDays !== null || maxDays !== null)
  const hasAny = $derived(
    !!search ||
      statusFilters.length > 0 ||
      certTypeFilter !== 'all' ||
      hasMountFilter ||
      hasExpiryRange,
  )

  const expiryLabel = $derived.by(() => {
    if (minDays === 0 && maxDays !== null) {
      return i18n.t('timelineWithinDays', '≤ {days} days', { days: maxDays })
    }
    if (maxDays === null && minDays !== null) {
      return i18n.t('timelineBeyondDays', '> {days} days', { days: minDays - 1 })
    }
    return i18n.t('timelineRangeDays', '{from}–{to} days', { from: minDays ?? 0, to: maxDays ?? 0 })
  })
</script>

{#if hasAny}
  <div class="vcv-active-filters" aria-live="polite">
    {#if search}
      <button type="button" class="vcv-filter-tag" onclick={onClearSearch}>
        {i18n.t('filterChipSearch', 'Search')}: <strong>{search}</strong>
        <X class="h-2.5 w-2.5" />
      </button>
    {/if}
    {#each statusFilters as status (status)}
      <button type="button" class="vcv-filter-tag vcv-filter-tag-{status}" onclick={() => onRemoveStatus(status)}>
        {statusLabels[status] ?? status}
        <X class="h-2.5 w-2.5" />
      </button>
    {/each}
    {#if certTypeFilter !== 'all'}
      <button type="button" class="vcv-filter-tag" onclick={onClearCertType}>
        {i18n.t('filterChipCertType', 'Type')}: <strong>{certTypeLabels[certTypeFilter] ?? certTypeFilter}</strong>
        <X class="h-2.5 w-2.5" />
      </button>
    {/if}
    {#if hasExpiryRange}
      <button type="button" class="vcv-filter-tag" onclick={onClearExpiry}>
        {i18n.t('filterChipExpiry', 'Expiry')}: <strong>{expiryLabel}</strong>
        <X class="h-2.5 w-2.5" />
      </button>
    {/if}
    {#if hasMountFilter}
      <button type="button" class="vcv-filter-tag" onclick={onClearMounts}>
        {i18n.t('filterChipSources', 'Sources')}: <strong>{mountFilter?.length ?? 0} / {allMountsCount}</strong>
        <X class="h-2.5 w-2.5" />
      </button>
    {/if}
    <button type="button" class="vcv-filter-tag-clear" onclick={onClearAll}>
      {i18n.t('filterChipReset', 'Clear all')}
    </button>
  </div>
{/if}
