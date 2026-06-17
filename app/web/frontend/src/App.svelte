<script lang="ts">
  import { onMount } from 'svelte'
  import { toast } from 'svelte-sonner'
  import ShieldCheck from '@lucide/svelte/icons/shield-check'
  import ChevronRight from '@lucide/svelte/icons/chevron-right'
  import Moon from '@lucide/svelte/icons/moon'
  import RefreshCw from '@lucide/svelte/icons/refresh-cw'
  import Search from '@lucide/svelte/icons/search'
  import Sun from '@lucide/svelte/icons/sun'
  import { Toaster } from '$lib/components/ui/sonner'
  import { Skeleton } from '$lib/components/ui/skeleton'
  import * as Select from '$lib/components/ui/select'
  import CertDetailModal from '$lib/components/CertDetailModal.svelte'
  import CAModal from '$lib/components/CAModal.svelte'
  import CertTypeSelect from '$lib/components/CertTypeSelect.svelte'
  import VaultStatusPill from '$lib/components/VaultStatusPill.svelte'
  import ActiveFilters from '$lib/components/ActiveFilters.svelte'
  import ErrorBanner from '$lib/components/ErrorBanner.svelte'
  import MountSelectorDialog from '$lib/components/MountSelectorDialog.svelte'
  import StatusOverview from '$lib/components/StatusOverview.svelte'
  import CertCard from '$lib/components/CertCard.svelte'
  import { createCertsStore } from '$lib/stores/certs.svelte'
  import { createStatusStore } from '$lib/stores/status.svelte'
  import { createThemeStore } from '$lib/stores/theme.svelte'
  import { createI18nStore, setI18nContext, LANGUAGES } from '$lib/stores/i18n.svelte'
  import {
    certStatus,
    daysUntilExpiry,
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
    formatDate,
    formatTime,
    type CertTypeFilter,
    type SortDirection,
    type SortKey,
  } from '$lib/utils/cert-filter'
  import type { Certificate, CertStatus } from '$lib/types'

  const i18n = setI18nContext(createI18nStore())
  const certs = createCertsStore(i18n)
  const status = createStatusStore()
  const theme = createThemeStore()

  type StatusKey = 'valid' | 'warning' | 'critical' | 'expired' | 'revoked'
  const statusMeta = $derived<Record<StatusKey, { label: string; desc: string }>>({
    valid: { label: i18n.t('statusLabelValid', 'Valid'), desc: i18n.t('statusDescValid', 'All good') },
    warning: {
      label: i18n.t('statusLabelWarning', 'Warning'),
      desc: i18n.t('statusDescWarning', '≤ {days} days', { days: DEFAULT_THRESHOLDS.warning }),
    },
    critical: {
      label: i18n.t('statusLabelCritical', 'Critical'),
      desc: i18n.t('statusDescCritical', '≤ {days} days', { days: DEFAULT_THRESHOLDS.critical }),
    },
    expired: { label: i18n.t('statusLabelExpired', 'Expired'), desc: i18n.t('statusDescExpired', 'Past expiry') },
    revoked: { label: i18n.t('statusLabelRevoked', 'Revoked'), desc: i18n.t('statusDescRevoked', 'Revoked by CA') },
  })
  const langName = $derived(LANGUAGES.find((l) => l.code === i18n.lang)?.name ?? i18n.lang.toUpperCase())

  /** Localized expiry label for the table: compact "{n}d" ahead, descriptive when due/past. */
  function expiryLabel(cert: Certificate): string {
    const days = daysUntilExpiry(cert)
    if (days > 0) return i18n.t('daysRemainingShort', '{days}d', { days })
    if (days === 0) return i18n.t('expiringToday', 'Expires today')
    const ago = Math.abs(days)
    return i18n.t(ago === 1 ? 'expiredDaysSingular' : 'expiredDays', 'Expired {days} days ago', { days: ago })
  }

  let search = $state('')
  let statusFilters = $state<CertStatus[]>([])
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
  let dismissedFetchError = $state<string | null>(null)

  const filtered = $derived(
    certs.certificates.filter((cert) =>
      matchesFilters(cert, { search, statuses: statusFilters, certType: certTypeFilter, mounts: mountFilter }),
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
    void (async () => {
      await i18n.ready
      await load(true)
    })()
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
      if (!certs.error) notifyExpiry()
      return
    }
    await Promise.all(promises)
    if (!certs.error) notifyExpiry()
  }

  /** Toast a single expiry summary after a load (critical takes precedence over warning). */
  function notifyExpiry(): void {
    const c = dashboardCounts(certs.certificates, DEFAULT_THRESHOLDS)
    if (c.critical > 0) {
      toast.warning(
        i18n.t('notificationCritical', '{count} certificate(s) expiring within {threshold} days or fewer!', {
          count: c.critical,
          threshold: DEFAULT_THRESHOLDS.critical,
        }),
      )
    } else if (c.warning > 0) {
      toast(
        i18n.t('notificationWarning', '{count} certificate(s) expiring within {threshold} days or fewer', {
          count: c.warning,
          threshold: DEFAULT_THRESHOLDS.warning,
        }),
      )
    }
  }

  async function manualRefresh(): Promise<void> {
    toast.promise(load(), {
      loading: i18n.t('toastRefreshing', 'Refreshing…'),
      success: () => i18n.t('loadSuccess', 'Certificates loaded'),
      error: i18n.t('toastRefreshFailed', 'Refresh failed'),
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

  function toggleStatus(next: CertStatus): void {
    statusFilters = statusFilters.includes(next)
      ? statusFilters.filter((s) => s !== next)
      : [...statusFilters, next]
    pageIndex = 0
  }

  function removeStatus(key: CertStatus): void {
    statusFilters = statusFilters.filter((s) => s !== key)
    pageIndex = 0
  }

  function clearAllFilters(): void {
    search = ''
    statusFilters = []
    certTypeFilter = 'all'
    mountFilter = null
    pageIndex = 0
  }

  function pageInfoText(): string {
    if (sorted.length === 0) return i18n.t('paginationResults', '{count} results', { count: 0 })
    if (pageSize === 'all') return i18n.t('paginationResults', '{count} results', { count: sorted.length })
    const start = safePage * (pageSize as number) + 1
    const end = Math.min(start + (pageSize as number) - 1, sorted.length)
    return i18n.t('paginationRange', '{start}–{end} of {total}', { start, end, total: sorted.length })
  }

  function onSearchKeydown(event: KeyboardEvent): void {
    if (event.key === '/' && document.activeElement?.tagName !== 'INPUT' && document.activeElement?.tagName !== 'TEXTAREA') {
      event.preventDefault()
      const el = document.getElementById('vcv-search') as HTMLInputElement | null
      el?.focus()
    }
  }

  $effect(() => {
    if (certs.error && certs.error !== dismissedFetchError) {
      toast.error(certs.error)
    }
  })

  // Toast vault connectivity transitions detected by the status poll.
  const prevConnected = new Map<string, boolean>()
  let statusSeeded = false
  $effect(() => {
    const vaults = status.status?.vaults
    if (!vaults) return
    for (const vault of vaults) {
      const was = prevConnected.get(vault.id)
      if (statusSeeded && was !== undefined && was !== vault.connected) {
        const name = vault.display_name || vault.id
        if (vault.connected) {
          toast.success(`${i18n.t('vaultConnectionRestored', 'Vault connection restored')} — ${name}`)
        } else {
          toast.error(`${i18n.t('vaultConnectionLost', 'Vault connection lost')} — ${name}`)
        }
      }
      prevConnected.set(vault.id, vault.connected)
    }
    statusSeeded = true
  })
</script>

<svelte:window onkeydown={onSearchKeydown} />

<Toaster richColors position="bottom-right" />

<a href="#vcv-main-content" class="vcv-skip-link">{i18n.t('skipToContent', 'Skip to main content')}</a>

<div class="vcv-layout">
  <header class="vcv-header">
    <div class="vcv-header-bar">
      <div class="vcv-header-brand">
        <span class="vcv-brand-mark" aria-hidden="true">
          <ShieldCheck class="h-5 w-5" />
        </span>
        <div class="vcv-brand-text">
          <h1 class="vcv-title">
            VaultCertsViewer
            {#if status.status}<span class="vcv-title-version">v{status.status.version}</span>{/if}
          </h1>
          <p class="vcv-title-subtitle">{i18n.t('appSubtitle', 'Inspect certificates across Vault / OpenBao PKI mounts')}</p>
        </div>
      </div>
      <div class="vcv-header-actions">
        <VaultStatusPill status={status.status} loading={status.loading} onRefresh={() => void status.refresh()} />
        <button
          class="vcv-button vcv-button-icon"
          type="button"
          title={i18n.t('buttonRefresh', 'Refresh')}
          aria-label={i18n.t('buttonRefresh', 'Refresh')}
          onclick={manualRefresh}
          disabled={certs.loading}
        >
          <RefreshCw class="h-4 w-4 {certs.loading ? 'animate-spin' : ''}" />
        </button>
        <button
          class="vcv-button vcv-button-icon vcv-theme-toggle"
          type="button"
          title={i18n.t('buttonToggleTheme', 'Toggle dark mode')}
          aria-label={i18n.t('buttonToggleTheme', 'Toggle dark mode')}
          onclick={theme.toggle}
        >
          {#if theme.theme === 'dark'}
            <Sun class="h-4 w-4" />
          {:else}
            <Moon class="h-4 w-4" />
          {/if}
        </button>
        <Select.Root type="single" value={i18n.lang} onValueChange={(value) => value && void i18n.setLang(value)}>
          <Select.Trigger class="vcv-select vcv-lang-select h-9" aria-label={i18n.t('labelLanguage', 'Language')}>
            <span class="vcv-lang-icon" aria-hidden="true">🌐</span>
            <span class="vcv-lang-name">{langName}</span>
          </Select.Trigger>
          <Select.Content>
            {#each LANGUAGES as language (language.code)}
              <Select.Item value={language.code}>{language.name}</Select.Item>
            {/each}
          </Select.Content>
        </Select.Root>
      </div>
    </div>

    <div id="vcv-filter-bar" class="vcv-filter-bar">
      <div class="vcv-filter-bar-inner">
        <div class="vcv-filter-palette">
          <div class="vcv-palette-item">
            <span class="vcv-palette-label">{i18n.t('filterChipSources', 'Sources')}</span>
            <button type="button" class="vcv-mount-filter" onclick={() => (mountModalOpen = true)}>
              {#if mountFilter === null || mountFilter.length === allMounts.length}
                {i18n.t('sourcesButtonAll', 'All mounts ({total})', { total: allMounts.length })}
              {:else}
                {i18n.t('sourcesButtonPartial', '{selected} / {total} mounts', {
                  selected: mountFilter.length,
                  total: allMounts.length,
                })}
              {/if}
            </button>
          </div>
          <span class="vcv-palette-separator" aria-hidden="true"></span>
          <div class="vcv-palette-item">
            <span class="vcv-palette-label">{i18n.t('filterChipCertType', 'Type')}</span>
            <CertTypeSelect value={certTypeFilter} onChange={(next) => { certTypeFilter = next; pageIndex = 0 }} />
          </div>
        </div>
        <div class="vcv-search-wrapper">
          <Search class="vcv-search-icon h-[18px] w-[18px]" aria-hidden="true" />
          <input
            id="vcv-search"
            class="vcv-input vcv-input-search"
            type="search"
            placeholder={i18n.t('searchPlaceholder', 'Search certificates, serials, SANs…')}
            bind:value={search}
            oninput={() => (pageIndex = 0)}
          />
          <kbd class="vcv-search-shortcut" aria-label={i18n.t('searchShortcutHint', 'Press / to focus search')}>/</kbd>
        </div>
      </div>
    </div>

    <ActiveFilters
      {search}
      {statusFilters}
      {certTypeFilter}
      {mountFilter}
      allMountsCount={allMounts.length}
      onClearSearch={() => (search = '')}
      onRemoveStatus={removeStatus}
      onClearCertType={() => (certTypeFilter = 'all')}
      onClearMounts={() => (mountFilter = null)}
      onClearAll={clearAllFilters}
    />
  </header>

  {#if certs.error && certs.error !== dismissedFetchError}
    <ErrorBanner message={certs.error} onDismiss={() => (dismissedFetchError = certs.error)} />
  {/if}

  {#if certs.vaultErrors.length > 0}
    <div class="vcv-vault-error-banner" role="status" aria-live="polite">
      <strong>{i18n.t('vaultsUnreachable', '{count} vault(s) unreachable', { count: certs.vaultErrors.length })}</strong>
      {i18n.t('vaultsUnreachableHint', 'Showing partial results.')}
      <details>
        <summary>{i18n.t('buttonDetails', 'Details')}</summary>
        <ul>
          {#each certs.vaultErrors as vaultError (vaultError.vaultId)}
            <li><code>{vaultError.vaultId}</code>: {vaultError.message}</li>
          {/each}
        </ul>
      </details>
    </div>
  {/if}

  <main id="vcv-main-content">
    <StatusOverview
      counts={{ valid: counts.valid, warning: counts.warning, critical: counts.critical, expired: counts.expired, revoked: counts.revoked }}
      meta={statusMeta}
      {statusFilters}
      donutLabel={i18n.t('dashboardCertsLabel', 'certs')}
      regionLabel={i18n.t('dashboardOverviewLabel', 'Certificate status overview')}
      onSelect={toggleStatus}
    />

    <div class="vcv-table-footer">
      <div class="vcv-page-size">
        <span id="vcv-page-size-label">{i18n.t('paginationPageSizeLabel', 'Results per page')}</span>
        <Select.Root
          type="single"
          value={String(pageSize)}
          onValueChange={(value) => {
            if (!value) return
            pageSize = value === 'all' ? 'all' : Number(value)
            pageIndex = 0
          }}
        >
          <Select.Trigger class="vcv-select vcv-page-size-select h-9" aria-labelledby="vcv-page-size-label">
            {pageSize === 'all' ? i18n.t('paginationPageSizeAll', 'All') : String(pageSize)}
          </Select.Trigger>
          <Select.Content>
            <Select.Item value="25">25</Select.Item>
            <Select.Item value="50">50</Select.Item>
            <Select.Item value="100">100</Select.Item>
            <Select.Item value="all">{i18n.t('paginationPageSizeAll', 'All')}</Select.Item>
          </Select.Content>
        </Select.Root>
      </div>
      <div class="vcv-page-buttons">
        <button
          type="button"
          class="vcv-button vcv-button-small vcv-button-ghost vcv-button-pill"
          disabled={safePage === 0}
          onclick={() => (pageIndex = Math.max(0, safePage - 1))}
        >
          {i18n.t('paginationPrev', 'Previous')}
        </button>
        <span class="vcv-page-info">{pageInfoText()}</span>
        <span class="vcv-badge vcv-badge-neutral"
          >{i18n.t('paginationInfo', 'Page {current} / {total}', { current: safePage + 1, total: totalPages })}</span
        >
        <button
          type="button"
          class="vcv-button vcv-button-small vcv-button-ghost vcv-button-pill"
          disabled={safePage >= totalPages - 1}
          onclick={() => (pageIndex = Math.min(totalPages - 1, safePage + 1))}
        >
          {i18n.t('paginationNext', 'Next')}
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
              <th
                scope="col"
                aria-sort={sortKey === 'commonName' ? (sortDir === 'asc' ? 'ascending' : 'descending') : 'none'}
              >
                <button
                  type="button"
                  class="vcv-sort"
                  data-direction={sortKey === 'commonName' ? sortDir : 'asc'}
                  onclick={() => toggleSort('commonName')}
                >
                  <span class="vcv-sort-label">{i18n.t('columnCommonName', 'Common Name')}</span>
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
                      <span class="vcv-sort-label">{i18n.t('labelVault', 'Vault')}</span>
                      <span class="vcv-sort-indicator" aria-hidden="true"></span>
                    </button>
                    <button
                      type="button"
                      class="vcv-sort"
                      data-direction={sortKey === 'pki' ? sortDir : 'asc'}
                      onclick={() => toggleSort('pki')}
                    >
                      <span class="vcv-sort-label">{i18n.t('labelPki', 'PKI')}</span>
                      <span class="vcv-sort-indicator" aria-hidden="true"></span>
                    </button>
                  </div>
                {/if}
              </th>
              <th
                scope="col"
                aria-sort={sortKey === 'expiresAt' ? (sortDir === 'asc' ? 'ascending' : 'descending') : 'none'}
              >
                <button
                  type="button"
                  class="vcv-sort"
                  data-direction={sortKey === 'expiresAt' ? sortDir : 'asc'}
                  onclick={() => toggleSort('expiresAt')}
                >
                  <span class="vcv-sort-label">{i18n.t('columnExpiresAt', 'Expires')}</span>
                  <span class="vcv-sort-indicator" aria-hidden="true"></span>
                </button>
              </th>
              <th scope="col" class="vcv-col-status">{i18n.t('columnStatus', 'Status')}</th>
            </tr>
          </thead>
          <tbody>
            {#if paged.length === 0}
              <tr>
                <td colspan="3" class="vcv-empty-row">
                  {certs.loading
                    ? i18n.t('labelLoading', 'Loading…')
                    : i18n.t('tableNoMatch', 'No certificates match the current filters.')}
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
                    <div class="vcv-expiry-count vcv-days-{s}">{expiryLabel(cert)}</div>
                    <div class="vcv-expiry-datetime">
                      <span class="vcv-expiry-date">{formatDate(cert.expiresAt)}</span>
                      <span class="vcv-date-secondary">· {formatTime(cert.expiresAt)} UTC</span>
                    </div>
                  </td>
                  <td class="vcv-col-status">
                    <div class="vcv-status-cell">
                      <div class="vcv-status-badges">
                        <span class={statusBadgeClass(s)}>{statusMeta[s].label}</span>
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

    <div class="vcv-certs-mobile-cards">
      {#if initialLoad && certs.certificates.length === 0}
        {#each Array(6) as _, i (i)}
          <div class="vcv-cert-card">
            <Skeleton class="h-5 w-3/4" />
            <Skeleton class="h-4 w-1/2" />
          </div>
        {/each}
      {:else if paged.length === 0}
        <p class="vcv-certs-mobile-empty">
          {certs.loading
            ? i18n.t('labelLoading', 'Loading…')
            : i18n.t('tableNoMatch', 'No certificates match the current filters.')}
        </p>
      {:else}
        {#each paged as cert (cert.id)}
          {@const s = certStatus(cert, DEFAULT_THRESHOLDS)}
          <CertCard {cert} {showVaultMount} statusLabel={statusMeta[s].label} onSelect={selectCert} />
        {/each}
      {/if}
    </div>
  </main>

  <footer class="vcv-footer" aria-label="Footer">
    <div class="vcv-footer-legend">
      <div class="vcv-legend-item"><span class="vcv-badge vcv-badge-valid">{statusMeta.valid.label}</span><span class="vcv-legend-text">{statusMeta.valid.desc}</span></div>
      <div class="vcv-legend-item"><span class="vcv-badge vcv-badge-warning">{statusMeta.warning.label}</span><span class="vcv-legend-text">{statusMeta.warning.desc}</span></div>
      <div class="vcv-legend-item"><span class="vcv-badge vcv-badge-critical">{statusMeta.critical.label}</span><span class="vcv-legend-text">{statusMeta.critical.desc}</span></div>
      <div class="vcv-legend-item"><span class="vcv-badge vcv-badge-expired">{statusMeta.expired.label}</span><span class="vcv-legend-text">{statusMeta.expired.desc}</span></div>
      <div class="vcv-legend-item"><span class="vcv-badge vcv-badge-revoked">{statusMeta.revoked.label}</span><span class="vcv-legend-text">{statusMeta.revoked.desc}</span></div>
    </div>
    <div class="vcv-footer-bottom">
      <div class="vcv-footer-brand">
        <div class="vcv-footer-title">
          VaultCertsViewer
          {#if status.status}<span class="vcv-footer-version">v{status.status.version}</span>{/if}
        </div>
        <div class="vcv-footer-meta">
          <span>{i18n.t('footerLicense', 'License')}: <a class="vcv-footer-inline-link" href="https://github.com/julienhmmt/vcv/blob/main/LICENSE" target="_blank" rel="noopener">GNU Affero GPL v3.0</a></span>
          <span class="vcv-footer-divider">•</span>
          <span>Imagined and designed by <a href="https://j.hommet.net" target="_blank" rel="noopener">Julien HOMMET</a>, developed by AI.</span>
        </div>
      </div>
      <div class="vcv-footer-links" aria-label="External links">
        <a class="vcv-footer-link" href="https://hub.docker.com/r/jhmmt/vcv" target="_blank" rel="noopener">
          Docker Hub
        </a>
        <a class="vcv-footer-link" href="https://github.com/julienhmmt/vcv" target="_blank" rel="noopener">
          GitHub
        </a>
        <a class="vcv-footer-link" href="https://j.hommet.net/vcv" target="_blank" rel="noopener">{i18n.t('footerMoreInfo', 'More info')}</a>
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
