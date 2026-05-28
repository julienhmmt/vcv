<script lang="ts">
  import { untrack } from 'svelte'
  import { Button } from '$lib/components/ui/button'
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card'
  import { Input } from '$lib/components/ui/input'
  import { Label } from '$lib/components/ui/label'
  import VaultEditor from './VaultEditor.svelte'
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

  let working = $state<SettingsFile>(untrack(() => structuredClone(settings)))
  let lastSyncedRef: SettingsFile | null = null

  $effect(() => {
    if (settings !== lastSyncedRef) {
      lastSyncedRef = settings
      working = structuredClone(settings)
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
    if (target.id && statusById.has(target.id)) {
      onRemoveVault(target.id)
      return
    }
    const vaults = working.vaults.filter((_, i) => i !== index)
    working = { ...working, vaults }
  }

  function corsText(): string {
    return (working.cors.allowed_origins ?? []).join(', ')
  }

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

  function submit(event: SubmitEvent): void {
    event.preventDefault()
    onSave(working)
  }
</script>

<div class="space-y-6">
  <header class="flex items-center justify-between">
    <h1 class="text-2xl font-semibold tracking-tight">VCV Admin</h1>
    <div class="flex gap-2">
      <Button variant="outline" onclick={onInvalidateCache}>Invalidate cache</Button>
      <Button variant="ghost" onclick={onLogout}>Sign out</Button>
    </div>
  </header>

  {#if error}
    <p class="rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">
      {error}
    </p>
  {/if}
  {#if successMessage}
    <p class="rounded-md border border-emerald-500/40 bg-emerald-500/10 p-3 text-sm text-emerald-600 dark:text-emerald-400">
      {successMessage}
    </p>
  {/if}

  <form class="space-y-6" onsubmit={submit}>
    <Card>
      <CardHeader>
        <CardTitle>Expiration thresholds (days)</CardTitle>
      </CardHeader>
      <CardContent class="grid gap-3 md:grid-cols-2">
        <div class="space-y-1">
          <Label>Critical</Label>
          <Input
            type="number"
            min="1"
            max="3650"
            value={working.certificates.expiration_thresholds.critical}
            oninput={(event) =>
              (working = {
                ...working,
                certificates: {
                  ...working.certificates,
                  expiration_thresholds: {
                    ...working.certificates.expiration_thresholds,
                    critical: Number((event.target as HTMLInputElement).value),
                  },
                },
              })}
          />
        </div>
        <div class="space-y-1">
          <Label>Warning</Label>
          <Input
            type="number"
            min="1"
            max="3650"
            value={working.certificates.expiration_thresholds.warning}
            oninput={(event) =>
              (working = {
                ...working,
                certificates: {
                  ...working.certificates,
                  expiration_thresholds: {
                    ...working.certificates.expiration_thresholds,
                    warning: Number((event.target as HTMLInputElement).value),
                  },
                },
              })}
          />
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>Metrics</CardTitle>
      </CardHeader>
      <CardContent class="space-y-2 text-sm">
        <label class="flex items-center gap-2">
          <input
            type="checkbox"
            checked={working.metrics.per_certificate ?? false}
            onchange={(event) =>
              (working = {
                ...working,
                metrics: { ...working.metrics, per_certificate: (event.target as HTMLInputElement).checked },
              })}
          />
          Per-certificate metrics (high cardinality)
        </label>
        <label class="flex items-center gap-2">
          <input
            type="checkbox"
            checked={working.metrics.enhanced_metrics ?? true}
            onchange={(event) =>
              (working = {
                ...working,
                metrics: { ...working.metrics, enhanced_metrics: (event.target as HTMLInputElement).checked },
              })}
          />
          Enhanced metrics
        </label>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>CORS allowed origins</CardTitle>
      </CardHeader>
      <CardContent>
        <Input
          value={corsText()}
          placeholder="https://app.example.com, https://other.example.com"
          oninput={(event) => updateCors((event.target as HTMLInputElement).value)}
        />
      </CardContent>
    </Card>

    <Card>
      <CardHeader class="flex flex-row items-center justify-between">
        <CardTitle>Vaults</CardTitle>
        <Button type="button" variant="outline" size="sm" onclick={onAddVault}>Add vault</Button>
      </CardHeader>
      <CardContent class="space-y-4">
        {#each working.vaults as vault, index (vault.id || index)}
          <VaultEditor
            {vault}
            status={statusById.get(vault.id)}
            onChange={(next) => updateVault(index, next)}
            onRemove={() => removeVault(index)}
          />
        {/each}
        {#if working.vaults.length === 0}
          <p class="text-sm text-muted-foreground">No vaults configured.</p>
        {/if}
      </CardContent>
    </Card>

    <div class="flex justify-end">
      <Button type="submit" disabled={loading}>
        {loading ? 'Saving…' : 'Save settings'}
      </Button>
    </div>
  </form>
</div>
