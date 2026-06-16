<script lang="ts">
  import { getI18n } from '$lib/stores/i18n.svelte'
  import type { AdminVaultStatus, VaultInstance } from '$lib/types'

  interface Props {
    vault: VaultInstance
    status?: AdminVaultStatus
    onChange: (next: VaultInstance) => void
    onRemove: () => void
  }

  const { vault, status, onChange, onRemove }: Props = $props()
  const i18n = getI18n()
  const uid = $props.id()

  let expanded = $state(true)
  let showToken = $state(false)

  function toggleExpanded(): void {
    expanded = !expanded
  }

  function onToggleClick(event: MouseEvent): void {
    event.stopPropagation()
    toggleExpanded()
  }

  function toggleToken(): void {
    showToken = !showToken
  }

  function onSummaryKeydown(event: KeyboardEvent): void {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      toggleExpanded()
    }
  }

  const mountsText = $derived((vault.pki_mounts ?? []).join(', '))

  function update<K extends keyof VaultInstance>(field: K, value: VaultInstance[K]): void {
    onChange({ ...vault, [field]: value })
  }

  function updateMounts(value: string): void {
    const mounts = value
      .split(',')
      .map((part) => part.trim())
      .filter(Boolean)
    onChange({
      ...vault,
      pki_mounts: mounts,
      pki_mount: mounts[0] ?? vault.pki_mount ?? 'pki',
    })
  }

  const enabled = $derived(vault.enabled !== false)

  type StatusKind = 'connected' | 'disconnected' | 'disabled' | 'unknown'

  function statusKind(): StatusKind {
    if (!status) return 'unknown'
    if (!status.enabled) return 'disabled'
    return status.connected ? 'connected' : 'disconnected'
  }

  const statusLabels = $derived<Record<StatusKind, string>>({
    connected: i18n.t('adminVaultConnected', 'Connected'),
    disconnected: i18n.t('adminVaultDisconnected', 'Disconnected'),
    disabled: i18n.t('adminVaultDisabled', 'Disabled'),
    unknown: i18n.t('adminVaultUnknown', 'Unknown'),
  })

  const kind = $derived(statusKind())
  const vaultLabel = $derived(vault.display_name || vault.id || i18n.t('adminVaultNew', 'new vault'))
</script>

<div class="ve-card" class:ve-card--collapsed={!expanded}>
  <!-- Summary row (always visible) -->
  <div class="ve-summary" role="button" tabindex="0" aria-expanded={expanded} aria-controls="ve-body-{uid}" onclick={toggleExpanded} onkeydown={onSummaryKeydown}>
    <span class="ve-status-dot ve-status-dot--{kind}" title={statusLabels[kind]}></span>
    <div class="ve-summary-info">
      <span class="ve-summary-id">{vaultLabel}</span>
      {#if vault.address}
        <span class="ve-summary-addr">{vault.address}</span>
      {/if}
    </div>
    <span class="ve-status-label ve-status-label--{kind}">{statusLabels[kind]}</span>
    <button
      type="button"
      class="ve-toggle-btn"
      tabindex="-1"
      onclick={onToggleClick}
      aria-label={expanded ? 'Collapse' : 'Expand'}
    >
      <svg class="ve-toggle-icon" class:ve-toggle-icon--open={expanded} viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <path d="M4 6l4 4 4-4"/>
      </svg>
    </button>
  </div>

  <!-- Expanded content -->
  {#if expanded}
    <div class="ve-body" id="ve-body-{uid}">
      <!-- Row: enabled toggle + remove -->
      <div class="ve-control-row">
        <label class="ve-enabled-toggle">
          <input
            type="checkbox"
            name="ve-enabled-{uid}"
            checked={enabled}
            onchange={(event) => update('enabled', (event.target as HTMLInputElement).checked)}
          />
          <span>{i18n.t('adminVaultEnabled', 'Enabled')}</span>
        </label>
        <button type="button" class="ve-remove-btn" aria-label="{i18n.t('adminVaultRemove', 'Remove vault')}: {vaultLabel}" onclick={onRemove}>
          {i18n.t('adminVaultRemove', 'Remove vault')}
        </button>
      </div>

      <!-- Identity fields -->
      <div class="ve-grid ve-grid--2">
        <div class="ve-field">
          <label class="ve-label" for="ve-id-{uid}">{i18n.t('adminVaultID', 'ID')}</label>
          <input
            id="ve-id-{uid}"
            class="ve-input"
            type="text"
            value={vault.id}
            oninput={(event) => update('id', (event.target as HTMLInputElement).value)}
          />
        </div>
        <div class="ve-field">
          <label class="ve-label" for="ve-name-{uid}">{i18n.t('adminVaultDisplayName', 'Display name')}</label>
          <input
            id="ve-name-{uid}"
            class="ve-input"
            type="text"
            value={vault.display_name ?? ''}
            oninput={(event) => update('display_name', (event.target as HTMLInputElement).value)}
          />
        </div>
        <div class="ve-field ve-field--full">
          <label class="ve-label" for="ve-addr-{uid}">{i18n.t('adminVaultAddress', 'Address')}</label>
          <input
            id="ve-addr-{uid}"
            class="ve-input"
            type="text"
            value={vault.address}
            placeholder="https://vault.example.com"
            oninput={(event) => update('address', (event.target as HTMLInputElement).value)}
          />
        </div>
      </div>

      <!-- Token -->
      <div class="ve-field">
        <div class="ve-label-row">
          <label class="ve-label" for="ve-token-{uid}">{i18n.t('adminVaultToken', 'Token')}</label>
          <button
            type="button"
            class="ve-reveal-btn"
            onclick={toggleToken}
          >
            {showToken ? i18n.t('adminVaultTokenHide', 'Hide') : i18n.t('adminVaultTokenReveal', 'Reveal')}
          </button>
        </div>
        <input
          id="ve-token-{uid}"
          class="ve-input ve-input--mono"
          type={showToken ? 'text' : 'password'}
          value={vault.token}
          placeholder="(unchanged)"
          oninput={(event) => update('token', (event.target as HTMLInputElement).value)}
        />
        <p class="ve-hint">{i18n.t('adminVaultTokenHint', 'Leave blank to keep the existing token.')}</p>
      </div>

      <!-- PKI Mounts -->
      <div class="ve-field">
        <label class="ve-label" for="ve-mounts-{uid}">{i18n.t('adminVaultPKIMounts', 'PKI mounts')}</label>
        <input
          id="ve-mounts-{uid}"
          class="ve-input"
          type="text"
          value={mountsText}
          placeholder="pki, pki_int"
          oninput={(event) => updateMounts((event.target as HTMLInputElement).value)}
        />
        <p class="ve-hint">Comma-separated. First mount is the default.</p>
      </div>

      <!-- TLS section -->
      <details class="ve-tls-details">
        <summary class="ve-tls-summary">TLS options</summary>
        <div class="ve-tls-body">
          <div class="ve-field">
            <label class="ve-label" for="ve-tls-ca-b64-{uid}">{i18n.t('adminVaultTLSCABase64', 'CA cert (base64)')}</label>
            <input
              id="ve-tls-ca-b64-{uid}"
              class="ve-input ve-input--mono"
              type="text"
              value={vault.tls_ca_cert_base64 ?? ''}
              oninput={(event) => update('tls_ca_cert_base64', (event.target as HTMLInputElement).value)}
            />
          </div>
          <div class="ve-grid ve-grid--2">
            <div class="ve-field">
              <label class="ve-label" for="ve-tls-ca-{uid}">{i18n.t('adminVaultTLSCAFile', 'CA cert file path')}</label>
              <input
                id="ve-tls-ca-{uid}"
                class="ve-input"
                type="text"
                value={vault.tls_ca_cert ?? ''}
                oninput={(event) => update('tls_ca_cert', (event.target as HTMLInputElement).value)}
              />
            </div>
            <div class="ve-field">
              <label class="ve-label" for="ve-tls-capath-{uid}">{i18n.t('adminVaultTLSCAPath', 'CA directory path')}</label>
              <input
                id="ve-tls-capath-{uid}"
                class="ve-input"
                type="text"
                value={vault.tls_ca_path ?? ''}
                oninput={(event) => update('tls_ca_path', (event.target as HTMLInputElement).value)}
              />
            </div>
          </div>
          <div class="ve-field">
            <label class="ve-label" for="ve-tls-sni-{uid}">{i18n.t('adminVaultTLSServerName', 'SNI server name')}</label>
            <input
              id="ve-tls-sni-{uid}"
              class="ve-input"
              type="text"
              value={vault.tls_server_name ?? ''}
              oninput={(event) => update('tls_server_name', (event.target as HTMLInputElement).value)}
            />
          </div>
          <label class="ve-enabled-toggle">
            <input
              type="checkbox"
              name="ve-tls-insecure-{uid}"
              checked={vault.tls_insecure ?? false}
              onchange={(event) => update('tls_insecure', (event.target as HTMLInputElement).checked)}
            />
            <span>{i18n.t('adminVaultTLSInsecure', 'Skip TLS verification')}</span>
          </label>
          <p class="ve-hint ve-hint--warn">{i18n.t('adminVaultTLSTip', 'Do not disable TLS verification in production.')}</p>
        </div>
      </details>
    </div>
  {/if}
</div>

<style>
  .ve-card {
    border: 1px solid var(--vcv-color-border);
    border-radius: var(--vcv-radius-md);
    background: var(--vcv-color-surface);
    overflow: hidden;
  }

  /* Summary row */
  .ve-summary {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem 1rem;
    cursor: pointer;
    user-select: none;
    transition: background 0.1s;
  }

  .ve-summary:hover {
    background: var(--vcv-color-bg-hover);
  }

  .ve-summary-info {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
  }

  .ve-summary-id {
    font-size: 0.8125rem;
    font-weight: 500;
    color: var(--vcv-color-text-strong);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .ve-summary-addr {
    font-size: 0.7rem;
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    color: var(--vcv-color-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* Status dot */
  .ve-status-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .ve-status-dot--connected { background: var(--vcv-color-primary); }
  .ve-status-dot--disconnected { background: var(--vcv-color-danger); }
  .ve-status-dot--disabled { background: var(--vcv-color-muted); }
  .ve-status-dot--unknown { background: var(--vcv-color-border-strong); }

  /* Status label */
  .ve-status-label {
    font-size: 0.7rem;
    font-weight: 500;
    letter-spacing: 0.03em;
    text-transform: uppercase;
    padding: 0.15rem 0.45rem;
    border-radius: var(--vcv-radius-sm);
  }

  .ve-status-label--connected {
    color: var(--vcv-color-success-text);
    background: var(--vcv-color-success-surface);
  }

  .ve-status-label--disconnected {
    color: var(--vcv-color-danger-text);
    background: var(--vcv-color-danger-surface);
  }

  .ve-status-label--disabled {
    color: var(--vcv-color-muted);
    background: var(--vcv-color-surface-muted);
  }

  .ve-status-label--unknown {
    color: var(--vcv-color-muted);
    background: var(--vcv-color-surface-muted);
  }

  /* Toggle button */
  .ve-toggle-btn {
    background: none;
    border: none;
    padding: 0.25rem;
    cursor: pointer;
    color: var(--vcv-color-muted);
    display: flex;
    align-items: center;
    border-radius: var(--vcv-radius-sm);
    transition: color 0.1s, background 0.1s;
  }

  .ve-toggle-btn:hover {
    color: var(--vcv-color-text);
    background: var(--vcv-color-bg-hover);
  }

  .ve-toggle-icon {
    width: 16px;
    height: 16px;
    transition: transform 0.18s ease-out;
  }

  .ve-toggle-icon--open {
    transform: rotate(180deg);
  }

  /* Body */
  .ve-body {
    padding: 1rem;
    border-top: 1px solid var(--vcv-color-border);
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  /* Control row */
  .ve-control-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .ve-enabled-toggle {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.8125rem;
    color: var(--vcv-color-text);
    cursor: pointer;
  }

  .ve-enabled-toggle input {
    accent-color: var(--vcv-color-primary);
  }

  .ve-remove-btn {
    font-size: 0.72rem;
    color: var(--vcv-color-danger);
    background: none;
    border: 1px solid transparent;
    cursor: pointer;
    padding: 0.2rem 0.5rem;
    border-radius: var(--vcv-radius-sm);
    transition: background 0.1s, border-color 0.1s;
  }

  .ve-remove-btn:hover {
    background: var(--vcv-color-danger-surface);
    border-color: var(--vcv-color-danger-border);
  }

  /* Grid */
  .ve-grid {
    display: grid;
    gap: 0.75rem;
  }

  .ve-grid--2 {
    grid-template-columns: repeat(2, 1fr);
  }

  @media (max-width: 540px) {
    .ve-grid--2 {
      grid-template-columns: 1fr;
    }
  }

  /* Field */
  .ve-field {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
  }

  .ve-field--full {
    grid-column: 1 / -1;
  }

  .ve-label {
    font-size: 0.7rem;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--vcv-color-muted);
  }

  .ve-label-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .ve-input {
    height: 2rem;
    padding: 0 0.625rem;
    font-size: 0.8125rem;
    background: var(--vcv-color-bg);
    border: 1px solid var(--vcv-color-border-strong);
    border-radius: var(--vcv-radius-sm);
    color: var(--vcv-color-text);
    transition: border-color 0.12s;
    width: 100%;
  }

  .ve-input:focus {
    outline: none;
    border-color: var(--vcv-color-primary);
    box-shadow: 0 0 0 3px var(--vcv-color-focus-ring);
  }

  .ve-input--mono {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 0.75rem;
  }

  .ve-reveal-btn {
    font-size: 0.7rem;
    color: var(--vcv-color-muted);
    background: none;
    border: none;
    cursor: pointer;
    padding: 0;
    transition: color 0.1s;
  }

  .ve-reveal-btn:hover {
    color: var(--vcv-color-text);
  }

  .ve-hint {
    font-size: 0.7rem;
    color: var(--vcv-color-muted);
    margin: 0;
  }

  .ve-hint--warn {
    color: var(--vcv-color-warning-strong);
  }

  /* TLS details */
  .ve-tls-details {
    border: 1px solid var(--vcv-color-border);
    border-radius: var(--vcv-radius-sm);
  }

  .ve-tls-summary {
    padding: 0.5rem 0.75rem;
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--vcv-color-muted);
    cursor: pointer;
    list-style: none;
    user-select: none;
  }

  .ve-tls-summary::-webkit-details-marker {
    display: none;
  }

  .ve-tls-summary:hover {
    color: var(--vcv-color-text);
  }

  .ve-tls-body {
    padding: 0.75rem;
    border-top: 1px solid var(--vcv-color-border);
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
</style>
