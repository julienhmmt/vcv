<script lang="ts">
  import Server from '@lucide/svelte/icons/server'
  import ChevronDown from '@lucide/svelte/icons/chevron-down'
  import { getI18n } from '$lib/stores/i18n.svelte'
  import ToggleSwitch from './ToggleSwitch.svelte'
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
  // The real token is never sent to the browser (the backend masks it), so the
  // field starts empty and only forwards a value the admin actually types.
  let tokenInput = $state('')

  // After a save the server returns a masked (empty) token; reset the input
  // so the password field doesn't show stale typed text.
  $effect(() => {
    if (!vault.token) tokenInput = ''
  })

  function toggleExpanded(): void {
    expanded = !expanded
  }

  function onToggleClick(event: MouseEvent): void {
    event.stopPropagation()
    toggleExpanded()
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

  // Reuses the same badge classes CertStatusBadge applies to cert rows, so a
  // vault's connection state reads with the exact same color language.
  const statusBadgeClasses: Record<StatusKind, string> = {
    connected: 'vcv-badge-valid',
    disconnected: 'vcv-badge-critical',
    disabled: 'vcv-badge-revoked',
    unknown: 'vcv-badge-revoked',
  }

  const kind = $derived(statusKind())
  const vaultLabel = $derived(vault.display_name || vault.id || i18n.t('adminVaultNew', 'new vault'))
</script>

<div class="ve-card" class:ve-card--collapsed={!expanded}>
  <!-- Summary row (always visible) -->
  <div class="ve-summary" role="button" tabindex="0" aria-expanded={expanded} aria-controls="ve-body-{uid}" onclick={toggleExpanded} onkeydown={onSummaryKeydown}>
    <Server class="h-4 w-4 ve-summary-icon" aria-hidden="true" />
    <div class="ve-summary-info">
      <span class="ve-summary-id">{vaultLabel}</span>
      {#if vault.address}
        <span class="ve-summary-addr">{vault.address}</span>
      {/if}
    </div>
    <span class="vcv-badge {statusBadgeClasses[kind]} ve-status-badge">
      <span class="ve-status-dot ve-status-dot--{kind}" aria-hidden="true"></span>
      {statusLabels[kind]}
    </span>
    <button
      type="button"
      class="ve-toggle-btn"
      tabindex="-1"
      onclick={onToggleClick}
      aria-label={expanded ? i18n.t('adminVaultCollapse', 'Collapse') : i18n.t('adminVaultExpand', 'Expand')}
    >
      <ChevronDown class="ve-toggle-icon {expanded ? 've-toggle-icon--open' : ''}" aria-hidden="true" />
    </button>
  </div>

  <!-- Expanded content -->
  {#if expanded}
    <div class="ve-body" id="ve-body-{uid}">
      <!-- Row: enabled toggle + remove -->
      <div class="ve-control-row">
        <label class="ve-enabled-toggle">
          <ToggleSwitch
            name="ve-enabled-{uid}"
            checked={enabled}
            onCheckedChange={(checked) => update('enabled', checked)}
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
        </div>
        <input
          id="ve-token-{uid}"
          class="ve-input ve-input--mono"
          type="password"
          autocomplete="new-password"
          value={tokenInput}
          placeholder={i18n.t('adminVaultTokenPlaceholder', 'Enter new token to replace; leave blank to keep current')}
          oninput={(event) => {
            tokenInput = (event.target as HTMLInputElement).value
            update('token', tokenInput)
          }}
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
        <p class="ve-hint">{i18n.t('adminVaultPKIMountsHint', 'Comma-separated. First mount is the default.')}</p>
      </div>

      <!-- TLS section -->
      <details class="ve-tls-details">
        <summary class="ve-tls-summary">
          {i18n.t('adminVaultTLSOptions', 'TLS options')}
          <ChevronDown class="ve-tls-chevron" aria-hidden="true" />
        </summary>
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
            <ToggleSwitch
              name="ve-tls-insecure-{uid}"
              checked={vault.tls_insecure ?? false}
              onCheckedChange={(checked) => update('tls_insecure', checked)}
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

  /* Server icon */
  .ve-summary :global(.ve-summary-icon) {
    color: var(--vcv-color-muted);
    flex-shrink: 0;
  }

  /* Status dot (nested inside the badge) */
  .ve-status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .ve-status-dot--connected { background: var(--vcv-status-valid-text); }
  .ve-status-dot--disconnected { background: var(--vcv-status-critical-text); }
  .ve-status-dot--disabled { background: var(--vcv-status-revoked-text); }
  .ve-status-dot--unknown { background: var(--vcv-color-muted); }

  /* Status badge: color comes from the shared .vcv-badge-{tone} classes */
  .ve-status-badge {
    flex-shrink: 0;
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

  .ve-toggle-btn :global(.ve-toggle-icon) {
    width: 16px;
    height: 16px;
    transition: transform 0.18s ease-out;
  }

  .ve-toggle-btn :global(.ve-toggle-icon--open) {
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

  .ve-enabled-toggle :global(.tgl-switch) {
    flex-shrink: 0;
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
    display: flex;
    align-items: center;
    justify-content: space-between;
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

  .ve-tls-summary :global(.ve-tls-chevron) {
    width: 14px;
    height: 14px;
    flex-shrink: 0;
    transition: transform 0.18s ease-out;
  }

  .ve-tls-details[open] .ve-tls-summary :global(.ve-tls-chevron) {
    transform: rotate(180deg);
  }

  .ve-tls-body {
    padding: 0.75rem;
    border-top: 1px solid var(--vcv-color-border);
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
</style>
