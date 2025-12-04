<script lang="ts">
  import { onMount } from 'svelte';

  const API_BASE_URL: string = 'http://localhost:52000';

  type CertificateStatus = 'valid' | 'expired' | 'revoked';
  type SortKey = 'commonName' | 'createdAt' | 'expiresAt';
  type SortDirection = 'asc' | 'desc';

  type Certificate = {
    id: string;
    commonName: string;
    sans: string[];
    createdAt: string;
    expiresAt: string;
    revoked: boolean;
  };

  let certificates: Certificate[] = [];
  let visibleCertificates: Certificate[] = [];
  let isLoading: boolean = false;
  let loadError: string = '';
  let searchTerm: string = '';
  let statusFilter: 'all' | CertificateStatus = 'all';
  let sortKey: SortKey = 'expiresAt';
  let sortDirection: SortDirection = 'asc';
  let showRevokeModal: boolean = false;
  let revokeToken: string = '';
  let revokeInProgress: boolean = false;
  let revokeError: string = '';
  let selectedCertificate: Certificate | null = null;

  const getStatus = (certificate: Certificate): CertificateStatus => {
    const now: number = Date.now();
    const expiresAtTime: number = new Date(certificate.expiresAt).getTime();
    if (certificate.revoked) {
      return 'revoked';
    }
    if (expiresAtTime <= now) {
      return 'expired';
    }
    return 'valid';
  };

  const loadCertificates = async (): Promise<void> => {
    isLoading = true;
    loadError = '';
    try {
      const response: Response = await fetch(`${API_BASE_URL}/api/certs`);
      if (!response.ok) {
        loadError = `Failed to load certificates (${response.status})`;
        return;
      }
      const data: Certificate[] = await response.json();
      certificates = data;
    } catch {
      loadError = 'Network error while loading certificates';
    } finally {
      isLoading = false;
      visibleCertificates = applyFilters(certificates);
    }
  };

  const sortCertificates = (items: Certificate[]): Certificate[] => {
    const sorted: Certificate[] = [...items];
    sorted.sort((left: Certificate, right: Certificate): number => {
      let leftValue: string = '';
      let rightValue: string = '';
      if (sortKey === 'commonName') {
        leftValue = left.commonName.toLowerCase();
        rightValue = right.commonName.toLowerCase();
      } else if (sortKey === 'createdAt') {
        leftValue = left.createdAt;
        rightValue = right.createdAt;
      } else {
        leftValue = left.expiresAt;
        rightValue = right.expiresAt;
      }
      if (leftValue < rightValue) {
        return sortDirection === 'asc' ? -1 : 1;
      }
      if (leftValue > rightValue) {
        return sortDirection === 'asc' ? 1 : -1;
      }
      return 0;
    });
    return sorted;
  };

  const applyFilters = (items: Certificate[]): Certificate[] => {
    const loweredTerm: string = searchTerm.trim().toLowerCase();
    const filtered: Certificate[] = items.filter((certificate: Certificate): boolean => {
      const status: CertificateStatus = getStatus(certificate);
      if (statusFilter !== 'all' && status !== statusFilter) {
        return false;
      }
      if (loweredTerm === '') {
        return true;
      }
      const sanJoined: string = certificate.sans.join(' ').toLowerCase();
      if (certificate.commonName.toLowerCase().includes(loweredTerm)) {
        return true;
      }
      return sanJoined.includes(loweredTerm);
    });
    return sortCertificates(filtered);
  };

  const updateVisibleCertificates = (): void => {
    visibleCertificates = applyFilters(certificates);
  };

  const setSort = (key: SortKey): void => {
    if (sortKey === key) {
      sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
    } else {
      sortKey = key;
      sortDirection = 'asc';
    }
    updateVisibleCertificates();
  };

  const handleSearchChange = (value: string): void => {
    searchTerm = value;
    updateVisibleCertificates();
  };

  const handleStatusFilterChange = (value: 'all' | CertificateStatus): void => {
    statusFilter = value;
    updateVisibleCertificates();
  };

  const openRevokeModal = (certificate: Certificate): void => {
    selectedCertificate = certificate;
    revokeToken = '';
    revokeError = '';
    showRevokeModal = true;
  };

  const closeRevokeModal = (): void => {
    showRevokeModal = false;
    revokeToken = '';
    revokeError = '';
    selectedCertificate = null;
  };

  const confirmRevocation = async (): Promise<void> => {
    if (selectedCertificate === null) {
      return;
    }
    if (revokeToken.trim() === '') {
      revokeError = 'Write token is required';
      return;
    }
    revokeInProgress = true;
    revokeError = '';
    try {
      const response: Response = await fetch(`${API_BASE_URL}/api/certs/${encodeURIComponent(selectedCertificate.id)}/revoke`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ writeToken: revokeToken })
      });
      if (response.status === 403) {
        revokeError = 'Revocation is disabled on the server';
        return;
      }
      if (response.status === 400) {
        revokeError = 'Invalid revoke request';
        return;
      }
      if (!response.ok) {
        revokeError = `Failed to revoke certificate (${response.status})`;
        return;
      }
      await loadCertificates();
      closeRevokeModal();
    } catch {
      revokeError = 'Network error while revoking certificate';
    } finally {
      revokeInProgress = false;
    }
  };

  const formatDate = (value: string): string => {
    const date: Date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return value;
    }
    return date.toISOString().slice(0, 19).replace('T', ' ');
  };

  const statusLabel = (certificate: Certificate): string => {
    const status: CertificateStatus = getStatus(certificate);
    if (status === 'valid') {
      return 'Valid';
    }
    if (status === 'expired') {
      return 'Expired';
    }
    return 'Revoked';
  };

  onMount((): void => {
    void loadCertificates();
  });
</script>

<section class="vcv-layout">
  <header class="vcv-header">
    <h2>Certificates</h2>
    <div class="vcv-filters">
      <input
        class="vcv-input"
        type="search"
        placeholder="Search by CN or SAN"
        value={searchTerm}
        on:input={(event) => handleSearchChange((event.currentTarget as HTMLInputElement).value)}
      />
      <select
        class="vcv-select"
        bind:value={statusFilter}
        on:change={(event) => handleStatusFilterChange((event.currentTarget as HTMLSelectElement).value as 'all' | CertificateStatus)}
      >
        <option value="all">All</option>
        <option value="valid">Valid</option>
        <option value="expired">Expired</option>
        <option value="revoked">Revoked</option>
      </select>
    </div>
  </header>

  {#if isLoading}
    <p>Loading certificates…</p>
  {:else if loadError}
    <p class="vcv-error">{loadError}</p>
  {:else if visibleCertificates.length === 0}
    <p>No certificates found.</p>
  {:else}
    <div class="vcv-table-wrapper">
      <table class="vcv-table">
        <thead>
          <tr>
            <th>
              <button type="button" class="vcv-sort" on:click={() => setSort('commonName')}>
                Common name
              </button>
            </th>
            <th>SAN</th>
            <th>
              <button type="button" class="vcv-sort" on:click={() => setSort('createdAt')}>
                Created at
              </button>
            </th>
            <th>
              <button type="button" class="vcv-sort" on:click={() => setSort('expiresAt')}>
                Expires at
              </button>
            </th>
            <th>Status</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each visibleCertificates as certificate}
            <tr>
              <td>{certificate.commonName}</td>
              <td>{certificate.sans.join(', ')}</td>
              <td>{formatDate(certificate.createdAt)}</td>
              <td>{formatDate(certificate.expiresAt)}</td>
              <td>{statusLabel(certificate)}</td>
              <td>
                {#if getStatus(certificate) !== 'revoked'}
                  <button type="button" class="vcv-button" on:click={() => openRevokeModal(certificate)}>
                    Revoke
                  </button>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}

  {#if showRevokeModal && selectedCertificate}
    <div class="vcv-modal-backdrop">
      <div class="vcv-modal">
        <h3>Confirm revocation</h3>
        <p>
          You are about to revoke certificate
          <strong>{selectedCertificate.commonName}</strong>.
        </p>
        <p>SAN: {selectedCertificate.sans.join(', ')}</p>
        <p>
          Expires at:
          {formatDate(selectedCertificate.expiresAt)}
        </p>
        <label class="vcv-label">
          Vault write token
          <input
            class="vcv-input"
            type="password"
            bind:value={revokeToken}
            autocomplete="off"
          />
        </label>
        {#if revokeError}
          <p class="vcv-error">{revokeError}</p>
        {/if}
        <div class="vcv-modal-actions">
          <button type="button" class="vcv-button" on:click={closeRevokeModal} disabled={revokeInProgress}>
            Cancel
          </button>
          <button type="button" class="vcv-button vcv-button-danger" on:click={() => void confirmRevocation()} disabled={revokeInProgress}>
            {#if revokeInProgress}
              Revoking…
            {:else}
              Confirm revocation
            {/if}
          </button>
        </div>
      </div>
    </div>
  {/if}
</section>

<style>
  .vcv-layout {
    max-width: 960px;
    margin: 0 auto;
    padding: 1.5rem;
    font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  }

  .vcv-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .vcv-filters {
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .vcv-input,
  .vcv-select {
    padding: 0.4rem 0.6rem;
    border-radius: 0.25rem;
    border: 1px solid #cbd5e1;
    font-size: 0.9rem;
  }

  .vcv-table-wrapper {
    overflow-x: auto;
  }

  .vcv-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.9rem;
  }

  .vcv-table th,
  .vcv-table td {
    padding: 0.5rem 0.75rem;
    border-bottom: 1px solid #e2e8f0;
    text-align: left;
    vertical-align: top;
  }

  .vcv-table th {
    background-color: #f8fafc;
  }

  .vcv-sort {
    background: none;
    border: none;
    padding: 0;
    margin: 0;
    font: inherit;
    cursor: pointer;
    color: #0f172a;
  }

  .vcv-button {
    padding: 0.35rem 0.7rem;
    border-radius: 0.25rem;
    border: 1px solid #0f172a;
    background-color: #0f172a;
    color: white;
    font-size: 0.8rem;
    cursor: pointer;
  }

  .vcv-button:disabled {
    opacity: 0.7;
    cursor: default;
  }

  .vcv-button-danger {
    border-color: #b91c1c;
    background-color: #b91c1c;
  }

  .vcv-error {
    color: #b91c1c;
  }

  .vcv-modal-backdrop {
    position: fixed;
    inset: 0;
    background-color: rgba(15, 23, 42, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 10;
  }

  .vcv-modal {
    background-color: white;
    padding: 1.25rem;
    border-radius: 0.5rem;
    max-width: 480px;
    width: 100%;
    box-shadow: 0 10px 25px rgba(15, 23, 42, 0.25);
  }

  .vcv-label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    margin-top: 0.75rem;
  }

  .vcv-modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    margin-top: 1rem;
  }
</style>
