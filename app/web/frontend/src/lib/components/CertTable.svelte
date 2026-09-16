<script lang="ts">
  import { get } from 'svelte/store'
  import { createVirtualizer } from '@tanstack/svelte-virtual'
  import { Skeleton } from '$lib/components/ui/skeleton'
  import { getI18n } from '$lib/stores/i18n.svelte'
  import { certStatus, parseCertID, statusBadgeClass, rowClassForStatus } from '$lib/utils/cert-status'
  import { formatDate, formatTime, type SortDirection, type SortKey } from '$lib/utils/cert-filter'
  import { certDisplayName, certExpiryLabel } from '$lib/utils/cert-label'
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
    sortKey: SortKey
    sortDir: SortDirection
    onSort: (key: SortKey) => void
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
    sortKey,
    sortDir,
    onSort,
    onSelect,
    onClearFilters,
  }: Props = $props()

  const i18n = getI18n()

  /**
   * Above this many rows the tbody renders a virtualized window (spacer rows
   * + visible slice) instead of the full list. Small lists keep the plain
   * render path, which is also what jsdom tests exercise (no layout engine
   * to measure against). Server paging means this triggers almost
   * exclusively for pageSize 'all'.
   */
  const VIRTUALIZE_AFTER = 100
  /** Initial per-row height guess until rows are measured. */
  const ESTIMATED_ROW_PX = 84

  const virtualized = $derived(certs.length > VIRTUALIZE_AFTER)

  let scrollEl = $state<HTMLDivElement>()
  const virtualizer = createVirtualizer<HTMLDivElement, HTMLTableRowElement>({
    count: 0,
    getScrollElement: () => scrollEl ?? null,
    estimateSize: () => ESTIMATED_ROW_PX,
    overscan: 8,
    getItemKey: (index) => certs[index]?.id ?? index,
  })

  $effect(() => {
    // Track list, scroll host, and order inputs; read the store one-shot via
    // get() so setOptions cannot re-trigger this effect.
    const count = virtualized ? certs.length : 0
    const host = scrollEl
    void sortKey
    void sortDir
    const firstKey = certs[0]?.id
    const instance = get(virtualizer)
    instance.setOptions({
      count,
      getScrollElement: () => host ?? null,
      estimateSize: () => ESTIMATED_ROW_PX,
      overscan: 8,
      getItemKey: (index) => certs[index]?.id ?? index,
    })
    if (virtualized && host && firstKey !== prevFirstKey) {
      // New result set (sort/filter/page turn): restart at the top.
      host.scrollTop = 0
    }
    prevFirstKey = firstKey
  })
  let prevFirstKey = $state<string | undefined>(undefined)

  const virtualItems = $derived($virtualizer.getVirtualItems())
  const totalVirtualSize = $derived($virtualizer.getTotalSize())
  const virtualWindow = $derived(
    virtualItems.length > 0
      ? virtualItems
      : certs.slice(0, 25).map((_, index) => ({ index, start: 0, end: 0, key: index, size: 0 })),
  )
  const virtualPadding = $derived.by(() => {
    if (!virtualized || virtualItems.length === 0) return { top: 0, bottom: 0 }
    const first = virtualItems[0]
    const last = virtualItems[virtualItems.length - 1]
    return { top: Math.max(0, first.start), bottom: Math.max(0, totalVirtualSize - last.end) }
  })

  function measureRow(node: HTMLTableRowElement): void {
    if (!virtualized) return
    get(virtualizer).measureElement(node)
  }

  function ariaSort(key: SortKey): 'ascending' | 'descending' | 'none' {
    if (sortKey !== key) return 'none'
    return sortDir === 'asc' ? 'ascending' : 'descending'
  }

  function sortIcon(key: SortKey): string {
    if (sortKey !== key) return '↕'
    return sortDir === 'asc' ? '↑' : '↓'
  }

  function sortLabel(column: string): string {
    return i18n.t('sortByColumn', 'Sort by {column}', { column })
  }
</script>

{#snippet certRow(cert: Certificate, dataIndex: number | undefined)}
  {@const s = certStatus(cert, thresholds)}
  {@const parts = parseCertID(cert.id)}
  <!-- Whole-row button: role="button" + aria-label gives SR users a single
       focusable target named by the common name; Enter activates the detail
       modal. Kept over a nested cell button to preserve one tab stop per row. -->
  <tr
    class="{rowClassForStatus(s)} vcv-row-clickable"
    data-index={dataIndex}
    use:measureRow
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
          <div class="vcv-expiry-count vcv-days-{s}">{certExpiryLabel(cert, i18n.t)}</div>
          <div class="vcv-expiry-datetime">
            <span class="vcv-expiry-date">{formatDate(cert.expiresAt)}</span>
            <span class="vcv-date-secondary">· {formatTime(cert.expiresAt)} UTC</span>
          </div>
        </div>
        <span class="vcv-row-action" aria-hidden="true">{i18n.t('buttonDetails', 'Details')}</span>
      </div>
    </td>
  </tr>
{/snippet}

<div class="vcv-table-wrapper" bind:this={scrollEl}>
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
    <table class="vcv-table" aria-rowcount={virtualized ? certs.length + 1 : undefined}>
      <colgroup>
        <col class="vcv-col-cert" />
        <col class="vcv-col-expiry" />
      </colgroup>
      <thead>
        <tr>
          <th scope="col" aria-sort={ariaSort('commonName')}>
            <button
              type="button"
              class="vcv-th-sort"
              class:vcv-th-sort-active={sortKey === 'commonName'}
              title={sortLabel(i18n.t('columnCommonName', 'Common Name'))}
              onclick={() => onSort('commonName')}
            >
              {i18n.t('columnCommonName', 'Common Name')}
              <span class="vcv-th-sort-icon" aria-hidden="true">{sortIcon('commonName')}</span>
            </button>
          </th>
          <th scope="col" class="vcv-col-expiry-head" aria-sort={ariaSort('expiresAt')}>
            <button
              type="button"
              class="vcv-th-sort"
              class:vcv-th-sort-active={sortKey === 'expiresAt'}
              title={sortLabel(i18n.t('columnExpiresAt', 'Expires'))}
              onclick={() => onSort('expiresAt')}
            >
              {i18n.t('columnExpiresAt', 'Expires')}
              <span class="vcv-th-sort-icon" aria-hidden="true">{sortIcon('expiresAt')}</span>
            </button>
          </th>
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
        {:else if !virtualized}
          {#each certs as cert (cert.id)}
            {@render certRow(cert, undefined)}
          {/each}
        {:else}
          {#if virtualPadding.top > 0}
            <tr class="vcv-spacer" aria-hidden="true"><td colspan="2" style="height: {virtualPadding.top}px;"></td></tr>
          {/if}
          {#each virtualWindow as item (certs[item.index]?.id ?? item.index)}
            {@render certRow(certs[item.index], item.index)}
          {/each}
          {#if virtualPadding.bottom > 0}
            <tr class="vcv-spacer" aria-hidden="true"><td colspan="2" style="height: {virtualPadding.bottom}px;"></td></tr>
          {/if}
        {/if}
      </tbody>
    </table>
  {/if}
</div>
