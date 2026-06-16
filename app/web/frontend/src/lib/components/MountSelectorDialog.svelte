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

<style>
  :global([data-slot="dialog-content"].vcv-msd) {
    display: flex;
    flex-direction: column;
    gap: 0;
    width: 100%;
    max-width: 36rem;
    padding: 0;
    overflow: hidden;
    border-radius: var(--vcv-radius-xl);
    background: var(--vcv-color-modal-surface);
    box-shadow: var(--vcv-shadow-card-hover);
  }

  :global(.vcv-msd-head) {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
    padding: 1.25rem 1.5rem 1rem;
  }

  :global(.vcv-msd-head-text) {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    min-width: 0;
  }

  :global(.vcv-msd-title) {
    font-size: 1.0625rem;
    font-weight: 650;
    letter-spacing: -0.01em;
    color: var(--vcv-color-text-strong);
  }

  :global(.vcv-msd-desc) {
    font-size: 0.8125rem;
    color: var(--vcv-color-muted);
  }

  :global(.vcv-msd-close) {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    border: 0;
    border-radius: var(--vcv-radius-full);
    background: transparent;
    color: var(--vcv-color-muted);
    cursor: pointer;
    transition:
      background-color 0.15s ease,
      color 0.15s ease;
  }

  :global(.vcv-msd-close:hover) {
    background: var(--vcv-color-modal-close-hover-bg);
    color: var(--vcv-color-text-strong);
  }

  :global(.vcv-msd-search) {
    position: relative;
    display: flex;
    align-items: center;
    margin: 0 1.5rem 0.25rem;
  }

  :global(.vcv-msd-search-icon) {
    position: absolute;
    left: 0.875rem;
    color: var(--vcv-color-muted);
    pointer-events: none;
  }

  :global(.vcv-msd-search-input) {
    width: 100%;
    padding: 0.625rem 2.25rem 0.625rem 2.5rem;
    border: 1px solid var(--vcv-color-border);
    border-radius: var(--vcv-radius-md);
    background: var(--vcv-color-surface-muted);
    color: var(--vcv-color-text-strong);
    font-size: 0.875rem;
    outline: none;
    transition:
      border-color 0.15s ease,
      box-shadow 0.15s ease,
      background-color 0.15s ease;
  }

  :global(.vcv-msd-search-input::placeholder) {
    color: var(--vcv-color-muted);
  }

  :global(.vcv-msd-search-input:focus) {
    border-color: var(--vcv-color-primary);
    background: var(--vcv-color-surface);
    box-shadow: 0 0 0 3px var(--vcv-color-focus-ring);
  }

  :global(.vcv-msd-search-clear) {
    position: absolute;
    right: 0.625rem;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 1.5rem;
    height: 1.5rem;
    border: 0;
    border-radius: var(--vcv-radius-full);
    background: transparent;
    color: var(--vcv-color-muted);
    cursor: pointer;
    transition:
      background-color 0.15s ease,
      color 0.15s ease;
  }

  :global(.vcv-msd-search-clear:hover) {
    background: var(--vcv-color-modal-close-hover-bg);
    color: var(--vcv-color-text-strong);
  }

  :global(.vcv-msd-list) {
    flex: 1;
    min-height: 0;
    max-height: min(70vh, 32rem);
    overflow-y: auto;
    padding: 0.5rem 1.5rem 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.625rem;
    scrollbar-width: thin;
  }

  :global(.vcv-msd-empty) {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    padding: 2.5rem 1rem;
    color: var(--vcv-color-muted);
    font-size: 0.875rem;
  }

  :global(.vcv-msd-group) {
    flex: 0 0 auto;
    border: 1px solid var(--vcv-color-border);
    border-radius: var(--vcv-radius-lg);
    background: var(--vcv-color-surface);
    overflow: hidden;
  }

  :global(.vcv-msd-vault) {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    width: 100%;
    min-height: 2.75rem;
    padding: 0.625rem 0.875rem;
    border: 0;
    background: var(--vcv-color-surface-muted);
    color: var(--vcv-color-text-strong);
    font-size: 0.8125rem;
    font-weight: 600;
    cursor: pointer;
    text-align: left;
    transition: background-color 0.15s ease;
  }

  :global(.vcv-msd-vault:hover) {
    background: var(--vcv-color-bg-hover);
  }

  :global(.vcv-msd-vault-icon) {
    flex-shrink: 0;
    color: var(--vcv-color-muted);
    transition: color 0.15s ease;
  }

  :global(.vcv-msd-vault.is-active .vcv-msd-vault-icon) {
    color: var(--vcv-color-primary);
  }

  :global(.vcv-msd-vault-name) {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  }

  :global(.vcv-msd-vault-count) {
    flex-shrink: 0;
    padding: 0.125rem 0.5rem;
    border-radius: var(--vcv-radius-full);
    background: var(--vcv-color-border);
    color: var(--vcv-color-text-subtle);
    font-size: 0.6875rem;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }

  :global(.vcv-msd-vault.is-active .vcv-msd-vault-count) {
    background: var(--vcv-color-primary-soft);
    color: var(--vcv-color-primary-strong);
  }

  :global(.vcv-msd-mounts) {
    display: flex;
    flex-direction: column;
    padding: 0.25rem;
  }

  :global(.vcv-msd-mount) {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    width: 100%;
    min-height: 2.75rem;
    padding: 0.5rem 0.625rem;
    border: 0;
    border-radius: var(--vcv-radius-md);
    background: transparent;
    color: var(--vcv-color-text);
    font-size: 0.8125rem;
    cursor: pointer;
    text-align: left;
    transition:
      background-color 0.12s ease,
      color 0.12s ease;
  }

  :global(.vcv-msd-mount:hover) {
    background: var(--vcv-color-bg-hover);
  }

  :global(.vcv-msd-mount.is-selected) {
    color: var(--vcv-color-text-strong);
  }

  :global(.vcv-msd-mount-icon) {
    flex-shrink: 0;
    color: var(--vcv-color-muted);
  }

  :global(.vcv-msd-mount.is-selected .vcv-msd-mount-icon) {
    color: var(--vcv-color-primary);
  }

  :global(.vcv-msd-mount-name) {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  }

  :global(.vcv-msd-check) {
    flex-shrink: 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 1.125rem;
    height: 1.125rem;
    border: 1.5px solid var(--vcv-color-border-strong);
    border-radius: 0.375rem;
    background: var(--vcv-color-surface);
    color: var(--vcv-color-on-accent);
    transition:
      background-color 0.15s ease,
      border-color 0.15s ease,
      transform 0.12s ease;
  }

  :global(.vcv-msd-check[data-state="all"]) {
    background: var(--vcv-color-primary);
    border-color: var(--vcv-color-primary-strong);
  }

  :global(.vcv-msd-check[data-state="some"]) {
    background: var(--vcv-color-primary-soft);
    border-color: var(--vcv-color-primary);
    color: var(--vcv-color-primary-strong);
  }

  :global(.vcv-msd-mount:active .vcv-msd-check),
  :global(.vcv-msd-vault:active .vcv-msd-check) {
    transform: scale(0.88);
  }

  :global(.vcv-msd-foot) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    flex-wrap: wrap;
    padding: 0.875rem 1.5rem;
    border-top: 1px solid var(--vcv-color-modal-header-border);
    background: var(--vcv-color-modal-actions-bg);
  }

  :global(.vcv-msd-stat) {
    font-size: 0.8125rem;
    color: var(--vcv-color-muted);
    font-variant-numeric: tabular-nums;
  }

  :global(.vcv-msd-stat strong) {
    color: var(--vcv-color-text-strong);
    font-weight: 650;
  }

  :global(.vcv-msd-foot-actions) {
    display: flex;
    gap: 0.5rem;
  }

  @media (520px >= width) {
    :global(.vcv-msd-foot-actions) {
      width: 100%;
    }

    :global(.vcv-msd-foot-actions .vcv-button) {
      flex: 1;
    }
  }
</style>
