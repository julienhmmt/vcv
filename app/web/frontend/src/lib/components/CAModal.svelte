<script lang="ts">
  import Modal from './Modal.svelte'
  import { api, ApiError } from '$lib/api'
  import type { DetailedCertificate } from '$lib/types'

  interface Props {
    certId: string | null
    open: boolean
    onClose: () => void
  }

  const { certId, open, onClose }: Props = $props()

  let ca = $state<DetailedCertificate | null>(null)
  let loading = $state(false)
  let error = $state<string | null>(null)

  $effect(() => {
    if (!open || !certId) {
      ca = null
      error = null
      return
    }
    loading = true
    error = null
    api
      .getCertificateCA(certId)
      .then((data) => {
        ca = data
      })
      .catch((err: unknown) => {
        error = err instanceof ApiError ? err.message : 'Failed to load CA'
      })
      .finally(() => {
        loading = false
      })
  })

  async function copyPem(): Promise<void> {
    if (!ca?.pem) return
    await navigator.clipboard.writeText(ca.pem)
  }
</script>

<Modal {open} title={ca?.caType === 'root' ? 'Root CA' : 'Intermediate CA'} {onClose}>
  <div class="vcv-details-content">
    {#if loading}
      <p>Loading…</p>
    {:else if error}
      <p class="vcv-error">{error}</p>
    {:else if ca}
      <div class="vcv-detail-grid">
        <div class="vcv-detail-row"><span class="vcv-detail-label">Subject</span><span class="vcv-detail-value">{ca.subject}</span></div>
        <div class="vcv-detail-row"><span class="vcv-detail-label">Issuer</span><span class="vcv-detail-value">{ca.issuer}</span></div>
        <div class="vcv-detail-row"><span class="vcv-detail-label">Serial</span><span class="vcv-detail-value vcv-mono">{ca.serialNumber}</span></div>
        <div class="vcv-detail-row"><span class="vcv-detail-label">SHA256</span><span class="vcv-detail-value vcv-mono">{ca.fingerprintSHA256}</span></div>
      </div>
      {#if ca.usage?.length}
        <div class="vcv-detail-section">
          <h3 class="vcv-detail-section-title">Usage</h3>
          <div class="vcv-tag-list">
            {#each ca.usage as use}
              <span class="vcv-tag">{use}</span>
            {/each}
          </div>
        </div>
      {/if}
      {#if ca.pem}
        <div class="vcv-detail-section">
          <div class="vcv-detail-section-head">
            <h3 class="vcv-detail-section-title">PEM</h3>
            <button type="button" class="vcv-button vcv-button-small" onclick={copyPem}>Copy</button>
          </div>
          <pre class="vcv-pem">{ca.pem}</pre>
        </div>
      {/if}
    {/if}
  </div>

  {#snippet actions()}
    <button type="button" class="vcv-button vcv-button-secondary" onclick={onClose}>Close</button>
  {/snippet}
</Modal>
