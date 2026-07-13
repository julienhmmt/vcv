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
  }

  const { pageSize, safePage, totalPages, pageInfoText, onPageSizeChange, onPrev, onNext }: Props = $props()

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
      disabled={safePage === 0}
      onclick={onPrev}
    >
      {i18n.t('paginationPrev', 'Previous')}
    </button>
    <button
      type="button"
      class="vcv-button vcv-button-small vcv-button-ghost vcv-button-pill"
      disabled={safePage >= totalPages - 1}
      onclick={onNext}
    >
      {i18n.t('paginationNext', 'Next')}
    </button>
  </div>
</div>
