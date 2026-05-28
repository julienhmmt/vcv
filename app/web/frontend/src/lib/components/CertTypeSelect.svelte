<script lang="ts">
  import * as Select from '$lib/components/ui/select'
  import type { CertTypeFilter } from '$lib/utils/cert-filter'

  interface Props {
    value: CertTypeFilter
    onChange: (value: CertTypeFilter) => void
  }

  const { value, onChange }: Props = $props()

  const options: { value: CertTypeFilter; label: string }[] = [
    { value: 'all', label: 'All types' },
    { value: 'machine', label: 'Machine' },
    { value: 'user', label: 'User' },
    { value: 'both', label: 'Both' },
    { value: 'unknown', label: 'Unknown' },
  ]

  const currentLabel = $derived(options.find((o) => o.value === value)?.label ?? 'All types')
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
