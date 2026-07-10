<script lang="ts">
  import { statusIcon } from '$lib/utils/cert-icons'
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
    regionLabel: string
    onSelect: (key: StatusKey) => void
  }

  const { counts, meta, statusFilters, regionLabel, onSelect }: Props = $props()

  // Ascending severity. Revoked is an orthogonal state and sits last.
  const ORDER: StatusKey[] = ['valid', 'warning', 'critical', 'expired', 'revoked']
  // Only the time-bounded statuses carry a threshold worth surfacing inline.
  const WITH_DESC: StatusKey[] = ['warning', 'critical']
</script>

<section class="vcv-overview" aria-label={regionLabel}>
  <div class="vcv-overview-stats" role="group" aria-label={regionLabel}>
    {#each ORDER as key (key)}
      {@const Icon = statusIcon(key as CertStatus)}
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
        {#if WITH_DESC.includes(key)}
          <span class="vcv-stat-desc">{meta[key].desc}</span>
        {/if}
      </button>
    {/each}
  </div>
</section>
