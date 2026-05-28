<script lang="ts">
  import { Badge } from '$lib/components/ui/badge'
  import { Button } from '$lib/components/ui/button'
  import type { Certificate } from '$lib/types'

  interface Props {
    certificates: Certificate[]
    selected: string[] | null
    onChange: (mounts: string[] | null) => void
  }

  const { certificates, selected, onChange }: Props = $props()

  const mounts = $derived.by(() => {
    const set = new Set<string>()
    for (const cert of certificates) {
      const parts = cert.id.split('|')
      const mountSerial = parts.length === 2 ? parts[1] : cert.id
      const mountName = mountSerial.split(':')[0]?.trim()
      const key = parts.length === 2 ? `${parts[0]}|${mountName}` : mountName
      if (key) set.add(key)
    }
    return Array.from(set).sort()
  })

  function toggle(key: string): void {
    if (selected === null) {
      onChange(mounts.filter((m) => m !== key))
      return
    }
    if (selected.includes(key)) {
      onChange(selected.filter((m) => m !== key))
      return
    }
    onChange([...selected, key])
  }

  function selectAll(): void {
    onChange(null)
  }

  function clear(): void {
    onChange([])
  }
</script>

{#if mounts.length > 1}
  <div class="flex flex-wrap items-center gap-2">
    <span class="text-xs uppercase tracking-wide text-muted-foreground">Mounts</span>
    {#each mounts as key}
      {@const active = selected === null || selected.includes(key)}
      <button type="button" onclick={() => toggle(key)} class="appearance-none">
        <Badge variant={active ? 'default' : 'outline'}>{key}</Badge>
      </button>
    {/each}
    <Button size="sm" variant="ghost" onclick={selectAll}>All</Button>
    <Button size="sm" variant="ghost" onclick={clear}>None</Button>
  </div>
{/if}
