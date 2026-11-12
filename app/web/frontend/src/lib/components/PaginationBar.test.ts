// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/svelte'

vi.mock('$lib/stores/i18n.svelte', () => ({
  getI18n: () => ({
    t: (_key: string, fallback?: string, params?: Record<string, string | number>) =>
      Object.entries(params ?? {}).reduce(
        (acc, [name, value]) => acc.replaceAll(`{${name}}`, String(value)),
        fallback ?? _key,
      ),
  }),
}))

import PaginationBar from '$lib/components/PaginationBar.svelte'

function baseProps(overrides: Record<string, unknown> = {}) {
  return {
    pageSize: 25,
    safePage: 2,
    totalPages: 10,
    pageInfoText: '51–75 of 240',
    onPageSizeChange: vi.fn(),
    onPrev: vi.fn(),
    onNext: vi.fn(),
    onFirst: vi.fn(),
    onLast: vi.fn(),
    onJump: vi.fn(),
    ...overrides,
  }
}

describe('PaginationBar', () => {
  it('shows the current page and the total page count', () => {
    render(PaginationBar, { props: baseProps() })

    expect(screen.getByRole('spinbutton', { name: 'Page' })).toHaveValue(3)
    expect(screen.getByText('of 10')).toBeInTheDocument()
  })

  it('jumps to the typed page (1-based input, 0-based callback)', async () => {
    const onJump = vi.fn()
    render(PaginationBar, { props: baseProps({ onJump }) })

    const input = screen.getByRole('spinbutton', { name: 'Page' })
    await fireEvent.change(input, { target: { value: '7' } })

    expect(onJump).toHaveBeenCalledWith(6)
  })

  it('calls the first and last handlers and disables them at the ends', async () => {
    const onFirst = vi.fn()
    const onLast = vi.fn()
    render(PaginationBar, { props: baseProps({ onFirst, onLast }) })

    await fireEvent.click(screen.getByRole('button', { name: 'First page' }))
    await fireEvent.click(screen.getByRole('button', { name: 'Last page' }))

    expect(onFirst).toHaveBeenCalledTimes(1)
    expect(onLast).toHaveBeenCalledTimes(1)
  })

  it('disables first and previous on the first page', () => {
    render(PaginationBar, { props: baseProps({ safePage: 0 }) })

    expect(screen.getByRole('button', { name: 'First page' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Previous' })).toBeDisabled()
  })

  it('disables last and next on the last page', () => {
    render(PaginationBar, { props: baseProps({ safePage: 9 }) })

    expect(screen.getByRole('button', { name: 'Last page' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Next' })).toBeDisabled()
  })
})
