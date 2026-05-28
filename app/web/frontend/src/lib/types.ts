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
