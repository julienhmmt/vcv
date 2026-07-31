import { daysUntilExpiry } from '$lib/utils/cert-status'
import type { Certificate } from '$lib/types'

/** Minimal translator signature shared with the i18n store (avoids a store dependency). */
export type TranslateFn = (key: string, fallback?: string, params?: Record<string, string | number>) => string

/**
 * Return a non-empty display name for a certificate, suitable for an
 * accessible name (aria-label). Falls back to the first SAN, then the
 * serial number, then the provided `fallback` (a localized "unnamed"
 * string) so the result is never empty.
 */
export function certDisplayName(cert: Certificate, fallback: string): string {
  if (cert.commonName.trim() !== '') return cert.commonName
  if (cert.sans.length > 0 && cert.sans[0].trim() !== '') return cert.sans[0]
  if (cert.serialNumber.trim() !== '') return cert.serialNumber
  return fallback
}

/**
 * Localized expiry label shared by the table and card views:
 * compact "{n}d" while valid, descriptive once due or past.
 */
export function certExpiryLabel(cert: Certificate, t: TranslateFn): string {
  const days = daysUntilExpiry(cert)
  if (days > 0) return t('daysRemainingShort', '{days}d', { days })
  if (days === 0) return t('expiringToday', 'Expires today')
  const ago = Math.abs(days)
  return t(ago === 1 ? 'expiredDaysSingular' : 'expiredDays', 'Expired {days} days ago', { days: ago })
}
