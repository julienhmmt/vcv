<script lang="ts">
  import { statusBadgeClass } from '$lib/utils/cert-status'
  import type { CertStatus } from '$lib/types'

  interface Props {
    status: CertStatus
    label?: string
  }

  const { status, label }: Props = $props()
</script>

<span class={statusBadgeClass(status)}>{label ?? status}</span>
