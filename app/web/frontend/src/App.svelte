<script lang="ts">
  import { onMount } from 'svelte'
  import { toast } from 'svelte-sonner'
  import ShieldCheck from '@lucide/svelte/icons/shield-check'
  import Moon from '@lucide/svelte/icons/moon'
  import RefreshCw from '@lucide/svelte/icons/refresh-cw'
  import Search from '@lucide/svelte/icons/search'
  import Sun from '@lucide/svelte/icons/sun'
  import { Toaster } from '$lib/components/ui/sonner'
  import * as Select from '$lib/components/ui/select'
  import CertDetailModal from '$lib/components/CertDetailModal.svelte'
  import CertTypeSelect from '$lib/components/CertTypeSelect.svelte'
  import VaultStatusPill from '$lib/components/VaultStatusPill.svelte'
  import ActiveFilters from '$lib/components/ActiveFilters.svelte'
  import ErrorBanner from '$lib/components/ErrorBanner.svelte'
  import MountSelectorDialog from '$lib/components/MountSelectorDialog.svelte'
  import StatusOverview from '$lib/components/StatusOverview.svelte'
  import CertTable from '$lib/components/CertTable.svelte'
  import CertMobileList from '$lib/components/CertMobileList.svelte'
  import PaginationBar from '$lib/components/PaginationBar.svelte'
  import CommandPalette from '$lib/components/CommandPalette.svelte'
  import { createCertsStore } from '$lib/stores/certs.svelte'
  import { createConfigStore } from '$lib/stores/config.svelte'
  import { createStatusStore } from '$lib/stores/status.svelte'
  import { createThemeStore } from '$lib/stores/theme.svelte'
  import { createI18nStore, setI18nContext, LANGUAGES } from '$lib/stores/i18n.svelte'
  import { parseCertID } from '$lib/utils/cert-status'
  import {
    matchesFilters,
    sortCerts,
    paginate,
    dashboardCounts,
    type CertTypeFilter,
    type SortDirection,
    type SortKey,
  } from '$lib/utils/cert-filter'
  import { parseUrlState, writeUrlState, type UrlState } from '$lib/utils/url-state'
  import { downloadExport, type ExportFormat } from '$lib/utils/export'
  import { expiryTier, shouldNotifyExpiry, type ExpiryTier } from '$lib/utils/expiry-notify'
  import type { Certificate, CertStatus } from '$lib/types'

  const i18n = setI18nContext(createI18nStore())
  const certs = createCertsStore(i18n)
  const config = createConfigStore()
  const status = createStatusStore()
  const theme = createThemeStore()
  const thresholds = $derived(config.thresholds)

  type StatusKey = 'valid' | 'warning' | 'critical' | 'expired' | 'revoked'
  const statusMeta = $derived<Record<StatusKey, { label: string; desc: string }>>({
    valid: { label: i18n.t('statusLabelValid', 'Valid'), desc: i18n.t('statusDescValid', 'All good') },
    warning: {
      label: i18n.t('statusLabelWarning', 'Warning'),
      desc: i18n.t('statusDescWarning', '≤ {days} days', { days: thresholds.warning }),
    },
    critical: {
      label: i18n.t('statusLabelCritical', 'Critical'),
      desc: i18n.t('statusDescCritical', '≤ {days} days', { days: thresholds.critical }),
    },
    expired: { label: i18n.t('statusLabelExpired', 'Expired'), desc: i18n.t('statusDescExpired', 'Past expiry') },
    revoked: { label: i18n.t('statusLabelRevoked', 'Revoked'), desc: i18n.t('statusDescRevoked', 'Revoked by CA') },
  })
  const langName = $derived(LANGUAGES.find((l) => l.code === i18n.lang)?.name ?? i18n.lang.toUpperCase())

  let search = $state('')
  /** Debounced search used by the filter pipeline (input stays snappy). */
  let searchForFilter = $state('')
  let statusFilters = $state<CertStatus[]>([])
  let certTypeFilter = $state<CertTypeFilter>('all')
  let mountFilter = $state<string[] | null>(null)
  let sortKey = $state<SortKey>('expiresAt')
  let sortDir = $state<SortDirection>('asc')
  let pageIndex = $state(0)
  let pageSize = $state<number | 'all'>(25)

  const urlDefaults: UrlState = {
    search: '',
    statusFilters: [],
    certTypeFilter: 'all',
    mountFilter: null,
    sortKey: 'expiresAt',
    sortDir: 'asc',
    pageSize: 25,
    pageIndex: 0,
  }

  let selected = $state<Certificate | null>(null)
  let certModalOpen = $state(false)
  let mountModalOpen = $state(false)
  let commandOpen = $state(false)
  let initialLoad = $state(true)
  let dismissedFetchError = $state<string | null>(null)
  let lastUpdated = $state<Date | null>(null)
  // Opt-in certificate auto-refresh interval in seconds; 0 = off (default).
  let autoRefreshSec = $state(0)
  const AUTO_REFRESH_OPTIONS = [0, 30, 60, 300]
  /** Last expiry tier we toasted, so auto-refresh does not spam identical alerts. */
  let lastNotifiedTier = $state<ExpiryTier>('none')

  const filtered = $derived(
    certs.certificates.filter((cert) =>
      matchesFilters(
        cert,
        {
          search: searchForFilter,
          statuses: statusFilters,
          certType: certTypeFilter,
          mounts: mountFilter,
        },
        thresholds,
      ),
    ),
  )
  const sorted = $derived(sortCerts(filtered, sortKey, sortDir))
  const pageSizeNum = $derived(pageSize === 'all' ? sorted.length || 1 : pageSize)
  const totalPages = $derived(Math.max(1, Math.ceil(sorted.length / pageSizeNum)))
  const safePage = $derived(Math.min(pageIndex, totalPages - 1))
  const paged = $derived(paginate(sorted, safePage, pageSize))
  function currentCounts() {
    return dashboardCounts(certs.certificates, thresholds)
  }
  const counts = $derived(currentCounts())
  const hasActiveFilters = $derived(
    !!search || statusFilters.length > 0 || certTypeFilter !== 'all' || mountFilter !== null,
  )

  const allMounts = $derived.by(() => {
    const set = new Set<string>()
    for (const cert of certs.certificates) {
      set.add(parseCertID(cert.id).mountKey)
    }
    return Array.from(set).sort()
  })
  const showVaultMount = $derived(allMounts.length > 1)

  let urlHydrated = $state(false)

  /** True only when the tab is visible; used to skip background polls. */
  function tabVisible(): boolean {
    return typeof document === 'undefined' || document.visibilityState === 'visible'
  }

  onMount(() => {
    const restored = parseUrlState(urlDefaults)
    search = restored.search
    searchForFilter = restored.search
    statusFilters = restored.statusFilters
    certTypeFilter = restored.certTypeFilter
    mountFilter = restored.mountFilter
    sortKey = restored.sortKey
    sortDir = restored.sortDir
    pageSize = restored.pageSize
    pageIndex = restored.pageIndex
    urlHydrated = true

    void (async () => {
      await i18n.ready
      try {
        await load(true)
      } catch {
        // Hard cert failure: ErrorBanner shows certs.error; avoid unhandled rejection.
      }
    })()
    const id = setInterval(() => {
      if (tabVisible()) void status.refresh()
    }, 10_000)
    return () => clearInterval(id)
  })

  // Debounce search used for filtering; keep the input bound to `search` for snappy typing.
  // Status / type / mount filters stay immediate (discrete UI).
  $effect(() => {
    const q = search
    const id = setTimeout(() => {
      if (searchForFilter !== q) {
        searchForFilter = q
        pageIndex = 0
      }
    }, 150)
    return () => clearTimeout(id)
  })

  // Opt-in certificate auto-refresh: re-poll on the chosen interval while not already loading.
  $effect(() => {
    if (autoRefreshSec <= 0) return
    const id = setInterval(() => {
      if (tabVisible() && !certs.loading) {
        void load().catch(() => {
          // Hard cert failure: ErrorBanner shows certs.error.
        })
      }
    }, autoRefreshSec * 1000)
    return () => clearInterval(id)
  })

  // Sync view state to the URL once initial state is restored, so links are shareable.
  // Debounced so rapid typing collapses into one replaceState call.
  $effect(() => {
    const snapshot: UrlState = {
      search,
      statusFilters,
      certTypeFilter,
      mountFilter,
      sortKey,
      sortDir,
      pageSize,
      pageIndex: safePage,
    }
    if (!urlHydrated) return
    // Read the reactive deps above; defer the history write so a burst of
    // keystrokes produces one replaceState instead of one per character.
    const id = setTimeout(() => writeUrlState(snapshot, urlDefaults), 150)
    return () => clearTimeout(id)
  })

  async function load(initial = false): Promise<void> {
    // Always reload public config so admin threshold edits land without a full page reload.
    const promises: Promise<void>[] = [certs.refresh(), status.refresh(), config.refresh()]
    if (initial) {
      try {
        await Promise.all(promises)
      } finally {
        initialLoad = false
      }
    } else {
      await Promise.all(promises)
    }
    // Config failure keeps last-good thresholds (config store). Cert hard-fail rejects for toast.
    if (certs.error) throw new Error(certs.error)
    lastUpdated = new Date()
    notifyExpiry(initial)
  }

  /**
   * Toast a single expiry summary after a load.
   * Initial load always notifies when tier ≠ none; later loads only when tier increases.
   */
  function notifyExpiry(isInitial: boolean): void {
    // Same algorithm as StatusOverview counts (avoid divergent threshold logic).
    const c = currentCounts()
    const tier = expiryTier(c)
    if (tier === 'none') {
      lastNotifiedTier = 'none'
      return
    }
    if (!shouldNotifyExpiry({ isInitial, tier, lastNotifiedTier })) return
    if (tier === 'critical') {
      toast.warning(
        i18n.t('notificationCritical', '{count} certificate(s) expiring within {threshold} days or fewer!', {
          count: c.critical,
          threshold: thresholds.critical,
        }),
      )
    } else {
      toast(
        i18n.t('notificationWarning', '{count} certificate(s) expiring within {threshold} days or fewer', {
          count: c.warning,
          threshold: thresholds.warning,
        }),
      )
    }
    lastNotifiedTier = tier
  }

  async function manualRefresh(): Promise<void> {
    toast.promise(load(), {
      loading: i18n.t('toastRefreshing', 'Refreshing…'),
      success: () => i18n.t('loadSuccess', 'Certificates loaded'),
      error: (err) =>
        err instanceof Error ? err.message : i18n.t('toastRefreshFailed', 'Refresh failed'),
    })
  }

  function exportCerts(format: ExportFormat): void {
    if (sorted.length === 0) {
      toast.error(i18n.t('exportEmpty', 'Nothing to export'))
      return
    }
    downloadExport(sorted, format, thresholds)
    toast.success(i18n.t('exportSuccess', 'Exported {count} certificate(s)', { count: sorted.length }))
  }

  // Bumped after each export so the Select remounts and the same format can be picked again.
  let exportNonce = $state(0)
  function onExportSelect(value: string | undefined): void {
    if (!value) return
    exportCerts(value as ExportFormat)
    exportNonce++
  }

  const SORT_OPTIONS = $derived<{ key: SortKey; label: string }[]>([
    { key: 'expiresAt', label: i18n.t('columnExpiresAt', 'Expires') },
    { key: 'commonName', label: i18n.t('columnCommonName', 'Common Name') },
    { key: 'vault', label: i18n.t('labelVault', 'Vault') },
    { key: 'pki', label: i18n.t('labelPki', 'PKI') },
  ])
  const sortKeyLabel = $derived(SORT_OPTIONS.find((o) => o.key === sortKey)?.label ?? sortKey)
  const resultCountText = $derived(
    i18n.t('dashboardResultCount', '{count} certificates', { count: sorted.length }),
  )

  function setSortKey(key: SortKey): void {
    sortKey = key
    sortDir = 'asc'
  }

  function toggleSortDir(): void {
    sortDir = sortDir === 'asc' ? 'desc' : 'asc'
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
    searchForFilter = ''
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

  function autoRefreshOptionLabel(seconds: number): string {
    if (seconds === 0) return i18n.t('autoRefreshOff', 'Off')
    if (seconds < 60) return `${seconds}s`
    return `${seconds / 60}m`
  }

  const lastUpdatedText = $derived(
    lastUpdated
      ? i18n.t('lastUpdatedLabel', 'Updated {time}', { time: lastUpdated.toLocaleTimeString() })
      : '',
  )

  function onSearchKeydown(event: KeyboardEvent): void {
    if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
      event.preventDefault()
      commandOpen = !commandOpen
      return
    }
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

<Toaster richColors position="bottom-right" theme={theme.theme} />

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
        <div class="vcv-auto-refresh">
          <span id="vcv-auto-refresh-label" class="vcv-auto-refresh-label">{i18n.t('autoRefreshLabel', 'Auto-refresh')}</span>
          <Select.Root
            type="single"
            value={String(autoRefreshSec)}
            onValueChange={(value) => value && (autoRefreshSec = Number(value))}
          >
            <Select.Trigger class="vcv-select vcv-auto-refresh-select h-9" aria-labelledby="vcv-auto-refresh-label">
              {autoRefreshOptionLabel(autoRefreshSec)}
            </Select.Trigger>
            <Select.Content>
              {#each AUTO_REFRESH_OPTIONS as seconds (seconds)}
                <Select.Item value={String(seconds)}>{autoRefreshOptionLabel(seconds)}</Select.Item>
              {/each}
            </Select.Content>
          </Select.Root>
          {#if lastUpdatedText}
            <span class="vcv-last-updated" aria-live="polite">{lastUpdatedText}</span>
          {/if}
        </div>
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
            aria-label={i18n.t('searchLabel', 'Search certificates')}
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
      onClearSearch={() => {
        search = ''
        searchForFilter = ''
        pageIndex = 0
      }}
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
      regionLabel={i18n.t('dashboardOverviewLabel', 'Certificate status overview')}
      onSelect={toggleStatus}
    />

    <div class="vcv-results-bar">
      <span class="vcv-results-count" aria-live="polite">{resultCountText}</span>
      <div class="vcv-results-actions">
        <div class="vcv-sort-control">
          <span id="vcv-sort-label" class="vcv-sort-control-label">{i18n.t('sortLabel', 'Sort')}</span>
          <Select.Root
            type="single"
            value={sortKey}
            onValueChange={(value) => value && setSortKey(value as SortKey)}
          >
            <Select.Trigger class="vcv-select vcv-sort-select h-9" aria-labelledby="vcv-sort-label">
              {sortKeyLabel}
            </Select.Trigger>
            <Select.Content>
              {#each SORT_OPTIONS as option (option.key)}
                <Select.Item value={option.key}>{option.label}</Select.Item>
              {/each}
            </Select.Content>
          </Select.Root>
          <button
            type="button"
            class="vcv-button vcv-button-icon vcv-sort-dir"
            data-direction={sortDir}
            aria-label={i18n.t('sortDirectionToggle', 'Toggle sort direction')}
            title={i18n.t('sortDirectionToggle', 'Toggle sort direction')}
            onclick={toggleSortDir}
          >
            {sortDir === 'asc' ? '↑' : '↓'}
          </button>
        </div>
        {#key exportNonce}
          <Select.Root type="single" onValueChange={onExportSelect}>
            <Select.Trigger
              class="vcv-select vcv-export-select h-9"
              aria-label={i18n.t('buttonExport', 'Export')}
              disabled={sorted.length === 0}
            >
              {i18n.t('buttonExport', 'Export')}
            </Select.Trigger>
            <Select.Content>
              <Select.Item value="csv">{i18n.t('exportCSV', 'Export CSV')}</Select.Item>
              <Select.Item value="json">{i18n.t('exportJSON', 'Export JSON')}</Select.Item>
            </Select.Content>
          </Select.Root>
        {/key}
      </div>
    </div>

    <CertTable
      certs={paged}
      loading={certs.loading}
      {initialLoad}
      hasInventory={certs.certificates.length > 0}
      {hasActiveFilters}
      {showVaultMount}
      {statusMeta}
      {thresholds}
      onSelect={selectCert}
      onClearFilters={clearAllFilters}
    />

    <CertMobileList
      certs={paged}
      loading={certs.loading}
      {initialLoad}
      hasInventory={certs.certificates.length > 0}
      {hasActiveFilters}
      {showVaultMount}
      {statusMeta}
      {thresholds}
      onSelect={selectCert}
      onClearFilters={clearAllFilters}
    />

    <PaginationBar
      {pageSize}
      {safePage}
      {totalPages}
      pageInfoText={pageInfoText()}
      onPageSizeChange={(size) => {
        pageSize = size
        pageIndex = 0
      }}
      onPrev={() => (pageIndex = Math.max(0, safePage - 1))}
      onNext={() => (pageIndex = Math.min(totalPages - 1, safePage + 1))}
    />
  </main>

  <footer class="vcv-footer" aria-label={i18n.t('footerLabel', 'Site information')}>
    <div class="vcv-footer-row">
      {#if status.status}
        <span class="vcv-footer-version">v{status.status.version}</span>
      {/if}
      <nav class="vcv-footer-links">
        <a class="vcv-footer-link" href="https://github.com/julienhmmt/vcv" target="_blank" rel="noopener noreferrer">GitHub</a>
        <span class="vcv-footer-sep" aria-hidden="true"></span>
        <a class="vcv-footer-link" href="https://github.com/julienhmmt/vcv/blob/main/LICENSE" target="_blank" rel="noopener noreferrer">{i18n.t('footerLicense', 'License')}</a>
        <span class="vcv-footer-sep" aria-hidden="true"></span>
        <a class="vcv-footer-link" href="https://j.hommet.net/vcv" target="_blank" rel="noopener noreferrer">{i18n.t('footerMoreInfo', 'More info')}</a>
      </nav>
    </div>
  </footer>
</div>

<CertDetailModal
  cert={selected}
  open={certModalOpen}
  onOpenChange={(value) => (certModalOpen = value)}
  {thresholds}
/>

<CommandPalette
  open={commandOpen}
  onOpenChange={(value) => (commandOpen = value)}
  certs={certs.certificates}
  theme={theme.theme}
  onSelectCert={selectCert}
  onToggleStatus={toggleStatus}
  onToggleTheme={theme.toggle}
  onSetLang={(code) => void i18n.setLang(code)}
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
