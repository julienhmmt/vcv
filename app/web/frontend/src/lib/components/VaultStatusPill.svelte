<script lang="ts">
  import * as HoverCard from '$lib/components/ui/hover-card'
  import Activity from '@lucide/svelte/icons/activity'
  import CircleAlert from '@lucide/svelte/icons/circle-alert'
  import CircleCheck from '@lucide/svelte/icons/circle-check'
  import { getI18n } from '$lib/stores/i18n.svelte'
  import type { StatusResponse } from '$lib/types'

  interface Props {
    status: StatusResponse | null
    loading: boolean
    onRefresh: () => void
  }

  const { status, loading, onRefresh }: Props = $props()
  const i18n = getI18n()

  const summary = $derived.by(() => {
    if (!status)
      return { text: i18n.t('statusConnecting', 'connecting…'), cls: 'vcv-status-state-neutral', up: 0, total: 0 }
    const total = status.vaults.length
    const up = status.vaults.filter((v) => v.connected).length
    return {
      text:
        total === 0
          ? i18n.t('statusNoVaults', 'no vaults')
          : total === 1
            ? status.vaults[0].display_name || status.vaults[0].id
            : i18n.t('footerVaultSummary', '{up}/{total} vaults', { up, total }),
      cls: total === 0 ? 'vcv-status-state-neutral' : up === total ? 'vcv-status-state-ok' : 'vcv-status-state-error',
      up,
      total,
    }
  })
</script>

<HoverCard.Root openDelay={150}>
  <HoverCard.Trigger
    class="vcv-vault-status-pill {summary.cls}"
    onclick={onRefresh}
    aria-label={i18n.t('modalVaultStatusTitle', 'Vault status')}
  >
    <Activity class="h-3.5 w-3.5" />
    <span>{summary.text}</span>
    {#if loading}<span class="vcv-pulse" aria-hidden="true">·</span>{/if}
  </HoverCard.Trigger>
  <HoverCard.Content class="w-72 vcv-hover-card">
    <div class="vcv-hover-card-head">
      <span class="vcv-hover-card-title">{i18n.t('mountStatsVaults', 'Vaults')}</span>
      <button type="button" class="vcv-button vcv-button-small vcv-button-ghost" onclick={onRefresh}>
        {i18n.t('buttonRefresh', 'Refresh')}
      </button>
    </div>
    {#if status?.vaults?.length}
      <ul class="vcv-hover-card-list">
        {#each status.vaults as vault (vault.id)}
          <li class="vcv-hover-card-row">
            {#if vault.connected}
              <CircleCheck class="h-4 w-4 text-emerald-500" />
            {:else}
              <CircleAlert class="h-4 w-4 text-red-500" />
            {/if}
            <div class="vcv-hover-card-info">
              <span class="vcv-hover-card-name">{vault.display_name || vault.id}</span>
              {#if !vault.connected && vault.error}
                <span class="vcv-hover-card-error">{vault.error}</span>
              {/if}
            </div>
          </li>
        {/each}
      </ul>
    {:else}
      <p class="vcv-hover-card-empty">{i18n.t('statusNoVaultsConfigured', 'No vaults configured.')}</p>
    {/if}
  </HoverCard.Content>
</HoverCard.Root>
