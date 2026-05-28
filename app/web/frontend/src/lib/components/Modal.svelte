<script lang="ts">
  import type { Snippet } from 'svelte'

  interface Props {
    open: boolean
    title: string
    large?: boolean
    onClose: () => void
    children: Snippet
    actions?: Snippet
  }

  const { open, title, large, onClose, children, actions }: Props = $props()

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === 'Escape') onClose()
  }

  function onBackdropClick(event: MouseEvent): void {
    if (event.target === event.currentTarget) onClose()
  }
</script>

<svelte:window onkeydown={onKeydown} />

{#if open}
  <div
    class="vcv-modal-backdrop"
    role="dialog"
    aria-modal="true"
    aria-label={title}
    onclick={onBackdropClick}
    onkeydown={onKeydown}
    tabindex="-1"
  >
    <div class="vcv-modal{large ? ' vcv-modal-lg' : ''}">
      <div class="vcv-modal-header">
        <h2 class="vcv-modal-title">{title}</h2>
        <button type="button" class="vcv-modal-close" aria-label="Close" onclick={onClose}>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M18 6 6 18" />
            <path d="m6 6 12 12" />
          </svg>
        </button>
      </div>
      <div class="vcv-modal-content">
        {@render children()}
      </div>
      {#if actions}
        <div class="vcv-modal-actions">
          {@render actions()}
        </div>
      {/if}
    </div>
  </div>
{/if}
