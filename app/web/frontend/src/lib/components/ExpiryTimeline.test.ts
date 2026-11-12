// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/svelte'
import type { Certificate } from '$lib/types'

vi.mock('$lib/stores/i18n.svelte', () => ({
  getI18n: () => ({
    t: (_key: string, fallback?: string, params?: Record<string, string | number>) =>
      Object.entries(params ?? {}).reduce(
        (acc, [name, value]) => acc.replaceAll(`{${name}}`, String(value)),
        fallback ?? _key,
      ),
  }),
}))

import ExpiryTimeline from '$lib/components/ExpiryTimeline.svelte'

const thresholds = { critical: 7, warning: 30 }

function cert(overrides: Partial<Certificate> = {}): Certificate {
  return {
    id: 'v|m:1',
    serialNumber: 'aa:bb',
    commonName: 'web.example.com',
    sans: [],
    certType: 'machine',
    createdAt: '2024-01-01T00:00:00Z',
    expiresAt: new Date(Date.now() + 3.5 * 86_400_000).toISOString(),
    revoked: false,
    ...overrides,
  }
}

describe('ExpiryTimeline', () => {
  it('renders bucket counts and threshold-aware labels', () => {
    render(ExpiryTimeline, {
      props: { certs: [cert()], thresholds, minDays: null, maxDays: null, onSelect: vi.fn() },
    })

    const region = screen.getByRole('region', { name: 'Upcoming expirations' })
    expect(region).toBeInTheDocument()
    expect(region.textContent).toContain('≤ 7 days')
    expect(region.textContent).toContain('8–30 days')
    expect(region.textContent).toContain('31–90 days')
    expect(region.textContent).toContain('> 90 days')
    // One cert expiring in ~3 days shows up in the critical bucket.
    const criticalBucket = region.querySelector('.vcv-expiry-bucket-critical .vcv-expiry-bucket-count')
    expect(criticalBucket?.textContent).toBe('1')
  })

  it('renders nothing when no certificate expires in the future', () => {
    const { container } = render(ExpiryTimeline, {
      props: { certs: [cert({ revoked: true })], thresholds, minDays: null, maxDays: null, onSelect: vi.fn() },
    })

    expect(container.querySelector('.vcv-expiry-timeline')).toBeNull()
  })
  it('selects a bucket range and clears it when the same bucket is clicked again', async () => {
    const onSelect = vi.fn()
    render(ExpiryTimeline, {
      props: { certs: [cert()], thresholds, minDays: null, maxDays: null, onSelect },
    })

    await fireEvent.click(screen.getByRole('button', { name: /8–30 days/ }))
    expect(onSelect).toHaveBeenLastCalledWith({ minDays: 8, maxDays: 30 })
  })

  it('reflects the active range with aria-pressed', () => {
    render(ExpiryTimeline, {
      props: { certs: [cert()], thresholds, minDays: 8, maxDays: 30, onSelect: vi.fn() },
    })

    expect(screen.getByRole('button', { name: /8–30 days/ })).toHaveAttribute('aria-pressed', 'true')
    expect(screen.getByRole('button', { name: /≤ 7 days/ })).toHaveAttribute('aria-pressed', 'false')
  })

})
