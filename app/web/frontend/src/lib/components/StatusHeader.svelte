<script lang="ts">
  import { Badge } from '$lib/components/ui/badge'
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card'
  import type { StatusResponse } from '$lib/types'

  interface Props {
    status: StatusResponse | null
    loading: boolean
    error: string | null
  }

  const { status, loading, error }: Props = $props()
</script>

<Card>
  <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
    <CardTitle class="text-base">Vault Status</CardTitle>
    {#if status}
      <span class="text-xs text-muted-foreground">v{status.version}</span>
    {/if}
  </CardHeader>
  <CardContent>
    {#if loading && !status}
      <p class="text-sm text-muted-foreground">Loading…</p>
    {:else if error}
      <p class="text-sm text-destructive">{error}</p>
    {:else if status}
      <div class="flex flex-wrap gap-2">
        {#each status.vaults as vault (vault.id)}
          <Badge variant={vault.connected ? 'default' : 'destructive'}>
            {vault.display_name}{vault.connected ? '' : ` — ${vault.error ?? 'down'}`}
          </Badge>
        {/each}
        {#if status.vaults.length === 0}
          <p class="text-sm text-muted-foreground">No vaults configured.</p>
        {/if}
      </div>
    {/if}
  </CardContent>
</Card>
