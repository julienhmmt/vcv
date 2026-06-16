import type { Component } from 'svelte'
import CheckCircle from '@lucide/svelte/icons/check-circle'
import AlertTriangle from '@lucide/svelte/icons/alert-triangle'
import AlertOctagon from '@lucide/svelte/icons/alert-octagon'
import XCircle from '@lucide/svelte/icons/x-circle'
import Ban from '@lucide/svelte/icons/ban'
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

export function statusBadgeClass(status: CertStatus): string {
  switch (status) {
    case 'valid':
      return 'vcv-badge vcv-badge-valid'
    case 'warning':
      return 'vcv-badge vcv-badge-warning'
    case 'critical':
      return 'vcv-badge vcv-badge-critical'
    case 'expired':
      return 'vcv-badge vcv-badge-expired'
    case 'revoked':
      return 'vcv-badge vcv-badge-revoked'
  }
}

export function statusIcon(status: CertStatus): Component {
  switch (status) {
    case 'valid':
      return CheckCircle
    case 'warning':
      return AlertTriangle
    case 'critical':
      return AlertOctagon
    case 'expired':
      return XCircle
    case 'revoked':
      return Ban
  }
}

export function rowClassForStatus(status: CertStatus): string {
  return `vcv-row-${status}`
}

export interface CertParts {
  vault: string
  mount: string
  mountKey: string
}

export function parseCertID(id: string): CertParts {
  const parts = id.split('|')
  if (parts.length === 2) {
    const mountName = parts[1].split(':')[0]?.trim() ?? ''
    return { vault: parts[0], mount: mountName, mountKey: `${parts[0]}|${mountName}` }
  }
  const mountName = id.split(':')[0]?.trim() ?? ''
  return { vault: '', mount: mountName, mountKey: mountName }
}
