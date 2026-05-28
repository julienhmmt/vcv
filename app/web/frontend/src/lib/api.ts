import type {
  Certificate,
  DetailedCertificate,
  PemResponse,
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
}
