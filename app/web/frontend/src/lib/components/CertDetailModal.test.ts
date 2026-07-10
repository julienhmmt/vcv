// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte'
import type { Certificate, DetailedCertificate } from '$lib/types'

const { getCertificateDetails, getCertificateCA } = vi.hoisted(() => ({
  getCertificateDetails: vi.fn(),
  getCertificateCA: vi.fn(),
}))

vi.mock('$lib/api', () => ({
  api: { getCertificateDetails, getCertificateCA },
  ApiError: class ApiError extends Error {},
}))

vi.mock('$lib/stores/i18n.svelte', () => ({
  getI18n: () => ({ t: (_key: string, fallback?: string) => fallback ?? _key, lang: 'en' }),
}))

import CertDetailModal from '$lib/components/CertDetailModal.svelte'

const cert: Certificate = {
  id: 'vault1|pki-int:aa:bb',
  serialNumber: 'aa:bb',
  commonName: 'web.example.com',
  sans: ['web.example.com', 'api.example.com'],
  certType: 'machine',
  createdAt: '2024-01-01T00:00:00Z',
  expiresAt: '2999-01-01T00:00:00Z',
  revoked: false,
}

function detailed(overrides: Partial<DetailedCertificate> = {}): DetailedCertificate {
  return {
    ...cert,
    issuer: 'Example Intermediate CA',
    subject: 'CN=web.example.com',
    keyAlgorithm: 'RSA',
    keySize: 2048,
    fingerprintSHA1: '11:22',
    fingerprintSHA256: '33:44',
    usage: ['Server Authentication'],
    pem: '-----BEGIN CERTIFICATE-----\nX\n-----END CERTIFICATE-----',
    caType: '',
    ...overrides,
  }
}

beforeEach(() => {
  getCertificateDetails.mockReset()
  getCertificateCA.mockReset()
})

describe('CertDetailModal', () => {
  it('fetches details on open and shows issuer, usage and SANs', async () => {
    getCertificateDetails.mockResolvedValue(detailed())
    render(CertDetailModal, { props: { cert, open: true, onOpenChange: vi.fn() } })

    await waitFor(() => expect(getCertificateDetails).toHaveBeenCalledWith(cert.id))
    expect(await screen.findByText('Example Intermediate CA')).toBeInTheDocument()
    expect(screen.getByText('Server Authentication')).toBeInTheDocument()
    expect(screen.getByText('api.example.com')).toBeInTheDocument()
  })

  it('keeps technical identifiers in a closed-by-default disclosure', async () => {
    getCertificateDetails.mockResolvedValue(detailed())
    render(CertDetailModal, { props: { cert, open: true, onOpenChange: vi.fn() } })

    await screen.findByText('Example Intermediate CA')
    const disclosure = document.querySelector('.vcv-cd-technical') as HTMLDetailsElement
    expect(disclosure).not.toBeNull()
    expect(disclosure.hasAttribute('open')).toBe(false)
    expect(screen.getByText('Technical details')).toBeInTheDocument()
  })

  it('loads the issuer in the same dialog without re-fetching the certificate', async () => {
    getCertificateDetails.mockResolvedValue(detailed())
    getCertificateCA.mockResolvedValue(detailed({ commonName: 'Example Intermediate CA', caType: 'intermediate' }))
    render(CertDetailModal, { props: { cert, open: true, onOpenChange: vi.fn() } })

    await screen.findByText('Example Intermediate CA')
    await fireEvent.click(screen.getByText('View issuer CA'))

    await waitFor(() => expect(getCertificateCA).toHaveBeenCalledWith(cert.id))
    expect(await screen.findByText('Back to certificate')).toBeInTheDocument()

    await fireEvent.click(screen.getByText('Back to certificate'))
    // Certificate view restored from cache — no second details fetch.
    expect(getCertificateDetails).toHaveBeenCalledTimes(1)
  })

  it('shows a Retry action when the detail request fails', async () => {
    getCertificateDetails.mockRejectedValueOnce(new Error('boom')).mockResolvedValueOnce(detailed())
    render(CertDetailModal, { props: { cert, open: true, onOpenChange: vi.fn() } })

    const retry = await screen.findByText('Retry')
    await fireEvent.click(retry)
    expect(await screen.findByText('Example Intermediate CA')).toBeInTheDocument()
    expect(getCertificateDetails).toHaveBeenCalledTimes(2)
  })
})
