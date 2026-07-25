// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/svelte'
import type { Certificate } from '$lib/types'

vi.mock('$lib/stores/i18n.svelte', () => ({
  getI18n: () => ({ t: (_key: string, fallback?: string) => fallback ?? _key }),
  LANGUAGES: [{ code: 'en', label: 'English' }],
}))

import CommandPalette from '$lib/components/CommandPalette.svelte'

function cert(overrides: Partial<Certificate> = {}): Certificate {
  return {
    id: 'vault1|pki-int:aa:bb',
    serialNumber: 'aa:bb',
    commonName: 'web.example.com',
    sans: ['web.example.com', 'api.example.com'],
    certType: 'machine',
    createdAt: '2024-01-01T00:00:00Z',
    expiresAt: '2999-01-01T00:00:00Z',
    revoked: false,
    ...overrides,
  }
}

function baseProps(certs: Certificate[]) {
  return {
    open: true,
    onOpenChange: vi.fn(),
    certs,
    theme: 'light' as const,
    onSelectCert: vi.fn(),
    onToggleStatus: vi.fn(),
    onToggleTheme: vi.fn(),
    onSetLang: vi.fn(),
  }
}

describe('CommandPalette', () => {
  it('matches a certificate by common name, serial, or SAN substring', async () => {
    const certs = [
      cert({ id: 'a', commonName: 'web.example.com', serialNumber: 'aa:bb', sans: [] }),
      cert({ id: 'b', commonName: 'other.internal', serialNumber: 'cc:dd', sans: ['san-only.example.com'] }),
    ]
    render(CommandPalette, { props: baseProps(certs) })

    const input = screen.getByPlaceholderText('Search certificates or commands…')
    await fireEvent.input(input, { target: { value: 'san-only' } })

    expect(screen.getByText('other.internal')).toBeInTheDocument()
    expect(screen.queryByText('web.example.com')).not.toBeInTheDocument()
  })

  it('re-derives matches when the certificate list changes without a stale haystack', async () => {
    const initial = [cert({ id: 'a', commonName: 'first.example.com' })]
    const { rerender } = render(CommandPalette, { props: baseProps(initial) })

    const input = screen.getByPlaceholderText('Search certificates or commands…')
    await fireEvent.input(input, { target: { value: 'second' } })
    expect(screen.queryByText('second.example.com')).not.toBeInTheDocument()

    const updated = [cert({ id: 'b', commonName: 'second.example.com' })]
    await rerender({ ...baseProps(updated) })

    expect(screen.getByText('second.example.com')).toBeInTheDocument()
  })
})
