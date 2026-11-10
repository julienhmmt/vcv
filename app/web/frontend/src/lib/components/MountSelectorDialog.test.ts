// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/svelte'

vi.mock('$lib/stores/i18n.svelte', () => ({
  getI18n: () => ({
    t: (_key: string, fallback?: string) => fallback ?? _key,
  }),
}))

import MountSelectorDialog from '$lib/components/MountSelectorDialog.svelte'

function renderDialog(selected: string[] | null, onChange = vi.fn()) {
  render(MountSelectorDialog, {
    props: {
      open: true,
      onOpenChange: vi.fn(),
      allMounts: ['a|pki', 'b|pki'],
      selected,
      onChange,
    },
  })
  return onChange
}

describe('MountSelectorDialog', () => {
  it('lists every mount even when the selection is a subset', () => {
    renderDialog(['a|pki'])
    expect(screen.getAllByText('pki', { selector: '.vcv-msd-mount-name' })).toHaveLength(2)
  })

  it('selecting the remaining mount reports null (all selected)', async () => {
    const onChange = renderDialog(['a|pki'])
    const rows = screen.getAllByRole('button').filter((b) => b.classList.contains('vcv-msd-mount'))
    await fireEvent.click(rows[1])
    expect(onChange).toHaveBeenCalledWith(null)
  })

  it('deselect all never emits an empty filter', async () => {
    const onChange = renderDialog(['a|pki'])
    await fireEvent.click(screen.getByRole('button', { name: 'Deselect all' }))
    expect(onChange).not.toHaveBeenCalled()
  })

  it('deselecting the last selected mount never emits an empty filter', async () => {
    const onChange = renderDialog(['a|pki'])
    const rows = screen.getAllByRole('button').filter((b) => b.classList.contains('vcv-msd-mount'))
    await fireEvent.click(rows[0])
    expect(onChange).not.toHaveBeenCalledWith([])
  })
})
