import type { Component } from 'svelte'
import CheckCircle from '@lucide/svelte/icons/check-circle'
import AlertTriangle from '@lucide/svelte/icons/alert-triangle'
import AlertOctagon from '@lucide/svelte/icons/alert-octagon'
import XCircle from '@lucide/svelte/icons/x-circle'
import Ban from '@lucide/svelte/icons/ban'
import type { CertStatus } from '$lib/types'

/** Lucide icon component for a status. Kept apart from cert-status.ts so the
    pure status logic stays free of Svelte/component imports (and unit-testable
    in a plain Node environment). */
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
