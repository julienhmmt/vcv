// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/svelte'

vi.mock('$lib/stores/i18n.svelte', () => ({
  getI18n: () => ({ t: (_key: string, fallback?: string) => fallback ?? _key }),
}))

import SavedViewsDialog from '$lib/components/SavedViewsDialog.svelte'
import type { SavedView, ViewSnapshot } from '$lib/stores/saved-views.svelte'

const snapshot: ViewSnapshot = {
  search: 'web',
  statusFilters: ['critical'],
  certTypeFilter: 'all',
  mountFilter: null,
  sortKey: 'expiresAt',
  sortDir: 'asc',
  pageSize: 25,
  minDays: 0,
  maxDays: 7,
}

const view: SavedView = { id: 'v1', name: 'Critical', snapshot }

function baseProps(overrides: Record<string, unknown> = {}) {
  return {
    open: true,
    onOpenChange: vi.fn(),
    views: [] as SavedView[],
    current: snapshot,
    onSave: vi.fn(),
    onRemove: vi.fn(),
    onApply: vi.fn(),
    ...overrides,
  }
}

describe('SavedViewsDialog', () => {
  it('shows the empty state', () => {
    render(SavedViewsDialog, { props: baseProps() })

    expect(screen.getByText(/No saved views yet/)).toBeInTheDocument()
  })

  it('saves the current snapshot under the typed name', async () => {
    const onSave = vi.fn()
    render(SavedViewsDialog, { props: baseProps({ onSave }) })

    await fireEvent.input(screen.getByLabelText('View name'), { target: { value: '  Critical  ' } })
    await fireEvent.click(screen.getByRole('button', { name: 'Save current view' }))

    expect(onSave).toHaveBeenCalledWith('  Critical  ', snapshot)
  })

  it('applies a view and closes the dialog', async () => {
    const onApply = vi.fn()
    const onOpenChange = vi.fn()
    render(SavedViewsDialog, { props: baseProps({ views: [view], onApply, onOpenChange }) })

    await fireEvent.click(screen.getByRole('button', { name: 'Apply' }))

    expect(onApply).toHaveBeenCalledWith(snapshot)
    expect(onOpenChange).toHaveBeenCalledWith(false)
  })

  it('deletes a view', async () => {
    const onRemove = vi.fn()
    render(SavedViewsDialog, { props: baseProps({ views: [view], onRemove }) })

    await fireEvent.click(screen.getByRole('button', { name: 'Delete view' }))

    expect(onRemove).toHaveBeenCalledWith('v1')
  })
})
