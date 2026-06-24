import { certStatus, parseCertID, DEFAULT_THRESHOLDS } from '$lib/utils/cert-status'
import type { Certificate, ExpirationThresholds } from '$lib/types'

export type ExportFormat = 'csv' | 'json'

interface ExportRow {
  commonName: string
  sans: string
  vault: string
  mount: string
  certType: string
  status: string
  expiresAt: string
  serialNumber: string
}

const COLUMNS: readonly (keyof ExportRow)[] = [
  'commonName',
  'sans',
  'vault',
  'mount',
  'certType',
  'status',
  'expiresAt',
  'serialNumber',
]

/** Flatten a certificate into the export row shape (status derived, IDs parsed). */
function toRow(cert: Certificate, thresholds: ExpirationThresholds): ExportRow {
  const parts = parseCertID(cert.id)
  return {
    commonName: cert.commonName,
    sans: cert.sans.join(' '),
    vault: parts.vault,
    mount: parts.mount,
    certType: cert.certType,
    status: certStatus(cert, thresholds),
    expiresAt: cert.expiresAt,
    serialNumber: cert.serialNumber,
  }
}

/**
 * Prefix a single quote when a cell could be interpreted as a spreadsheet
 * formula. Covers the characters Excel/Sheets/LibreOffice treat as formula
 * starters: = + - @ tab (0x09) carriage return (0x0d). Leading whitespace is
 * trimmed for the trigger check so a value like " =cmd" is still neutralized.
 * The leading quote is stripped by spreadsheet apps on display but prevents
 * formula evaluation.
 */
function sanitizeCsvCell(value: string): string {
  const trimmed = value.replace(/^ +/, '')
  if (trimmed.length === 0) return value
  const first = trimmed[0]
  if (first === '=' || first === '+' || first === '-' || first === '@' || first === '\t' || first === '\r') {
    return `'${value}`
  }
  return value
}

/** Escape a value for RFC 4180 CSV: wrap in quotes when it contains a comma, quote, or newline. */
function escapeCsv(value: string): string {
  if (/[",\n\r]/.test(value)) return `"${value.replace(/"/g, '""')}"`
  return value
}

function toCsv(rows: ExportRow[]): string {
  const header = COLUMNS.join(',')
  const lines = rows.map((row) => COLUMNS.map((col) => escapeCsv(sanitizeCsvCell(row[col]))).join(','))
  return [header, ...lines].join('\r\n')
}

/** Serialize the given certificates to the chosen format using derived status fields. */
export function buildExport(
  certs: readonly Certificate[],
  format: ExportFormat,
  thresholds: ExpirationThresholds = DEFAULT_THRESHOLDS,
): { content: string; mime: string; extension: string } {
  const rows = certs.map((cert) => toRow(cert, thresholds))
  if (format === 'json') {
    return { content: JSON.stringify(rows, null, 2), mime: 'application/json', extension: 'json' }
  }
  return { content: toCsv(rows), mime: 'text/csv;charset=utf-8', extension: 'csv' }
}

/** Trigger a client-side download of the certificate inventory in the chosen format. */
export function downloadExport(
  certs: readonly Certificate[],
  format: ExportFormat,
  thresholds: ExpirationThresholds = DEFAULT_THRESHOLDS,
): void {
  const { content, mime, extension } = buildExport(certs, format, thresholds)
  const stamp = new Date().toISOString().slice(0, 10)
  const blob = new Blob([content], { type: mime })
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = `vcv-certificates-${stamp}.${extension}`
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
  URL.revokeObjectURL(url)
}
