const MOUNTS_ALL_SENTINEL = "__all__";

const state = {
  availableMounts: [],
  hasSyncedInitialUrl: false,
  messages: {},
  selectedMounts: [],
  suppressUrlUpdateUntilNextSuccess: false,
  vaultConnected: null,
  vaultMountGroups: [],
};

function buildVaultMountKey(vaultId, mount) {
  return `${vaultId}|${mount}`;
}

function formatMountGroupTitle(group) {
  if (group && typeof group.displayName === "string" && group.displayName !== "") {
    return group.displayName;
  }
  if (group && typeof group.id === "string" && group.id !== "") {
    return group.id;
  }
  return "Vault";
}

function shouldSuppressErrorToast(detail) {
  const xhr = detail && detail.xhr;
  if (!xhr) {
    return false;
  }
  const isAbort = xhr.status === 0 && (xhr.statusText === "abort" || xhr.statusText === "");
  return isAbort;
}

function buildCertsPageUrl() {
  const values = getCertsHtmxValues();
  const langSelect = document.getElementById("vcv-lang-select");
  const params = new URLSearchParams();
  if (langSelect && typeof langSelect.value === "string" && langSelect.value !== "") {
    params.set("lang", langSelect.value);
  }
  params.set("expiry", values.expiry);
  params.set("mounts", values.mounts);
  params.set("page", values.page);
  params.set("pageSize", values.pageSize);
  params.set("search", values.search);
  params.set("sortDir", values.sortDir);
  params.set("sortKey", values.sortKey);
  params.set("status", values.status);
  return `/?${params.toString()}`;
}

function applyCertsStateFromUrl() {
  const params = new URLSearchParams(window.location.search || "");
  
  // Sync language select if present in URL
  const langParam = params.get("lang");
  const langSelect = document.getElementById("vcv-lang-select");
  if (langParam && langSelect && langSelect.value !== langParam) {
    console.log(`[VCV] URL lang param detected: ${langParam}. Updating select.`);
    langSelect.value = langParam;
  }

  const searchInput = document.getElementById("vcv-search");
  if (searchInput && params.has("search")) {
    searchInput.value = params.get("search") || "";
  }
  const statusSelect = document.getElementById("vcv-status-filter");
  if (statusSelect && params.has("status")) {
    statusSelect.value = params.get("status") || "all";
  }
  const expirySelect = document.getElementById("vcv-expiry-filter");
  if (expirySelect && params.has("expiry")) {
    expirySelect.value = params.get("expiry") || "all";
  }
  const pageSizeSelect = document.getElementById("vcv-page-size");
  if (pageSizeSelect && params.has("pageSize")) {
    pageSizeSelect.value = params.get("pageSize") || "25";
  }
  const pageInput = document.getElementById("vcv-page");
  if (pageInput && params.has("page")) {
    pageInput.value = params.get("page") || "0";
  }
  const sortKeyInput = document.getElementById("vcv-sort-key");
  if (sortKeyInput && params.has("sortKey")) {
    sortKeyInput.value = params.get("sortKey") || "commonName";
  }
  const sortDirInput = document.getElementById("vcv-sort-dir");
  if (sortDirInput && params.has("sortDir")) {
    sortDirInput.value = params.get("sortDir") || "asc";
  }
  const mountsValue = params.get("mounts");
  if (typeof mountsValue === "string") {
    if (mountsValue === MOUNTS_ALL_SENTINEL) {
      state.selectedMounts = [...state.availableMounts];
    } else if (mountsValue === "") {
      state.selectedMounts = [];
    } else {
      const requested = mountsValue.split(",").map((value) => value.trim()).filter((value) => value !== "");
      state.selectedMounts = requested.filter((value) => state.availableMounts.includes(value));
    }
  }
}


// HTMX Error Handler with translation support
function initHtmxErrorHandler() {
  const handleErrorEvent = function(evt, kind) {
    const detail = evt.detail;
    if (shouldSuppressErrorToast(detail)) {
      return;
    }
    const xhr = detail.xhr;
    const status = xhr ? xhr.status : 0;
    const statusText = xhr ? xhr.statusText : kind;
    const url = xhr ? xhr.responseURL : "";
    console.error('HTMX request failed:', {status, statusText, url});
    const messages = state.messages;
    let errorMessage = messages.errorGeneric || "Request failed";
    if (status === 404) {
      errorMessage = messages.errorNotFound || "Resource not found";
    } else if (status >= 500) {
      errorMessage = messages.errorServer || "Server error occurred";
    } else if (status === 0) {
      errorMessage = messages.errorNetwork || "Network error";
    }
    showErrorToast(errorMessage);
  };

  document.body.addEventListener('htmx:responseError', function(evt) {
    handleErrorEvent(evt, "responseError");
  });

  document.body.addEventListener('htmx:sendError', function(evt) {
    handleErrorEvent(evt, "sendError");
  });

  document.body.addEventListener('htmx:timeout', function(evt) {
    handleErrorEvent(evt, "timeout");
  });
  
}

function initModalHandlers() {
  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape") {
      closeCertificateModal();
      closeMountModal();
      closeDocumentationModal();
      closeVaultStatusModal();
    }
  });

  const backdrops = document.querySelectorAll(".vcv-modal-backdrop");
  backdrops.forEach((backdrop) => {
    backdrop.addEventListener("click", (event) => {
      if (event.target === backdrop) {
        closeCertificateModal();
        closeMountModal();
        closeDocumentationModal();
        closeVaultStatusModal();
      }
    });
  });
}

function initUrlSync() {
  if (!window.htmx) {
    return;
  }
  document.body.addEventListener('htmx:afterRequest', function(evt) {
    const detail = evt.detail;
    if (!detail || !detail.successful) {
      return;
    }
    if (detail.path !== '/ui/certs') {
      return;
    }
    const target = detail.target;
    if (!target || target.id !== 'vcv-certs-body') {
      return;
    }
    if (state.suppressUrlUpdateUntilNextSuccess) {
      state.suppressUrlUpdateUntilNextSuccess = false;
      return;
    }
    const url = buildCertsPageUrl();
    const triggeringEvent = (detail.requestConfig && detail.requestConfig.triggeringEvent) || detail.triggeringEvent;
    const isInputEvent = triggeringEvent && triggeringEvent.type === 'input';
    if (!state.hasSyncedInitialUrl) {
      state.hasSyncedInitialUrl = true;
      window.history.replaceState({}, '', url);
      return;
    }
    if (isInputEvent) {
      window.history.replaceState({}, '', url);
      return;
    }
    window.history.pushState({}, '', url);
  });

  window.addEventListener('popstate', async function() {
    console.log("[VCV] popstate detected. Syncing UI state.");
    state.suppressUrlUpdateUntilNextSuccess = true;
    applyCertsStateFromUrl();
    await loadMessages();
    applyTranslations();
    renderMountSelector();
    setMountsHiddenField();
    refreshHtmxCertsTable();
    
    // If documentation modal is open, reload it
    const docModal = document.getElementById("vcv-documentation-modal");
    if (docModal && !docModal.classList.contains("vcv-hidden")) {
      loadDocumentation();
    }
  });
}

function initVaultConnectionNotifications() {
	document.body.addEventListener('htmx:afterSwap', function(evt) {
		const detail = evt.detail;
		const requestConfig = detail && detail.requestConfig;
		const requestPath = (requestConfig && typeof requestConfig.path === 'string') ? requestConfig.path : '';
		if (requestPath === '' || !requestPath.startsWith('/ui/status')) {
			return;
		}
		setTimeout(() => {
			const container = document.getElementById('vcv-footer-vaults');
			if (!container) {
				return;
			}
			const summaryPill = container.querySelector('.vcv-footer-pill-summary');
			if (summaryPill) {
				const isOk = summaryPill.classList.contains('vcv-footer-pill-ok');
				const isError = summaryPill.classList.contains('vcv-footer-pill-error');
				if (!isOk && !isError) {
					return;
				}
				const nextState = isOk;
				if (state.vaultConnected === null) {
					state.vaultConnected = nextState;
					return;
				}
				if (state.vaultConnected === nextState) {
					return;
				}
				state.vaultConnected = nextState;
				const messages = state.messages || {};
				if (nextState) {
					const restored = messages.vaultConnectionRestored || "Vault connection restored";
					showSuccessToast(restored);
					return;
				}
				const lost = messages.vaultConnectionLost || "Vault connection lost";
				showErrorToast(lost);
				return;
			}
			const connectedCount = container.querySelectorAll('.vcv-footer-pill-ok').length;
			const disconnectedCount = container.querySelectorAll('.vcv-footer-pill-error').length;
			if (connectedCount === 0 && disconnectedCount === 0) {
				return;
			}
			const nextState = disconnectedCount === 0;
			if (state.vaultConnected === null) {
				state.vaultConnected = nextState;
				return;
			}
			if (state.vaultConnected === nextState) {
				return;
			}
			state.vaultConnected = nextState;
			const messages = state.messages || {};
			if (nextState) {
				const restored = messages.vaultConnectionRestored || "Vault connection restored";
				showSuccessToast(restored);
				return;
			}
			const lost = messages.vaultConnectionLost || "Vault connection lost";
			showErrorToast(lost);
		}, 0);
	});
}

// Client-side validation
function initClientValidation() {
  document.body.addEventListener('htmx:beforeRequest', function(evt) {
    const detail = evt.detail;
    if (!detail) {
      return;
    }
    const params = detail.parameters || {};
    
    // Validate search input
    if (typeof params.search === 'string' && params.search.length > 0) {
      if (params.search.length < 2) {
        evt.preventDefault();
        const messages = state.messages || {};
        const validationError = messages.errorSearchTooShort || "Search term must be at least 2 characters";
        showErrorToast(validationError);
        return;
      }
      
      // Check for potentially dangerous patterns
      const dangerousPatterns = /[<>\"'&]/;
      if (dangerousPatterns.test(params.search)) {
        evt.preventDefault();
        const messages = state.messages || {};
        const validationError = messages.errorInvalidChars || "Search contains invalid characters";
        showErrorToast(validationError);
        return;
      }
    }
    
    // Validate page size
    if (params.pageSize) {
      const validSizes = ['25', '50', '100', 'all'];
      if (!validSizes.includes(params.pageSize)) {
        evt.preventDefault();
        const messages = state.messages || {};
        const validationError = messages.errorInvalidPageSize || "Invalid page size";
        showErrorToast(validationError);
        return;
      }
    }
    
    // Validate date range for expiry filter
    if (params.expiry && params.expiry !== 'all') {
      const days = parseInt(params.expiry, 10);
      if (isNaN(days) || days < 1 || days > 365) {
        evt.preventDefault();
        const messages = state.messages || {};
        const validationError = messages.errorInvalidExpiry || "Expiry days must be between 1 and 365";
        showErrorToast(validationError);
        return;
      }
    }
  });
}

// Toast notification system
function showErrorToast(message) {
  showToast(message, 'error', 5000);
}

function showInfoToast(message) {
  showToast(message, 'info', 3000);
}

function showSuccessToast(message) {
  showToast(message, 'success', 3000);
}

function showToast(message, type = 'info', duration = 5000) {
  const toastContainer = document.getElementById('toast-container');
  if (!toastContainer) return;
  
  const toast = document.createElement('div');
  toast.className = `vcv-toast vcv-toast-${type}`;
  toast.innerHTML = `
    <span>${message}</span>
    <button class="vcv-toast-close" onclick="this.parentElement.remove()">×</button>
  `;
  
  toastContainer.appendChild(toast);
  
  // Auto-remove after duration
  if (duration > 0) {
    setTimeout(() => {
      if (toast.parentElement) {
        toast.remove();
      }
    }, duration);
  }
}

function getCurrentLanguage() {
  const params = new URLSearchParams(window.location.search || "");
  const langParam = params.get("lang");
  if (langParam) {
    console.log(`[VCV] Detected language from URL: ${langParam}`);
    return langParam;
  }
  const langSelect = document.getElementById("vcv-lang-select");
  if (langSelect && langSelect.value) {
    console.log(`[VCV] Detected language from select: ${langSelect.value}`);
    return langSelect.value;
  }
  const htmlLang = document.documentElement.lang;
  if (htmlLang) {
    console.log(`[VCV] Detected language from html tag: ${htmlLang}`);
    return htmlLang;
  }
  return "en";
}

async function loadMessages() {
  const lang = getCurrentLanguage();
  console.log(`[VCV] Loading messages for: ${lang}`);
  try {
    const url = `/api/i18n?lang=${encodeURIComponent(lang)}`;
    const response = await fetch(url);
    if (!response.ok) {
      console.error(`[VCV] Failed to load messages: ${response.status}`);
      return;
    }
    const payload = await response.json();
    if (!payload || !payload.messages) {
      return;
    }
    state.messages = payload.messages;
    window.vcvMessages = payload.messages;
    console.log(`[VCV] Messages loaded. Server reported language: ${payload.language}`);
    
    // Sync language select with what the server actually returned
    const langSelect = document.getElementById("vcv-lang-select");
    if (langSelect && payload.language && langSelect.value !== payload.language) {
      console.log(`[VCV] Syncing language select: ${langSelect.value} -> ${payload.language}`);
      langSelect.value = payload.language;
    }
    
    // Update html lang attribute if it differs
    if (payload.language && document.documentElement.lang !== payload.language) {
      document.documentElement.lang = payload.language;
    }
  } catch (err) {
    console.error("[VCV] Error loading messages:", err);
  }
}

function setText(element, value) {
  if (!element || typeof value !== "string" || value === "") {
    return;
  }
  element.textContent = value;
}

function applyTranslations() {
  const messages = state.messages;
  if (!messages) {
    return;
  }
  setText(document.getElementById("certificate-modal-close"), messages.buttonClose);
  setText(document.getElementById("chart-expiry-title"), messages.chartExpiryTimeline);
  setText(document.getElementById("chart-status-title"), messages.chartStatusDistribution);
  setText(document.getElementById("dashboard-expired-label"), messages.dashboardExpired);
  setText(document.getElementById("dashboard-expiring-label"), messages.dashboardExpiring);
  setText(document.getElementById("dashboard-total-label"), messages.dashboardTotal);
  setText(document.getElementById("dashboard-valid-label"), messages.dashboardValid);
  setText(document.getElementById("mount-close"), messages.buttonClose);
  setText(document.getElementById("mount-deselect-all"), messages.deselectAll);
  setText(document.getElementById("mount-modal-title"), messages.mountSelectorTitle);
  setText(document.getElementById("mount-select-all"), messages.selectAll);
  setText(document.getElementById("mount-stats-selected-label"), messages.mountStatsSelected);
  setText(document.getElementById("mount-stats-total-label"), messages.mountStatsTotal);
  setText(document.getElementById("vcv-page-size-label"), messages.paginationPageSizeLabel);
  setText(document.getElementById("vcv-status-filter-label"), messages.statusFilterTitle);
  const searchInput = document.getElementById("vcv-search");
  if (searchInput && typeof messages.searchPlaceholder === "string" && messages.searchPlaceholder !== "") {
    searchInput.setAttribute("placeholder", messages.searchPlaceholder);
  }
  const mountSearchInput = document.getElementById("mount-search");
  if (mountSearchInput && typeof messages.mountSearchPlaceholder === "string" && messages.mountSearchPlaceholder !== "") {
    mountSearchInput.setAttribute("placeholder", messages.mountSearchPlaceholder);
  }
  const statusSelect = document.getElementById("vcv-status-filter");
  if (statusSelect) {
    setText(statusSelect.querySelector("option[value='all']"), messages.statusFilterAll);
    setText(statusSelect.querySelector("option[value='expired']"), messages.statusFilterExpired);
    setText(statusSelect.querySelector("option[value='revoked']"), messages.statusFilterRevoked);
    setText(statusSelect.querySelector("option[value='valid']"), messages.statusFilterValid);
  }
  const expirySelect = document.getElementById("vcv-expiry-filter");
  if (expirySelect) {
    setText(expirySelect.querySelector("option[value='all']"), messages.expiryFilterAll);
    setText(expirySelect.querySelector("option[value='7']"), messages.expiryFilter7Days);
    setText(expirySelect.querySelector("option[value='30']"), messages.expiryFilter30Days);
    setText(expirySelect.querySelector("option[value='90']"), messages.expiryFilter90Days);
  }
  const pageSizeSelect = document.getElementById("vcv-page-size");
  if (pageSizeSelect) {
    setText(pageSizeSelect.querySelector("option[value='all']"), messages.paginationAll);
  }
  setText(document.getElementById("vcv-page-prev"), messages.paginationPrev);
  setText(document.getElementById("vcv-documentation-modal-close"), messages.buttonClose);
  setText(document.getElementById("vcv-documentation-modal-title"), messages.buttonDocumentation);
  setText(document.getElementById("vault-status-modal-close"), messages.buttonClose);
  setText(document.getElementById("vault-status-modal-title"), messages.vaultStatusTitle || "Vault status");
  
  const refreshBtn = document.getElementById("refresh-btn");
  if (refreshBtn) {
    refreshBtn.setAttribute("title", messages.buttonRefresh);
  }

  const docBtn = document.getElementById("vcv-documentation-btn");
  if (docBtn) {
    docBtn.setAttribute("title", messages.buttonDocumentation);
    docBtn.setAttribute("aria-label", messages.buttonDocumentation);
  }
  const themeToggle = document.getElementById("theme-toggle");
  if (themeToggle) {
    themeToggle.setAttribute("title", messages.buttonToggleTheme || "Toggle theme");
    themeToggle.setAttribute("aria-label", messages.buttonToggleTheme || "Toggle theme");
  }
  const langSelect = document.getElementById("vcv-lang-select");
  if (langSelect) {
    langSelect.setAttribute("aria-label", messages.labelLanguage || "Language");
  }
  const loadingIndicator = document.getElementById("vcv-loading-indicator");
  if (loadingIndicator) {
    setText(loadingIndicator.querySelector(".vcv-loading-text"), messages.labelLoading || "Loading...");
  }
  const legend = document.querySelector(".vcv-legend");
  if (legend) {
    const validBadge = legend.querySelector(".vcv-badge-valid");
    const validItem = validBadge ? validBadge.closest(".vcv-legend-item") : null;
    if (validItem) {
      setText(validItem.querySelector(".vcv-badge-valid"), messages.legendValidTitle);
      setText(validItem.querySelector(".vcv-legend-text"), messages.legendValidText);
    }
    const expiredBadge = legend.querySelector(".vcv-badge-expired");
    const expiredItem = expiredBadge ? expiredBadge.closest(".vcv-legend-item") : null;
    if (expiredItem) {
      setText(expiredItem.querySelector(".vcv-badge-expired"), messages.legendExpiredTitle);
      setText(expiredItem.querySelector(".vcv-legend-text"), messages.legendExpiredText);
    }
    const revokedBadge = legend.querySelector(".vcv-badge-revoked");
    const revokedItem = revokedBadge ? revokedBadge.closest(".vcv-legend-item") : null;
    if (revokedItem) {
      setText(revokedItem.querySelector(".vcv-badge-revoked"), messages.legendRevokedTitle);
      setText(revokedItem.querySelector(".vcv-legend-text"), messages.legendRevokedText);
    }
  }
}

function setMountsHiddenField() {
  const mountsInput = document.getElementById("vcv-mounts");
  if (!mountsInput) {
    return;
  }
  if (state.availableMounts.length === 0) {
    mountsInput.value = MOUNTS_ALL_SENTINEL;
    return;
  }
  if (state.selectedMounts.length === 0) {
    mountsInput.value = "";
    return;
  }
  if (state.selectedMounts.length === state.availableMounts.length) {
    mountsInput.value = MOUNTS_ALL_SENTINEL;
    return;
  }
  mountsInput.value = state.selectedMounts.join(",");
}

function getCertsHtmxValues() {
  const expiryFilter = document.getElementById("vcv-expiry-filter");
  const mountsInput = document.getElementById("vcv-mounts");
  const pageInput = document.getElementById("vcv-page");
  const pageSizeSelect = document.getElementById("vcv-page-size");
  const searchInput = document.getElementById("vcv-search");
  const sortDirInput = document.getElementById("vcv-sort-dir");
  const sortKeyInput = document.getElementById("vcv-sort-key");
  const statusFilter = document.getElementById("vcv-status-filter");
  const langSelect = document.getElementById("vcv-lang-select");
  return {
    expiry: expiryFilter ? expiryFilter.value : "all",
    mounts: mountsInput ? mountsInput.value : "",
    page: pageInput ? pageInput.value : "0",
    pageSize: pageSizeSelect ? pageSizeSelect.value : "25",
    search: searchInput ? searchInput.value : "",
    sortDir: sortDirInput ? sortDirInput.value : "asc",
    sortKey: sortKeyInput ? sortKeyInput.value : "commonName",
    status: statusFilter ? statusFilter.value : "all",
    lang: langSelect ? langSelect.value : "",
  };
}

function handleSortClick(event) {
  const button = event.target && event.target.closest ? event.target.closest(".vcv-sort") : null;
  if (!button) {
    return;
  }
  const sortKey = button.getAttribute("data-sort-key") || "";
  if (sortKey === "") {
    return;
  }
  const activeKeyInput = document.getElementById("vcv-sort-key");
  const activeDirInput = document.getElementById("vcv-sort-dir");
  if (!activeKeyInput || !activeDirInput) {
    return;
  }
  const currentKey = activeKeyInput.value || "commonName";
  const currentDir = activeDirInput.value || "asc";
  if (sortKey === currentKey) {
    activeDirInput.value = currentDir === "asc" ? "desc" : "asc";
  } else {
    activeKeyInput.value = sortKey;
    activeDirInput.value = "asc";
  }
  const pageInput = document.getElementById("vcv-page");
  if (pageInput) {
    pageInput.value = "0";
  }
  refreshHtmxCertsTable();
}

function refreshHtmxCertsTable() {
  const certsBody = document.getElementById("vcv-certs-body");
  if (!certsBody || !window.htmx) {
    return;
  }
  setMountsHiddenField();
  window.htmx.ajax("GET", "/ui/certs", {
    indicator: "#vcv-loading-indicator",
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
  const label = typeof state.messages.mountSelectorTitle === "string" && state.messages.mountSelectorTitle !== "" ? state.messages.mountSelectorTitle : "PKI Engines";
  const tooltip = typeof state.messages.mountSelectorTooltip === "string" && state.messages.mountSelectorTooltip !== "" ? state.messages.mountSelectorTooltip : "Filter certificates by Vault instance and PKI mount";
  container.innerHTML = `
    <button type="button" class="vcv-button vcv-button-ghost vcv-mount-trigger" onclick="openMountModal()" title="${tooltip}">
      <span class="vcv-mount-trigger-label">${label}</span>
    </button>
  `;
}

function updateMountStats() {
  const selectedEl = document.getElementById("mount-stats-selected");
  const totalEl = document.getElementById("mount-stats-total");
  if (selectedEl) {
    selectedEl.textContent = state.selectedMounts.length;
  }
  if (totalEl) {
    totalEl.textContent = state.availableMounts.length;
  }
}

function toggleVaultSection(vaultId) {
  const section = document.querySelector(`.vcv-mount-modal-section[data-vault-section="${vaultId}"]`);
  if (!section) {
    return;
  }
  section.classList.toggle("vcv-collapsed");
}

function renderMountModalList() {
  const listContainer = document.getElementById("mount-modal-list");
  if (!listContainer) {
    return;
  }
  const messages = state.messages || {};
  const deselectAllLabel = typeof messages.deselectAll === "string" && messages.deselectAll !== "" ? messages.deselectAll : "None";
  const groups = Array.isArray(state.vaultMountGroups) ? state.vaultMountGroups : [];
  const selectAllLabel = typeof messages.selectAll === "string" && messages.selectAll !== "" ? messages.selectAll : "All";
  const selectedSet = new Set(state.selectedMounts);
  if (groups.length > 0) {
    const content = groups
      .map((group) => {
        const title = formatMountGroupTitle(group);
        const mounts = Array.isArray(group.mounts) ? group.mounts : [];
        const selectedCount = mounts.filter((m) => selectedSet.has(buildVaultMountKey(group.id, m))).length;
        const totalCount = mounts.length;
        const countBadge = `<span class="vcv-badge vcv-badge-neutral" style="font-size: 0.75rem; padding: 0.125rem 0.5rem;">${selectedCount}/${totalCount}</span>`;
        const options = mounts
          .map((mountName) => {
            const key = buildVaultMountKey(group.id, mountName);
            const checkedAttr = selectedSet.has(key) ? "checked" : "";
            const selectedClass = selectedSet.has(key) ? " vcv-mount-option-selected" : "";
            return `<label class="vcv-mount-option${selectedClass}" data-vault="${group.id}" data-mount="${mountName}"><input type="checkbox" ${checkedAttr} onchange="toggleMount('${key}')" /><span class="vcv-mount-name">${mountName}</span></label>`;
          })
          .join("");
        const headerActions = `<div class="vcv-mount-modal-section-actions"><button type="button" class="vcv-button vcv-button-small vcv-button-secondary" onclick="selectAllVaultMounts('${group.id}', event)">${selectAllLabel}</button><button type="button" class="vcv-button vcv-button-small vcv-button-secondary" onclick="deselectAllVaultMounts('${group.id}', event)">${deselectAllLabel}</button></div>`;
        return `<div class="vcv-mount-modal-section" data-vault-section="${group.id}"><div class="vcv-mount-modal-section-header" onclick="toggleVaultSection('${group.id}')"><div class="vcv-mount-modal-section-title">${title} ${countBadge}</div>${headerActions}</div><div class="vcv-mount-modal-section-options">${options}</div></div>`;
      })
      .join("");
    listContainer.innerHTML = content;
    updateMountStats();
    return;
  }
  const items = state.availableMounts.map((mount) => {
    const isSelected = selectedSet.has(mount);
    const checkedAttr = isSelected ? "checked" : "";
    const selectedClass = isSelected ? "selected" : "";
    return `<label class="vcv-mount-modal-option ${selectedClass}" data-mount="${mount}"><input type="checkbox" ${checkedAttr} onchange="toggleMount('${mount}')" /><span class="vcv-mount-modal-name">${mount}</span></label>`;
  });
  const emptyText = typeof messages.noData === "string" && messages.noData !== "" ? messages.noData : "No data";
  listContainer.innerHTML = items.join("") || `<p class="vcv-empty">${emptyText}</p>`;
  updateMountStats();
}

function filterMountList(searchTerm) {
  const term = searchTerm.toLowerCase().trim();
  const sections = document.querySelectorAll(".vcv-mount-modal-section");
  const options = document.querySelectorAll(".vcv-mount-option, .vcv-mount-modal-option");
  
  if (term === "") {
    sections.forEach((section) => section.classList.remove("vcv-hidden"));
    options.forEach((option) => option.classList.remove("vcv-hidden"));
    return;
  }
  
  sections.forEach((section) => {
    const vaultId = section.getAttribute("data-vault-section") || "";
    const vaultMatches = vaultId.toLowerCase().includes(term);
    const sectionOptions = section.querySelectorAll(".vcv-mount-option");
    let hasVisibleOptions = false;
    
    sectionOptions.forEach((option) => {
      const mountName = option.getAttribute("data-mount") || "";
      const mountMatches = mountName.toLowerCase().includes(term);
      
      if (vaultMatches || mountMatches) {
        option.classList.remove("vcv-hidden");
        hasVisibleOptions = true;
      } else {
        option.classList.add("vcv-hidden");
      }
    });
    
    if (vaultMatches || hasVisibleOptions) {
      section.classList.remove("vcv-hidden");
    } else {
      section.classList.add("vcv-hidden");
    }
  });
  
  options.forEach((option) => {
    if (!option.closest(".vcv-mount-modal-section")) {
      const mountName = option.getAttribute("data-mount") || "";
      if (mountName.toLowerCase().includes(term)) {
        option.classList.remove("vcv-hidden");
      } else {
        option.classList.add("vcv-hidden");
      }
    }
  });
}

function openMountModal() {
  const modal = document.getElementById("mount-modal");
  if (!modal) {
    return;
  }
  const searchInput = document.getElementById("mount-search");
  if (searchInput) {
    searchInput.value = "";
  }
  renderMountModalList();
  updateMountStats();
  modal.classList.remove("vcv-hidden");
  if (searchInput) {
    setTimeout(() => searchInput.focus(), 100);
  }
}

function closeMountModal() {
  const modal = document.getElementById("mount-modal");
  if (!modal) {
    return;
  }
  modal.classList.add("vcv-hidden");
}

function openVaultStatusModal() {
  const modal = document.getElementById("vault-status-modal");
  if (!modal) {
    return;
  }
  modal.classList.remove("vcv-hidden");
}

function closeVaultStatusModal() {
  const modal = document.getElementById("vault-status-modal");
  if (!modal) {
    return;
  }
  modal.classList.add("vcv-hidden");
}

let vaultRefreshLastTime = 0;
const vaultRefreshCooldown = 5000;

function handleVaultRefresh(event) {
  const now = Date.now();
  const button = event.target;
  if (!button) {
    return;
  }
  if (now - vaultRefreshLastTime < vaultRefreshCooldown) {
    event.preventDefault();
    event.stopPropagation();
    const remaining = Math.ceil((vaultRefreshCooldown - (now - vaultRefreshLastTime)) / 1000);
    const messages = state.messages || {};
    const msg = messages.cacheInvalidateFailed || "Please wait";
    showErrorToast(`${msg} (${remaining}s)`);
    return false;
  }
  vaultRefreshLastTime = now;
  button.disabled = true;
  setTimeout(() => {
    button.disabled = false;
  }, vaultRefreshCooldown);
  return true;
}

function toggleMount(mountKey) {
  const index = state.selectedMounts.indexOf(mountKey);
  if (index === -1) {
    state.selectedMounts.push(mountKey);
  } else {
    state.selectedMounts.splice(index, 1);
  }
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

function selectAllVaultMounts(vaultId, event) {
  if (event) {
    event.stopPropagation();
  }
  const groups = Array.isArray(state.vaultMountGroups) ? state.vaultMountGroups : [];
  const group = groups.find((item) => item.id === vaultId);
  if (!group || !Array.isArray(group.mounts)) {
    return;
  }
  const keysToAdd = group.mounts.map((mount) => buildVaultMountKey(vaultId, mount));
  const selectedSet = new Set(state.selectedMounts);
  keysToAdd.forEach((key) => selectedSet.add(key));
  state.selectedMounts = Array.from(selectedSet);
  renderMountSelector();
  renderMountModalList();
  refreshHtmxCertsTable();
}

function deselectAllVaultMounts(vaultId, event) {
  if (event) {
    event.stopPropagation();
  }
  const prefix = `${vaultId}|`;
  state.selectedMounts = state.selectedMounts.filter((key) => !key.startsWith(prefix));
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

let currentDocType = "user";

function openDocumentationModal(type = "user") {
  const modal = document.getElementById("vcv-documentation-modal");
  const title = document.getElementById("vcv-documentation-modal-title");
  if (!modal) {
    return;
  }
  
  currentDocType = type;
  
  if (title) {
    const messages = state.messages || {};
    title.textContent = type === "admin" 
      ? (messages.adminDocsTitle || "Admin documentation - VCV") 
      : (messages.buttonDocumentation || "Documentation");
  }
  
  modal.classList.remove("vcv-hidden");
  loadDocumentation(type);
}

function closeDocumentationModal() {
  const modal = document.getElementById("vcv-documentation-modal");
  if (!modal) {
    return;
  }
  modal.classList.add("vcv-hidden");
}

async function loadDocumentation(type = null) {
  const content = document.getElementById("vcv-documentation-content");
  if (!content) {
    return;
  }
  
  const docType = type || currentDocType || "user";
  
  // Show loading spinner
  content.innerHTML = `
    <div class="vcv-loading-spinner-container">
      <div class="vcv-loading-spinner"></div>
    </div>
  `;

  const lang = getCurrentLanguage() || "en";
  console.log(`[VCV] Loading documentation (${docType}) for language: ${lang}`);
  
  try {
    const endpoint = docType === "admin" ? "/ui/docs/configuration" : "/ui/docs/user-guide";
    const response = await fetch(`${endpoint}?lang=${lang}&_=${Date.now()}`);
    if (!response.ok) {
      content.innerHTML = `<p class="vcv-error">Failed to load documentation (${response.status})</p>`;
      return;
    }
    const html = await response.text();
    content.innerHTML = html;
  } catch (err) {
    console.error("[VCV] Error loading documentation:", err);
    content.innerHTML = `<p class="vcv-error">Error: ${err.message}</p>`;
  }
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
    if (!data) {
      return;
    }
    const vaults = Array.isArray(data.vaults) ? data.vaults : [];
    if (vaults.length > 0) {
      state.vaultMountGroups = vaults
        .map((vault) => {
          const id = (vault && typeof vault.id === "string") ? vault.id : "";
          const displayName = (vault && typeof vault.displayName === "string") ? vault.displayName : "";
          const mounts = Array.isArray(vault && vault.pkiMounts) ? vault.pkiMounts : [];
          const normalizedMounts = mounts.map((mount) => String(mount)).map((mount) => mount.trim()).filter((mount) => mount !== "");
          return { id, displayName, mounts: normalizedMounts };
        })
        .filter((vault) => vault.id !== "" && vault.mounts.length > 0);
      state.availableMounts = state.vaultMountGroups
        .map((vault) => vault.mounts.map((mount) => buildVaultMountKey(vault.id, mount)))
        .reduce((acc, keys) => acc.concat(keys), []);
      state.selectedMounts = [...state.availableMounts];
      applyTranslations();
      return;
    }
    if (!Array.isArray(data.pkiMounts)) {
      return;
    }
    state.availableMounts = data.pkiMounts;
    state.selectedMounts = [...data.pkiMounts];
    applyTranslations();
  } catch {
  }
}

function initEventHandlers() {
  document.querySelectorAll(".vcv-sort").forEach((button) => {
    button.addEventListener("click", handleSortClick);
  });
}

function dismissNotifications() {
  const banner = document.getElementById("vcv-notifications");
  if (banner) {
    banner.classList.add("vcv-hidden");
  }
}

async function main() {
  // Sync language first to ensure all subsequent loads (messages, config, etc.) use correct language
  initLanguageFromURL();
  const messagesPromise = loadMessages();
  applyThemeFromStorage();
  initEventHandlers();

  // Initialize HTMX enhancements
  initHtmxErrorHandler();
  initClientValidation();
  
  const isCertsPage = !!document.getElementById("vcv-certs-body");
  if (isCertsPage) {
    initUrlSync();
    initModalHandlers();
    initVaultConnectionNotifications();
    // Load remaining non-critical startup data
    await messagesPromise;
    applyTranslations();
    await loadConfig();
    applyCertsStateFromUrl();
    renderMountSelector();
    setMountsHiddenField();
  } else {
    // Admin page or other pages
    initModalHandlers();
    await messagesPromise;
    applyTranslations();
  }
}

main();

window.openMountModal = openMountModal;
window.closeMountModal = closeMountModal;
window.toggleMount = toggleMount;
window.selectAllMounts = selectAllMounts;
window.deselectAllMounts = deselectAllMounts;
window.selectAllVaultMounts = selectAllVaultMounts;
window.deselectAllVaultMounts = deselectAllVaultMounts;
window.filterMountList = filterMountList;
window.toggleVaultSection = toggleVaultSection;
window.openCertificateModal = openCertificateModal;
window.closeCertificateModal = closeCertificateModal;
window.openDocumentationModal = openDocumentationModal;
window.closeDocumentationModal = closeDocumentationModal;
window.dismissNotifications = dismissNotifications;
