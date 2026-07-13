import type { ExpirationThresholds } from '$lib/types'
import { DEFAULT_THRESHOLDS } from '$lib/utils/cert-status'

/**
 * Parse expiration thresholds from a public config payload.
 * Invalid or missing values fall back to DEFAULT_THRESHOLDS.
 */
export function thresholdsFromConfig(
  raw: { critical?: number; warning?: number } | undefined,
  fallback: ExpirationThresholds = DEFAULT_THRESHOLDS,
): ExpirationThresholds {
  if (
    raw != null &&
    typeof raw.critical === 'number' &&
    typeof raw.warning === 'number' &&
    Number.isFinite(raw.critical) &&
    Number.isFinite(raw.warning) &&
    raw.critical > 0 &&
    raw.warning > 0
  ) {
    return { critical: raw.critical, warning: raw.warning }
  }
  return fallback
}
