// Mirrors Go JSON shapes from app/internal/certs and app/cmd/server/main.go.

export interface Certificate {
  id: string
  serialNumber: string
  commonName: string
  sans: string[]
  certType: string
  createdAt: string
  expiresAt: string
  revoked: boolean
}

export interface DetailedCertificate extends Certificate {
  issuer: string
  subject: string
  keyAlgorithm: string
  keySize: number
  fingerprintSHA1: string
  fingerprintSHA256: string
  usage: string[]
  pem: string
  caType: 'intermediate' | 'root' | ''
}

export interface VaultListError {
  vaultId: string
  message: string
}

export interface CertificatesEnvelope {
  certificates: Certificate[]
  errors: VaultListError[]
}

export interface PemResponse {
  serialNumber: string
  pem: string
}

export interface VaultStatusEntry {
  id: string
  display_name: string
  connected: boolean
  error?: string
}

export interface StatusResponse {
  version: string
  vault_connected: boolean
  vault_error?: string
  vaults: VaultStatusEntry[]
}

export interface VersionInfo {
  version: string
  commit?: string
  buildDate?: string
}

export type CertStatus = 'valid' | 'expired' | 'revoked' | 'critical' | 'warning'

export interface ExpirationThresholds {
  critical: number
  warning: number
}

export interface VaultInstance {
  id: string
  original_id?: string
  address: string
  token?: string
  pki_mount?: string
  pki_mounts?: string[]
  display_name?: string
  tls_insecure?: boolean
  tls_ca_cert?: string
  tls_ca_cert_base64?: string
  tls_ca_path?: string
  tls_server_name?: string
  enabled?: boolean | null
}

export interface CertificateSettings {
  expiration_thresholds: ExpirationThresholds
}

export interface MetricsSettings {
  per_certificate?: boolean | null
  enhanced_metrics?: boolean | null
  pinned_certificates?: string[]
}

export interface CORSSettings {
  allowed_origins?: string[]
  allow_credentials?: boolean
}

export interface AppSettings {
  env: string
  port: number
}

export interface AdminSettings {
  password?: string
}

export interface SettingsFile {
  app: AppSettings
  admin?: AdminSettings
  certificates: CertificateSettings
  metrics: MetricsSettings
  cors: CORSSettings
  vaults: VaultInstance[]
}

export interface AdminVaultStatus {
  id: string
  enabled: boolean
  connected: boolean
}

export interface AdminSettingsResponse {
  settings: SettingsFile
  vault_statuses: AdminVaultStatus[]
}

export interface I18nResponse {
  language: string
  messages: Record<string, string>
}

export interface AdminSessionResponse {
  authenticated: boolean
}

export interface AdminDocsResponse {
  html: string
}

export interface AdminVaultAddedResponse {
  key: string
  vault: VaultInstance
}

/** Public config from GET /api/config (no secrets). */
export interface PublicConfigResponse {
  expirationThresholds: ExpirationThresholds
  metrics?: { per_certificate?: boolean; enhanced_metrics?: boolean }
  pkiMounts?: string[]
  vaults?: { id: string; displayName: string; pkiMounts: string[] }[]
}
