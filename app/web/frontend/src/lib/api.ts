import type {
  AdminDocsResponse,
  AdminSessionResponse,
  AdminSettingsResponse,
  AdminVaultAddedResponse,
  CertificatesEnvelope,
  DetailedCertificate,
  I18nResponse,
  PemResponse,
  PublicConfigResponse,
  SettingsFile,
  StatusResponse,
} from './types'

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

async function parseErrorMessage(response: Response, path: string, method: string): Promise<string> {
  let message = `${method} ${path} failed: ${response.status}`
  try {
    const body = (await response.json()) as { error?: string }
    if (body?.error) message = body.error
  } catch {
    // non-JSON body
  }
  return message
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const { headers: initHeaders, ...restInit } = init ?? {}
  const method = init?.method ?? 'GET'
  const response = await fetch(path, {
    credentials: 'same-origin',
    ...restInit,
    headers: {
      Accept: 'application/json',
      ...initHeaders,
    },
  })
  if (!response.ok) {
    throw new ApiError(response.status, await parseErrorMessage(response, path, method))
  }
  return (await response.json()) as T
}

/** POST/DELETE with no response body on success; still parses body.error on failure. */
async function requestVoid(path: string, init?: RequestInit): Promise<void> {
  const { headers: initHeaders, ...restInit } = init ?? {}
  const method = init?.method ?? 'GET'
  const response = await fetch(path, {
    credentials: 'same-origin',
    ...restInit,
    headers: {
      Accept: 'application/json',
      ...initHeaders,
    },
  })
  if (!response.ok) {
    throw new ApiError(response.status, await parseErrorMessage(response, path, method))
  }
}

/** Server-side list query mirroring GET /api/certs params. Page is 1-based. */
export interface CertsQuery {
  mounts?: string[]
  search?: string
  statuses?: string[]
  certType?: string
  sort?: string
  order?: 'asc' | 'desc'
  page?: number
  pageSize?: number | 'all'
}

export const api = {
  listCertificates(query: CertsQuery = {}): Promise<CertificatesEnvelope> {
    const params = new URLSearchParams()
    if (query.mounts !== undefined) params.set('mounts', query.mounts.join(','))
    if (query.search) params.set('search', query.search)
    if (query.statuses && query.statuses.length > 0) params.set('status', query.statuses.join(','))
    if (query.certType && query.certType !== 'all') params.set('cert_type', query.certType)
    if (query.sort) params.set('sort', query.sort)
    if (query.order) params.set('order', query.order)
    if (query.page !== undefined) params.set('page', String(query.page))
    if (query.pageSize !== undefined) params.set('page_size', String(query.pageSize))
    const qs = params.toString()
    return request<CertificatesEnvelope>(`/api/certs${qs ? `?${qs}` : ''}`)
  },
  getCertificateDetails(id: string): Promise<DetailedCertificate> {
    return request<DetailedCertificate>(`/api/certs/${encodeURIComponent(id)}/details`)
  },
  getCertificatePem(id: string): Promise<PemResponse> {
    return request<PemResponse>(`/api/certs/${encodeURIComponent(id)}/pem`)
  },
  status(): Promise<StatusResponse> {
    return request<StatusResponse>('/api/status')
  },
  config(): Promise<PublicConfigResponse> {
    return request<PublicConfigResponse>('/api/config')
  },
  i18n(lang?: string): Promise<I18nResponse> {
    const qs = lang ? `?lang=${encodeURIComponent(lang)}` : ''
    return request<I18nResponse>(`/api/i18n${qs}`)
  },
  getCertificateCA(id: string): Promise<DetailedCertificate> {
    return request<DetailedCertificate>(`/api/certs/${encodeURIComponent(id)}/ca`)
  },
  adminSession(): Promise<AdminSessionResponse> {
    return request<AdminSessionResponse>('/api/admin/session')
  },
  adminDocs(): Promise<AdminDocsResponse> {
    return request<AdminDocsResponse>('/api/admin/docs')
  },
  adminLogin(username: string, password: string): Promise<AdminSessionResponse> {
    return request<AdminSessionResponse>('/api/admin/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    })
  },
  adminLogout(): Promise<void> {
    return requestVoid('/api/admin/logout', { method: 'POST' })
  },
  adminGetSettings(): Promise<AdminSettingsResponse> {
    return request<AdminSettingsResponse>('/api/admin/settings')
  },
  adminPutSettings(settings: SettingsFile): Promise<AdminSettingsResponse> {
    return request<AdminSettingsResponse>('/api/admin/settings', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(settings),
    })
  },
  adminAddVault(): Promise<AdminVaultAddedResponse> {
    return request<AdminVaultAddedResponse>('/api/admin/vault', { method: 'POST' })
  },
  adminDeleteVault(id: string): Promise<void> {
    return requestVoid(`/api/admin/vault/${encodeURIComponent(id)}`, { method: 'DELETE' })
  },
  adminInvalidateCache(): Promise<void> {
    return requestVoid('/api/cache/invalidate', { method: 'POST' })
  },
}
