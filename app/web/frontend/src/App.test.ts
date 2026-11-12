// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/svelte'
import { toast } from 'svelte-sonner'
import type {
  Certificate,
  CertificatesEnvelope,
  I18nResponse,
  PublicConfigResponse,
  StatusResponse,
} from '$lib/types'
import type { CertsQuery } from '$lib/api'

const { listCertificates, config, status, i18n, getCertificateDetails } = vi.hoisted(() => ({
  listCertificates: vi.fn(),
  config: vi.fn(),
  status: vi.fn(),
  i18n: vi.fn(),
  getCertificateDetails: vi.fn(),
}))

vi.mock('$lib/api', () => ({
  api: { listCertificates, config, status, i18n, getCertificateDetails },
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

function envelopeFor(certificates: Certificate[], page: number, pageSize: number): CertificatesEnvelope {
  const totalPages = Math.max(1, Math.ceil(certificates.length / Math.max(1, pageSize)))
  return {
    certificates,
    errors: [],
    total: certificates.length,
    page,
    page_size: pageSize,
    total_pages: totalPages,
    counts: {
      valid: certificates.length,
      warning: 0,
      critical: 0,
      expired: 0,
      revoked: 0,
      total: certificates.length,
    },
  }
}

/** Mini-fake of GET /api/certs: honors search and numeric paging like the server. */
function mockOk(certificates: Certificate[]): void {
  listCertificates.mockImplementation((query: CertsQuery = {}) => {
    const q = (query.search ?? '').toLowerCase()
    const scoped = q
      ? certificates.filter((c) => c.commonName.toLowerCase().includes(q))
      : [...certificates]
    if (query.pageSize === undefined || query.pageSize === 'all') {
      return Promise.resolve(envelopeFor(scoped, 1, scoped.length))
    }
    const size = query.pageSize
    const page = query.page ?? 1
    const items = scoped.slice((page - 1) * size, page * size)
    return Promise.resolve({ ...envelopeFor(items, page, size), total: scoped.length })
  })
  config.mockResolvedValue({
    expirationThresholds: { critical: 7, warning: 30 },
  } satisfies PublicConfigResponse)
  status.mockResolvedValue({
    version: '1.9',
    vault_connected: true,
    vaults: [{ id: 'vault1', display_name: 'vault1', connected: true }],
  } satisfies StatusResponse)
  i18n.mockResolvedValue({ language: 'en', messages: {} } satisfies I18nResponse)
}

const originalMatchMedia = window.matchMedia

/** Fakes the viewport: `mobile` true makes `(max-width: 768px)` match. */
function setMobileViewport(mobile: boolean): void {
  window.matchMedia = vi.fn((query: string) => ({
    matches: mobile && query === '(max-width: 768px)',
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  })) as unknown as typeof window.matchMedia
}

beforeEach(() => {
  setMobileViewport(false)
  listCertificates.mockReset()
  config.mockReset()
  status.mockReset()
  i18n.mockReset()
  getCertificateDetails.mockReset()
  window.history.replaceState(null, '', '/')
  window.localStorage.clear()
})

afterEach(() => {
  vi.useRealTimers()
  vi.restoreAllMocks()
  window.matchMedia = originalMatchMedia
})

describe('App', () => {
  it('loads certificates on mount and renders them in the desktop table only', async () => {
    mockOk([cert({ id: 'a', commonName: 'web.example.com' }), cert({ id: 'b', commonName: 'api.example.com' })])

    const { container } = render(App)

    expect(await screen.findAllByText('web.example.com')).toHaveLength(1)
    expect(screen.getAllByText('api.example.com')).toHaveLength(1)
    expect(container.querySelector('table.vcv-table')).not.toBeNull()
    expect(container.querySelector('.vcv-cert-card')).toBeNull()
    expect(window.matchMedia).toHaveBeenCalledWith('(max-width: 768px)')
    // Initial load fetches the table page plus the full inventory.
    expect(listCertificates).toHaveBeenCalledTimes(2)
  })

  it('mounts only the mobile card list at the 768px breakpoint', async () => {
    setMobileViewport(true)
    mockOk([cert({ id: 'a', commonName: 'web.example.com' }), cert({ id: 'b', commonName: 'api.example.com' })])

    const { container } = render(App)

    expect(await screen.findAllByText('web.example.com')).toHaveLength(1)
    expect(container.querySelectorAll('.vcv-cert-card')).toHaveLength(2)
    expect(container.querySelector('table.vcv-table')).toBeNull()
  })

  it('debounces search and updates the result count only after the timer fires', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    mockOk([cert({ id: 'a', commonName: 'web.example.com' }), cert({ id: 'b', commonName: 'other.internal' })])

    render(App)
    await vi.waitFor(() => expect(listCertificates).toHaveBeenCalledTimes(2))

    const searchInput = screen.getByLabelText('Search certificates') as HTMLInputElement
    await fireEvent.input(searchInput, { target: { value: 'web' } })

    // Immediately after typing, filtering hasn't happened yet (still debounced).
    expect(screen.getByText('2 certificates')).toBeInTheDocument()

    await vi.advanceTimersByTimeAsync(200)

    expect(screen.getByText('1 certificates')).toBeInTheDocument()
  })

  it('keeps the Sources chip visible when the inventory only holds the selected mount', async () => {
    window.history.replaceState(null, '', `/?mounts=${encodeURIComponent('a|pki')}`)
    mockOk([cert({ id: 'a|pki:aa', commonName: 'web.example.com' })])
    config.mockResolvedValue({
      expirationThresholds: { critical: 7, warning: 30 },
      vaults: [
        { id: 'a', displayName: 'A', pkiMounts: ['pki'] },
        { id: 'b', displayName: 'B', pkiMounts: ['pki'] },
      ],
    } satisfies PublicConfigResponse)

    render(App)

    expect(await screen.findByText('1 / 2')).toBeInTheDocument()
  })

  it('restores filter/sort/page state from the URL on mount', async () => {
    window.history.replaceState(null, '', '/?q=web&status=critical&sort=commonName&dir=desc')
    mockOk([cert({ id: 'a', commonName: 'web.example.com' })])

    render(App)
    await vi.waitFor(() => expect(listCertificates).toHaveBeenCalledTimes(2))

    const searchInput = screen.getByLabelText('Search certificates') as HTMLInputElement
    expect(searchInput.value).toBe('web')
  })

  it('writes filter state back to the URL after the debounce, once hydrated', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
    mockOk([cert({ id: 'a', commonName: 'web.example.com' })])

    render(App)
    await vi.waitFor(() => expect(listCertificates).toHaveBeenCalledTimes(2))

    const searchInput = screen.getByLabelText('Search certificates') as HTMLInputElement
    await fireEvent.input(searchInput, { target: { value: 'web' } })

    await vi.advanceTimersByTimeAsync(400)

    expect(window.location.search).toContain('q=web')
  })

  it('shows the error banner and no error toast after a failed manual refresh', async () => {
    mockOk([cert({ id: 'a', commonName: 'web.example.com' })])
    render(App)
    await screen.findAllByText('web.example.com')

    const errorSpy = vi.spyOn(toast, 'error')
    mockDown()
    await fireEvent.click(screen.getByRole('button', { name: 'Refresh' }))

    expect(
      await screen.findByText('Network error loading certificates. Please try again.'),
    ).toBeInTheDocument()
    expect(errorSpy).not.toHaveBeenCalled()
  })

  it('shows the last updated time without enabling auto-refresh', async () => {
    mockOk([cert({ id: 'a', commonName: 'web.example.com' })])
    render(App)
    await screen.findAllByText('web.example.com')

    expect(screen.getByText(/^Updated /)).toBeInTheDocument()
  })
  it('opens the detail dialog from a ?cert= deep link', async () => {
    const c = cert()
    mockOk([c])
    getCertificateDetails.mockResolvedValue({ ...c, issuer: 'CA', subject: 'CN=web', keyAlgorithm: 'RSA', keySize: 2048, fingerprintSHA1: 'aa', fingerprintSHA256: 'bb', usage: [], pem: '', caType: '' })
    window.history.replaceState(null, '', `/?cert=${encodeURIComponent(c.id)}`)

    render(App)

    expect(await screen.findByRole('dialog')).toBeInTheDocument()
    expect(getCertificateDetails).toHaveBeenCalledWith(c.id)
  })

  it('toggles compact rows from the results bar', async () => {
    mockOk([cert()])
    render(App)
    await screen.findByText('web.example.com')

    await fireEvent.click(screen.getByRole('button', { name: 'Row density' }))

    expect(document.querySelector('.vcv-table')?.classList.contains('vcv-table-compact')).toBe(true)
    expect(window.localStorage.getItem('vcv-density')).toBe('compact')
  })

  it('points first-run users at the admin panel when no vault is configured', async () => {
    mockOk([])
    status.mockResolvedValue({ version: '1.9', vault_connected: false, vaults: [] } satisfies StatusResponse)

    render(App)

    const link = await screen.findByRole('link', { name: 'Open admin panel' })
    expect(link).toHaveAttribute('href', '/admin')
    expect(screen.getByText('No vaults configured yet')).toBeInTheDocument()
  })

  it('shows the no-mounts guidance when a vault has no PKI mount', async () => {
    mockOk([])
    status.mockResolvedValue({
      version: '1.9',
      vault_connected: true,
      vaults: [{ id: 'vault1', display_name: 'vault1', connected: true }],
    } satisfies StatusResponse)

    render(App)

    expect(await screen.findByText('No PKI mounts configured yet')).toBeInTheDocument()
  })

  it('does not show first-run guidance when certificates are listed', async () => {
    mockOk([cert()])

    render(App)

    await screen.findByText('web.example.com')
    expect(screen.queryByRole('link', { name: 'Open admin panel' })).toBeNull()
  })

})

function mockDown(): void {
  listCertificates.mockImplementation(() => Promise.reject(new Error('backend down')))
}
