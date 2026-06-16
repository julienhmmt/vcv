<script lang="ts">
  import Donut from '$lib/components/Donut.svelte'
  import { statusIcon } from '$lib/utils/cert-status'
  import type { CertStatus } from '$lib/types'

  type StatusKey = 'valid' | 'warning' | 'critical' | 'expired' | 'revoked'

  interface Counts {
    valid: number
    warning: number
    critical: number
    expired: number
    revoked: number
  }

  interface Props {
    counts: Counts
    meta: Record<StatusKey, { label: string; desc: string }>
    statusFilters: StatusKey[]
    donutLabel: string
    regionLabel: string
    onSelect: (key: StatusKey) => void
  }

  const { counts, meta, statusFilters, donutLabel, regionLabel, onSelect }: Props = $props()

  // Ascending severity, matching the donut gradient order. Revoked is an
  // orthogonal state and sits last.
  const ORDER: StatusKey[] = ['valid', 'warning', 'critical', 'expired', 'revoked']
</script>

<section class="vcv-overview" aria-label={regionLabel}>
  <div class="vcv-overview-chart">
    <Donut {counts} label={donutLabel} />
  </div>
  <div class="vcv-overview-stats" role="group" aria-label={regionLabel}>
    {#each ORDER as key, i (key)}
      {@const Icon = statusIcon(key as CertStatus)}
      {#if i > 0}
        <span class="vcv-stat-divider" aria-hidden="true"></span>
      {/if}
      <button
        type="button"
        class="vcv-stat vcv-stat-{key}"
        class:vcv-stat-active={statusFilters.includes(key)}
        aria-pressed={statusFilters.includes(key)}
        onclick={() => onSelect(key)}
      >
        <span class="vcv-stat-label">
          <Icon class="vcv-stat-icon h-3 w-3" aria-hidden="true" />
          {meta[key].label}
        </span>
        <span class="vcv-stat-count">{counts[key]}</span>
      </button>
    {/each}
  </div>
</section>
