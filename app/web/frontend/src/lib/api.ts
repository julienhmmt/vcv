import type {
  AdminSessionResponse,
  AdminSettingsResponse,
  AdminVaultAddedResponse,
  Certificate,
  DetailedCertificate,
  PemResponse,
  SettingsFile,
  StatusResponse,
  VersionInfo,
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

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    credentials: 'same-origin',
    headers: {
      Accept: 'application/json',
      ...(init?.headers ?? {}),
    },
    ...init,
  })
  if (!response.ok) {
    throw new ApiError(response.status, `${init?.method ?? 'GET'} ${path} failed: ${response.status}`)
  }
  return (await response.json()) as T
}

export const api = {
  listCertificates(mounts?: string[]): Promise<Certificate[]> {
    const qs = mounts === undefined ? '' : `?mounts=${encodeURIComponent(mounts.join(','))}`
    return request<Certificate[]>(`/api/certs${qs}`)
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
  version(): Promise<VersionInfo> {
    return request<VersionInfo>('/api/version')
  },
  i18n(lang?: string): Promise<Record<string, string>> {
    const qs = lang ? `?lang=${encodeURIComponent(lang)}` : ''
    return request<Record<string, string>>(`/api/i18n${qs}`)
  },
  getCertificateCA(id: string): Promise<DetailedCertificate> {
    return request<DetailedCertificate>(`/api/certs/${encodeURIComponent(id)}/ca`)
  },
  adminSession(): Promise<AdminSessionResponse> {
    return request<AdminSessionResponse>('/api/admin/session')
  },
  adminLogin(username: string, password: string): Promise<AdminSessionResponse> {
    return request<AdminSessionResponse>('/api/admin/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    })
  },
  async adminLogout(): Promise<void> {
    await fetch('/api/admin/logout', { method: 'POST', credentials: 'same-origin' })
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
  async adminDeleteVault(id: string): Promise<void> {
    const response = await fetch(`/api/admin/vault/${encodeURIComponent(id)}`, {
      method: 'DELETE',
      credentials: 'same-origin',
    })
    if (!response.ok) {
      throw new ApiError(response.status, `DELETE /api/admin/vault/${id} failed`)
    }
  },
  async adminInvalidateCache(): Promise<void> {
    const response = await fetch('/api/cache/invalidate', {
      method: 'POST',
      credentials: 'same-origin',
    })
    if (!response.ok) {
      throw new ApiError(response.status, 'cache invalidate failed')
    }
  },
}
