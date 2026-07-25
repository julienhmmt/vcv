<script lang="ts">
  import Search from '@lucide/svelte/icons/search'
  import CertTypeSelect from '$lib/components/CertTypeSelect.svelte'
  import { getI18n } from '$lib/stores/i18n.svelte'
  import type { CertTypeFilter } from '$lib/utils/cert-filter'

  interface Props {
    search: string
    certTypeFilter: CertTypeFilter
    mountFilter: string[] | null
    allMountsCount: number
    onSearchInput: (value: string) => void
    onCertTypeChange: (value: CertTypeFilter) => void
    onOpenMountModal: () => void
  }

  const { search, certTypeFilter, mountFilter, allMountsCount, onSearchInput, onCertTypeChange, onOpenMountModal }: Props =
    $props()
  const i18n = getI18n()
</script>

<div id="vcv-filter-bar" class="vcv-filter-bar">
  <div class="vcv-filter-bar-inner">
    <div class="vcv-filter-palette">
      <div class="vcv-palette-item">
        <span class="vcv-palette-label">{i18n.t('filterChipSources', 'Sources')}</span>
        <button type="button" class="vcv-mount-filter" onclick={onOpenMountModal}>
          {#if mountFilter === null || mountFilter.length === allMountsCount}
            {i18n.t('sourcesButtonAll', 'All mounts ({total})', { total: allMountsCount })}
          {:else}
            {i18n.t('sourcesButtonPartial', '{selected} / {total} mounts', {
              selected: mountFilter.length,
              total: allMountsCount,
            })}
          {/if}
        </button>
      </div>
      <span class="vcv-palette-separator" aria-hidden="true"></span>
      <div class="vcv-palette-item">
        <span class="vcv-palette-label">{i18n.t('filterChipCertType', 'Type')}</span>
        <CertTypeSelect value={certTypeFilter} onChange={onCertTypeChange} />
      </div>
    </div>
    <div class="vcv-search-wrapper">
      <Search class="vcv-search-icon h-[18px] w-[18px]" aria-hidden="true" />
      <input
        id="vcv-search"
        class="vcv-input vcv-input-search"
        type="search"
        aria-label={i18n.t('searchLabel', 'Search certificates')}
        placeholder={i18n.t('searchPlaceholder', 'Search certificates, serials, SANs…')}
        value={search}
        oninput={(event) => onSearchInput((event.target as HTMLInputElement).value)}
      />
      <kbd class="vcv-search-shortcut" aria-label={i18n.t('searchShortcutHint', 'Press / to focus search')}>/</kbd>
    </div>
  </div>
</div>
