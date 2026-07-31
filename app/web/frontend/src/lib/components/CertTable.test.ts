// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/svelte'
import type { Certificate, ExpirationThresholds } from '$lib/types'

vi.mock('$lib/stores/i18n.svelte', () => ({
  getI18n: () => ({
    t: (_key: string, fallback?: string, params?: Record<string, string | number>) =>
      Object.entries(params ?? {}).reduce(
        (acc, [name, value]) => acc.replaceAll(`{${name}}`, String(value)),
        fallback ?? _key,
      ),
  }),
}))

import CertTable from '$lib/components/CertTable.svelte'

const thresholds: ExpirationThresholds = { critical: 7, warning: 30 }

const cert: Certificate = {
  id: 'vault1|pki-int:aa:bb',
  serialNumber: 'aa:bb',
  commonName: 'web.example.com',
  sans: ['api.example.com'],
  certType: 'machine',
  createdAt: '2024-01-01T00:00:00Z',
  expiresAt: '2999-01-01T00:00:00Z',
  revoked: false,
}

const statusMeta = {
  valid: { label: 'Valid', desc: 'All good' },
  warning: { label: 'Warning', desc: '≤ 30 days' },
  critical: { label: 'Critical', desc: '≤ 7 days' },
  expired: { label: 'Expired', desc: 'Past expiry' },
  revoked: { label: 'Revoked', desc: 'Revoked by CA' },
}

const baseProps = {
  certs: [cert],
  loading: false,
  initialLoad: false,
  hasInventory: true,
  hasActiveFilters: false,
  showVaultMount: false,
  statusMeta,
  thresholds,
  sortKey: 'expiresAt' as const,
  sortDir: 'asc' as const,
  onSort: vi.fn(),
  onSelect: vi.fn(),
  onClearFilters: vi.fn(),
}

describe('CertTable', () => {
  it('renders a clickable row with certificate name and details label', () => {
    render(CertTable, { props: baseProps })

    expect(screen.getByRole('button', { name: 'web.example.com: Details' })).toBeInTheDocument()
    expect(screen.getByText('web.example.com')).toBeInTheDocument()
    expect(screen.getByText('api.example.com')).toBeInTheDocument()
  })

  it('marks the active sort column with aria-sort and the direction icon', () => {
    render(CertTable, { props: baseProps })

    const commonNameHead = screen.getByRole('columnheader', { name: /common name/i })
    const expiresHead = screen.getByRole('columnheader', { name: /expires/i })
    expect(commonNameHead).toHaveAttribute('aria-sort', 'none')
    expect(expiresHead).toHaveAttribute('aria-sort', 'ascending')
    expect(expiresHead.querySelector('.vcv-th-sort-icon')?.textContent).toBe('↑')
  })

  it('reports the sort key when a column header is clicked', async () => {
    const onSort = vi.fn()
    render(CertTable, { props: { ...baseProps, onSort } })

    await fireEvent.click(screen.getByRole('button', { name: 'Common Name' }))
    expect(onSort).toHaveBeenCalledWith('commonName')

    await fireEvent.click(screen.getByRole('button', { name: 'Expires' }))
    expect(onSort).toHaveBeenCalledWith('expiresAt')
  })

  it('reflects descending direction on the active column', () => {
    render(CertTable, { props: { ...baseProps, sortKey: 'commonName', sortDir: 'desc' } })

    const commonNameHead = screen.getByRole('columnheader', { name: /common name/i })
    expect(commonNameHead).toHaveAttribute('aria-sort', 'descending')
    expect(commonNameHead.querySelector('.vcv-th-sort-icon')?.textContent).toBe('↓')
  })
})
