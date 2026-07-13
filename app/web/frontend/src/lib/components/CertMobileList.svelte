<script lang="ts">
  import { Skeleton } from '$lib/components/ui/skeleton'
  import CertCard from '$lib/components/CertCard.svelte'
  import { getI18n } from '$lib/stores/i18n.svelte'
  import { certStatus } from '$lib/utils/cert-status'
  import type { Certificate, CertStatus, ExpirationThresholds } from '$lib/types'

  interface Props {
    certs: Certificate[]
    loading: boolean
    initialLoad: boolean
    hasInventory: boolean
    hasActiveFilters: boolean
    showVaultMount: boolean
    statusMeta: Record<CertStatus, { label: string; desc: string }>
    thresholds: ExpirationThresholds
    onSelect: (cert: Certificate) => void
    onClearFilters: () => void
  }

  const {
    certs,
    loading,
    initialLoad,
    hasInventory,
    hasActiveFilters,
    showVaultMount,
    statusMeta,
    thresholds,
    onSelect,
    onClearFilters,
  }: Props = $props()

  const i18n = getI18n()
</script>

<div class="vcv-certs-mobile-cards">
  {#if initialLoad && !hasInventory}
    {#each Array(6) as _, i (i)}
      <div class="vcv-cert-card">
        <Skeleton class="h-5 w-3/4" />
        <Skeleton class="h-4 w-1/2" />
      </div>
    {/each}
  {:else if certs.length === 0}
    <div class="vcv-certs-mobile-empty">
      {#if loading}
        {i18n.t('labelLoading', 'Loading…')}
      {:else if hasActiveFilters}
        <p class="vcv-empty-title">{i18n.t('tableNoMatch', 'No certificates match the current filters.')}</p>
        <button type="button" class="vcv-button vcv-button-small vcv-button-ghost vcv-button-pill" onclick={onClearFilters}>
          {i18n.t('filterChipReset', 'Clear all')}
        </button>
      {:else}
        <p class="vcv-empty-title">{i18n.t('tableEmpty', 'No certificates found.')}</p>
        <p class="vcv-empty-hint">{i18n.t('tableEmptyHint', 'No PKI mount returned any certificates yet.')}</p>
      {/if}
    </div>
  {:else}
    {#each certs as cert (cert.id)}
      {@const s = certStatus(cert, thresholds)}
      <CertCard {cert} {showVaultMount} statusLabel={statusMeta[s].label} {thresholds} onSelect={onSelect} />
    {/each}
  {/if}
</div>
