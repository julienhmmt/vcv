<script lang="ts">
  import Trash2 from '@lucide/svelte/icons/trash-2'
  import * as Dialog from '$lib/components/ui/dialog'
  import { Input } from '$lib/components/ui/input'
  import { getI18n } from '$lib/stores/i18n.svelte'
  import type { SavedView, ViewSnapshot } from '$lib/stores/saved-views.svelte'

  interface Props {
    open: boolean
    onOpenChange: (open: boolean) => void
    views: SavedView[]
    current: ViewSnapshot
    onSave: (name: string, snapshot: ViewSnapshot) => void
    onRemove: (id: string) => void
    onApply: (snapshot: ViewSnapshot) => void
  }

  const { open, onOpenChange, views, current, onSave, onRemove, onApply }: Props = $props()
  const i18n = getI18n()

  let name = $state('')

  function save(): void {
    if (!name.trim()) return
    onSave(name, current)
    name = ''
  }

  function apply(snapshot: ViewSnapshot): void {
    onApply(snapshot)
    onOpenChange(false)
  }
</script>

<Dialog.Root {open} {onOpenChange}>
  <Dialog.Content class="max-w-lg">
    <Dialog.Title>{i18n.t('savedViewsTitle', 'Saved views')}</Dialog.Title>

    <div class="vcv-saved-views">
      <div class="vcv-saved-views-save">
        <Input
          bind:value={name}
          placeholder={i18n.t('savedViewsNamePlaceholder', 'View name')}
          aria-label={i18n.t('savedViewsNamePlaceholder', 'View name')}
          onkeydown={(event) => {
            if (event.key === 'Enter') save()
          }}
        />
        <button type="button" class="vcv-button vcv-button-primary" disabled={!name.trim()} onclick={save}>
          {i18n.t('savedViewsSave', 'Save current view')}
        </button>
      </div>

      {#if views.length === 0}
        <p class="vcv-saved-views-empty">
          {i18n.t('savedViewsEmpty', 'No saved views yet. Save the current filters to reuse them later.')}
        </p>
      {:else}
        <ul class="vcv-saved-views-list">
          {#each views as view (view.id)}
            <li class="vcv-saved-views-row">
              <span class="vcv-saved-views-name">{view.name}</span>
              <button type="button" class="vcv-button vcv-button-small vcv-button-secondary" onclick={() => apply(view.snapshot)}>
                {i18n.t('savedViewsApply', 'Apply')}
              </button>
              <button
                type="button"
                class="vcv-icon-action"
                aria-label={i18n.t('savedViewsDelete', 'Delete view')}
                title={i18n.t('savedViewsDelete', 'Delete view')}
                onclick={() => onRemove(view.id)}
              >
                <Trash2 class="h-4 w-4" />
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  </Dialog.Content>
</Dialog.Root>
