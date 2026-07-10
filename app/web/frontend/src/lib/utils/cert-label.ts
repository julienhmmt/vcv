import type { Certificate } from '$lib/types'

/**
 * Return a non-empty display name for a certificate, suitable for an
 * accessible name (aria-label). Falls back to the first SAN, then the
 * serial number, then the provided `fallback` (a localized "unnamed"
 * string) so the result is never empty.
 */
export function certDisplayName(cert: Certificate, fallback: string): string {
  if (cert.commonName.trim() !== '') return cert.commonName
  if (cert.sans.length > 0 && cert.sans[0].trim() !== '') return cert.sans[0]
  if (cert.serialNumber.trim() !== '') return cert.serialNumber
  return fallback
}
