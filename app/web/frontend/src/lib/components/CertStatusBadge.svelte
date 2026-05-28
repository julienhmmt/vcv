<script lang="ts">
  import { Badge } from '$lib/components/ui/badge'
  import { statusVariant } from '$lib/utils/cert-status'
  import type { CertStatus } from '$lib/types'

  interface Props {
    status: CertStatus
    days?: number
  }

  const { status, days }: Props = $props()

  const label = $derived.by(() => {
    switch (status) {
      case 'valid':
        return 'Valid'
      case 'warning':
        return days != null ? `Expires in ${days}d` : 'Expiring soon'
      case 'critical':
        return days != null ? `Expires in ${days}d` : 'Critical'
      case 'expired':
        return 'Expired'
      case 'revoked':
        return 'Revoked'
    }
  })
</script>

<Badge variant={statusVariant(status)}>{label}</Badge>
