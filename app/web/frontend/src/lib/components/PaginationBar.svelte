<script lang="ts">
  import * as Select from '$lib/components/ui/select'
  import { getI18n } from '$lib/stores/i18n.svelte'

  interface Props {
    pageSize: number | 'all'
    safePage: number
    totalPages: number
    pageInfoText: string
    onPageSizeChange: (size: number | 'all') => void
    onPrev: () => void
    onNext: () => void
    onFirst: () => void
    onLast: () => void
    onJump: (pageIndex: number) => void
  }

  const { pageSize, safePage, totalPages, pageInfoText, onPageSizeChange, onPrev, onNext, onFirst, onLast, onJump }: Props =
    $props()

  const i18n = getI18n()
</script>

<div class="vcv-pagination-bar">
  <div class="vcv-page-size">
    <span id="vcv-page-size-label">{i18n.t('paginationPageSizeLabel', 'Results per page')}</span>
    <Select.Root
      type="single"
      value={String(pageSize)}
      onValueChange={(value) => {
        if (!value) return
        onPageSizeChange(value === 'all' ? 'all' : Number(value))
      }}
    >
      <Select.Trigger class="vcv-select vcv-page-size-select h-9" aria-labelledby="vcv-page-size-label">
        {pageSize === 'all' ? i18n.t('paginationPageSizeAll', 'All') : String(pageSize)}
      </Select.Trigger>
      <Select.Content>
        <Select.Item value="25">25</Select.Item>
        <Select.Item value="50">50</Select.Item>
        <Select.Item value="100">100</Select.Item>
        <Select.Item value="all">{i18n.t('paginationPageSizeAll', 'All')}</Select.Item>
      </Select.Content>
    </Select.Root>
  </div>
  <span class="vcv-page-info">{pageInfoText}</span>
  <div class="vcv-page-buttons">
    <button
      type="button"
      class="vcv-button vcv-button-small vcv-button-ghost vcv-button-pill"
      aria-label={i18n.t('paginationFirst', 'First page')}
      title={i18n.t('paginationFirst', 'First page')}
      disabled={safePage === 0}
      onclick={onFirst}
    >
      «
    </button>
    <button
      type="button"
      class="vcv-button vcv-button-small vcv-button-ghost vcv-button-pill"
      disabled={safePage === 0}
      onclick={onPrev}
    >
      {i18n.t('paginationPrev', 'Previous')}
    </button>
    <label class="vcv-page-jump">
      <span class="sr-only">{i18n.t('paginationJumpLabel', 'Page')}</span>
      <input
        class="vcv-input vcv-page-jump-input"
        type="number"
        min="1"
        max={totalPages}
        value={safePage + 1}
        aria-label={i18n.t('paginationJumpLabel', 'Page')}
        onchange={(event) => {
          const value = Number((event.currentTarget as HTMLInputElement).value)
          if (Number.isInteger(value)) onJump(value - 1)
        }}
      />
      <span class="vcv-page-jump-total">{i18n.t('paginationPageOf', 'of {total}', { total: totalPages })}</span>
    </label>
    <button
      type="button"
      class="vcv-button vcv-button-small vcv-button-ghost vcv-button-pill"
      disabled={safePage >= totalPages - 1}
      onclick={onNext}
    >
      {i18n.t('paginationNext', 'Next')}
    </button>
    <button
      type="button"
      class="vcv-button vcv-button-small vcv-button-ghost vcv-button-pill"
      aria-label={i18n.t('paginationLast', 'Last page')}
      title={i18n.t('paginationLast', 'Last page')}
      disabled={safePage >= totalPages - 1}
      onclick={onLast}
    >
      »
    </button>
  </div>
</div>
