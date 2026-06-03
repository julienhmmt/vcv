<script lang="ts">
  import * as Dialog from '$lib/components/ui/dialog'
  import Check from '@lucide/svelte/icons/check'
  import Minus from '@lucide/svelte/icons/minus'
  import Search from '@lucide/svelte/icons/search'
  import Server from '@lucide/svelte/icons/server'
  import Layers from '@lucide/svelte/icons/layers'
  import X from '@lucide/svelte/icons/x'
  import { getI18n } from '$lib/stores/i18n.svelte'

  interface Props {
    open: boolean
    onOpenChange: (open: boolean) => void
    allMounts: string[]
    selected: string[] | null
    onChange: (selected: string[] | null) => void
  }

  const { open, onOpenChange, allMounts, selected, onChange }: Props = $props()
  const i18n = getI18n()

  let query = $state('')

  const selectedSet = $derived(
    selected === null ? new Set(allMounts) : new Set(selected),
  )

  interface VaultGroup {
    vault: string
    mounts: { key: string; mount: string }[]
  }

  // Group flat "vault|mount" keys under their vault, preserving sort order.
  const groups = $derived.by<VaultGroup[]>(() => {
    const byVault = new Map<string, VaultGroup>()
    for (const key of allMounts) {
      const sep = key.indexOf('|')
      const vault = sep >= 0 ? key.slice(0, sep) : key
      const mount = sep >= 0 ? key.slice(sep + 1) : key
      let group = byVault.get(vault)
      if (!group) {
        group = { vault, mounts: [] }
        byVault.set(vault, group)
      }
      group.mounts.push({ key, mount })
    }
    return Array.from(byVault.values())
  })

  const filteredGroups = $derived.by<VaultGroup[]>(() => {
    const q = query.trim().toLowerCase()
    if (!q) return groups
    const out: VaultGroup[] = []
    for (const group of groups) {
      const vaultHit = group.vault.toLowerCase().includes(q)
      const mounts = vaultHit
        ? group.mounts
        : group.mounts.filter((m) => m.key.toLowerCase().includes(q))
      if (mounts.length > 0) out.push({ vault: group.vault, mounts })
    }
    return out
  })

  const hasMatches = $derived(filteredGroups.length > 0)

  function vaultState(group: VaultGroup): 'all' | 'some' | 'none' {
    let count = 0
    for (const m of group.mounts) if (selectedSet.has(m.key)) count++
    if (count === 0) return 'none'
    if (count === group.mounts.length) return 'all'
    return 'some'
  }

  function commit(next: Set<string>): void {
    onChange(next.size === allMounts.length ? null : Array.from(next))
  }

  function toggle(key: string): void {
    const next = new Set(selectedSet)
    if (next.has(key)) next.delete(key)
    else next.add(key)
    commit(next)
  }

  function toggleVault(group: VaultGroup): void {
    const next = new Set(selectedSet)
    if (vaultState(group) === 'all') {
      for (const m of group.mounts) next.delete(m.key)
    } else {
      for (const m of group.mounts) next.add(m.key)
    }
    commit(next)
  }

  function selectAll(): void {
    onChange(null)
  }

  function deselectAll(): void {
    onChange([])
  }
</script>

<Dialog.Root {open} {onOpenChange}>
  <Dialog.Content class="vcv-msd" showCloseButton={false}>
    <header class="vcv-msd-head">
      <div class="vcv-msd-head-text">
        <Dialog.Title class="vcv-msd-title">
          {i18n.t('mountSelectorTitle', 'Sources')}
        </Dialog.Title>
        <Dialog.Description class="vcv-msd-desc">
          {i18n.t('mountSelectorTooltip', 'Filter the table by Vault and PKI mount.')}
        </Dialog.Description>
      </div>
      <button
        type="button"
        class="vcv-msd-close"
        aria-label={i18n.t('buttonClose', 'Close')}
        onclick={() => onOpenChange(false)}
      >
        <X class="h-4 w-4" />
      </button>
    </header>

    <div class="vcv-msd-search">
      <Search class="vcv-msd-search-icon h-4 w-4" />
      <input
        type="text"
        class="vcv-msd-search-input"
        placeholder={i18n.t('mountSearchPlaceholder', 'Search mounts…')}
        bind:value={query}
        autocomplete="off"
        spellcheck="false"
      />
      {#if query}
        <button
          type="button"
          class="vcv-msd-search-clear"
          aria-label={i18n.t('buttonClear', 'Clear')}
          onclick={() => (query = '')}
        >
          <X class="h-3.5 w-3.5" />
        </button>
      {/if}
    </div>

    <div class="vcv-msd-list">
      {#if !hasMatches}
        <div class="vcv-msd-empty">
          <Search class="h-5 w-5" />
          <span>{i18n.t('mountNoMatch', 'No mount matches.')}</span>
        </div>
      {:else}
        {#each filteredGroups as group (group.vault)}
          {@const state = vaultState(group)}
          {@const sel = group.mounts.filter((m) => selectedSet.has(m.key)).length}
          <section class="vcv-msd-group">
            <button
              type="button"
              class="vcv-msd-vault"
              class:is-active={state !== 'none'}
              onclick={() => toggleVault(group)}
            >
              <span class="vcv-msd-check" data-state={state}>
                {#if state === 'all'}
                  <Check class="h-3.5 w-3.5" />
                {:else if state === 'some'}
                  <Minus class="h-3.5 w-3.5" />
                {/if}
              </span>
              <Server class="vcv-msd-vault-icon h-4 w-4" />
              <span class="vcv-msd-vault-name">{group.vault}</span>
              <span class="vcv-msd-vault-count">{sel}/{group.mounts.length}</span>
            </button>

            <div class="vcv-msd-mounts">
              {#each group.mounts as m (m.key)}
                {@const isSelected = selectedSet.has(m.key)}
                <button
                  type="button"
                  class="vcv-msd-mount"
                  class:is-selected={isSelected}
                  onclick={() => toggle(m.key)}
                >
                  <span class="vcv-msd-check" data-state={isSelected ? 'all' : 'none'}>
                    {#if isSelected}
                      <Check class="h-3.5 w-3.5" />
                    {/if}
                  </span>
                  <Layers class="vcv-msd-mount-icon h-3.5 w-3.5" />
                  <span class="vcv-msd-mount-name">{m.mount}</span>
                </button>
              {/each}
            </div>
          </section>
        {/each}
      {/if}
    </div>

    <footer class="vcv-msd-foot">
      <span class="vcv-msd-stat">
        <strong>{selectedSet.size}</strong> / {allMounts.length}
        {i18n.t('mountStatsSelected', 'selected')}
      </span>
      <div class="vcv-msd-foot-actions">
        <button type="button" class="vcv-button vcv-button-small vcv-button-ghost" onclick={deselectAll}>
          {i18n.t('deselectAll', 'Deselect all')}
        </button>
        <button type="button" class="vcv-button vcv-button-small vcv-button-ghost" onclick={selectAll}>
          {i18n.t('selectAll', 'Select all')}
        </button>
        <button type="button" class="vcv-button vcv-button-primary vcv-button-small" onclick={() => onOpenChange(false)}>
          {i18n.t('buttonDone', 'Done')}
        </button>
      </div>
    </footer>
  </Dialog.Content>
</Dialog.Root>
