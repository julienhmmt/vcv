<script lang="ts">
  import { onMount } from 'svelte'
  import CertDetailModal from '$lib/components/CertDetailModal.svelte'
  import CAModal from '$lib/components/CAModal.svelte'
  import Donut from '$lib/components/Donut.svelte'
  import Modal from '$lib/components/Modal.svelte'
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
  let vaultModalOpen = $state(false)

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

  const vaultSummary = $derived.by(() => {
    if (!status.status) return { text: '—', cls: 'vcv-status-state-neutral' }
    const total = status.status.vaults.length
    const up = status.status.vaults.filter((v) => v.connected).length
    if (total === 0) return { text: 'No vaults', cls: 'vcv-status-state-neutral' }
    return {
      text: total > 1 ? `${up}/${total} vaults` : (status.status.vaults[0].display_name || status.status.vaults[0].id),
      cls: up === total ? 'vcv-status-state-ok' : 'vcv-status-state-error',
    }
  })

  onMount(() => {
    void certs.refresh()
    void status.refresh()
    const id = setInterval(() => void status.refresh(), 10_000)
    return () => clearInterval(id)
  })

  function refresh(): void {
    void certs.refresh()
    void status.refresh()
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

  function clearStatus(): void {
    statusFilter = 'all'
    pageIndex = 0
  }

  function toggleMount(key: string): void {
    if (mountFilter === null) {
      mountFilter = allMounts.filter((m) => m !== key)
    } else if (mountFilter.includes(key)) {
      mountFilter = mountFilter.filter((m) => m !== key)
    } else {
      mountFilter = [...mountFilter, key]
    }
    pageIndex = 0
  }

  function selectAllMounts(): void {
    mountFilter = null
    pageIndex = 0
  }

  function deselectAllMounts(): void {
    mountFilter = []
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
</script>

<svelte:window onkeydown={onSearchKeydown} />

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
        <div id="vcv-vault-status">
          <button type="button" class="vcv-vault-status-pill {vaultSummary.cls}" onclick={() => (vaultModalOpen = true)}>
            <span class="vcv-status-dot"></span>
            <span>{vaultSummary.text}</span>
          </button>
        </div>
        <button
          class="vcv-button vcv-button-icon"
          type="button"
          title={i18n.t('common.refresh', 'Refresh')}
          onclick={refresh}
          disabled={certs.loading}
        >
          <span class="vcv-refresh-icon">↻</span>
        </button>
        <button
          class="vcv-button vcv-button-icon vcv-theme-toggle"
          type="button"
          title="Toggle dark mode"
          aria-label="Toggle dark mode"
          onclick={theme.toggle}
        >
          <span aria-hidden="true">{theme.theme === 'dark' ? '☀️' : '🌙'}</span>
        </button>
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

    <div id="vcv-filter-bar" class="vcv-filter-bar">
      <div class="vcv-filter-bar-inner">
        <div class="vcv-filter-palette">
          <div class="vcv-palette-item">
            <span class="vcv-palette-label">{i18n.t('filter.sources', 'Sources')}</span>
            <button type="button" class="vcv-mount-filter" onclick={() => (mountModalOpen = true)}>
              {#if mountFilter === null}
                All mounts ({allMounts.length})
              {:else if mountFilter.length === allMounts.length}
                All mounts ({allMounts.length})
              {:else}
                {mountFilter.length} / {allMounts.length} mounts
              {/if}
            </button>
          </div>
          <span class="vcv-palette-separator" aria-hidden="true"></span>
          <div class="vcv-palette-item">
            <label class="vcv-palette-label" for="vcv-cert-type-filter">{i18n.t('label.certType', 'Type')}</label>
            <select
              id="vcv-cert-type-filter"
              class="vcv-select vcv-select-compact"
              bind:value={certTypeFilter}
              onchange={() => (pageIndex = 0)}
            >
              <option value="all">All</option>
              <option value="machine">Machine</option>
              <option value="user">User</option>
              <option value="both">Both</option>
              <option value="unknown">Unknown</option>
            </select>
          </div>
        </div>
        <div class="vcv-search-wrapper">
          <svg class="vcv-search-icon" aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
          <input
            id="vcv-search"
            class="vcv-input vcv-input-search"
            type="search"
            placeholder={i18n.t('search.placeholder', 'Search certificates…')}
            bind:value={search}
            oninput={() => (pageIndex = 0)}
          />
          <kbd class="vcv-search-shortcut" aria-label="Focus search">/</kbd>
        </div>
      </div>
    </div>
  </header>

  <main id="vcv-main-content">
    <div id="vcv-dashboard" class="vcv-dashboard">
      <div class="vcv-dashboard-row">
        <div class="vcv-dashboard-stats">
          <div class="vcv-stat-group vcv-stat-group-attention">
            <span class="vcv-stat-group-label">{i18n.t('dashboard.attention', 'Attention')}</span>
            <div class="vcv-stat-group-cards">
              <button
                type="button"
                class="vcv-stat vcv-stat-risk vcv-stat-expired vcv-stat-clickable {statusFilter === 'expired' ? 'vcv-stat-active' : ''}"
                onclick={() => setStatus('expired')}
              >
                <div class="vcv-stat-header">
                  <span class="vcv-stat-dot"></span>
                  <span class="vcv-stat-label">Expired</span>
                </div>
                <span class="vcv-stat-value">{counts.expired}</span>
                <span class="vcv-stat-desc">Past expiry date</span>
              </button>
              <button
                type="button"
                class="vcv-stat vcv-stat-risk vcv-stat-critical vcv-stat-clickable {statusFilter === 'critical' ? 'vcv-stat-active' : ''}"
                onclick={() => setStatus('critical')}
              >
                <div class="vcv-stat-header">
                  <span class="vcv-stat-dot"></span>
                  <span class="vcv-stat-label">Critical</span>
                </div>
                <span class="vcv-stat-value">{counts.critical}</span>
                <span class="vcv-stat-desc">≤ {DEFAULT_THRESHOLDS.critical} days left</span>
              </button>
              <button
                type="button"
                class="vcv-stat vcv-stat-risk vcv-stat-revoked vcv-stat-clickable {statusFilter === 'revoked' ? 'vcv-stat-active' : ''}"
                onclick={() => setStatus('revoked')}
              >
                <div class="vcv-stat-header">
                  <span class="vcv-stat-dot"></span>
                  <span class="vcv-stat-label">Revoked</span>
                </div>
                <span class="vcv-stat-value">{counts.revoked}</span>
                <span class="vcv-stat-desc">Marked revoked</span>
              </button>
            </div>
          </div>
          <div class="vcv-stat-group vcv-stat-group-healthy">
            <span class="vcv-stat-group-label">{i18n.t('dashboard.healthy', 'Healthy')}</span>
            <div class="vcv-stat-group-cards">
              <button
                type="button"
                class="vcv-stat vcv-stat-quiet vcv-stat-valid vcv-stat-clickable {statusFilter === 'valid' ? 'vcv-stat-active' : ''}"
                onclick={() => setStatus('valid')}
              >
                <div class="vcv-stat-header">
                  <span class="vcv-stat-dot"></span>
                  <span class="vcv-stat-label">Valid</span>
                </div>
                <span class="vcv-stat-value">{counts.valid}</span>
                <span class="vcv-stat-desc">All good</span>
              </button>
              <button
                type="button"
                class="vcv-stat vcv-stat-quiet vcv-stat-warning vcv-stat-clickable {statusFilter === 'warning' ? 'vcv-stat-active' : ''}"
                onclick={() => setStatus('warning')}
              >
                <div class="vcv-stat-header">
                  <span class="vcv-stat-dot"></span>
                  <span class="vcv-stat-label">Warning</span>
                </div>
                <span class="vcv-stat-value">{counts.warning}</span>
                <span class="vcv-stat-desc">≤ {DEFAULT_THRESHOLDS.warning} days left</span>
              </button>
            </div>
          </div>
        </div>
        <Donut counts={{ valid: counts.valid, warning: counts.warning, critical: counts.critical, expired: counts.expired, revoked: counts.revoked }} label="certs" />
      </div>
      <div class="vcv-dashboard-actions">
        <span class="vcv-dashboard-hint">Click a status to filter the table</span>
        {#if statusFilter !== 'all'}
          <button type="button" class="vcv-button vcv-button-small vcv-clear-filter" onclick={clearStatus}>
            ✕ Clear filter
          </button>
        {/if}
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
              <td colspan="3" style="text-align:center;padding:24px;color:var(--vcv-text-muted)">
                {certs.loading ? 'Loading…' : 'No certificates'}
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
                    <span class="vcv-expiry-date">{new Date(cert.expiresAt).toISOString().split('T')[0]}</span>
                    <span class="vcv-date-secondary">· {new Date(cert.expiresAt).toISOString().split('T')[1].slice(0, 5)} UTC</span>
                  </div>
                </td>
                <td class="vcv-col-status">
                  <div class="vcv-status-cell">
                    <div class="vcv-status-badges">
                      <span class={statusBadgeClass(s)}>{s}</span>
                    </div>
                    <span class="vcv-row-chevron" aria-hidden="true">›</span>
                  </div>
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
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
  onClose={() => (certModalOpen = false)}
  onShowCA={(id) => {
    caCertId = id
    caModalOpen = true
  }}
/>

<CAModal certId={caCertId} open={caModalOpen} onClose={() => (caModalOpen = false)} />

<Modal open={mountModalOpen} title="Sources" large onClose={() => (mountModalOpen = false)}>
  <div class="vcv-mount-modal-body">
    <div class="vcv-mount-actions" style="display:flex;gap:12px;margin-bottom:12px">
      <button type="button" class="vcv-mount-link-btn" onclick={selectAllMounts}>Select all</button>
      <span class="vcv-mount-header-divider"></span>
      <button type="button" class="vcv-mount-link-btn" onclick={deselectAllMounts}>Deselect all</button>
    </div>
    <div class="vcv-mount-modal-list">
      {#each allMounts as key}
        {@const selected = mountFilter === null || mountFilter.includes(key)}
        <label class="vcv-mount-row" style="display:flex;align-items:center;gap:8px;padding:6px 0">
          <input
            type="checkbox"
            checked={selected}
            onchange={() => toggleMount(key)}
          />
          <span class="vcv-mono">{key}</span>
        </label>
      {/each}
      {#if allMounts.length === 0}
        <p>No mounts available.</p>
      {/if}
    </div>
  </div>
  {#snippet actions()}
    <button type="button" class="vcv-button vcv-button-primary" onclick={() => (mountModalOpen = false)}>Done</button>
  {/snippet}
</Modal>

<Modal open={vaultModalOpen} title="Vault Status" onClose={() => (vaultModalOpen = false)}>
  {#if status.status}
    <div class="vcv-details-content">
      {#each status.status.vaults as vault (vault.id)}
        <div class="vcv-detail-row">
          <span class="vcv-detail-label">{vault.display_name || vault.id}</span>
          <span class={statusBadgeClass(vault.connected ? 'valid' : 'expired')}>
            {vault.connected ? 'Connected' : (vault.error || 'Down')}
          </span>
        </div>
      {/each}
      {#if status.status.vaults.length === 0}
        <p>No vaults configured.</p>
      {/if}
    </div>
  {:else}
    <p>Loading…</p>
  {/if}
  {#snippet actions()}
    <button type="button" class="vcv-button vcv-button-secondary" onclick={refresh}>Refresh</button>
    <button type="button" class="vcv-button" onclick={() => (vaultModalOpen = false)}>Close</button>
  {/snippet}
</Modal>
