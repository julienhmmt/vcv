const mountsAllSentinel = "__all__";

const state = {
  availableMounts: [],
  selectedMounts: [],
  messages: {},
  retryCount: new Map(),
  maxRetries: 3,
};

// HTMX Error Handler with translation support
function initHtmxErrorHandler() {
  document.body.addEventListener('htmx:responseError', function(evt) {
    const errorInfo = {
      status: evt.detail.xhr.status,
      statusText: evt.detail.xhr.statusText,
      url: evt.detail.xhr.responseURL,
      errorId: evt.detail.target.id || 'unknown'
    };
    
    console.error('HTMX request failed:', errorInfo);
    
    // Show translated error message
    const messages = state.messages;
    let errorMessage = messages.errorGeneric || "Request failed";
    
    if (evt.detail.xhr.status === 404) {
      errorMessage = messages.errorNotFound || "Resource not found";
    } else if (evt.detail.xhr.status === 500) {
      errorMessage = messages.errorServer || "Server error occurred";
    } else if (evt.detail.xhr.status === 0) {
      errorMessage = messages.errorNetwork || "Network error";
    }
    
    // Show error toast
    showErrorToast(errorMessage);
    
    // Implement retry logic for network errors
    if (evt.detail.xhr.status === 0 || evt.detail.xhr.status >= 500) {
      handleRetry(evt.detail);
    }
  });
  
  document.body.addEventListener('htmx:afterRequest', function(evt) {
    // Reset retry count on success
    if (evt.detail.successful) {
      state.retryCount.delete(evt.detail.target.id);
    }
  });
}

// Retry strategy with exponential backoff
function handleRetry(detail) {
  const targetId = detail.target.id;
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
    // Retry the original request
    if (detail.requestConfig) {
      window.htmx.ajax(detail.requestConfig.verb, detail.requestConfig.path, detail.requestConfig);
    }
  }, delay);
}

// Loading indicators
function initLoadingIndicators() {
  // Show loading spinner
  document.body.addEventListener('htmx:beforeRequest', function(evt) {
    const target = evt.detail.target;
    const loadingIndicator = document.getElementById('vcv-loading-indicator');
    if (loadingIndicator) {
      loadingIndicator.classList.remove('vcv-hidden');
    }
    
    // Show loading state on specific elements
    if (target.id === 'vcv-certs-body') {
      target.classList.add('vcv-loading');
    }
  });
  
  // Hide loading spinner
  document.body.addEventListener('htmx:afterRequest', function(evt) {
    const loadingIndicator = document.getElementById('vcv-loading-indicator');
    if (loadingIndicator) {
      loadingIndicator.classList.add('vcv-hidden');
    }
    
    const target = evt.detail.target;
    if (target.id === 'vcv-certs-body') {
      target.classList.remove('vcv-loading');
    }
  });
  
  // Hide loading on error
  document.body.addEventListener('htmx:responseError', function(evt) {
    const loadingIndicator = document.getElementById('vcv-loading-indicator');
    if (loadingIndicator) {
      loadingIndicator.classList.add('vcv-hidden');
    }
    
    const target = evt.detail.target;
    if (target.id === 'vcv-certs-body') {
      target.classList.remove('vcv-loading');
    }
  });
}

// Client-side validation
function initClientValidation() {
  document.body.addEventListener('htmx:beforeRequest', function(evt) {
    const target = evt.detail.target;
    const params = evt.detail.parameters;
    
    // Validate search input
    if (params.search && params.search.length > 0) {
      if (params.search.length < 2) {
        evt.preventDefault();
        const messages = state.messages;
        const validationError = messages.errorSearchTooShort || "Search term must be at least 2 characters";
        showErrorToast(validationError);
        return;
      }
      
      // Check for potentially dangerous patterns
      const dangerousPatterns = /[<>\"'&]/;
      if (dangerousPatterns.test(params.search)) {
        evt.preventDefault();
        const messages = state.messages;
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
        const messages = state.messages;
        const validationError = messages.errorInvalidPageSize || "Invalid page size";
        showErrorToast(validationError);
        return;
      }
    }
    
    // Validate date range for expiry filter
    if (params.expiry && params.expiry !== 'all') {
      const days = parseInt(params.expiry);
      if (isNaN(days) || days < 1 || days > 365) {
        evt.preventDefault();
        const messages = state.messages;
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
  setText(document.getElementById("mount-modal-title"), messages.mountSelectorTitle);
  setText(document.getElementById("mount-deselect-all"), messages.deselectAll);
  setText(document.getElementById("mount-select-all"), messages.selectAll);
  setText(document.getElementById("mount-close"), messages.buttonClose);
  setText(document.getElementById("dashboard-total-label"), messages.dashboardTotal);
  setText(document.getElementById("dashboard-valid-label"), messages.dashboardValid);
  setText(document.getElementById("dashboard-expiring-label"), messages.dashboardExpiring);
  setText(document.getElementById("dashboard-expired-label"), messages.dashboardExpired);
  setText(document.getElementById("chart-status-title"), messages.chartStatusDistribution);
  setText(document.getElementById("chart-expiry-title"), messages.chartExpiryTimeline);
  setText(document.getElementById("vcv-status-filter-label"), messages.statusFilterTitle);
  setText(document.getElementById("vcv-page-size-label"), messages.paginationPageSizeLabel);
  setText(document.querySelector("#certificate-modal .vcv-modal-title"), messages.modalDetailsTitle);
  setText(document.getElementById("certificate-modal-close"), messages.buttonClose);
  const searchInput = document.getElementById("vcv-search");
  if (searchInput && typeof messages.searchPlaceholder === "string" && messages.searchPlaceholder !== "") {
    searchInput.setAttribute("placeholder", messages.searchPlaceholder);
  }
  const statusSelect = document.getElementById("vcv-status-filter");
  if (statusSelect) {
    setText(statusSelect.querySelector("option[value='all']"), messages.statusFilterAll);
    setText(statusSelect.querySelector("option[value='valid']"), messages.statusFilterValid);
    setText(statusSelect.querySelector("option[value='expired']"), messages.statusFilterExpired);
    setText(statusSelect.querySelector("option[value='revoked']"), messages.statusFilterRevoked);
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
  const emptyText = typeof state.messages.noData === "string" && state.messages.noData !== "" ? state.messages.noData : "No data";
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
  await loadMessages();
  applyTranslations();
  initEventHandlers();
  await loadConfig();
  renderMountSelector();
  setMountsHiddenField();
  
  // Initialize HTMX enhancements
  initHtmxErrorHandler();
  initLoadingIndicators();
  initClientValidation();
  initCacheManagement();
  
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
