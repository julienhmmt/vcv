<script lang="ts">
  import Donut from '$lib/components/Donut.svelte'

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
  <div class="vcv-overview-segments" role="group" aria-label={regionLabel}>
    {#each ORDER as key (key)}
      <button
        type="button"
        class="vcv-seg vcv-seg-{key}"
        class:vcv-seg-active={statusFilters.includes(key)}
        aria-pressed={statusFilters.includes(key)}
        onclick={() => onSelect(key)}
      >
        <span class="vcv-seg-top">
          <span class="vcv-seg-dot" aria-hidden="true"></span>
          <span class="vcv-seg-label">{meta[key].label}</span>
        </span>
        <span class="vcv-seg-value">{counts[key]}</span>
        <span class="vcv-seg-desc">{meta[key].desc}</span>
      </button>
    {/each}
  </div>
</section>
