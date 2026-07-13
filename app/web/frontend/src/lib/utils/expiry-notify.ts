export type ExpiryTier = 'none' | 'warning' | 'critical'

const TIER_RANK: Record<ExpiryTier, number> = {
  none: 0,
  warning: 1,
  critical: 2,
}

/** Derive the toast tier from dashboard expiry counts (critical beats warning). */
export function expiryTier(counts: { critical: number; warning: number }): ExpiryTier {
  if (counts.critical > 0) return 'critical'
  if (counts.warning > 0) return 'warning'
  return 'none'
}

/**
 * Whether to show an expiry toast: always on initial load when tier ≠ none,
 * and on later loads only when the tier increases (warning → critical).
 */
export function shouldNotifyExpiry(args: {
  isInitial: boolean
  tier: ExpiryTier
  lastNotifiedTier: ExpiryTier
}): boolean {
  if (args.tier === 'none') return false
  if (args.isInitial) return true
  return TIER_RANK[args.tier] > TIER_RANK[args.lastNotifiedTier]
}
