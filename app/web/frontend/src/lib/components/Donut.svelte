<script lang="ts">
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
  }

  const { counts, label }: Props = $props()

  const total = $derived(counts.valid + counts.warning + counts.critical + counts.expired + counts.revoked)
  const style = $derived(
    `--chart-valid: ${counts.valid}; --chart-warning: ${counts.warning}; --chart-critical: ${counts.critical}; --chart-expired: ${counts.expired}; --chart-revoked: ${counts.revoked}; --chart-total: ${total};`,
  )
</script>

<div class="vcv-dashboard-donut vcv-dashboard-donut-compact" {style}>
  <div class="vcv-donut-wrapper">
    <div class="vcv-donut{total === 0 ? ' vcv-donut-empty' : ''}"></div>
    <div class="vcv-donut-center">
      <span class="vcv-donut-center-value">{total}</span>
      <span class="vcv-donut-center-label">{label}</span>
    </div>
  </div>
</div>
