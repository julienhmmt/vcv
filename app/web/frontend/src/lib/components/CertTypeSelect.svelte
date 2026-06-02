<script lang="ts">
  import * as Select from '$lib/components/ui/select'
  import { getI18n } from '$lib/stores/i18n.svelte'
  import type { CertTypeFilter } from '$lib/utils/cert-filter'

  interface Props {
    value: CertTypeFilter
    onChange: (value: CertTypeFilter) => void
  }

  const { value, onChange }: Props = $props()
  const i18n = getI18n()

  const options = $derived<{ value: CertTypeFilter; label: string }[]>([
    { value: 'all', label: i18n.t('certTypeFilterAll', 'All types') },
    { value: 'machine', label: i18n.t('certTypeFilterMachine', 'Machine') },
    { value: 'user', label: i18n.t('certTypeFilterUser', 'User') },
    { value: 'both', label: i18n.t('certTypeFilterBoth', 'Both') },
    { value: 'unknown', label: i18n.t('certTypeFilterUnknown', 'Unknown') },
  ])

  const currentLabel = $derived(options.find((o) => o.value === value)?.label ?? options[0].label)
</script>

<Select.Root type="single" value={value} onValueChange={(next) => onChange(next as CertTypeFilter)}>
  <Select.Trigger class="vcv-select vcv-select-compact h-9 min-w-[140px]">
    {currentLabel}
  </Select.Trigger>
  <Select.Content>
    {#each options as option}
      <Select.Item value={option.value}>{option.label}</Select.Item>
    {/each}
  </Select.Content>
</Select.Root>
