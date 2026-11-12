<script lang="ts">
  import { get } from 'svelte/store'
  import { toast } from 'svelte-sonner'
  import Copy from '@lucide/svelte/icons/copy'
  import { copyToClipboard } from '$lib/utils/clipboard'
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
    noVaultsConfigured: boolean
    noMountsConfigured: boolean
    refreshing: boolean
    compact: boolean
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
    noVaultsConfigured,
    noMountsConfigured,
    refreshing,
    compact,
    statusMeta,
    thresholds,
    sortKey,
    sortDir,
    onSort,
    onSelect,
    onClearFilters,
  }: Props = $props()

  const i18n = getI18n()

  const certTypeLabels = $derived<Record<string, string>>({
    machine: i18n.t('certTypeFilterMachine', 'Machine'),
    user: i18n.t('certTypeFilterUser', 'User'),
    both: i18n.t('certTypeFilterBoth', 'Both'),
    unknown: i18n.t('certTypeFilterUnknown', 'Unknown'),
  })

  function typeLabel(type: string): string {
    return certTypeLabels[type] ?? type
  }

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

  async function copyValue(value: string): Promise<void> {
    if (!value) return
    if (!(await copyToClipboard(value))) {
      toast.error(i18n.t('copyFailed', 'Copy failed — clipboard unavailable'))
      return
    }
    toast.success(i18n.t('labelCopied', 'Copied!'))
  }

  /** Arrow keys move between row name buttons; spacer rows have none, so focus stops there. */
  function onNameKeydown(event: KeyboardEvent): void {
    if (event.key !== 'ArrowDown' && event.key !== 'ArrowUp') return
    const row = (event.currentTarget as HTMLElement).closest('tr')
    const sibling = event.key === 'ArrowDown' ? row?.nextElementSibling : row?.previousElementSibling
    const next = sibling?.querySelector<HTMLButtonElement>('.vcv-cn-button')
    if (!next) return
    event.preventDefault()
    next.focus()
  }
</script>

{#snippet certRow(cert: Certificate, dataIndex: number | undefined)}
  {@const s = certStatus(cert, thresholds)}
  {@const parts = parseCertID(cert.id)}
  <!-- The name is the real focusable control; the row keeps a mouse-only
       click handler so the whole line stays a large hit target. -->
  <tr
    class="{rowClassForStatus(s)} vcv-row-clickable"
    data-index={dataIndex}
    use:measureRow
    onclick={() => onSelect(cert)}
  >
    <td class="vcv-col-cert">
      <div class="vcv-cert-header">
        <div class="vcv-cert-name-row">
          <button
            type="button"
            class="vcv-cn-name vcv-cn-button"
            aria-label={`${certDisplayName(cert, i18n.t('certUnnamed', 'Unnamed certificate'))}: ${i18n.t('buttonDetails', 'Details')}`}
            onclick={(event) => {
              event.stopPropagation()
              onSelect(cert)
            }}
            onkeydown={onNameKeydown}
          >
            {cert.commonName || '—'}
          </button>
          <span class="vcv-row-quick-actions">
            <button
              type="button"
              class="vcv-icon-action"
              aria-label={i18n.t('labelCopyCommonName', 'Copy common name')}
              title={i18n.t('labelCopyCommonName', 'Copy common name')}
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
              title={i18n.t('labelCopySerial', 'Copy serial number')}
              onclick={(event) => {
                event.stopPropagation()
                void copyValue(cert.serialNumber)
              }}
            >
              <Copy class="h-3.5 w-3.5" />
            </button>
          </span>
        </div>
        <div class="vcv-cert-meta-row">
          <span class="vcv-cert-meta-item vcv-cert-type">{typeLabel(cert.certType)}</span>
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

<div
  class="vcv-table-wrapper"
  class:vcv-table-scroll={virtualized}
  class:vcv-table-stale={refreshing}
  aria-busy={refreshing}
  bind:this={scrollEl}
>
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
    <table class="vcv-table" class:vcv-table-compact={compact} aria-rowcount={virtualized ? certs.length + 1 : undefined}>
      <caption class="sr-only">{i18n.t('tableCaption', 'Certificates')}</caption>
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
              {:else if noVaultsConfigured || noMountsConfigured}
                <div class="vcv-empty-state vcv-empty-first-run">
                  <p class="vcv-empty-title">
                    {noVaultsConfigured
                      ? i18n.t('emptyNoVaultsTitle', 'No vaults configured yet')
                      : i18n.t('emptyNoMountsTitle', 'No PKI mounts configured yet')}
                  </p>
                  <p class="vcv-empty-hint">
                    {noVaultsConfigured
                      ? i18n.t('emptyNoVaultsHint', 'Add a HashiCorp Vault or OpenBao instance in the admin panel, then enable its PKI mount.')
                      : i18n.t('emptyNoMountsHint', 'Enable at least one PKI mount for a configured vault in the admin panel.')}
                  </p>
                  <a class="vcv-button vcv-button-primary vcv-empty-cta" href="/admin">
                    {i18n.t('emptyOpenAdmin', 'Open admin panel')}
                  </a>
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
