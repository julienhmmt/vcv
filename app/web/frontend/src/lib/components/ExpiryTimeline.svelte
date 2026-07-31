<script lang="ts">
  import { getI18n } from '$lib/stores/i18n.svelte'
  import { buildExpiryTimeline, type ExpiryBucket } from '$lib/utils/expiry-timeline'
  import type { Certificate, ExpirationThresholds } from '$lib/types'

  interface Props {
    certs: Certificate[]
    thresholds: ExpirationThresholds
  }

  const { certs, thresholds }: Props = $props()
  const i18n = getI18n()

  const buckets = $derived(buildExpiryTimeline(certs, thresholds))
  const total = $derived(buckets.reduce((sum, bucket) => sum + bucket.count, 0))

  function bucketLabel(bucket: ExpiryBucket): string {
    if (bucket.key === 'critical') {
      return i18n.t('timelineWithinDays', '≤ {days} days', { days: bucket.to ?? 0 })
    }
    if (bucket.key === 'later') {
      return i18n.t('timelineBeyondDays', '> {days} days', { days: bucket.from - 1 })
    }
    return i18n.t('timelineRangeDays', '{from}–{to} days', { from: bucket.from, to: bucket.to ?? 0 })
  }
</script>

{#if total > 0}
  <section class="vcv-expiry-timeline" aria-label={i18n.t('expiryTimelineLabel', 'Upcoming expirations')}>
    <span class="vcv-expiry-timeline-title">{i18n.t('expiryTimelineLabel', 'Upcoming expirations')}</span>
    <div class="vcv-expiry-timeline-buckets">
      {#each buckets as bucket (bucket.key)}
        <div class="vcv-expiry-bucket vcv-expiry-bucket-{bucket.tone}">
          <span class="vcv-expiry-bucket-count">{bucket.count}</span>
          <span class="vcv-expiry-bucket-label">{bucketLabel(bucket)}</span>
        </div>
      {/each}
    </div>
  </section>
{/if}
