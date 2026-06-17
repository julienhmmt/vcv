// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/svelte'
import Donut from '$lib/components/Donut.svelte'

const labels = {
  valid: 'Valid',
  warning: 'Warning',
  critical: 'Critical',
  expired: 'Expired',
  revoked: 'Revoked',
}

describe('Donut', () => {
  it('shows the total and renders one interactive segment per non-zero status', () => {
    render(Donut, {
      props: {
        counts: { valid: 3, warning: 0, critical: 1, expired: 0, revoked: 0 },
        label: 'certs',
        segmentLabels: labels,
        onSelect: vi.fn(),
      },
    })

    expect(screen.getByText('4')).toBeInTheDocument()
    expect(screen.getByText('certs')).toBeInTheDocument()

    const segments = screen.getAllByRole('button')
    expect(segments).toHaveLength(2) // valid + critical only
    expect(screen.getByLabelText('Valid: 3')).toBeInTheDocument()
    expect(screen.getByLabelText('Critical: 1')).toBeInTheDocument()
  })

  it('toggles the status filter when a segment is clicked', async () => {
    const onSelect = vi.fn()
    render(Donut, {
      props: {
        counts: { valid: 2, warning: 0, critical: 0, expired: 0, revoked: 0 },
        label: 'certs',
        segmentLabels: labels,
        onSelect,
      },
    })

    await fireEvent.click(screen.getByLabelText('Valid: 2'))
    expect(onSelect).toHaveBeenCalledWith('valid')
  })

  it('renders no segments and an empty ring when the total is zero', () => {
    const { container } = render(Donut, {
      props: {
        counts: { valid: 0, warning: 0, critical: 0, expired: 0, revoked: 0 },
        label: 'certs',
      },
    })

    expect(screen.queryAllByRole('button')).toHaveLength(0)
    expect(container.querySelector('.vcv-donut-empty')).not.toBeNull()
  })
})
