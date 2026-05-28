<script lang="ts">
  import { X } from '@lucide/svelte'
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
        Search: <strong>{search}</strong>
        <X class="h-3 w-3" />
      </button>
    {/if}
    {#if statusFilter !== 'all'}
      <button type="button" class="vcv-filter-chip vcv-filter-chip-{statusFilter}" onclick={onClearStatus}>
        Status: <strong>{statusFilter}</strong>
        <X class="h-3 w-3" />
      </button>
    {/if}
    {#if certTypeFilter !== 'all'}
      <button type="button" class="vcv-filter-chip" onclick={onClearCertType}>
        Type: <strong>{certTypeFilter}</strong>
        <X class="h-3 w-3" />
      </button>
    {/if}
    {#if hasMountFilter}
      <button type="button" class="vcv-filter-chip" onclick={onClearMounts}>
        Mounts: <strong>{mountFilter?.length ?? 0} / {allMountsCount}</strong>
        <X class="h-3 w-3" />
      </button>
    {/if}
    <button type="button" class="vcv-button vcv-button-small vcv-button-ghost" onclick={onClearAll}>
      Clear all
    </button>
  </div>
{/if}
