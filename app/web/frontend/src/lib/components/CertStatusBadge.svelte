<script lang="ts">
  import { statusBadgeClass, statusIcon } from '$lib/utils/cert-status'
  import type { CertStatus } from '$lib/types'

  interface Props {
    status: CertStatus
    label?: string
  }

  const { status, label }: Props = $props()
  const Icon = $derived(statusIcon(status))
</script>

<span class={statusBadgeClass(status)}>
  <Icon class="h-3 w-3" aria-hidden="true" />
  {label ?? status}
</span>
