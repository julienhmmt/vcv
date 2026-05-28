<script lang="ts">
  import { onMount } from 'svelte'
  import { toast } from 'svelte-sonner'
  import BookOpen from '@lucide/svelte/icons/book-open'
  import ChevronRight from '@lucide/svelte/icons/chevron-right'
  import Globe from '@lucide/svelte/icons/globe'
  import Moon from '@lucide/svelte/icons/moon'
  import RefreshCw from '@lucide/svelte/icons/refresh-cw'
  import Search from '@lucide/svelte/icons/search'
  import Sun from '@lucide/svelte/icons/sun'
  import { Toaster } from '$lib/components/ui/sonner'
  import { Skeleton } from '$lib/components/ui/skeleton'
  import CertDetailModal from '$lib/components/CertDetailModal.svelte'
  import CAModal from '$lib/components/CAModal.svelte'
  import CertTypeSelect from '$lib/components/CertTypeSelect.svelte'
  import VaultStatusPill from '$lib/components/VaultStatusPill.svelte'
  import ActiveFilters from '$lib/components/ActiveFilters.svelte'
  import MountSelectorDialog from '$lib/components/MountSelectorDialog.svelte'
  import Donut from '$lib/components/Donut.svelte'
  import { createCertsStore } from '$lib/stores/certs.svelte'
  import { createStatusStore } from '$lib/stores/status.svelte'
  import { createThemeStore } from '$lib/stores/theme.svelte'
  import { createI18nStore } from '$lib/stores/i18n.svelte'
  import {
    certStatus,
    parseCertID,
    statusBadgeClass,
    rowClassForStatus,
    DEFAULT_THRESHOLDS,
  } from '$lib/utils/cert-status'
  import {
    matchesFilters,
    sortCerts,
    paginate,
    dashboardCounts,
    daysRemainingLabel,
    formatDate,
    formatTime,
    type CertTypeFilter,
    type SortDirection,
    type SortKey,
    type StatusFilter,
  } from '$lib/utils/cert-filter'
  import type { Certificate } from '$lib/types'

  const certs = createCertsStore()
  const status = createStatusStore()
  const theme = createThemeStore()
  const i18n = createI18nStore()

  let search = $state('')
  let statusFilter = $state<StatusFilter>('all')
  let certTypeFilter = $state<CertTypeFilter>('all')
  let mountFilter = $state<string[] | null>(null)
  let sortKey = $state<SortKey>('expiresAt')
  let sortDir = $state<SortDirection>('asc')
  let pageIndex = $state(0)
  let pageSize = $state<number | 'all'>(25)

  let selected = $state<Certificate | null>(null)
  let certModalOpen = $state(false)
  let caCertId = $state<string | null>(null)
  let caModalOpen = $state(false)
  let mountModalOpen = $state(false)
  let initialLoad = $state(true)

  const filtered = $derived(
    certs.certificates.filter((cert) =>
      matchesFilters(cert, { search, status: statusFilter, certType: certTypeFilter, mounts: mountFilter }),
    ),
  )
  const sorted = $derived(sortCerts(filtered, sortKey, sortDir))
  const pageSizeNum = $derived(pageSize === 'all' ? sorted.length || 1 : pageSize)
  const totalPages = $derived(Math.max(1, Math.ceil(sorted.length / pageSizeNum)))
  const safePage = $derived(Math.min(pageIndex, totalPages - 1))
  const paged = $derived(paginate(sorted, safePage, pageSize))
  const counts = $derived(dashboardCounts(certs.certificates, DEFAULT_THRESHOLDS))

  const allMounts = $derived.by(() => {
    const set = new Set<string>()
    for (const cert of certs.certificates) {
      set.add(parseCertID(cert.id).mountKey)
    }
    return Array.from(set).sort()
  })
  const showVaultMount = $derived(allMounts.length > 1)

  onMount(() => {
    void load(true)
    const id = setInterval(() => void status.refresh(), 10_000)
    return () => clearInterval(id)
  })

  async function load(initial = false): Promise<void> {
    const promises = [certs.refresh(), status.refresh()]
    if (initial) {
      try {
        await Promise.all(promises)
      } finally {
        initialLoad = false
      }
      return
    }
    await Promise.all(promises)
  }

  async function manualRefresh(): Promise<void> {
    toast.promise(load(), {
      loading: 'Refreshing…',
      success: () => `Loaded ${certs.certificates.length} certificates`,
      error: 'Refresh failed',
    })
  }

  function toggleSort(key: SortKey): void {
    if (sortKey === key) {
      sortDir = sortDir === 'asc' ? 'desc' : 'asc'
    } else {
      sortKey = key
      sortDir = 'asc'
    }
  }

  function selectCert(cert: Certificate): void {
    selected = cert
    certModalOpen = true
  }

  function setStatus(next: StatusFilter): void {
    statusFilter = statusFilter === next ? 'all' : next
    pageIndex = 0
  }

  function clearAllFilters(): void {
    search = ''
    statusFilter = 'all'
    certTypeFilter = 'all'
    mountFilter = null
    pageIndex = 0
  }

  function pageInfoText(): string {
    if (sorted.length === 0) return '0 results'
    if (pageSize === 'all') return `${sorted.length} results`
    const start = safePage * (pageSize as number) + 1
    const end = Math.min(start + (pageSize as number) - 1, sorted.length)
    return `${start}–${end} of ${sorted.length}`
  }

  function onSearchKeydown(event: KeyboardEvent): void {
    if (event.key === '/' && document.activeElement?.tagName !== 'INPUT' && document.activeElement?.tagName !== 'TEXTAREA') {
      event.preventDefault()
      const el = document.getElementById('vcv-search') as HTMLInputElement | null
      el?.focus()
    }
  }

  $effect(() => {
    if (certs.error) toast.error(certs.error)
  })
</script>

<svelte:window onkeydown={onSearchKeydown} />

<Toaster richColors position="bottom-right" />

<a href="#vcv-main-content" class="vcv-skip-link">Skip to main content</a>

<div class="vcv-layout">
  <header class="vcv-header">
    <div class="vcv-header-bar">
      <div class="vcv-header-brand">
        <h1 class="vcv-title">
          VaultCertsViewer
          {#if status.status}<span class="vcv-title-version">v{status.status.version}</span>{/if}
        </h1>
        <p class="vcv-title-subtitle">{i18n.t('app.subtitle', 'Inspect certificates across Vault / OpenBao PKI mounts')}</p>
      </div>
      <div class="vcv-header-actions">
        <VaultStatusPill status={status.status} loading={status.loading} onRefresh={() => void status.refresh()} />
        <button
          class="vcv-button vcv-button-icon"
          type="button"
          title="Refresh"
          onclick={manualRefresh}
          disabled={certs.loading}
        >
          <RefreshCw class="h-4 w-4 {certs.loading ? 'animate-spin' : ''}" />
        </button>
        <button
          class="vcv-button vcv-button-icon vcv-theme-toggle"
          type="button"
          title="Toggle dark mode"
          aria-label="Toggle dark mode"
          onclick={theme.toggle}
        >
          {#if theme.theme === 'dark'}
            <Sun class="h-4 w-4" />
          {:else}
            <Moon class="h-4 w-4" />
          {/if}
        </button>
        <a class="vcv-button vcv-button-icon" href="/admin" title="Admin">
          <BookOpen class="h-4 w-4" />
        </a>
        <div class="vcv-lang-wrapper">
          <Globe class="vcv-lang-icon h-4 w-4" />
          <select
            class="vcv-select vcv-lang-select"
            aria-label="Language"
            value={i18n.lang}
            onchange={(event) => void i18n.setLang((event.target as HTMLSelectElement).value)}
          >
            <option value="de">DE</option>
            <option value="en">EN</option>
            <option value="es">ES</option>
            <option value="fr">FR</option>
            <option value="it">IT</option>
          </select>
        </div>
      </div>
    </div>

    <div id="vcv-filter-bar" class="vcv-filter-bar">
      <div class="vcv-filter-bar-inner">
        <div class="vcv-filter-palette">
          <div class="vcv-palette-item">
            <span class="vcv-palette-label">{i18n.t('filter.sources', 'Sources')}</span>
            <button type="button" class="vcv-mount-filter" onclick={() => (mountModalOpen = true)}>
              {#if mountFilter === null || mountFilter.length === allMounts.length}
                All mounts ({allMounts.length})
              {:else}
                {mountFilter.length} / {allMounts.length} mounts
              {/if}
            </button>
          </div>
          <span class="vcv-palette-separator" aria-hidden="true"></span>
          <div class="vcv-palette-item">
            <span class="vcv-palette-label">{i18n.t('label.certType', 'Type')}</span>
            <CertTypeSelect value={certTypeFilter} onChange={(next) => { certTypeFilter = next; pageIndex = 0 }} />
          </div>
        </div>
        <div class="vcv-search-wrapper">
          <Search class="vcv-search-icon h-[18px] w-[18px]" aria-hidden="true" />
          <input
            id="vcv-search"
            class="vcv-input vcv-input-search"
            type="search"
            placeholder={i18n.t('search.placeholder', 'Search certificates, serials, SANs…')}
            bind:value={search}
            oninput={() => (pageIndex = 0)}
          />
          <kbd class="vcv-search-shortcut" aria-label="Focus search">/</kbd>
        </div>
      </div>
    </div>

    <ActiveFilters
      {search}
      {statusFilter}
      {certTypeFilter}
      {mountFilter}
      allMountsCount={allMounts.length}
      onClearSearch={() => (search = '')}
      onClearStatus={() => (statusFilter = 'all')}
      onClearCertType={() => (certTypeFilter = 'all')}
      onClearMounts={() => (mountFilter = null)}
      onClearAll={clearAllFilters}
    />
  </header>

  <main id="vcv-main-content">
    <div id="vcv-dashboard" class="vcv-dashboard">
      <div class="vcv-dashboard-row">
        <div class="vcv-dashboard-stats">
          <div class="vcv-stat-group vcv-stat-group-attention">
            <span class="vcv-stat-group-label">{i18n.t('dashboard.attention', 'Attention')}</span>
            <div class="vcv-stat-group-cards">
              {#each [
                { key: 'expired', label: 'Expired', desc: 'Past expiry' },
                { key: 'critical', label: 'Critical', desc: `≤ ${DEFAULT_THRESHOLDS.critical} days` },
                { key: 'revoked', label: 'Revoked', desc: 'Marked revoked' },
              ] as item (item.key)}
                <button
                  type="button"
                  class="vcv-stat vcv-stat-risk vcv-stat-{item.key} vcv-stat-clickable {statusFilter === item.key ? 'vcv-stat-active' : ''}"
                  onclick={() => setStatus(item.key as StatusFilter)}
                >
                  <div class="vcv-stat-header">
                    <span class="vcv-stat-dot"></span>
                    <span class="vcv-stat-label">{item.label}</span>
                  </div>
                  <span class="vcv-stat-value">{counts[item.key as keyof typeof counts]}</span>
                  <span class="vcv-stat-desc">{item.desc}</span>
                </button>
              {/each}
            </div>
          </div>
          <div class="vcv-stat-group vcv-stat-group-healthy">
            <span class="vcv-stat-group-label">{i18n.t('dashboard.healthy', 'Healthy')}</span>
            <div class="vcv-stat-group-cards">
              {#each [
                { key: 'valid', label: 'Valid', desc: 'All good' },
                { key: 'warning', label: 'Warning', desc: `≤ ${DEFAULT_THRESHOLDS.warning} days` },
              ] as item (item.key)}
                <button
                  type="button"
                  class="vcv-stat vcv-stat-quiet vcv-stat-{item.key} vcv-stat-clickable {statusFilter === item.key ? 'vcv-stat-active' : ''}"
                  onclick={() => setStatus(item.key as StatusFilter)}
                >
                  <div class="vcv-stat-header">
                    <span class="vcv-stat-dot"></span>
                    <span class="vcv-stat-label">{item.label}</span>
                  </div>
                  <span class="vcv-stat-value">{counts[item.key as keyof typeof counts]}</span>
                  <span class="vcv-stat-desc">{item.desc}</span>
                </button>
              {/each}
            </div>
          </div>
        </div>
        <Donut
          counts={{ valid: counts.valid, warning: counts.warning, critical: counts.critical, expired: counts.expired, revoked: counts.revoked }}
          label="certs"
        />
      </div>
    </div>

    <div class="vcv-table-footer">
      <div class="vcv-page-size">
        <label for="vcv-page-size">Results per page</label>
        <select
          id="vcv-page-size"
          class="vcv-select vcv-page-size-select"
          value={String(pageSize)}
          onchange={(event) => {
            const value = (event.target as HTMLSelectElement).value
            pageSize = value === 'all' ? 'all' : Number(value)
            pageIndex = 0
          }}
        >
          <option value="25">25</option>
          <option value="50">50</option>
          <option value="100">100</option>
          <option value="all">All</option>
        </select>
      </div>
      <div class="vcv-page-buttons">
        <button
          type="button"
          class="vcv-button vcv-button-small vcv-button-ghost vcv-button-pill"
          disabled={safePage === 0}
          onclick={() => (pageIndex = Math.max(0, safePage - 1))}
        >
          Previous
        </button>
        <span class="vcv-page-info">{pageInfoText()}</span>
        <span class="vcv-badge vcv-badge-neutral">page {safePage + 1} / {totalPages}</span>
        <button
          type="button"
          class="vcv-button vcv-button-small vcv-button-ghost vcv-button-pill"
          disabled={safePage >= totalPages - 1}
          onclick={() => (pageIndex = Math.min(totalPages - 1, safePage + 1))}
        >
          Next
        </button>
      </div>
    </div>

    <div class="vcv-table-wrapper">
      {#if initialLoad && certs.certificates.length === 0}
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
            <col class="vcv-col-status" />
          </colgroup>
          <thead>
            <tr>
              <th scope="col">
                <button
                  type="button"
                  class="vcv-sort"
                  data-direction={sortKey === 'commonName' ? sortDir : 'asc'}
                  onclick={() => toggleSort('commonName')}
                >
                  <span class="vcv-sort-label">Common Name</span>
                  <span class="vcv-sort-indicator" aria-hidden="true"></span>
                </button>
                {#if showVaultMount}
                  <div class="vcv-sort-meta">
                    <button
                      type="button"
                      class="vcv-sort"
                      data-direction={sortKey === 'vault' ? sortDir : 'asc'}
                      onclick={() => toggleSort('vault')}
                    >
                      <span class="vcv-sort-label">Vault</span>
                      <span class="vcv-sort-indicator" aria-hidden="true"></span>
                    </button>
                    <button
                      type="button"
                      class="vcv-sort"
                      data-direction={sortKey === 'pki' ? sortDir : 'asc'}
                      onclick={() => toggleSort('pki')}
                    >
                      <span class="vcv-sort-label">PKI</span>
                      <span class="vcv-sort-indicator" aria-hidden="true"></span>
                    </button>
                  </div>
                {/if}
              </th>
              <th scope="col">
                <button
                  type="button"
                  class="vcv-sort"
                  data-direction={sortKey === 'expiresAt' ? sortDir : 'asc'}
                  onclick={() => toggleSort('expiresAt')}
                >
                  <span class="vcv-sort-label">Expires</span>
                  <span class="vcv-sort-indicator" aria-hidden="true"></span>
                </button>
              </th>
              <th scope="col" class="vcv-col-status">Status</th>
            </tr>
          </thead>
          <tbody>
            {#if paged.length === 0}
              <tr>
                <td colspan="3" class="vcv-empty-row">
                  {certs.loading ? 'Loading…' : 'No certificates match the current filters.'}
                </td>
              </tr>
            {:else}
              {#each paged as cert (cert.id)}
                {@const s = certStatus(cert, DEFAULT_THRESHOLDS)}
                {@const parts = parseCertID(cert.id)}
                <tr
                  class="{rowClassForStatus(s)} vcv-row-clickable"
                  onclick={() => selectCert(cert)}
                  onkeydown={(event) => event.key === 'Enter' && selectCert(cert)}
                  tabindex="0"
                  role="button"
                  aria-label={cert.commonName}
                >
                  <td class="vcv-col-cert">
                    <div class="vcv-cert-header">
                      <span class="vcv-cn-name">{cert.commonName || '—'}</span>
                      {#if showVaultMount}
                        <span class="vcv-cert-meta-item">{parts.vault || '—'}</span>
                        <span class="vcv-cert-meta-item">{parts.mount || '—'}</span>
                      {/if}
                    </div>
                    {#if cert.sans.length > 0}
                      <div class="vcv-san-row">
                        <span class="vcv-san-tag" title={cert.sans.join(', ')}>{cert.sans.join(', ')}</span>
                      </div>
                    {/if}
                  </td>
                  <td class="vcv-col-expiry">
                    <div class="vcv-expiry-count vcv-days-{s}">{daysRemainingLabel(cert)}</div>
                    <div class="vcv-expiry-datetime">
                      <span class="vcv-expiry-date">{formatDate(cert.expiresAt)}</span>
                      <span class="vcv-date-secondary">· {formatTime(cert.expiresAt)} UTC</span>
                    </div>
                  </td>
                  <td class="vcv-col-status">
                    <div class="vcv-status-cell">
                      <div class="vcv-status-badges">
                        <span class={statusBadgeClass(s)}>{s}</span>
                      </div>
                      <ChevronRight class="vcv-row-chevron h-4 w-4" aria-hidden="true" />
                    </div>
                  </td>
                </tr>
              {/each}
            {/if}
          </tbody>
        </table>
      {/if}
    </div>
  </main>

  <footer class="vcv-footer" aria-label="Footer">
    <div class="vcv-footer-legend">
      <div class="vcv-legend-item"><span class="vcv-badge vcv-badge-valid">Valid</span><span class="vcv-legend-text">All good</span></div>
      <div class="vcv-legend-item"><span class="vcv-badge vcv-badge-warning">Warning</span><span class="vcv-legend-text">≤ {DEFAULT_THRESHOLDS.warning} days</span></div>
      <div class="vcv-legend-item"><span class="vcv-badge vcv-badge-critical">Critical</span><span class="vcv-legend-text">≤ {DEFAULT_THRESHOLDS.critical} days</span></div>
      <div class="vcv-legend-item"><span class="vcv-badge vcv-badge-expired">Expired</span><span class="vcv-legend-text">Past expiry</span></div>
      <div class="vcv-legend-item"><span class="vcv-badge vcv-badge-revoked">Revoked</span><span class="vcv-legend-text">Revoked by CA</span></div>
    </div>
    <div class="vcv-footer-bottom">
      <div class="vcv-footer-brand">
        <div class="vcv-footer-title">
          VaultCertsViewer
          {#if status.status}<span class="vcv-footer-version">v{status.status.version}</span>{/if}
        </div>
        <div class="vcv-footer-meta">
          <span>License: <a class="vcv-footer-inline-link" href="https://github.com/julienhmmt/vcv/blob/main/LICENSE" target="_blank" rel="noopener">GNU Affero GPL v3.0</a></span>
          <span class="vcv-footer-divider">•</span>
          <span>Imagined and designed by <a href="https://j.hommet.net" target="_blank" rel="noopener">Julien HOMMET</a>, developed by AI.</span>
        </div>
      </div>
      <div class="vcv-footer-links" aria-label="External links">
        <a class="vcv-footer-link" href="https://hub.docker.com/r/jhmmt/vcv" target="_blank" rel="noopener">
          <img class="vcv-footer-icon" src="/docker.svg" alt="" /> Docker Hub
        </a>
        <a class="vcv-footer-link" href="https://github.com/julienhmmt/vcv" target="_blank" rel="noopener">
          <img class="vcv-footer-icon" src="/github.svg" alt="" /> GitHub
        </a>
        <a class="vcv-footer-link" href="https://j.hommet.net/vcv" target="_blank" rel="noopener">More info</a>
        <a class="vcv-footer-link" href="https://vcv.hommet.net" target="_blank" rel="noopener">Demo</a>
      </div>
    </div>
  </footer>
</div>

<CertDetailModal
  cert={selected}
  open={certModalOpen}
  onOpenChange={(value) => (certModalOpen = value)}
  onShowCA={(id) => {
    caCertId = id
    caModalOpen = true
  }}
/>

<CAModal
  certId={caCertId}
  open={caModalOpen}
  onOpenChange={(value) => (caModalOpen = value)}
/>

<MountSelectorDialog
  open={mountModalOpen}
  onOpenChange={(value) => (mountModalOpen = value)}
  {allMounts}
  selected={mountFilter}
  onChange={(next) => {
    mountFilter = next
    pageIndex = 0
  }}
/>
