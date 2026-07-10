// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/svelte'
import StatusOverview from '$lib/components/StatusOverview.svelte'

const meta = {
  valid: { label: 'Valid', desc: 'All good' },
  warning: { label: 'Warning', desc: '≤ 30 days' },
  critical: { label: 'Critical', desc: '≤ 7 days' },
  expired: { label: 'Expired', desc: 'Past expiry' },
  revoked: { label: 'Revoked', desc: 'Revoked by CA' },
}

const counts = { valid: 30, warning: 85, critical: 90, expired: 60, revoked: 15 }

function setup(statusFilters: Array<keyof typeof meta> = [], onSelect = vi.fn()) {
  render(StatusOverview, {
    props: { counts, meta, statusFilters, regionLabel: 'Certificate status overview', onSelect },
  })
  return { onSelect }
}

describe('StatusOverview', () => {
  it('renders five status controls with localized labels and counts', () => {
    setup()
    const buttons = screen.getAllByRole('button')
    expect(buttons).toHaveLength(5)
    for (const key of ['valid', 'warning', 'critical', 'expired', 'revoked'] as const) {
      expect(screen.getByText(meta[key].label)).toBeInTheDocument()
      expect(screen.getByText(String(counts[key]))).toBeInTheDocument()
    }
  })

  it('shows threshold descriptions only for Warning and Critical', () => {
    setup()
    expect(screen.getByText('≤ 30 days')).toBeInTheDocument()
    expect(screen.getByText('≤ 7 days')).toBeInTheDocument()
    expect(screen.queryByText('All good')).toBeNull()
    expect(screen.queryByText('Past expiry')).toBeNull()
  })

  it('calls onSelect with the clicked status key', async () => {
    const { onSelect } = setup()
    await fireEvent.click(screen.getByText('Critical'))
    expect(onSelect).toHaveBeenCalledWith('critical')
  })

  it('marks active status controls with aria-pressed', () => {
    setup(['warning'])
    const pressed = screen.getAllByRole('button', { pressed: true })
    expect(pressed).toHaveLength(1)
    expect(pressed[0]).toHaveTextContent('Warning')
  })

  it('renders no chart and no duplicated total', () => {
    const { container } = render(StatusOverview, {
      props: { counts, meta, statusFilters: [], regionLabel: 'overview', onSelect: vi.fn() },
    })
    // Status icons are SVGs, but the donut chart and its container must be gone.
    expect(container.querySelector('.vcv-donut-svg')).toBeNull()
    expect(container.querySelector('.vcv-overview-chart')).toBeNull()
    // The word "certs" (the old donut center label) must not appear.
    expect(screen.queryByText('certs')).toBeNull()
  })
})
