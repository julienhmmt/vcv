const mountsAllSentinel = "__all__";

const state = {
  availableMounts: [],
  selectedMounts: [],
};

function setMountsHiddenField() {
  const mountsInput = document.getElementById("vcv-mounts");
  if (!mountsInput) {
    return;
  }
  if (state.availableMounts.length === 0) {
    mountsInput.value = mountsAllSentinel;
    return;
  }
  if (state.selectedMounts.length === 0) {
    mountsInput.value = "";
    return;
  }
  if (state.selectedMounts.length === state.availableMounts.length) {
    mountsInput.value = mountsAllSentinel;
    return;
  }
  mountsInput.value = state.selectedMounts.join(",");
}

function getCertsHtmxValues() {
  const searchInput = document.getElementById("vcv-search");
  const statusFilter = document.getElementById("vcv-status-filter");
  const expiryFilter = document.getElementById("vcv-expiry-filter");
  const pageSizeSelect = document.getElementById("vcv-page-size");
  const pageInput = document.getElementById("vcv-page");
  const sortKeyInput = document.getElementById("vcv-sort-key");
  const sortDirInput = document.getElementById("vcv-sort-dir");
  const mountsInput = document.getElementById("vcv-mounts");
  return {
    search: searchInput ? searchInput.value : "",
    status: statusFilter ? statusFilter.value : "all",
    expiry: expiryFilter ? expiryFilter.value : "all",
    pageSize: pageSizeSelect ? pageSizeSelect.value : "25",
    page: pageInput ? pageInput.value : "0",
    sortKey: sortKeyInput ? sortKeyInput.value : "commonName",
    sortDir: sortDirInput ? sortDirInput.value : "asc",
    mounts: mountsInput ? mountsInput.value : "",
  };
}

function refreshHtmxCertsTable() {
  const certsBody = document.getElementById("vcv-certs-body");
  if (!certsBody || !window.htmx) {
    return;
  }
  setMountsHiddenField();
  window.htmx.ajax("GET", "/ui/certs", {
    target: "#vcv-certs-body",
    swap: "innerHTML",
    values: getCertsHtmxValues(),
  });
}

function renderMountSelector() {
  const container = document.getElementById("mount-selector");
  if (!container) {
    return;
  }
  const totalMounts = state.availableMounts.length;
  const selectedCount = state.selectedMounts.length;
  const label = "PKI Engines";
  const summary = totalMounts === 0 ? "—" : selectedCount === totalMounts ? "All" : selectedCount === 0 ? "None" : `${selectedCount}/${totalMounts}`;
  container.innerHTML = `
    <button type="button" class="vcv-button vcv-button-ghost vcv-mount-trigger" onclick="openMountModal()">
      <span class="vcv-mount-trigger-label">${label}</span>
      <span class="vcv-badge vcv-badge-neutral">${summary}</span>
    </button>
  `;
}

function renderMountModalList() {
  const listContainer = document.getElementById("mount-modal-list");
  if (!listContainer) {
    return;
  }
  const items = state.availableMounts.map((mount) => {
    const isSelected = state.selectedMounts.includes(mount);
    const checkedAttr = isSelected ? "checked" : "";
    const selectedClass = isSelected ? "selected" : "";
    return `
      <label class="vcv-mount-modal-option ${selectedClass}">
        <input type="checkbox" ${checkedAttr} onchange="toggleMount('${mount}')">
        <span class="vcv-mount-modal-name">${mount}</span>
      </label>
    `;
  });
  listContainer.innerHTML = items.join("") || `<p class="vcv-empty">No data</p>`;
}

function openMountModal() {
  const modal = document.getElementById("mount-modal");
  if (!modal) {
    return;
  }
  renderMountModalList();
  modal.classList.remove("vcv-hidden");
}

function closeMountModal() {
  const modal = document.getElementById("mount-modal");
  if (!modal) {
    return;
  }
  modal.classList.add("vcv-hidden");
}

function toggleMount(mount) {
  const index = state.selectedMounts.indexOf(mount);
  if (index > -1) {
    state.selectedMounts.splice(index, 1);
  } else {
    state.selectedMounts.push(mount);
  }
  renderMountSelector();
  renderMountModalList();
  refreshHtmxCertsTable();
}

function selectAllMounts() {
  state.selectedMounts = [...state.availableMounts];
  renderMountSelector();
  renderMountModalList();
  refreshHtmxCertsTable();
}

function deselectAllMounts() {
  state.selectedMounts = [];
  renderMountSelector();
  renderMountModalList();
  refreshHtmxCertsTable();
}

function openCertificateModal() {
  const modal = document.getElementById("certificate-modal");
  if (!modal) {
    return;
  }
  modal.classList.remove("vcv-hidden");
}

function closeCertificateModal() {
  const modal = document.getElementById("certificate-modal");
  if (!modal) {
    return;
  }
  modal.classList.add("vcv-hidden");
}

function applyThemeFromStorage() {
  const theme = localStorage.getItem("vcv-theme") || "light";
  document.documentElement.setAttribute("data-theme", theme);
  const themeValue = document.getElementById("vcv-theme-value");
  if (themeValue) {
    themeValue.value = theme;
  }
}

function initLanguageFromURL() {
  const langSelect = document.getElementById("vcv-lang-select");
  if (!langSelect) {
    return;
  }
  const params = new URLSearchParams(window.location.search || "");
  const lang = params.get("lang");
  if (!lang) {
    return;
  }
  langSelect.value = lang;
}

async function loadConfig() {
  try {
    const response = await fetch("/api/config");
    if (!response.ok) {
      return;
    }
    const data = await response.json();
    if (!data || !Array.isArray(data.pkiMounts)) {
      return;
    }
    state.availableMounts = data.pkiMounts;
    state.selectedMounts = [...data.pkiMounts];
  } catch {
  }
}

function initEventHandlers() {
  const mountModal = document.getElementById("mount-modal");
  if (mountModal) {
    mountModal.addEventListener("click", (e) => {
      if (e.target === mountModal) {
        closeMountModal();
      }
    });
  }
  const certModal = document.getElementById("certificate-modal");
  if (certModal) {
    certModal.addEventListener("click", (e) => {
      if (e.target === certModal) {
        closeCertificateModal();
      }
    });
  }
}

function dismissNotifications() {
  const banner = document.getElementById("vcv-notifications");
  if (banner) {
    banner.classList.add("vcv-hidden");
  }
}

async function main() {
  applyThemeFromStorage();
  initLanguageFromURL();
  initEventHandlers();
  await loadConfig();
  renderMountSelector();
  setMountsHiddenField();
  refreshHtmxCertsTable();
}

main();

window.openMountModal = openMountModal;
window.closeMountModal = closeMountModal;
window.toggleMount = toggleMount;
window.selectAllMounts = selectAllMounts;
window.deselectAllMounts = deselectAllMounts;
window.openCertificateModal = openCertificateModal;
window.closeCertificateModal = closeCertificateModal;
window.dismissNotifications = dismissNotifications;
