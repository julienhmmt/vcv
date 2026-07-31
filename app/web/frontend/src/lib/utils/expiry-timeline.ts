import { daysUntilExpiry } from '$lib/utils/cert-status'
import type { Certificate, ExpirationThresholds } from '$lib/types'

/** Upper edge of the "near" bucket; beyond this certs sit in the open-ended "later" bucket. */
export const TIMELINE_FAR_HORIZON_DAYS = 90

export type ExpiryBucketKey = 'critical' | 'warning' | 'near' | 'later'
export type ExpiryBucketTone = 'critical' | 'warning' | 'neutral' | 'muted'

export interface ExpiryBucket {
  key: ExpiryBucketKey
  /** Inclusive lower bound in days from now. */
  from: number
  /** Inclusive upper bound in days from now; null when open-ended. */
  to: number | null
  count: number
  tone: ExpiryBucketTone
}

/**
 * Bucket not-yet-expired, non-revoked certificates by time-to-expiry.
 * The first two buckets follow the configured critical/warning thresholds;
 * certificates already expired or with an unparseable expiry are excluded
 * (they are covered by the status overview, not a future timeline).
 */
export function buildExpiryTimeline(
  certs: Certificate[],
  thresholds: ExpirationThresholds,
  now: Date = new Date(),
): ExpiryBucket[] {
  const buckets: ExpiryBucket[] = [
    { key: 'critical', from: 0, to: thresholds.critical, count: 0, tone: 'critical' },
    { key: 'warning', from: thresholds.critical + 1, to: thresholds.warning, count: 0, tone: 'warning' },
  ]
  // The near bucket only exists when the warning threshold leaves room before the far horizon.
  if (thresholds.warning < TIMELINE_FAR_HORIZON_DAYS) {
    buckets.push({ key: 'near', from: thresholds.warning + 1, to: TIMELINE_FAR_HORIZON_DAYS, count: 0, tone: 'neutral' })
  }
  buckets.push({
    key: 'later',
    from: Math.max(thresholds.warning, TIMELINE_FAR_HORIZON_DAYS) + 1,
    to: null,
    count: 0,
    tone: 'muted',
  })

  for (const cert of certs) {
    if (cert.revoked) continue
    const days = daysUntilExpiry(cert, now)
    if (!Number.isFinite(days) || days < 0) continue
    const bucket = buckets.find((candidate) => days >= candidate.from && (candidate.to === null || days <= candidate.to))
    if (bucket) bucket.count++
  }
  return buckets
}
