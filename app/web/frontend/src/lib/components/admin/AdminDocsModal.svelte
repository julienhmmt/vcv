<script lang="ts">
  import BookOpen from '@lucide/svelte/icons/book-open'
  import * as Dialog from '$lib/components/ui/dialog'
  import { ScrollArea } from '$lib/components/ui/scroll-area'
  import { api, ApiError } from '$lib/api'
  import { getI18n } from '$lib/stores/i18n.svelte'

  interface Props {
    open: boolean
    onOpenChange: (open: boolean) => void
  }

  const { open, onOpenChange }: Props = $props()
  const i18n = getI18n()

  let html = $state<string | null>(null)
  let loading = $state(false)
  let error = $state<string | null>(null)

  $effect(() => {
    if (!open || html !== null) return
    loading = true
    error = null
    api
      .adminDocs()
      .then((data) => {
        html = data.html
      })
      .catch((err: unknown) => {
        error = err instanceof ApiError ? err.message : i18n.t('adminDocsError', 'Failed to load documentation')
      })
      .finally(() => {
        loading = false
      })
  })
</script>

<Dialog.Root {open} {onOpenChange}>
  <Dialog.Content class="max-w-3xl p-0 overflow-hidden">
    <Dialog.Header class="px-6 pt-6">
      <Dialog.Title class="flex items-center gap-2">
        <BookOpen class="h-5 w-5 text-primary" />
        {i18n.t('adminDocsTitle', 'Documentation')}
      </Dialog.Title>
    </Dialog.Header>

    {#if loading && html === null}
      <div class="px-6 py-8 text-sm text-muted-foreground">{i18n.t('labelLoading', 'Loading…')}</div>
    {:else if error}
      <div class="px-6 py-8 text-sm text-destructive">{error}</div>
    {:else if html !== null}
      <ScrollArea class="max-h-[75vh]">
        <!-- Trusted: rendered from the binary-embedded ADMIN.md, not user input. -->
        <div class="vcv-docs px-6 pb-6">{@html html}</div>
      </ScrollArea>
    {/if}
  </Dialog.Content>
</Dialog.Root>

<style>
  .vcv-docs :global(h1) {
    font-size: 1.4rem;
    font-weight: 600;
    margin: 0.5rem 0 1rem;
  }
  .vcv-docs :global(h2) {
    font-size: 1.1rem;
    font-weight: 600;
    margin: 1.5rem 0 0.5rem;
  }
  .vcv-docs :global(p) {
    margin: 0.6rem 0;
    font-size: 0.875rem;
    line-height: 1.6;
  }
  .vcv-docs :global(ul) {
    margin: 0.6rem 0;
    padding-left: 1.25rem;
    list-style: disc;
    font-size: 0.875rem;
    line-height: 1.6;
  }
  .vcv-docs :global(li) {
    margin: 0.25rem 0;
  }
  .vcv-docs :global(a) {
    color: var(--primary);
    text-decoration: underline;
  }
  .vcv-docs :global(code) {
    font-family: var(--font-mono, ui-monospace, monospace);
    font-size: 0.8125rem;
    background: color-mix(in oklab, var(--muted) 60%, transparent);
    padding: 0.1em 0.35em;
    border-radius: 0.25rem;
  }
  .vcv-docs :global(pre) {
    background: color-mix(in oklab, var(--muted) 60%, transparent);
    padding: 0.75rem 1rem;
    border-radius: 0.5rem;
    overflow-x: auto;
    margin: 0.75rem 0;
  }
  .vcv-docs :global(pre code) {
    background: none;
    padding: 0;
  }
  .vcv-docs :global(blockquote) {
    border-left: 3px solid var(--border);
    padding-left: 0.75rem;
    margin: 0.75rem 0;
    color: var(--muted-foreground);
    font-size: 0.85rem;
  }
</style>
