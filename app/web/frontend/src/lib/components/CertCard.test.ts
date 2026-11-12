// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/svelte'
import type { Certificate } from '$lib/types'

const { copyToClipboard } = vi.hoisted(() => ({ copyToClipboard: vi.fn() }))

vi.mock('$lib/utils/clipboard', () => ({ copyToClipboard }))

vi.mock('$lib/stores/i18n.svelte', () => ({
  getI18n: () => ({ t: (_key: string, fallback?: string) => fallback ?? _key }),
}))

import CertCard from '$lib/components/CertCard.svelte'

const cert: Certificate = {
  id: 'vault1|pki-int:aa:bb',
  serialNumber: 'aa:bb',
  commonName: 'web.example.com',
  sans: [],
  certType: 'machine',
  createdAt: '2024-01-01T00:00:00Z',
  expiresAt: '2999-01-01T00:00:00Z',
  revoked: false,
}

describe('CertCard', () => {
  it('explains the row action with text instead of an unlabeled chevron', () => {
    const { container } = render(CertCard, {
      props: { cert, showVaultMount: false, status: 'valid', statusLabel: 'Valid', onSelect: vi.fn() },
    })

    expect(screen.getByRole('button', { name: 'web.example.com: Details' })).toBeInTheDocument()
    expect(screen.getByText('Details')).toBeInTheDocument()
    expect(container.querySelector('.vcv-cert-card-action svg')).toBeNull()
  })

  it('places the status below the certificate name as a quiet inline badge', () => {
    const { container } = render(CertCard, {
      props: { cert, showVaultMount: false, status: 'critical', statusLabel: 'Critical', onSelect: vi.fn() },
    })
    const title = container.querySelector('.vcv-cert-card-title')

    expect(title?.querySelector('.vcv-cert-status-inline')).toHaveTextContent('Critical')
    expect(container.querySelector('.vcv-cert-card-header > .vcv-status-badges')).toBeNull()
  })
  it('shows the certificate type next to the name', () => {
    render(CertCard, {
      props: { cert, showVaultMount: false, status: 'valid', statusLabel: 'Valid', onSelect: vi.fn() },
    })

    expect(screen.getByText('Machine')).toBeInTheDocument()
  })

  it('copies the serial without opening the detail dialog', async () => {
    const onSelect = vi.fn()
    copyToClipboard.mockResolvedValue(true)
    render(CertCard, {
      props: { cert, showVaultMount: false, status: 'valid', statusLabel: 'Valid', onSelect },
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Copy serial number' }))

    expect(copyToClipboard).toHaveBeenCalledWith('aa:bb')
    expect(onSelect).not.toHaveBeenCalled()
  })

})
