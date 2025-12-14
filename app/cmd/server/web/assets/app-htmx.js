const mountsAllSentinel = "__all__";

const state = {
  availableMounts: [],
  hasSyncedInitialUrl: false,
  lastErrorAtByTargetId: new Map(),
  lastRequestByTargetId: new Map(),
  maxRetries: 3,
  messages: {},
  retryCount: new Map(),
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

function getRequestTargetId(detail) {
  const target = detail && detail.target;
  if (!target || !target.id) {
    return "unknown";
  }
  return target.id;
}

function shouldSuppressErrorToast(detail) {
  const xhr = detail && detail.xhr;
  if (!xhr) {
    return false;
  }
  const isAbort = xhr.status === 0 && (xhr.statusText === "abort" || xhr.statusText === "");
  return isAbort;
}

function isRetryable(detail) {
  const xhr = detail && detail.xhr;
  if (!xhr) {
    return true;
  }
  if (xhr.status === 0) {
    return true;
  }
  return xhr.status >= 500;
}

function setCertsBusy(isBusy) {
  const toolbar = document.querySelector(".vcv-toolbar");
  if (toolbar) {
    if (isBusy) {
      toolbar.setAttribute("aria-busy", "true");
    } else {
      toolbar.removeAttribute("aria-busy");
    }
  }
  const refreshButton = document.getElementById("refresh-btn");
  if (refreshButton) {
    refreshButton.disabled = isBusy;
    if (isBusy) {
      refreshButton.classList.add("vcv-button-loading");
    } else {
      refreshButton.classList.remove("vcv-button-loading");
    }
  }
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
  params.set("pki", values.pki);
  params.set("page", values.page);
  params.set("pageSize", values.pageSize);
  params.set("search", values.search);
  params.set("sortDir", values.sortDir);
  params.set("sortKey", values.sortKey);
  params.set("status", values.status);
  params.set("vault", values.vault);
  return `/?${params.toString()}`;
}

function applyCertsStateFromUrl() {
  const params = new URLSearchParams(window.location.search || "");
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
  const vaultFilter = document.getElementById("vcv-vault-filter");
  if (vaultFilter && params.has("vault")) {
    vaultFilter.value = params.get("vault") || "all";
  }
  const pkiFilter = document.getElementById("vcv-pki-filter");
  if (pkiFilter && params.has("pki")) {
    pkiFilter.value = params.get("pki") || "all";
  }
  const mountsValue = params.get("mounts");
  if (typeof mountsValue === "string") {
    if (mountsValue === mountsAllSentinel) {
      state.selectedMounts = [...state.availableMounts];
    } else if (mountsValue === "") {
      state.selectedMounts = [];
    } else {
      const requested = mountsValue.split(",").map((value) => value.trim()).filter((value) => value !== "");
      state.selectedMounts = requested.filter((value) => state.availableMounts.includes(value));
    }
  }
}

function updateSelectOptions(select, options) {
  if (!select) {
    return;
  }
  const current = typeof select.value === "string" && select.value !== "" ? select.value : "all";
  select.innerHTML = "";
  const allOption = document.createElement("option");
  allOption.value = "all";
  allOption.textContent = "All";
  select.appendChild(allOption);
  options.forEach((value) => {
    const option = document.createElement("option");
    option.value = value;
    option.textContent = value;
    select.appendChild(option);
  });
  select.value = current;
}

function updateVaultPkiFiltersVisibility(vaultOptions, pkiOptions) {
  const vaultGroup = document.getElementById("vcv-vault-filter-group");
  const pkiGroup = document.getElementById("vcv-pki-filter-group");
  if (vaultGroup) {
    vaultGroup.classList.toggle("vcv-hidden", !(Array.isArray(vaultOptions) && vaultOptions.length > 1));
  }
  if (pkiGroup) {
    pkiGroup.classList.toggle("vcv-hidden", !(Array.isArray(pkiOptions) && pkiOptions.length > 1));
  }
}

// HTMX Error Handler with translation support
function initHtmxErrorHandler() {
  document.body.addEventListener('htmx:configRequest', function(evt) {
    const detail = evt.detail;
    const targetId = getRequestTargetId(detail);
    if (targetId !== "unknown") {
      state.lastRequestByTargetId.set(targetId, {
        verb: detail.verb,
        path: detail.path,
        requestConfig: detail.requestConfig,
      });
    }
  });

  const handleErrorEvent = function(evt, kind) {
    const detail = evt.detail;
    const targetId = getRequestTargetId(detail);
    const now = Date.now();
    const lastAt = state.lastErrorAtByTargetId.get(targetId) || 0;
    if (now-lastAt < 200) {
      return;
    }
    state.lastErrorAtByTargetId.set(targetId, now);
    if (shouldSuppressErrorToast(detail)) {
      return;
    }
    const xhr = detail.xhr;
    const status = xhr ? xhr.status : 0;
    const statusText = xhr ? xhr.statusText : kind;
    const url = xhr ? xhr.responseURL : "";
    console.error('HTMX request failed:', {status, statusText, url, targetId});
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
    if (isRetryable(detail)) {
      handleRetry(targetId);
    }
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
  
  document.body.addEventListener('htmx:afterRequest', function(evt) {
    // Reset retry count on success
    if (evt.detail.successful) {
      state.retryCount.delete(evt.detail.target.id);
    }
  });
}

// Retry strategy with exponential backoff
function handleRetry(targetId) {
  const currentRetries = state.retryCount.get(targetId) || 0;
  
  if (currentRetries >= state.maxRetries) {
    state.retryCount.delete(targetId);
    const messages = state.messages;
    const finalError = messages.errorMaxRetries || "Maximum retry attempts reached";
    showErrorToast(finalError);
    return;
  }
  
  const delay = Math.pow(2, currentRetries) * 1000; // 1s, 2s, 4s
  state.retryCount.set(targetId, currentRetries + 1);
  
  const messages = state.messages;
  const retryMessage = messages.errorRetry || `Retrying... (${currentRetries + 1}/${state.maxRetries})`;
  showInfoToast(retryMessage);
  
  setTimeout(() => {
    const lastRequest = state.lastRequestByTargetId.get(targetId);
    if (!lastRequest || !lastRequest.requestConfig) {
      return;
    }
    window.htmx.ajax(lastRequest.verb, lastRequest.path, lastRequest.requestConfig);
  }, delay);
}

// Loading indicators
function initLoadingIndicators() {
  // Show loading spinner
  document.body.addEventListener('htmx:beforeRequest', function(evt) {
    const target = evt.detail.target;
    if (target.id !== 'vcv-certs-body') {
      return;
    }
    const loadingIndicator = document.getElementById('vcv-loading-indicator');
    if (loadingIndicator) {
      loadingIndicator.classList.remove('vcv-hidden');
    }
    target.classList.add('vcv-loading');
    setCertsBusy(true);
  });
  
  // Hide loading spinner
  document.body.addEventListener('htmx:afterRequest', function(evt) {
    const target = evt.detail.target;
    if (target.id !== 'vcv-certs-body') {
      return;
    }
    const loadingIndicator = document.getElementById('vcv-loading-indicator');
    if (loadingIndicator) {
      loadingIndicator.classList.add('vcv-hidden');
    }
    target.classList.remove('vcv-loading');
    setCertsBusy(false);
  });
  
  // Hide loading on error
  document.body.addEventListener('htmx:responseError', function(evt) {
    const target = evt.detail.target;
    if (target.id !== 'vcv-certs-body') {
      return;
    }
    const loadingIndicator = document.getElementById('vcv-loading-indicator');
    if (loadingIndicator) {
      loadingIndicator.classList.add('vcv-hidden');
    }
    target.classList.remove('vcv-loading');
    setCertsBusy(false);
  });

  document.body.addEventListener('htmx:sendError', function(evt) {
    const target = evt.detail.target;
    if (target.id !== 'vcv-certs-body') {
      return;
    }
    const loadingIndicator = document.getElementById('vcv-loading-indicator');
    if (loadingIndicator) {
      loadingIndicator.classList.add('vcv-hidden');
    }
    target.classList.remove('vcv-loading');
    setCertsBusy(false);
  });

  document.body.addEventListener('htmx:timeout', function(evt) {
    const target = evt.detail.target;
    if (target.id !== 'vcv-certs-body') {
      return;
    }
    const loadingIndicator = document.getElementById('vcv-loading-indicator');
    if (loadingIndicator) {
      loadingIndicator.classList.add('vcv-hidden');
    }
    target.classList.remove('vcv-loading');
    setCertsBusy(false);
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

  window.addEventListener('popstate', function() {
    state.suppressUrlUpdateUntilNextSuccess = true;
    applyCertsStateFromUrl();
    renderMountSelector();
    setMountsHiddenField();
    refreshHtmxCertsTable();
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

// Cache management
function initCacheManagement() {
  // Configure HTMX cache behavior
  document.body.addEventListener('htmx:configRequest', function(evt) {
    // Add cache control headers
    const headers = evt.detail.headers;
    
    // Cache GET requests for 5 minutes
    if (evt.detail.verb === 'GET') {
      headers['Cache-Control'] = 'max-age=300';
    }
    
    // Add ETag support
    headers['If-None-Match'] = localStorage.getItem('vcv-etag-' + evt.detail.path) || '';
  });
  
  // Handle cache responses
  document.body.addEventListener('htmx:afterRequest', function(evt) {
    const xhr = evt.detail.xhr;
    const etag = xhr.getResponseHeader('ETag');
    
    if (etag && evt.detail.path) {
      localStorage.setItem('vcv-etag-' + evt.detail.path, etag);
    }
    
    // Handle 304 Not Modified
    if (xhr.status === 304) {
      evt.preventDefault();
      console.log('Using cached version for:', evt.detail.path);
    }
  });
  
  // Clear cache on refresh
  document.addEventListener('htmx:afterRequest', function(evt) {
    if (evt.detail.path === '/ui/certs/refresh') {
      // Clear all cache when explicitly refreshing
      Object.keys(localStorage).forEach(key => {
        if (key.startsWith('vcv-etag-')) {
          localStorage.removeItem(key);
        }
      });
      
      const messages = state.messages;
      const cacheCleared = messages.cacheCleared || "Cache cleared";
      showInfoToast(cacheCleared);
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
  const lang = params.get("lang");
  return lang || "";
}

async function loadMessages() {
  try {
    const lang = getCurrentLanguage();
    const url = lang ? `/api/i18n?lang=${encodeURIComponent(lang)}` : "/api/i18n";
    const response = await fetch(url);
    if (!response.ok) {
      return;
    }
    const payload = await response.json();
    if (!payload || !payload.messages) {
      return;
    }
    state.messages = payload.messages;
  } catch {
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
  setText(document.getElementById("vcv-page-size-label"), messages.paginationPageSizeLabel);
  setText(document.getElementById("vcv-pki-filter-label"), "PKI");
  setText(document.getElementById("vcv-status-filter-label"), messages.statusFilterTitle);
  setText(document.getElementById("vcv-vault-filter-label"), "Vault");
  setText(document.querySelector("#certificate-modal .vcv-modal-title"), messages.modalDetailsTitle);
  const searchInput = document.getElementById("vcv-search");
  if (searchInput && typeof messages.searchPlaceholder === "string" && messages.searchPlaceholder !== "") {
    searchInput.setAttribute("placeholder", messages.searchPlaceholder);
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
  setText(document.getElementById("vcv-page-next"), messages.paginationNext);
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
  const expiryFilter = document.getElementById("vcv-expiry-filter");
  const mountsInput = document.getElementById("vcv-mounts");
  const pageInput = document.getElementById("vcv-page");
  const pageSizeSelect = document.getElementById("vcv-page-size");
  const pkiFilter = document.getElementById("vcv-pki-filter");
  const searchInput = document.getElementById("vcv-search");
  const sortDirInput = document.getElementById("vcv-sort-dir");
  const sortKeyInput = document.getElementById("vcv-sort-key");
  const statusFilter = document.getElementById("vcv-status-filter");
  const vaultFilter = document.getElementById("vcv-vault-filter");
  return {
    expiry: expiryFilter ? expiryFilter.value : "all",
    mounts: mountsInput ? mountsInput.value : "",
    page: pageInput ? pageInput.value : "0",
    pageSize: pageSizeSelect ? pageSizeSelect.value : "25",
    pki: pkiFilter ? pkiFilter.value : "all",
    search: searchInput ? searchInput.value : "",
    sortDir: sortDirInput ? sortDirInput.value : "asc",
    sortKey: sortKeyInput ? sortKeyInput.value : "commonName",
    status: statusFilter ? statusFilter.value : "all",
    vault: vaultFilter ? vaultFilter.value : "all",
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
  const label = typeof state.messages.mountSelectorTitle === "string" && state.messages.mountSelectorTitle !== "" ? state.messages.mountSelectorTitle : "PKI Engines";
  const summary = totalMounts === 0 ? "—" : `${selectedCount}/${totalMounts}`;
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
  const deselectAllLabel = typeof messages.deselectAll === "string" && messages.deselectAll !== "" ? messages.deselectAll : "None";
  const groups = Array.isArray(state.vaultMountGroups) ? state.vaultMountGroups : [];
  const messages = state.messages || {};
  const selectAllLabel = typeof messages.selectAll === "string" && messages.selectAll !== "" ? messages.selectAll : "All";
  const selectedSet = new Set(state.selectedMounts);
  if (groups.length > 0) {
    const content = groups
      .map((group) => {
        const title = formatMountGroupTitle(group);
        const mounts = Array.isArray(group.mounts) ? group.mounts : [];
        const options = mounts
          .map((mountName) => {
            const key = buildVaultMountKey(group.id, mountName);
            const checkedAttr = selectedSet.has(key) ? "checked" : "";
            const selectedClass = selectedSet.has(key) ? " vcv-mount-option-selected" : "";
            return `<label class="vcv-mount-option${selectedClass}"><input type="checkbox" ${checkedAttr} onchange="toggleMount('${key}')" /><span class="vcv-mount-name">${mountName}</span></label>`;
          })
          .join("");
        const headerActions = `<div class="vcv-mount-modal-section-actions"><button type="button" class="vcv-button vcv-button-small vcv-button-secondary" onclick="selectAllVaultMounts('${group.id}')">${selectAllLabel}</button><button type="button" class="vcv-button vcv-button-small vcv-button-secondary" onclick="deselectAllVaultMounts('${group.id}')">${deselectAllLabel}</button></div>`;
        return `<div class="vcv-mount-modal-section"><div class="vcv-mount-modal-section-header"><div class="vcv-mount-modal-section-title">${title}</div>${headerActions}</div><div class="vcv-mount-modal-section-options">${options}</div></div>`;
      })
      .join("");
    listContainer.innerHTML = content;
    return;
  }
  const items = state.availableMounts.map((mount) => {
    const checkedAttr = isSelected ? "checked" : "";
    const isSelected = selectedSet.has(mount);
    const selectedClass = isSelected ? "selected" : "";
    return `<label class="vcv-mount-modal-option ${selectedClass}"><input type="checkbox" ${checkedAttr} onchange="toggleMount('${mount}')"><span class="vcv-mount-modal-name">${mount}</span></label>`;
  });
  const emptyText = typeof messages.noData === "string" && messages.noData !== "" ? messages.noData : "No data";
  listContainer.innerHTML = items.join("") || `<p class="vcv-empty">${emptyText}</p>`;
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

function selectAllVaultMounts(vaultId) {
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

function deselectAllVaultMounts(vaultId) {
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

      const vaultFilter = document.getElementById("vcv-vault-filter");
      const pkiFilter = document.getElementById("vcv-pki-filter");
      const vaultOptions = state.vaultMountGroups.map((group) => group.id);
      const uniquePki = state.vaultMountGroups
        .map((group) => group.mounts)
        .reduce((acc, mounts) => acc.concat(mounts), [])
        .map((value) => String(value).trim())
        .filter((value) => value !== "");
      const pkiOptions = Array.from(new Set(uniquePki)).sort();
      updateSelectOptions(vaultFilter, vaultOptions);
      updateSelectOptions(pkiFilter, pkiOptions);
      updateVaultPkiFiltersVisibility(vaultOptions, pkiOptions);
      applyTranslations();
      return;
    }
    if (!Array.isArray(data.pkiMounts)) {
      return;
    }
    state.availableMounts = data.pkiMounts;
    state.selectedMounts = [...data.pkiMounts];

    const vaultFilter = document.getElementById("vcv-vault-filter");
    const pkiFilter = document.getElementById("vcv-pki-filter");
    updateSelectOptions(vaultFilter, []);
    updateSelectOptions(pkiFilter, state.availableMounts);
    updateVaultPkiFiltersVisibility([], state.availableMounts);
    applyTranslations();
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
  applyThemeFromStorage();
  initLanguageFromURL();
  initEventHandlers();

  // Initialize HTMX enhancements
  initHtmxErrorHandler();
  initLoadingIndicators();
  initClientValidation();
  initCacheManagement();
  initUrlSync();
  initVaultConnectionNotifications();

  // Apply URL state and trigger first table load ASAP.
  applyCertsStateFromUrl();
  refreshHtmxCertsTable();

  // Load non-critical startup data in parallel.
  await Promise.all([loadMessages(), loadConfig()]);
  applyTranslations();
  renderMountSelector();
  setMountsHiddenField();
}

main();

window.openMountModal = openMountModal;
window.closeMountModal = closeMountModal;
window.toggleMount = toggleMount;
window.selectAllMounts = selectAllMounts;
window.deselectAllMounts = deselectAllMounts;
window.selectAllVaultMounts = selectAllVaultMounts;
window.deselectAllVaultMounts = deselectAllVaultMounts;
window.openCertificateModal = openCertificateModal;
window.closeCertificateModal = closeCertificateModal;
window.dismissNotifications = dismissNotifications;
