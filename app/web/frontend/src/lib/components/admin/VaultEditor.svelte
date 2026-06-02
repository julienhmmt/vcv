<script lang="ts">
  import { Badge } from '$lib/components/ui/badge'
  import { Button } from '$lib/components/ui/button'
  import { Input } from '$lib/components/ui/input'
  import { Label } from '$lib/components/ui/label'
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

  function statusBadge(): { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' } {
    if (!status) return { label: i18n.t('adminVaultUnknown', 'Unknown'), variant: 'outline' }
    if (!status.enabled) return { label: i18n.t('adminVaultDisabled', 'Disabled'), variant: 'secondary' }
    return status.connected
      ? { label: i18n.t('adminVaultConnected', 'Connected'), variant: 'default' }
      : { label: i18n.t('adminVaultDisconnected', 'Disconnected'), variant: 'destructive' }
  }

  const badge = $derived(statusBadge())
  const enabled = $derived(vault.enabled !== false)
</script>

<div class="rounded-lg border bg-card p-4 space-y-3">
  <div class="flex flex-wrap items-center gap-2">
    <Badge variant={badge.variant}>{badge.label}</Badge>
    <span class="font-mono text-xs text-muted-foreground">{vault.id || '(new)'}</span>
    <div class="ml-auto flex gap-2">
      <label class="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          checked={enabled}
          onchange={(event) => update('enabled', (event.target as HTMLInputElement).checked)}
        />
        {i18n.t('adminVaultEnabled', 'Enabled')}
      </label>
      <Button variant="destructive" size="sm" onclick={onRemove}>{i18n.t('adminVaultRemove', 'Remove')}</Button>
    </div>
  </div>

  <div class="grid gap-3 md:grid-cols-2">
    <div class="space-y-1">
      <Label>{i18n.t('adminVaultID', 'ID')}</Label>
      <Input value={vault.id} oninput={(event) => update('id', (event.target as HTMLInputElement).value)} />
    </div>
    <div class="space-y-1">
      <Label>{i18n.t('adminVaultDisplayName', 'Display Name')}</Label>
      <Input
        value={vault.display_name ?? ''}
        oninput={(event) => update('display_name', (event.target as HTMLInputElement).value)}
      />
    </div>
    <div class="space-y-1 md:col-span-2">
      <Label>{i18n.t('adminVaultAddress', 'Address')}</Label>
      <Input
        value={vault.address}
        placeholder="https://vault.example.com"
        oninput={(event) => update('address', (event.target as HTMLInputElement).value)}
      />
    </div>
    <div class="space-y-1 md:col-span-2">
      <Label>{i18n.t('adminVaultToken', 'Token')}</Label>
      <Input
        type="password"
        value={vault.token}
        placeholder="(unchanged)"
        oninput={(event) => update('token', (event.target as HTMLInputElement).value)}
      />
    </div>
    <div class="space-y-1 md:col-span-2">
      <Label>{i18n.t('adminVaultPKIMounts', 'PKI Mounts (comma-separated)')}</Label>
      <Input value={mountsText} oninput={(event) => updateMounts((event.target as HTMLInputElement).value)} />
    </div>
    <label class="flex items-center gap-2 text-sm md:col-span-2">
      <input
        type="checkbox"
        checked={vault.tls_insecure ?? false}
        onchange={(event) => update('tls_insecure', (event.target as HTMLInputElement).checked)}
      />
      {i18n.t('adminVaultTLSInsecure', 'TLS Insecure (skip verification)')}
    </label>
  </div>
</div>
