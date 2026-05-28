<script lang="ts">
  import { Skeleton } from '$lib/components/ui/skeleton'

  interface Props {
    rows?: number
  }

  const { rows = 8 }: Props = $props()
</script>

<div class="space-y-2 px-4 py-3">
  {#each Array(rows) as _, i (i)}
    <div class="flex items-center gap-4">
      <Skeleton class="h-5 flex-1" />
      <Skeleton class="h-5 w-32" />
      <Skeleton class="h-5 w-24" />
    </div>
  {/each}
</div>
