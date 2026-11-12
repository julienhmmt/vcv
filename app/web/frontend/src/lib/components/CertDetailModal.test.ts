// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte'
import type { Certificate, DetailedCertificate } from '$lib/types'

const { getCertificateDetails, getCertificateCA, copyToClipboard } = vi.hoisted(() => ({
  getCertificateDetails: vi.fn(),
  getCertificateCA: vi.fn(),
  copyToClipboard: vi.fn(),
}))

vi.mock('$lib/api', () => ({
  api: { getCertificateDetails, getCertificateCA },
  ApiError: class ApiError extends Error {},
}))

vi.mock('$lib/utils/clipboard', () => ({ copyToClipboard }))

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
  copyToClipboard.mockReset()
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

  it('presents the issuer in a passport layout with technical identifiers collapsed', async () => {
    getCertificateDetails.mockResolvedValue(detailed())
    getCertificateCA.mockResolvedValue(detailed({ caType: 'root', subject: 'CN=Example Root CA' }))
    render(CertDetailModal, { props: { cert, open: true, onOpenChange: vi.fn() } })

    await screen.findByText('Example Intermediate CA')
    await fireEvent.click(screen.getByText('View issuer CA'))
    await screen.findByText('Root CA')

    expect(document.querySelector('.vcv-ca-passport .vcv-cd-passport-sidebar')).not.toBeNull()
    const disclosure = document.querySelector('.vcv-ca-passport .vcv-cd-technical') as HTMLDetailsElement
    expect(disclosure).not.toBeNull()
    expect(disclosure.hasAttribute('open')).toBe(false)
  })

  it('shows a Retry action when the detail request fails', async () => {
    getCertificateDetails.mockRejectedValueOnce(new Error('boom')).mockResolvedValueOnce(detailed())
    render(CertDetailModal, { props: { cert, open: true, onOpenChange: vi.fn() } })

    const retry = await screen.findByText('Retry')
    await fireEvent.click(retry)
    expect(await screen.findByText('Example Intermediate CA')).toBeInTheDocument()
    expect(getCertificateDetails).toHaveBeenCalledTimes(2)
  })

  it('ignores a stale details response after switching certificates', async () => {
    const certB: Certificate = {
      ...cert,
      id: 'vault1|pki-int:cc:dd',
      serialNumber: 'cc:dd',
      commonName: 'other.example.com',
    }
    let resolveA: (value: DetailedCertificate) => void = () => {}
    const deferredA = new Promise<DetailedCertificate>((resolve) => {
      resolveA = resolve
    })
    getCertificateDetails.mockImplementation((id: string) => {
      if (id === cert.id) return deferredA
      return Promise.resolve(
        detailed({
          id: certB.id,
          serialNumber: certB.serialNumber,
          commonName: certB.commonName,
          issuer: 'Issuer B',
          subject: 'CN=other.example.com',
        }),
      )
    })
    const { rerender } = render(CertDetailModal, {
      props: { cert, open: true, onOpenChange: vi.fn() },
    })
    await waitFor(() => expect(getCertificateDetails).toHaveBeenCalledWith(cert.id))
    await rerender({ cert: certB, open: true, onOpenChange: vi.fn() })
    expect(await screen.findByText('Issuer B')).toBeInTheDocument()
    resolveA(
      detailed({
        id: cert.id,
        issuer: 'Stale Issuer A',
        subject: 'CN=web.example.com',
      }),
    )
    await new Promise((r) => setTimeout(r, 20))
    expect(screen.queryByText('Stale Issuer A')).not.toBeInTheDocument()
    expect(screen.getByText('Issuer B')).toBeInTheDocument()
  })

  it('ignores a stale details error after a successful newer load', async () => {
    const certB: Certificate = {
      ...cert,
      id: 'vault1|pki-int:ee:ff',
      serialNumber: 'ee:ff',
      commonName: 'third.example.com',
    }
    let rejectA: (reason?: unknown) => void = () => {}
    const deferredA = new Promise<DetailedCertificate>((_resolve, reject) => {
      rejectA = reject
    })
    getCertificateDetails.mockImplementation((id: string) => {
      if (id === cert.id) return deferredA
      return Promise.resolve(
        detailed({
          id: certB.id,
          serialNumber: certB.serialNumber,
          commonName: certB.commonName,
          issuer: 'Issuer Fresh',
        }),
      )
    })
    const { rerender } = render(CertDetailModal, {
      props: { cert, open: true, onOpenChange: vi.fn() },
    })
    await waitFor(() => expect(getCertificateDetails).toHaveBeenCalledWith(cert.id))
    await rerender({ cert: certB, open: true, onOpenChange: vi.fn() })
    expect(await screen.findByText('Issuer Fresh')).toBeInTheDocument()
    rejectA(new Error('stale failure'))
    await new Promise((r) => setTimeout(r, 20))
    expect(screen.queryByText('Retry')).not.toBeInTheDocument()
    expect(screen.getByText('Issuer Fresh')).toBeInTheDocument()
  })
  it('copies a deep link that contains the certificate id', async () => {
    getCertificateDetails.mockResolvedValue(detailed())
    copyToClipboard.mockResolvedValue(true)
    render(CertDetailModal, { props: { cert, open: true, onOpenChange: vi.fn() } })

    await screen.findByText('Example Intermediate CA')
    await fireEvent.click(screen.getByText('Copy link'))

    expect(copyToClipboard).toHaveBeenCalledTimes(1)
    expect(String(copyToClipboard.mock.calls[0][0])).toContain('cert=vault1%7Cpki-int%3Aaa%3Abb')
  })

  it('navigates the signing chain level by level', async () => {
    const intermediate = detailed({ id: 'vault1|pki-int:ca1', commonName: 'Intermediate CA', caType: 'intermediate' })
    const root = detailed({ id: 'vault1|pki-int:ca2', commonName: 'Root CA', caType: 'root' })
    getCertificateDetails.mockResolvedValue(detailed())
    getCertificateCA.mockResolvedValueOnce(intermediate).mockResolvedValueOnce(root)
    render(CertDetailModal, { props: { cert, open: true, onOpenChange: vi.fn() } })

    await screen.findByText('Example Intermediate CA')
    await fireEvent.click(screen.getByText('View issuer CA'))
    await screen.findByRole('button', { name: 'Intermediate CA' })
    await fireEvent.click(screen.getByText('View issuer CA'))
    await screen.findByRole('button', { name: 'Root CA' })

    expect(getCertificateCA).toHaveBeenNthCalledWith(1, cert.id)
    expect(getCertificateCA).toHaveBeenNthCalledWith(2, intermediate.id)
    const chain = screen.getByRole('navigation', { name: 'Certificate chain' })
    expect(chain.textContent).toContain('Certificate')
    expect(chain.textContent).toContain('Intermediate CA')
    expect(chain.textContent).toContain('Root CA')
  })

  it('does not offer a further issuer for a root certificate', async () => {
    getCertificateDetails.mockResolvedValue(detailed())
    getCertificateCA.mockResolvedValue(detailed({ caType: 'root', commonName: 'Root CA' }))
    render(CertDetailModal, { props: { cert, open: true, onOpenChange: vi.fn() } })

    await screen.findByText('Example Intermediate CA')
    await fireEvent.click(screen.getByText('View issuer CA'))
    await screen.findAllByText('Root CA')

    expect(screen.queryByText('View issuer CA')).toBeNull()
  })

  it('shows the elapsed validity as a progress bar', async () => {
    const timed = { ...cert, createdAt: '2020-01-01T00:00:00Z', expiresAt: '2033-01-01T00:00:00Z' }
    getCertificateDetails.mockResolvedValue(detailed())
    render(CertDetailModal, { props: { cert: timed, open: true, onOpenChange: vi.fn() } })

    const bar = await screen.findByRole('progressbar', { name: 'Validity' })
    const value = Number(bar.getAttribute('aria-valuenow'))
    expect(value).toBeGreaterThan(30)
    expect(value).toBeLessThan(70)
  })

  it('copies the subject', async () => {
    getCertificateDetails.mockResolvedValue(detailed())
    copyToClipboard.mockResolvedValue(true)
    render(CertDetailModal, { props: { cert, open: true, onOpenChange: vi.fn() } })

    await screen.findByText('Example Intermediate CA')
    const subjectRow = screen.getByText('Subject').parentElement as HTMLElement
    await fireEvent.click(subjectRow.querySelector('button') as HTMLButtonElement)

    expect(copyToClipboard).toHaveBeenCalledWith('CN=web.example.com')
  })

})
