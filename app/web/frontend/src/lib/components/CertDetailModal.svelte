<script lang="ts">
  import Modal from './Modal.svelte'
  import { api, ApiError } from '$lib/api'
  import type { Certificate, DetailedCertificate } from '$lib/types'

  interface Props {
    cert: Certificate | null
    open: boolean
    onClose: () => void
    onShowCA?: (certId: string) => void
  }

  const { cert, open, onClose, onShowCA }: Props = $props()

  let details = $state<DetailedCertificate | null>(null)
  let loading = $state(false)
  let error = $state<string | null>(null)

  $effect(() => {
    if (!open || !cert) {
      details = null
      error = null
      return
    }
    loading = true
    error = null
    api
      .getCertificateDetails(cert.id)
      .then((data) => {
        details = data
      })
      .catch((err: unknown) => {
        error = err instanceof ApiError ? err.message : 'Failed to load details'
      })
      .finally(() => {
        loading = false
      })
  })

  async function copyPem(): Promise<void> {
    if (!details?.pem) return
    await navigator.clipboard.writeText(details.pem)
  }
</script>

<Modal {open} title={cert?.commonName || 'Certificate'} {onClose}>
  <div class="vcv-details-content">
    {#if loading}
      <p>Loading…</p>
    {:else if error}
      <p class="vcv-error">{error}</p>
    {:else if details}
      <div class="vcv-detail-grid">
        <div class="vcv-detail-row"><span class="vcv-detail-label">Serial</span><span class="vcv-detail-value vcv-mono">{details.serialNumber}</span></div>
        <div class="vcv-detail-row"><span class="vcv-detail-label">Issuer</span><span class="vcv-detail-value">{details.issuer || '—'}</span></div>
        <div class="vcv-detail-row"><span class="vcv-detail-label">Subject</span><span class="vcv-detail-value">{details.subject || '—'}</span></div>
        <div class="vcv-detail-row"><span class="vcv-detail-label">Key</span><span class="vcv-detail-value">{details.keyAlgorithm} {details.keySize ? `(${details.keySize} bits)` : ''}</span></div>
        <div class="vcv-detail-row"><span class="vcv-detail-label">SHA1</span><span class="vcv-detail-value vcv-mono">{details.fingerprintSHA1}</span></div>
        <div class="vcv-detail-row"><span class="vcv-detail-label">SHA256</span><span class="vcv-detail-value vcv-mono">{details.fingerprintSHA256}</span></div>
      </div>

      {#if details.sans?.length}
        <div class="vcv-detail-section">
          <h3 class="vcv-detail-section-title">SANs</h3>
          <div class="vcv-tag-list">
            {#each details.sans as san}
              <span class="vcv-tag">{san}</span>
            {/each}
          </div>
        </div>
      {/if}

      {#if details.usage?.length}
        <div class="vcv-detail-section">
          <h3 class="vcv-detail-section-title">Usage</h3>
          <div class="vcv-tag-list">
            {#each details.usage as use}
              <span class="vcv-tag">{use}</span>
            {/each}
          </div>
        </div>
      {/if}

      {#if details.pem}
        <div class="vcv-detail-section">
          <div class="vcv-detail-section-head">
            <h3 class="vcv-detail-section-title">PEM</h3>
            <button type="button" class="vcv-button vcv-button-small" onclick={copyPem}>Copy</button>
          </div>
          <pre class="vcv-pem">{details.pem}</pre>
        </div>
      {/if}
    {/if}
  </div>

  {#snippet actions()}
    {#if cert && onShowCA}
      <button
        type="button"
        class="vcv-button vcv-button-secondary"
        onclick={() => {
          onClose()
          onShowCA(cert.id)
        }}
      >
        View issuer CA
      </button>
    {/if}
    <button type="button" class="vcv-button vcv-button-secondary" onclick={onClose}>Close</button>
  {/snippet}
</Modal>
