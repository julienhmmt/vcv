<script lang="ts">
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
    label: string
    /** Per-status labels for segment tooltips/aria; enables interactivity when paired with onSelect. */
    segmentLabels?: Record<StatusKey, string>
    onSelect?: (key: StatusKey) => void
  }

  const { counts, label, segmentLabels, onSelect }: Props = $props()

  // Ascending severity, matching the legacy conic-gradient order.
  const ORDER: StatusKey[] = ['valid', 'warning', 'critical', 'expired', 'revoked']

  const total = $derived(counts.valid + counts.warning + counts.critical + counts.expired + counts.revoked)

  // SVG ring geometry: r chosen so circumference ≈ 100, letting dash lengths be raw percentages.
  const RADIUS = 15.915
  const interactive = $derived(!!onSelect)

  const segments = $derived.by(() => {
    if (total === 0) return []
    let acc = 0
    return ORDER.filter((key) => counts[key] > 0).map((key) => {
      const value = counts[key]
      const pct = (value / total) * 100
      const seg = { key, value, pct, offset: 25 - acc }
      acc += pct
      return seg
    })
  })

  function tooltip(key: StatusKey, value: number): string {
    const name = segmentLabels?.[key] ?? key
    return `${name}: ${value}`
  }
</script>

<div class="vcv-dashboard-donut vcv-dashboard-donut-compact">
  <div class="vcv-donut-wrapper">
    {#if total === 0}
      <div class="vcv-donut vcv-donut-empty"></div>
    {:else}
      <svg class="vcv-donut-svg" viewBox="0 0 36 36" role="img" aria-label={label}>
        {#each segments as seg (seg.key)}
          {#if interactive}
            <circle
              class="vcv-donut-segment vcv-donut-segment-{seg.key}"
              cx="18"
              cy="18"
              r={RADIUS}
              fill="none"
              stroke-width="4"
              stroke-dasharray="{seg.pct} {100 - seg.pct}"
              stroke-dashoffset={seg.offset}
              role="button"
              tabindex="0"
              aria-label={tooltip(seg.key, seg.value)}
              onclick={() => onSelect?.(seg.key)}
              onkeydown={(event) => {
                if (event.key === 'Enter' || event.key === ' ') {
                  event.preventDefault()
                  onSelect?.(seg.key)
                }
              }}
            >
              <title>{tooltip(seg.key, seg.value)}</title>
            </circle>
          {:else}
            <circle
              class="vcv-donut-segment-static vcv-donut-segment-{seg.key}"
              cx="18"
              cy="18"
              r={RADIUS}
              fill="none"
              stroke-width="4"
              stroke-dasharray="{seg.pct} {100 - seg.pct}"
              stroke-dashoffset={seg.offset}
            >
              <title>{tooltip(seg.key, seg.value)}</title>
            </circle>
          {/if}
        {/each}
      </svg>
    {/if}
    <div class="vcv-donut-center">
      <span class="vcv-donut-center-value">{total}</span>
      <span class="vcv-donut-center-label">{label}</span>
    </div>
  </div>
</div>
