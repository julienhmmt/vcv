<script lang="ts">
  import { untrack } from 'svelte'
  import { Button } from '$lib/components/ui/button'
  import { Input } from '$lib/components/ui/input'
  import { Label } from '$lib/components/ui/label'
  import VaultEditor from './VaultEditor.svelte'
  import AdminDocsModal from './AdminDocsModal.svelte'
  import { getI18n } from '$lib/stores/i18n.svelte'
  import type { AdminVaultStatus, SettingsFile, VaultInstance } from '$lib/types'

  interface Props {
    settings: SettingsFile
    statuses: AdminVaultStatus[]
    loading: boolean
    error: string | null
    successMessage: string | null
    onSave: (next: SettingsFile) => void
    onAddVault: () => void
    onRemoveVault: (id: string) => void
    onInvalidateCache: () => void
    onLogout: () => void
  }

  const {
    settings,
    statuses,
    loading,
    error,
    successMessage,
    onSave,
    onAddVault,
    onRemoveVault,
    onInvalidateCache,
    onLogout,
  }: Props = $props()

  const i18n = getI18n()

  let working = $state<SettingsFile>(untrack(() => $state.snapshot(settings)))
  let lastSyncedRef: SettingsFile | null = null

  $effect(() => {
    if (settings !== lastSyncedRef) {
      lastSyncedRef = settings
      working = $state.snapshot(settings)
    }
  })

  const statusById = $derived.by(() => {
    const map = new Map<string, AdminVaultStatus>()
    for (const status of statuses) map.set(status.id, status)
    return map
  })

  function updateVault(index: number, next: VaultInstance): void {
    const vaults = [...working.vaults]
    vaults[index] = next
    working = { ...working, vaults }
  }

  function removeVault(index: number): void {
    const target = working.vaults[index]
    const vaults = working.vaults.filter((_, i) => i !== index)
    working = { ...working, vaults }
    if (target.id && statusById.has(target.id)) {
      onRemoveVault(target.id)
    }
  }

  const corsText = $derived((working.cors.allowed_origins ?? []).join(', '))

  function updateCors(value: string): void {
    working = {
      ...working,
      cors: {
        ...working.cors,
        allowed_origins: value
          .split(',')
          .map((part) => part.trim())
          .filter(Boolean),
      },
    }
  }

  function updateThreshold(field: 'critical' | 'warning', value: number): void {
    if (Number.isNaN(value)) return
    working = {
      ...working,
      certificates: {
        ...working.certificates,
        expiration_thresholds: { ...working.certificates.expiration_thresholds, [field]: value },
      },
    }
  }

  function updateMetric(field: 'per_certificate' | 'enhanced_metrics', value: boolean): void {
    working = { ...working, metrics: { ...working.metrics, [field]: value } }
  }

  function submit(event: SubmitEvent): void {
    event.preventDefault()
    onSave(working)
  }

  let docsOpen = $state(false)

  function openDocs(): void {
    docsOpen = true
  }

  const navItems = $derived([
    { id: 'thresholds', label: i18n.t('adminNavThresholds', 'Thresholds') },
    { id: 'metrics', label: i18n.t('adminNavMetrics', 'Metrics') },
    { id: 'cors', label: i18n.t('adminNavCors', 'CORS') },
    { id: 'vaults', label: i18n.t('adminNavVaults', 'Vaults') },
  ])
</script>

<div class="adm-layout">
  <!-- Top bar -->
  <header class="adm-topbar">
    <div class="adm-topbar-left">
      <span class="adm-topbar-title">VCV Admin</span>
      <span class="adm-topbar-sep">/</span>
      <span class="adm-topbar-sub">{i18n.t('adminTitle', 'Settings')}</span>
    </div>
    <nav class="adm-topbar-actions">
      <a href="/" class="adm-action-link">{i18n.t('adminBackToVCV', 'Back to VCV')}</a>
      <button type="button" class="adm-action-link" aria-label={i18n.t('adminDocsTitle', 'Documentation')} onclick={openDocs}>
        {i18n.t('adminDocsTitle', 'Docs')}
      </button>
      <button type="button" class="adm-action-btn adm-action-btn--secondary" aria-label={i18n.t('adminInvalidateCache', 'Flush cache')} onclick={onInvalidateCache}>
        {i18n.t('adminInvalidateCache', 'Flush cache')}
      </button>
      <button type="button" class="adm-action-btn adm-action-btn--ghost" aria-label={i18n.t('adminLogout', 'Sign out')} onclick={onLogout}>
        {i18n.t('adminLogout', 'Sign out')}
      </button>
    </nav>
  </header>

  <!-- Feedback bar -->
  {#if error}
    <div class="adm-feedback adm-feedback--error" role="alert">{error}</div>
  {/if}
  {#if successMessage}
    <div class="adm-feedback adm-feedback--success" role="status">{successMessage}</div>
  {/if}

  <!-- Body: nav + content -->
  <div class="adm-body">
    <!-- Sticky left nav -->
    <aside class="adm-sidenav">
      {#each navItems as item}
        <a href="#{item.id}" class="adm-sidenav-item">{item.label}</a>
      {/each}
    </aside>

    <!-- Scrollable settings form -->
    <form class="adm-content" onsubmit={submit}>

      <!-- Section: Thresholds -->
      <section class="adm-section" id="thresholds">
        <div class="adm-section-head">
          <h2 class="adm-section-title">{i18n.t('adminThresholdsTitle', 'Expiration thresholds')}</h2>
          <p class="adm-section-hint">{i18n.t('adminThresholdsHint', 'Days before a certificate is flagged.')}</p>
        </div>
        <div class="adm-grid adm-grid--2">
          <div class="adm-field">
            <Label class="adm-label">{i18n.t('adminCriticalThreshold', 'Critical (days)')}</Label>
            <Input
              type="number"
              min="1"
              max="3650"
              value={working.certificates.expiration_thresholds.critical}
              oninput={(event) => updateThreshold('critical', Number((event.target as HTMLInputElement).value))}
            />
          </div>
          <div class="adm-field">
            <Label class="adm-label">{i18n.t('adminWarningThreshold', 'Warning (days)')}</Label>
            <Input
              type="number"
              min="1"
              max="3650"
              value={working.certificates.expiration_thresholds.warning}
              oninput={(event) => updateThreshold('warning', Number((event.target as HTMLInputElement).value))}
            />
          </div>
        </div>
      </section>

      <hr class="adm-divider" />

      <!-- Section: Metrics -->
      <section class="adm-section" id="metrics">
        <div class="adm-section-head">
          <h2 class="adm-section-title">{i18n.t('adminMetrics', 'Metrics')}</h2>
          <p class="adm-section-hint">{i18n.t('adminMetricsHint', 'Prometheus scrape configuration.')}</p>
        </div>
        <div class="adm-toggles">
          <label class="adm-toggle">
            <input
              type="checkbox"
              name="metrics_per_certificate"
              class="adm-toggle-input"
              checked={working.metrics.per_certificate ?? false}
              onchange={(event) => updateMetric('per_certificate', (event.target as HTMLInputElement).checked)}
            />
            <div class="adm-toggle-body">
              <span class="adm-toggle-label">{i18n.t('adminMetricsPerCertificate', 'Per-certificate metrics')}</span>
              <span class="adm-toggle-desc">{i18n.t('adminMetricsPerCertificateDesc', 'High cardinality. Enable only if Prometheus is configured for it.')}</span>
            </div>
          </label>
          <label class="adm-toggle">
            <input
              type="checkbox"
              name="metrics_enhanced"
              class="adm-toggle-input"
              checked={working.metrics.enhanced_metrics ?? true}
              onchange={(event) => updateMetric('enhanced_metrics', (event.target as HTMLInputElement).checked)}
            />
            <div class="adm-toggle-body">
              <span class="adm-toggle-label">{i18n.t('adminMetricsEnhanced', 'Enhanced metrics')}</span>
              <span class="adm-toggle-desc">Additional gauges and histograms beyond the base set.</span>
            </div>
          </label>
        </div>
      </section>

      <hr class="adm-divider" />

      <!-- Section: CORS -->
      <section class="adm-section" id="cors">
        <div class="adm-section-head">
          <h2 class="adm-section-title">{i18n.t('adminCORSOrigins', 'CORS origins')}</h2>
          <p class="adm-section-hint">{i18n.t('adminCORSOriginsHint', 'Comma-separated list of allowed origins.')}</p>
        </div>
        <div class="adm-field">
          <Input
            value={corsText}
            placeholder="https://app.example.com, https://other.example.com"
            oninput={(event) => updateCors((event.target as HTMLInputElement).value)}
          />
        </div>
      </section>

      <hr class="adm-divider" />

      <!-- Section: Vaults -->
      <section class="adm-section" id="vaults">
        <div class="adm-section-head">
          <div class="adm-section-head-row">
            <div>
              <h2 class="adm-section-title">{i18n.t('adminVaults', 'Vaults')}</h2>
              <p class="adm-section-hint">{i18n.t('adminVaultsHint', 'Vault and OpenBao instances to read from.')}</p>
            </div>
            <Button type="button" variant="outline" size="sm" onclick={onAddVault}>
              {i18n.t('adminAddVault', '+ Add vault')}
            </Button>
          </div>
        </div>
        <div class="adm-vault-list">
          {#each working.vaults as vault, index (vault.id || index)}
            <VaultEditor
              {vault}
              status={statusById.get(vault.id)}
              onChange={(next) => updateVault(index, next)}
              onRemove={() => removeVault(index)}
            />
          {/each}
          {#if working.vaults.length === 0}
            <p class="adm-empty">{i18n.t('adminVaultsEmpty', 'No vaults configured.')}</p>
          {/if}
        </div>
      </section>

      <!-- Footer actions -->
      <div class="adm-form-footer">
        <p class="adm-footer-note">{i18n.t('adminRestartNote', '')}</p>
        <Button type="submit" disabled={loading}>
          {loading ? i18n.t('adminSaving', 'Saving…') : i18n.t('adminSaveSettings', 'Save settings')}
        </Button>
      </div>

    </form>
  </div>
</div>

<AdminDocsModal open={docsOpen} onOpenChange={(open) => (docsOpen = open)} />

<style>
  /* Layout */
  .adm-layout {
    display: flex;
    flex-direction: column;
    min-height: 100svh;
    background: var(--vcv-color-bg);
    color: var(--vcv-color-text);
  }

  /* Top bar */
  .adm-topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 1.5rem;
    height: 3rem;
    border-bottom: 1px solid var(--vcv-color-border);
    background: var(--vcv-color-surface);
    flex-shrink: 0;
    position: sticky;
    top: 0;
    z-index: 20;
  }

  .adm-topbar-left {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.8125rem;
  }

  .adm-topbar-title {
    font-weight: 600;
    color: var(--vcv-color-text-strong);
    letter-spacing: 0.01em;
  }

  .adm-topbar-sep {
    color: var(--vcv-color-border-strong);
  }

  .adm-topbar-sub {
    color: var(--vcv-color-muted);
  }

  .adm-topbar-actions {
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }

  .adm-action-link {
    padding: 0.25rem 0.6rem;
    font-size: 0.75rem;
    color: var(--vcv-color-muted);
    text-decoration: none;
    border-radius: var(--vcv-radius-sm);
    background: none;
    border: none;
    cursor: pointer;
    transition: color 0.12s, background 0.12s;
  }

  .adm-action-link:hover {
    color: var(--vcv-color-text);
    background: var(--vcv-color-bg-hover);
  }

  .adm-action-btn {
    padding: 0.25rem 0.75rem;
    font-size: 0.75rem;
    border-radius: var(--vcv-radius-sm);
    cursor: pointer;
    transition: background 0.12s, border-color 0.12s;
  }

  .adm-action-btn--secondary {
    background: var(--vcv-color-surface);
    border: 1px solid var(--vcv-color-border-strong);
    color: var(--vcv-color-text);
  }

  .adm-action-btn--secondary:hover {
    background: var(--vcv-color-bg-hover);
  }

  .adm-action-btn--ghost {
    background: none;
    border: 1px solid transparent;
    color: var(--vcv-color-muted);
  }

  .adm-action-btn--ghost:hover {
    color: var(--vcv-color-text);
    background: var(--vcv-color-bg-hover);
  }

  /* Feedback */
  .adm-feedback {
    padding: 0.6rem 1.5rem;
    font-size: 0.8125rem;
    border-bottom: 1px solid;
  }

  .adm-feedback--error {
    background: var(--vcv-color-danger-surface);
    border-color: var(--vcv-color-danger-border);
    color: var(--vcv-color-danger-text);
  }

  .adm-feedback--success {
    background: var(--vcv-color-success-surface);
    border-color: var(--vcv-color-success-border);
    color: var(--vcv-color-success-text);
  }

  /* Body */
  .adm-body {
    display: flex;
    flex: 1;
    min-height: 0;
  }

  /* Sidenav */
  .adm-sidenav {
    width: 11rem;
    flex-shrink: 0;
    padding: 2rem 1rem 2rem 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
    position: sticky;
    top: 3rem;
    height: calc(100svh - 3rem);
    overflow-y: auto;
    border-right: 1px solid var(--vcv-color-border);
    background: var(--vcv-color-surface);
  }

  .adm-sidenav-item {
    display: block;
    padding: 0.35rem 0.6rem;
    font-size: 0.8125rem;
    color: var(--vcv-color-muted);
    text-decoration: none;
    border-radius: var(--vcv-radius-sm);
    transition: color 0.12s, background 0.12s;
  }

  .adm-sidenav-item:hover {
    color: var(--vcv-color-text);
    background: var(--vcv-color-bg-hover);
  }

  /* Content */
  .adm-content {
    flex: 1;
    padding: 2.5rem 2.5rem 4rem;
    max-width: 48rem;
    overflow-y: auto;
  }

  /* Section */
  .adm-section {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
    scroll-margin-top: 4rem;
  }

  .adm-section-head {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .adm-section-head-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
  }

  .adm-section-title {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--vcv-color-text-strong);
    margin: 0;
  }

  .adm-section-hint {
    font-size: 0.75rem;
    color: var(--vcv-color-muted);
    margin: 0;
  }

  .adm-divider {
    border: none;
    border-top: 1px solid var(--vcv-color-border);
    margin: 2rem 0;
  }

  /* Grid */
  .adm-grid {
    display: grid;
    gap: 1rem;
  }

  .adm-grid--2 {
    grid-template-columns: repeat(2, 1fr);
  }

  @media (max-width: 640px) {
    .adm-grid--2 {
      grid-template-columns: 1fr;
    }
  }

  /* Field */
  .adm-field {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
  }

  :global(.adm-label) {
    font-size: 0.75rem;
    font-weight: 500;
    letter-spacing: 0.03em;
    text-transform: uppercase;
    color: var(--vcv-color-muted);
  }

  /* Toggles */
  .adm-toggles {
    display: flex;
    flex-direction: column;
    gap: 0;
    border: 1px solid var(--vcv-color-border);
    border-radius: var(--vcv-radius-md);
    overflow: hidden;
  }

  .adm-toggle {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    padding: 0.875rem 1rem;
    cursor: pointer;
    transition: background 0.1s;
    border-bottom: 1px solid var(--vcv-color-border);
  }

  .adm-toggle:last-child {
    border-bottom: none;
  }

  .adm-toggle:hover {
    background: var(--vcv-color-bg-hover);
  }

  .adm-toggle-input {
    margin-top: 0.1rem;
    accent-color: var(--vcv-color-primary);
    flex-shrink: 0;
  }

  .adm-toggle-body {
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
  }

  .adm-toggle-label {
    font-size: 0.8125rem;
    font-weight: 500;
    color: var(--vcv-color-text);
  }

  .adm-toggle-desc {
    font-size: 0.72rem;
    color: var(--vcv-color-muted);
  }

  /* Vault list */
  .adm-vault-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .adm-empty {
    font-size: 0.8125rem;
    color: var(--vcv-color-muted);
    padding: 1.5rem;
    text-align: center;
    border: 1px dashed var(--vcv-color-border);
    border-radius: var(--vcv-radius-md);
    margin: 0;
  }

  /* Footer */
  .adm-form-footer {
    margin-top: 2.5rem;
    padding-top: 1.5rem;
    border-top: 1px solid var(--vcv-color-border);
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .adm-footer-note {
    font-size: 0.72rem;
    color: var(--vcv-color-muted);
    margin: 0;
  }

  /* Responsive: turn the sidenav into a horizontal scroll bar below md */
  @media (max-width: 768px) {
    .adm-body {
      flex-direction: column;
    }

    .adm-sidenav {
      position: static;
      top: auto;
      width: auto;
      height: auto;
      flex-direction: row;
      gap: 0.25rem;
      padding: 0.5rem 1rem;
      overflow-x: auto;
      overflow-y: hidden;
      border-right: none;
      border-bottom: 1px solid var(--vcv-color-border);
    }

    .adm-sidenav-item {
      white-space: nowrap;
      flex-shrink: 0;
    }

    .adm-content {
      padding: 1.5rem 1rem 3rem;
      max-width: none;
    }
  }
</style>
