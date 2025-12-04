const API_BASE_URL = "";

const state = {
  certificates: [],
  visible: [],
  searchTerm: "",
  statusFilter: "all",
  sortKey: "expiresAt",
  sortDirection: "asc",
  selectedCertificate: null,
  revokeInProgress: false,
};

function getStatus(certificate) {
  const now = Date.now();
  const expiresAtTime = new Date(certificate.expiresAt).getTime();
  if (certificate.revoked) {
    return "revoked";
  }
  if (Number.isFinite(expiresAtTime) && expiresAtTime <= now) {
    return "expired";
  }
  return "valid";
}

function statusLabel(certificate) {
  const status = getStatus(certificate);
  if (status === "valid") {
    return "Valid";
  }
  if (status === "expired") {
    return "Expired";
  }
  return "Revoked";
}

function formatDate(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toISOString().slice(0, 19).replace("T", " ");
}

function setStatusMessage(message) {
  const element = document.getElementById("vcv-status");
  if (!element) {
    return;
  }
  element.textContent = message || "";
}

async function loadCertificates() {
  setStatusMessage("");
  try {
    const response = await fetch(`${API_BASE_URL}/api/certs`);
    if (!response.ok) {
      setStatusMessage(`Failed to load certificates (${response.status})`);
      return;
    }
    const data = await response.json();
    if (!Array.isArray(data)) {
      setStatusMessage("Unexpected response format from server");
      return;
    }
    state.certificates = data;
    applyFiltersAndRender();
  } catch {
    setStatusMessage("Network error while loading certificates");
  }
}

function sortCertificates(items) {
  const sorted = [...items];
  sorted.sort((left, right) => {
    let leftValue = "";
    let rightValue = "";
    if (state.sortKey === "commonName") {
      leftValue = String(left.commonName || "").toLowerCase();
      rightValue = String(right.commonName || "").toLowerCase();
    } else if (state.sortKey === "createdAt") {
      leftValue = String(left.createdAt || "");
      rightValue = String(right.createdAt || "");
    } else {
      leftValue = String(left.expiresAt || "");
      rightValue = String(right.expiresAt || "");
    }
    if (leftValue < rightValue) {
      return state.sortDirection === "asc" ? -1 : 1;
    }
    if (leftValue > rightValue) {
      return state.sortDirection === "asc" ? 1 : -1;
    }
    return 0;
  });
  return sorted;
}

function applyFilters(items) {
  const loweredTerm = state.searchTerm.trim().toLowerCase();
  const filtered = items.filter((certificate) => {
    const status = getStatus(certificate);
    if (state.statusFilter !== "all" && status !== state.statusFilter) {
      return false;
    }
    if (loweredTerm === "") {
      return true;
    }
    const sanJoined = (certificate.sans || []).join(" ").toLowerCase();
    if (String(certificate.commonName || "").toLowerCase().includes(loweredTerm)) {
      return true;
    }
    return sanJoined.includes(loweredTerm);
  });
  return sortCertificates(filtered);
}

function applyFiltersAndRender() {
  state.visible = applyFilters(state.certificates);
  renderTable();
}

function renderTable() {
  const tbody = document.getElementById("vcv-certs-body");
  if (!tbody) {
    return;
  }
  tbody.textContent = "";
  state.visible.forEach((certificate) => {
    const row = document.createElement("tr");

    const cnCell = document.createElement("td");
    cnCell.textContent = certificate.commonName || "";
    row.appendChild(cnCell);

    const sanCell = document.createElement("td");
    sanCell.textContent = (certificate.sans || []).join(", ");
    row.appendChild(sanCell);

    const createdCell = document.createElement("td");
    createdCell.textContent = formatDate(certificate.createdAt || "");
    row.appendChild(createdCell);

    const expiresCell = document.createElement("td");
    expiresCell.textContent = formatDate(certificate.expiresAt || "");
    row.appendChild(expiresCell);

    const statusCell = document.createElement("td");
    const badge = document.createElement("span");
    const status = getStatus(certificate);
    badge.className = `vcv-badge vcv-badge-${status}`;
    badge.textContent = statusLabel(certificate);
    statusCell.appendChild(badge);
    row.appendChild(statusCell);

    const actionsCell = document.createElement("td");
    if (status !== "revoked") {
      const button = document.createElement("button");
      button.type = "button";
      button.className = "vcv-button";
      button.textContent = "Revoke";
      button.addEventListener("click", () => {
        openRevokeModal(certificate);
      });
      actionsCell.appendChild(button);
    }
    row.appendChild(actionsCell);

    tbody.appendChild(row);
  });
}

function handleSearchChange(value) {
  state.searchTerm = value;
  applyFiltersAndRender();
}

function handleStatusFilterChange(value) {
  state.statusFilter = value;
  applyFiltersAndRender();
}

function handleSortClick(key) {
  if (state.sortKey === key) {
    state.sortDirection = state.sortDirection === "asc" ? "desc" : "asc";
  } else {
    state.sortKey = key;
    state.sortDirection = "asc";
  }
  applyFiltersAndRender();
}

function openRevokeModal(certificate) {
  state.selectedCertificate = certificate;
  const modal = document.getElementById("vcv-revoke-modal");
  const summary = document.getElementById("vcv-revoke-summary");
  const tokenInput = document.getElementById("vcv-revoke-token");
  const errorElement = document.getElementById("vcv-revoke-error");
  if (!modal || !summary || !tokenInput || !errorElement) {
    return;
  }
  summary.textContent = `You are about to revoke certificate ${certificate.commonName || ""} (serial: ${
    certificate.id || ""
  }).`;
  tokenInput.value = "";
  errorElement.textContent = "";
  modal.classList.remove("vcv-hidden");
  tokenInput.focus();
}

function closeRevokeModal() {
  const modal = document.getElementById("vcv-revoke-modal");
  const errorElement = document.getElementById("vcv-revoke-error");
  if (!modal || !errorElement) {
    return;
  }
  state.selectedCertificate = null;
  state.revokeInProgress = false;
  errorElement.textContent = "";
  modal.classList.add("vcv-hidden");
}

async function confirmRevocation() {
  if (!state.selectedCertificate || state.revokeInProgress) {
    return;
  }
  const tokenInput = document.getElementById("vcv-revoke-token");
  const errorElement = document.getElementById("vcv-revoke-error");
  if (!tokenInput || !errorElement) {
    return;
  }
  const token = tokenInput.value.trim();
  if (token === "") {
    errorElement.textContent = "Write token is required";
    return;
  }
  state.revokeInProgress = true;
  errorElement.textContent = "";
  try {
    const response = await fetch(
      `${API_BASE_URL}/api/certs/${encodeURIComponent(state.selectedCertificate.id)}/revoke`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ writeToken: token }),
      }
    );
    if (response.status === 403) {
      errorElement.textContent = "Revocation is disabled on the server";
      return;
    }
    if (response.status === 400) {
      errorElement.textContent = "Invalid revoke request";
      return;
    }
    if (!response.ok) {
      errorElement.textContent = `Failed to revoke certificate (${response.status})`;
      return;
    }
    await loadCertificates();
    closeRevokeModal();
  } catch {
    errorElement.textContent = "Network error while revoking certificate";
  } finally {
    state.revokeInProgress = false;
  }
}

function attachEventListeners() {
  const searchInput = document.getElementById("vcv-search");
  const statusSelect = document.getElementById("vcv-status-filter");
  const sortButtons = document.querySelectorAll(".vcv-sort");
  const cancelButton = document.getElementById("vcv-revoke-cancel");
  const confirmButton = document.getElementById("vcv-revoke-confirm");

  if (searchInput) {
    searchInput.addEventListener("input", (event) => {
      const target = event.currentTarget;
      handleSearchChange(target.value || "");
    });
  }

  if (statusSelect) {
    statusSelect.addEventListener("change", (event) => {
      const target = event.currentTarget;
      handleStatusFilterChange(target.value || "all");
    });
  }

  sortButtons.forEach((button) => {
    button.addEventListener("click", () => {
      const key = button.getAttribute("data-sort-key");
      if (key) {
        handleSortClick(key);
      }
    });
  });

  if (cancelButton) {
    cancelButton.addEventListener("click", () => {
      closeRevokeModal();
    });
  }

  if (confirmButton) {
    confirmButton.addEventListener("click", () => {
      confirmRevocation();
    });
  }
}

window.addEventListener("DOMContentLoaded", () => {
  attachEventListeners();
  loadCertificates();
});
