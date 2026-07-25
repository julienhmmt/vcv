// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/svelte'
import type { Certificate, CertificatesEnvelope, I18nResponse, PublicConfigResponse, StatusResponse } from '$lib/types'

const { listCertificates, config, status, i18n } = vi.hoisted(() => ({
  listCertificates: vi.fn(),
  config: vi.fn(),
  status: vi.fn(),
  i18n: vi.fn(),
}))

vi.mock('$lib/api', () => ({
  api: { listCertificates, config, status, i18n },
  ApiError: class ApiError extends Error {},
}))

import App from './App.svelte'

function cert(overrides: Partial<Certificate> = {}): Certificate {
  return {
    id: 'vault1|pki:aa',
    serialNumber: 'aa',
    commonName: 'web.example.com',
    sans: [],
    certType: 'machine',
    createdAt: '2024-01-01T00:00:00Z',
    expiresAt: '2999-01-01T00:00:00Z',
    revoked: false,
    ...overrides,
  }
}

function mockOk(certificates: Certificate[]): void {
  listCertificates.mockResolvedValue({ certificates, errors: [] } satisfies CertificatesEnvelope)
  config.mockResolvedValue({
    expirationThresholds: { critical: 7, warning: 30 },
  } satisfies PublicConfigResponse)
  status.mockResolvedValue({
    version: '1.9',
    vault_connected: true,
    vaults: [],
  } satisfies StatusResponse)
  i18n.mockResolvedValue({ language: 'en', messages: {} } satisfies I18nResponse)
}

beforeEach(() => {
  listCertificates.mockReset()
  config.mockReset()
  status.mockReset()
  i18n.mockReset()
  window.history.replaceState(null, '', '/')
})

afterEach(() => {
  vi.useRealTimers()
})

describe('App', () => {
  it('loads certificates on mount and renders them (desktop table + mobile list both mount in jsdom)', async () => {
    mockOk([cert({ id: 'a', commonName: 'web.example.com' }), cert({ id: 'b', commonName: 'api.example.com' })])

    render(App)

    expect(await screen.findAllByText('web.example.com')).toHaveLength(2)
    expect(screen.getAllByText('api.example.com')).toHaveLength(2)
    expect(listCertificates).toHaveBeenCalledTimes(1)
  })

  it('debounces search and updates the result count only after the timer fires', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    mockOk([cert({ id: 'a', commonName: 'web.example.com' }), cert({ id: 'b', commonName: 'other.internal' })])

    render(App)
    await vi.waitFor(() => expect(listCertificates).toHaveBeenCalledTimes(1))

    const searchInput = screen.getByLabelText('Search certificates') as HTMLInputElement
    await fireEvent.input(searchInput, { target: { value: 'web' } })

    // Immediately after typing, filtering hasn't happened yet (still debounced).
    expect(screen.getByText('2 certificates')).toBeInTheDocument()

    await vi.advanceTimersByTimeAsync(200)

    expect(screen.getByText('1 certificates')).toBeInTheDocument()
  })

  it('restores filter/sort/page state from the URL on mount', async () => {
    window.history.replaceState(null, '', '/?q=web&status=critical&sort=commonName&dir=desc')
    mockOk([cert({ id: 'a', commonName: 'web.example.com' })])

    render(App)
    await vi.waitFor(() => expect(listCertificates).toHaveBeenCalledTimes(1))

    const searchInput = screen.getByLabelText('Search certificates') as HTMLInputElement
    expect(searchInput.value).toBe('web')
  })

  it('writes filter state back to the URL after the debounce, once hydrated', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    mockOk([cert({ id: 'a', commonName: 'web.example.com' })])

    render(App)
    await vi.waitFor(() => expect(listCertificates).toHaveBeenCalledTimes(1))

    const searchInput = screen.getByLabelText('Search certificates') as HTMLInputElement
    await fireEvent.input(searchInput, { target: { value: 'web' } })

    await vi.advanceTimersByTimeAsync(400)

    expect(window.location.search).toContain('q=web')
  })
})
