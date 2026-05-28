<script lang="ts">
  import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
  } from '$lib/components/ui/table'
  import CertStatusBadge from './CertStatusBadge.svelte'
  import { certStatus, daysUntilExpiry, DEFAULT_THRESHOLDS } from '$lib/utils/cert-status'
  import type { Certificate, ExpirationThresholds } from '$lib/types'

  interface Props {
    certificates: Certificate[]
    loading: boolean
    error: string | null
    thresholds?: ExpirationThresholds
    onSelect?: (cert: Certificate) => void
  }

  const {
    certificates,
    loading,
    error,
    thresholds = DEFAULT_THRESHOLDS,
    onSelect,
  }: Props = $props()

  function formatDate(iso: string): string {
    return new Date(iso).toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: '2-digit',
    })
  }
</script>

{#if error}
  <p class="text-sm text-destructive">{error}</p>
{:else if loading && certificates.length === 0}
  <p class="text-sm text-muted-foreground">Loading certificates…</p>
{:else if certificates.length === 0}
  <p class="text-sm text-muted-foreground">No certificates found.</p>
{:else}
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Common Name</TableHead>
        <TableHead>Serial</TableHead>
        <TableHead>Type</TableHead>
        <TableHead>Expires</TableHead>
        <TableHead>Status</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each certificates as cert (cert.id)}
        {@const status = certStatus(cert, thresholds)}
        {@const days = daysUntilExpiry(cert)}
        <TableRow
          class="cursor-pointer"
          onclick={() => onSelect?.(cert)}
        >
          <TableCell class="font-medium">{cert.commonName || '—'}</TableCell>
          <TableCell class="font-mono text-xs">{cert.serialNumber}</TableCell>
          <TableCell>{cert.certType}</TableCell>
          <TableCell>{formatDate(cert.expiresAt)}</TableCell>
          <TableCell>
            <CertStatusBadge {status} {days} />
          </TableCell>
        </TableRow>
      {/each}
    </TableBody>
  </Table>
{/if}
