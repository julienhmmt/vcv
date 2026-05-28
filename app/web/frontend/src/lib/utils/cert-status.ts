import type { Certificate, CertStatus, ExpirationThresholds } from '$lib/types'

export const DEFAULT_THRESHOLDS: ExpirationThresholds = {
  critical: 7,
  warning: 30,
}

export function daysUntilExpiry(cert: Certificate, now: Date = new Date()): number {
  const expires = new Date(cert.expiresAt).getTime()
  return Math.floor((expires - now.getTime()) / (1000 * 60 * 60 * 24))
}

export function certStatus(
  cert: Certificate,
  thresholds: ExpirationThresholds = DEFAULT_THRESHOLDS,
  now: Date = new Date(),
): CertStatus {
  if (cert.revoked) return 'revoked'
  const days = daysUntilExpiry(cert, now)
  if (days < 0) return 'expired'
  if (days <= thresholds.critical) return 'critical'
  if (days <= thresholds.warning) return 'warning'
  return 'valid'
}

export function statusVariant(status: CertStatus): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (status) {
    case 'valid':
      return 'default'
    case 'warning':
      return 'secondary'
    case 'critical':
    case 'expired':
    case 'revoked':
      return 'destructive'
    default:
      return 'outline'
  }
}
