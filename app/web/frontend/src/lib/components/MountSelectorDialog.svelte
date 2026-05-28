<script lang="ts">
  import * as Command from '$lib/components/ui/command'
  import * as Dialog from '$lib/components/ui/dialog'
  import { Check, Square, SquareCheck } from '@lucide/svelte'

  interface Props {
    open: boolean
    onOpenChange: (open: boolean) => void
    allMounts: string[]
    selected: string[] | null
    onChange: (selected: string[] | null) => void
  }

  const { open, onOpenChange, allMounts, selected, onChange }: Props = $props()

  const selectedSet = $derived(
    selected === null ? new Set(allMounts) : new Set(selected),
  )

  function toggle(key: string): void {
    const base = selected === null ? [...allMounts] : [...selected]
    const idx = base.indexOf(key)
    if (idx >= 0) {
      base.splice(idx, 1)
    } else {
      base.push(key)
    }
    onChange(base.length === allMounts.length ? null : base)
  }

  function selectAll(): void {
    onChange(null)
  }

  function deselectAll(): void {
    onChange([])
  }
</script>

<Dialog.Root {open} {onOpenChange}>
  <Dialog.Content class="max-w-xl">
    <Dialog.Header>
      <Dialog.Title>Sources</Dialog.Title>
      <Dialog.Description>
        Filter the table by Vault and PKI mount.
      </Dialog.Description>
    </Dialog.Header>

    <Command.Root class="rounded-md border">
      <Command.Input placeholder="Search mounts…" />
      <Command.List class="max-h-72">
        <Command.Empty>No mount matches.</Command.Empty>
        <Command.Group>
          {#each allMounts as key}
            {@const isSelected = selectedSet.has(key)}
            <Command.Item value={key} onSelect={() => toggle(key)}>
              {#if isSelected}
                <SquareCheck class="mr-2 h-4 w-4 text-primary" />
              {:else}
                <Square class="mr-2 h-4 w-4 text-muted-foreground" />
              {/if}
              <span class="font-mono text-xs">{key}</span>
              {#if isSelected}
                <Check class="ml-auto h-4 w-4 text-primary" />
              {/if}
            </Command.Item>
          {/each}
        </Command.Group>
      </Command.List>
    </Command.Root>

    <Dialog.Footer class="flex items-center justify-between gap-2">
      <div class="flex items-center gap-2 text-xs text-muted-foreground">
        <span>{selectedSet.size} / {allMounts.length} selected</span>
      </div>
      <div class="flex gap-2">
        <button type="button" class="vcv-button vcv-button-small vcv-button-ghost" onclick={deselectAll}>
          Deselect all
        </button>
        <button type="button" class="vcv-button vcv-button-small vcv-button-ghost" onclick={selectAll}>
          Select all
        </button>
        <button type="button" class="vcv-button vcv-button-primary vcv-button-small" onclick={() => onOpenChange(false)}>
          Done
        </button>
      </div>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
