// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/svelte'
import type { Certificate, ExpirationThresholds } from '$lib/types'

vi.mock('$lib/stores/i18n.svelte', () => ({
  getI18n: () => ({ t: (_key: string, fallback?: string) => fallback ?? _key }),
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

describe('CertTable', () => {
  it('renders a clickable row with certificate name and details label', () => {
    render(CertTable, {
      props: {
        certs: [cert],
        loading: false,
        initialLoad: false,
        hasInventory: true,
        hasActiveFilters: false,
        showVaultMount: false,
        statusMeta,
        thresholds,
        onSelect: vi.fn(),
        onClearFilters: vi.fn(),
      },
    })

    expect(screen.getByRole('button', { name: 'web.example.com: Details' })).toBeInTheDocument()
    expect(screen.getByText('web.example.com')).toBeInTheDocument()
    expect(screen.getByText('api.example.com')).toBeInTheDocument()
  })
})
