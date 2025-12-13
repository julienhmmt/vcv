const API_BASE_URL = "";

// Retry utility with exponential backoff
async function fetchWithRetry(url, options = {}, maxRetries = 3, baseDelay = 1000) {
  let lastError;
  
  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      const response = await fetch(url, options);
      
      // Retry on network errors or 5xx server errors
      if (!response.ok && response.status >= 500 && attempt < maxRetries) {
        throw new Error(`Server error: ${response.status}`);
      }
      
      return response;
    } catch (error) {
      lastError = error;
      
      // Don't retry on client errors (4xx) or if this was the last attempt
      if (attempt === maxRetries || error.message.includes('4')) {
        throw error;
      }
      
      // Exponential backoff with jitter
      const delay = baseDelay * Math.pow(2, attempt) + Math.random() * 1000;
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }
  
  throw lastError;
}

const state = {
  certificates: [],
  expiryFilter: "all",
  loading: false,
  loadingDetails: false,
  notificationsDismissed: false,
  searchTerm: "",
  selectedCertificate: null,
  sortDirection: "asc",
  sortKey: "expiresAt",
  pageIndex: 0,
  pageSize: "25",
  selectedMounts: [], // Array of selected mount names
  availableMounts: [], // Array of available mount names from config
  status: {
    version: "—",
    vaultConnected: null,
    vaultError: "",
  },
  statusFilter: "all",
  theme: localStorage.getItem('vcv-theme') || 'light',
  visible: [],
  pageVisible: [],
  expirationThresholds: {
    critical: 7,
    warning: 30,
  },
};

const toastState = {
  toasts: [],
  nextId: 1,
};

// Toast notification system
function showToast(message, type = 'info', duration = 5000) {
  const id = toastState.nextId++;
  const toast = { id, message, type };
  toastState.toasts.push(toast);
  renderToasts();
  
  if (duration > 0) {
    setTimeout(() => hideToast(id), duration);
  }
}

async function loadConfig() {
  try {
    const response = await fetchWithRetry(`${API_BASE_URL}/api/config`);
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    const data = await response.json();
    if (data.expirationThresholds) {
      state.expirationThresholds.critical = data.expirationThresholds.critical || 7;
      state.expirationThresholds.warning = data.expirationThresholds.warning || 30;
    }
    if (data.pkiMounts) {
      state.availableMounts = data.pkiMounts;
      // Initialize selectedMounts with all available mounts
      state.selectedMounts = [...data.pkiMounts];
    }
  } catch (error) {
    // Keep default values if config loading fails
  }
}

async function loadStatus() {
  try {
    const response = await fetchWithRetry(`${API_BASE_URL}/api/status`);
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    const data = await response.json();
    state.status.version = data.version || "—";
    state.status.vaultConnected = Boolean(data.vault_connected);
    state.status.vaultError = data.vault_error || "";
  } catch (error) {
    state.status.version = "—";
    state.status.vaultConnected = false;
    state.status.vaultError = error?.message || "unreachable";
  } finally {
    renderStatusFooter();
  }
}

function renderStatusFooter() {
  const versionEl = document.getElementById("vcv-footer-version");
  const vaultEl = document.getElementById("vcv-footer-vault");
  if (!versionEl || !vaultEl) return;

  versionEl.textContent = formatMessage("footerVersion", `VCV v${state.status.version}`, {
    version: state.status.version,
  });

  if (state.status.vaultConnected === null) {
    vaultEl.textContent = formatMessage("footerVaultLoading", "Vault: …");
    vaultEl.className = "vcv-footer-pill";
    return;
  }

  if (state.status.vaultConnected) {
    vaultEl.textContent = formatMessage("footerVaultConnected", "Vault: connected");
    vaultEl.className = "vcv-footer-pill vcv-footer-pill-ok";
  } else {
    vaultEl.textContent = formatMessage("footerVaultDisconnected", "Vault: disconnected");
    vaultEl.className = "vcv-footer-pill vcv-footer-pill-error";
    if (state.status.vaultError) {
      vaultEl.title = state.status.vaultError;
    }
  }
}

function hideToast(id) {
  const index = toastState.toasts.findIndex(t => t.id === id);
  if (index > -1) {
    toastState.toasts.splice(index, 1);
    renderToasts();
  }
}

function renderToasts() {
  const container = document.getElementById('toast-container');
  if (!container) return;
  
  container.innerHTML = toastState.toasts.map(toast => `
    <div class="vcv-toast vcv-toast-${toast.type}" data-toast-id="${toast.id}">
      <span>${toast.message}</span>
      <button class="vcv-toast-close" onclick="hideToast(${toast.id})">×</button>
    </div>
  `).join('');
}

// Keyboard shortcuts
function initKeyboardShortcuts() {
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      closeModal();
    }
  });
}

function closeModal() {
  const modal = document.getElementById('certificate-modal');
  if (modal) {
    modal.classList.remove('vcv-modal--open');
    state.selectedCertificate = null;
  }
}

// Theme management
function initTheme() {
  applyTheme(state.theme);
  const toggle = document.getElementById('theme-toggle');
  if (toggle) {
    toggle.addEventListener('click', toggleTheme);
  }
}

function toggleTheme() {
  state.theme = state.theme === 'light' ? 'dark' : 'light';
  localStorage.setItem('vcv-theme', state.theme);
  applyTheme(state.theme);
}

function applyTheme(theme) {
  document.documentElement.setAttribute('data-theme', theme);
  const icon = document.getElementById('theme-icon');
  if (icon) {
    icon.textContent = theme === 'dark' ? '☀️' : '🌙';
  }
}

// Expiration notifications
function checkExpirationNotifications() {
  if (state.notificationsDismissed) return;
  
  const now = new Date();
  const criticalCerts = [];
  const warningCerts = [];
  const criticalThreshold = state.expirationThresholds.critical;
  const warningThreshold = state.expirationThresholds.warning;
  
  state.certificates.forEach(cert => {
    const days = calculateDaysUntilExpiry(cert.expiresAt);
    if (days !== null && days > 0) {
      if (days <= criticalThreshold) {
        criticalCerts.push({ name: cert.commonName, days });
      } else if (days <= warningThreshold) {
        warningCerts.push({ name: cert.commonName, days });
      }
    }
  });
  
  const banner = document.getElementById('vcv-notifications');
  const text = document.getElementById('vcv-notifications-text');
  
  if (criticalCerts.length > 0) {
    banner.classList.remove('vcv-hidden');
    banner.classList.add('vcv-notifications-critical');
    text.textContent = formatMessage(
      'notificationCritical',
      `${criticalCerts.length} certificate(s) expiring within ${criticalThreshold} days!`,
      { count: criticalCerts.length, threshold: criticalThreshold }
    );
  } else if (warningCerts.length > 0) {
    banner.classList.remove('vcv-hidden');
    banner.classList.remove('vcv-notifications-critical');
    text.textContent = formatMessage(
      'notificationWarning',
      `${warningCerts.length} certificate(s) expiring within ${warningThreshold} days`,
      { count: warningCerts.length, threshold: warningThreshold }
    );
  } else {
    banner.classList.add('vcv-hidden');
  }
}

function dismissNotifications() {
  state.notificationsDismissed = true;
  const banner = document.getElementById('vcv-notifications');
  if (banner) {
    banner.classList.add('vcv-hidden');
  }
}

// Dashboard
function updateDashboard() {
  const stats = {
    total: state.certificates.length,
    valid: 0,
    expired: 0,
    revoked: 0,
    expiringSoon: 0,
  };
  
  const expiringCerts = [];
  const warningThreshold = state.expirationThresholds.warning;
  
  state.certificates.forEach(cert => {
    const statuses = getStatus(cert);
    
    // Count all applicable statuses for statistics
    if (statuses.includes('valid')) stats.valid++;
    if (statuses.includes('expired')) stats.expired++;
    if (statuses.includes('revoked')) stats.revoked++;
    
    const days = calculateDaysUntilExpiry(cert.expiresAt);
    if (days !== null && days > 0 && days <= warningThreshold) {
      stats.expiringSoon++;
      expiringCerts.push({ name: cert.commonName, days, id: cert.id });
    }
  });
  
  // Update cards
  document.getElementById('dashboard-total').textContent = stats.total;
  document.getElementById('dashboard-valid').textContent = stats.valid;
  document.getElementById('dashboard-expiring').textContent = stats.expiringSoon;
  document.getElementById('dashboard-expired').textContent = stats.expired;
  
  // Render donut chart
  renderDonutChart(stats);
  
  // Render timeline
  renderExpiryTimeline(expiringCerts);
}

function renderDonutChart(stats) {
  const container = document.getElementById('status-chart');
  if (!container) return;
  
  // For donut chart, use mutually exclusive categories (prioritized: revoked > expired > valid)
  const chartStats = {
    valid: 0,
    expired: 0,
    revoked: 0,
  };
  
  state.certificates.forEach(cert => {
    const statuses = getStatus(cert);
    if (statuses.includes('revoked')) {
      chartStats.revoked++;
    } else if (statuses.includes('expired')) {
      chartStats.expired++;
    } else {
      chartStats.valid++;
    }
  });
  
  const total = chartStats.valid + chartStats.expired + chartStats.revoked;
  if (total === 0) {
    container.innerHTML = `<div class="vcv-timeline-empty">${formatMessage("noData", "No data")}</div>`;
    return;
  }
  
  // Calculate circumference (2 * PI * radius)
  const circumference = 2 * Math.PI * 50;
  
  // Calculate dash lengths for each segment (dash + gap = full circumference)
  const validDash = (chartStats.valid / total) * circumference;
  const expiredDash = (chartStats.expired / total) * circumference;
  const revokedDash = (chartStats.revoked / total) * circumference;
  
  // SVG circles start at 3 o'clock, we want to start at 12 o'clock
  // stroke-dashoffset moves the start point counter-clockwise
  // To start at top: offset = circumference / 4 (90 degrees)
  const startOffset = circumference / 4;
  
  // Each segment starts where the previous one ended
  const validStart = startOffset;
  const expiredStart = startOffset - validDash;
  const revokedStart = startOffset - validDash - expiredDash;
  
  const validLabel = formatMessage("chartLegendValid", "Valid");
  const expiredLabel = formatMessage("chartLegendExpired", "Expired");
  const revokedLabel = formatMessage("chartLegendRevoked", "Revoked");
  
  // Count dual-status certificates for note
  const dualStatusCount = state.certificates.filter(cert => {
    const statuses = getStatus(cert);
    return statuses.includes('revoked') && statuses.includes('expired');
  }).length;
  
  container.innerHTML = `
    <div class="vcv-donut">
      <svg viewBox="0 0 120 120" width="140" height="140">
        <!-- Background ring -->
        <circle cx="60" cy="60" r="50" fill="none" stroke="#e2e8f0" stroke-width="16" />
        <!-- Valid segment (green) - starts at top -->
        ${chartStats.valid > 0 ? `<circle cx="60" cy="60" r="50" fill="none" stroke="#22c55e" stroke-width="16"
          stroke-dasharray="${validDash} ${circumference - validDash}" 
          stroke-dashoffset="${validStart}" />` : ''}
        <!-- Expired segment (amber) - starts after valid -->
        ${chartStats.expired > 0 ? `<circle cx="60" cy="60" r="50" fill="none" stroke="#f59e0b" stroke-width="16"
          stroke-dasharray="${expiredDash} ${circumference - expiredDash}" 
          stroke-dashoffset="${expiredStart}" />` : ''}
        <!-- Revoked segment (red) - starts after expired -->
        ${chartStats.revoked > 0 ? `<circle cx="60" cy="60" r="50" fill="none" stroke="#ef4444" stroke-width="16"
          stroke-dasharray="${revokedDash} ${circumference - revokedDash}" 
          stroke-dashoffset="${revokedStart}" />` : ''}
      </svg>
      <div class="vcv-donut-center">
        <div class="vcv-donut-value">${total}</div>
        <div class="vcv-donut-label">TOTAL</div>
      </div>
    </div>
    <div class="vcv-chart-legend">
      <div class="vcv-chart-legend-item">
        <span class="vcv-chart-legend-dot" style="background: #22c55e"></span>
        <span>${validLabel} (${chartStats.valid})</span>
      </div>
      <div class="vcv-chart-legend-item">
        <span class="vcv-chart-legend-dot" style="background: #f59e0b"></span>
        <span>${expiredLabel} (${chartStats.expired})</span>
      </div>
      <div class="vcv-chart-legend-item">
        <span class="vcv-chart-legend-dot" style="background: #ef4444"></span>
        <span>${revokedLabel} (${chartStats.revoked})</span>
      </div>
      ${dualStatusCount > 0 ? `<div class="vcv-chart-note" style="font-size: 0.75rem; color: #6b7280; margin-top: 0.5rem;">${formatMessage("DualStatusNote", `${dualStatusCount} certificate(s) are both expired and revoked`, { count: dualStatusCount })}</div>` : ''}
    </div>
  `;
}

function renderExpiryTimeline(certs) {
  const container = document.getElementById('expiry-timeline');
  if (!container) return;
  
  if (certs.length === 0) {
    container.innerHTML = `<div class="vcv-timeline-empty">${formatMessage("noCertsExpiringSoon", "No certificates expiring soon")}</div>`;
    return;
  }
  
  // Sort by days remaining
  certs.sort((a, b) => a.days - b.days);
  
  const criticalThreshold = state.expirationThresholds.critical;
  const warningThreshold = state.expirationThresholds.warning;
  
  container.innerHTML = certs.slice(0, 10).map(cert => {
    let dotClass = 'vcv-timeline-dot-normal';
    if (cert.days <= criticalThreshold) dotClass = 'vcv-timeline-dot-critical';
    else if (cert.days <= warningThreshold) dotClass = 'vcv-timeline-dot-warning';
    
    return `
      <div class="vcv-timeline-item" onclick="showCertificateDetails('${cert.id}')">
        <div class="vcv-timeline-dot ${dotClass}"></div>
        <div class="vcv-timeline-name">${escapeHtml(cert.name)}</div>
        <div class="vcv-timeline-days">${formatMessage("daysRemainingShort", `${cert.days}d`, { days: cert.days })}</div>
      </div>
    `;
  }).join('');
}

function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// Search highlight
function highlightText(text, searchTerm) {
  if (!searchTerm || searchTerm.length < 2) return escapeHtml(text);
  
  const escaped = escapeHtml(text);
  const regex = new RegExp(`(${escapeRegex(searchTerm)})`, 'gi');
  return escaped.replace(regex, '<mark class="vcv-highlight">$1</mark>');
}

function escapeRegex(string) {
  return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

const i18nState = {
  language: "en",
  messages: null,
};

function interpolate(template, values) {
  let result = template;
  const entries = Object.entries(values || {});
  entries.forEach(([key, value]) => {
    const pattern = new RegExp(`{{\\s*${key}\\s*}}`, "g");
    result = result.replace(pattern, String(value));
  });
  return result;
}

function t(key) {
  const messages = i18nState.messages;
  if (!messages || typeof messages[key] !== "string") {
    return "";
  }
  return messages[key];
}

function formatMessage(key, fallback, values) {
  const template = t(key);
  if (!template) {
    return fallback;
  }
  if (!values) {
    return template;
  }
  return interpolate(template, values);
}

function applyTranslations() {
  const appTitle = t("appTitle");
  if (appTitle) {
    document.title = appTitle;
    const titleElement = document.querySelector(".vcv-title");
    if (titleElement) {
      titleElement.textContent = appTitle;
    }
  }

  const subtitle = t("appSubtitle");
  if (subtitle) {
    const subtitleElement = document.querySelector(".vcv-subtitle");
    if (subtitleElement) {
      subtitleElement.textContent = subtitle;
    }
  }

  const rotateCrlButton = document.getElementById("rotate-crl-btn");
  if (rotateCrlButton) {
    rotateCrlButton.textContent = t("buttonRotateCRL") || rotateCrlButton.textContent;
  }

  const downloadCrlButton = document.getElementById("download-crl-btn");
  if (downloadCrlButton) {
    downloadCrlButton.textContent = t("buttonDownloadCRL") || downloadCrlButton.textContent;
  }

  const dashboardTotalLabel = document.getElementById("dashboard-total-label");
  if (dashboardTotalLabel) {
    dashboardTotalLabel.textContent = t("dashboardTotal") || dashboardTotalLabel.textContent;
  }

  const dashboardValidLabel = document.getElementById("dashboard-valid-label");
  if (dashboardValidLabel) {
    dashboardValidLabel.textContent = t("dashboardValid") || dashboardValidLabel.textContent;
  }

  const dashboardExpiringLabel = document.getElementById("dashboard-expiring-label");
  if (dashboardExpiringLabel) {
    dashboardExpiringLabel.textContent = t("dashboardExpiring") || dashboardExpiringLabel.textContent;
  }

  const dashboardExpiredLabel = document.getElementById("dashboard-expired-label");
  if (dashboardExpiredLabel) {
    dashboardExpiredLabel.textContent = t("dashboardExpired") || dashboardExpiredLabel.textContent;
  }

  const chartStatusTitle = document.getElementById("chart-status-title");
  if (chartStatusTitle) {
    chartStatusTitle.textContent = t("chartStatusDistribution") || chartStatusTitle.textContent;
  }

  const chartExpiryTitle = document.getElementById("chart-expiry-title");
  if (chartExpiryTitle) {
    chartExpiryTitle.textContent = t("chartExpiryTimeline") || chartExpiryTitle.textContent;
  }

  const pageSizeLabel = document.getElementById("vcv-page-size-label");
  if (pageSizeLabel) {
    pageSizeLabel.textContent = t("paginationPageSizeLabel") || pageSizeLabel.textContent;
  }

  const pagePrev = document.getElementById("vcv-page-prev");
  if (pagePrev) {
    pagePrev.textContent = t("paginationPrev") || pagePrev.textContent;
  }

  const pageNext = document.getElementById("vcv-page-next");
  if (pageNext) {
    pageNext.textContent = t("paginationNext") || pageNext.textContent;
  }

  const statusFilterLabel = document.getElementById("vcv-status-filter-label");
  if (statusFilterLabel) {
    statusFilterLabel.textContent = t("statusFilterTitle") || statusFilterLabel.textContent;
  }

  const statusFilterSelect = document.getElementById("vcv-status-filter");
  if (statusFilterSelect) {
    const allOption = statusFilterSelect.querySelector('option[value="all"]');
    const validOption = statusFilterSelect.querySelector('option[value="valid"]');
    const expiredOption = statusFilterSelect.querySelector('option[value="expired"]');
    const revokedOption = statusFilterSelect.querySelector('option[value="revoked"]');
    if (allOption) allOption.textContent = t("statusFilterAll") || allOption.textContent;
    if (validOption) validOption.textContent = t("statusFilterValid") || validOption.textContent;
    if (expiredOption) expiredOption.textContent = t("statusFilterExpired") || expiredOption.textContent;
    if (revokedOption) revokedOption.textContent = t("statusFilterRevoked") || revokedOption.textContent;
  }

  const pageSizeAll = document.querySelector('#vcv-page-size option[value="all"]');
  if (pageSizeAll) {
    pageSizeAll.textContent = t("paginationAll") || pageSizeAll.textContent;
  }

  const detailsTitle = document.getElementById("details-title");
  if (detailsTitle) {
    detailsTitle.textContent = t("modalDetailsTitle") || detailsTitle.textContent;
  }

  const expiryFilterSelect = document.getElementById("vcv-expiry-filter");
  if (expiryFilterSelect) {
    const allOption = expiryFilterSelect.querySelector('option[value="all"]');
    if (allOption) {
      allOption.textContent = t("expiryFilterAll") || allOption.textContent;
    }
    const option7 = expiryFilterSelect.querySelector('option[value="7"]');
    if (option7) {
      option7.textContent = t("expiryFilter7Days") || option7.textContent;
    }
    const option30 = expiryFilterSelect.querySelector('option[value=\"30\"]');
    if (option30) {
      option30.textContent = t("expiryFilter30Days") || option30.textContent;
    }
    const option90 = expiryFilterSelect.querySelector('option[value=\"90\"]');
    if (option90) {
      option90.textContent = t("expiryFilter90Days") || option90.textContent;
    }
  }

  const searchInput = document.getElementById("vcv-search");
  if (searchInput) {
    const placeholder = t("searchPlaceholder");
    if (placeholder) {
      searchInput.placeholder = placeholder;
    }
  }

  const statusSelect = document.getElementById("vcv-status-filter");
  if (statusSelect) {
    const allOption = statusSelect.querySelector('option[value="all"]');
    const validOption = statusSelect.querySelector('option[value="valid"]');
    const expiredOption = statusSelect.querySelector('option[value="expired"]');
    const revokedOption = statusSelect.querySelector('option[value="revoked"]');
    if (allOption) {
      allOption.textContent = t("statusFilterAll") || allOption.textContent;
    }
    if (validOption) {
      validOption.textContent = t("statusFilterValid") || validOption.textContent;
    }
    if (expiredOption) {
      expiredOption.textContent = t("statusFilterExpired") || expiredOption.textContent;
    }
    if (revokedOption) {
      revokedOption.textContent = t("statusFilterRevoked") || revokedOption.textContent;
    }
  }

  const commonNameHeader = document.querySelector('.vcv-sort[data-sort-key="commonName"] .vcv-sort-label');
  if (commonNameHeader) {
    commonNameHeader.textContent = t("columnCommonName") || commonNameHeader.textContent;
  }

  const sanHeader = document.querySelector(".vcv-table thead th:nth-child(2)");
  if (sanHeader) {
    sanHeader.textContent = t("columnSan") || sanHeader.textContent;
  }

  const createdHeader = document.querySelector('.vcv-sort[data-sort-key="createdAt"] .vcv-sort-label');
  if (createdHeader) {
    createdHeader.textContent = t("columnCreatedAt") || createdHeader.textContent;
  }

  const expiresHeader = document.querySelector('.vcv-sort[data-sort-key="expiresAt"] .vcv-sort-label');
  if (expiresHeader) {
    expiresHeader.textContent = t("columnExpiresAt") || expiresHeader.textContent;
  }

  const statusHeader = document.querySelector(".vcv-table thead th:nth-child(5)");
  if (statusHeader) {
    statusHeader.textContent = t("columnStatus") || statusHeader.textContent;
  }

  const actionsHeader = document.querySelector(".vcv-table thead th:nth-child(6)");
  if (actionsHeader) {
    actionsHeader.textContent = t("columnActions") || actionsHeader.textContent;
  }

  const legendItems = document.querySelectorAll(".vcv-legend-item");
  if (legendItems[0]) {
    const badge = legendItems[0].querySelector(".vcv-badge");
    const text = legendItems[0].querySelector(".vcv-legend-text");
    if (badge) {
      badge.textContent = t("legendValidTitle") || badge.textContent;
    }
    if (text) {
      text.textContent = t("legendValidText") || text.textContent;
    }
  }
  if (legendItems[1]) {
    const badge = legendItems[1].querySelector(".vcv-badge");
    const text = legendItems[1].querySelector(".vcv-legend-text");
    if (badge) {
      badge.textContent = t("legendExpiredTitle") || badge.textContent;
    }
    if (text) {
      text.textContent = t("legendExpiredText") || text.textContent;
    }
  }
  if (legendItems[2]) {
    const badge = legendItems[2].querySelector(".vcv-badge");
    const text = legendItems[2].querySelector(".vcv-legend-text");
    if (badge) {
      badge.textContent = t("legendRevokedTitle") || badge.textContent;
    }
    if (text) {
      text.textContent = t("legendRevokedText") || text.textContent;
    }
  }

  const modalTitle = document.querySelector(".vcv-modal-title");
  if (modalTitle) {
    modalTitle.textContent = t("modalConfirmTitle") || modalTitle.textContent;
  }

  const writeTokenLabel = document.getElementById("vcv-revoke-token-label");
  if (writeTokenLabel) {
    writeTokenLabel.textContent = t("modalWriteTokenLabel") || writeTokenLabel.textContent;
  }

  const cancelButton = document.getElementById("vcv-revoke-cancel");
  if (cancelButton) {
    cancelButton.textContent = t("modalCancelButton") || cancelButton.textContent;
  }

  const confirmButton = document.getElementById("vcv-revoke-confirm");
  if (confirmButton) {
    confirmButton.textContent = t("modalConfirmButton") || confirmButton.textContent;
  }

  const langSelect = document.getElementById("vcv-lang-select");
  if (langSelect && i18nState.language) {
    langSelect.value = i18nState.language;
  }
}

async function loadTranslations() {
  try {
    const params = new URLSearchParams(window.location.search || "");
    const lang = params.get("lang");
    const url = lang
      ? `${API_BASE_URL}/api/i18n?lang=${encodeURIComponent(lang)}`
      : `${API_BASE_URL}/api/i18n`;
    const response = await fetch(url);
    if (!response.ok) {
      return;
    }
    const data = await response.json();
    if (!data || !data.messages) {
      return;
    }
    i18nState.language = data.language || "en";
    i18nState.messages = data.messages;
    applyTranslations();
  } catch {
    // Keep built-in English strings if translation loading fails.
  }
}

function getStatus(certificate) {
  const now = Date.now();
  const expiresAtTime = new Date(certificate.expiresAt).getTime();
  const statuses = [];
  
  if (certificate.revoked) {
    statuses.push("revoked");
  }
  if (Number.isFinite(expiresAtTime) && expiresAtTime <= now) {
    statuses.push("expired");
  }
  
  if (statuses.length === 0) {
    statuses.push("valid");
  }
  
  return statuses;
}

function statusLabel(status) {
  if (status === "valid") {
    return formatMessage("statusLabelValid", "Valid");
  }
  if (status === "expired") {
    return formatMessage("statusLabelExpired", "Expired");
  }
  return formatMessage("statusLabelRevoked", "Revoked");
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

function updateSummary() {
  const element = document.getElementById("vcv-summary");
  if (!element) {
    return;
  }
  const total = state.certificates.length;
  const visible = state.visible.length;
  if (total === 0) {
    element.textContent = formatMessage("summaryNoCertificates", "No certificates.");
    return;
  }
  if (visible === total) {
    element.textContent = formatMessage(
      "summaryAll",
      `${total} certificate${total > 1 ? "s" : ""}`,
      { total },
    );
    return;
  }
  element.textContent = formatMessage(
    "summarySome",
    `${visible} of ${total} certificate${total > 1 ? "s" : ""} shown`,
    { visible, total },
  );
}

function updateSortIndicators() {
  const buttons = document.querySelectorAll(".vcv-sort");
  buttons.forEach((button) => {
    const key = button.getAttribute("data-sort-key");
    if (!key) {
      return;
    }
    const isActive = key === state.sortKey;
    button.setAttribute("data-active", isActive ? "true" : "false");
    if (isActive) {
      button.setAttribute("data-direction", state.sortDirection);
    } else {
      button.removeAttribute("data-direction");
    }
  });
}

// Mount selection functions
function toggleMount(mount) {
  const index = state.selectedMounts.indexOf(mount);
  if (index > -1) {
    state.selectedMounts.splice(index, 1);
  } else {
    state.selectedMounts.push(mount);
  }
  renderMountSelector();
  renderMountModalList();
  loadCertificates(); // Reload certificates with new mount filter
}

function selectAllMounts() {
  state.selectedMounts = [...state.availableMounts];
  renderMountSelector();
  renderMountModalList();
  loadCertificates();
}

function deselectAllMounts() {
  state.selectedMounts = [];
  renderMountSelector();
  renderMountModalList();
  loadCertificates();
}

function renderMountSelector() {
  const container = document.getElementById('mount-selector');
  if (!container) return;

  const totalMounts = state.availableMounts.length;
  const selectedCount = state.selectedMounts.length;
  const label = formatMessage("mountSelectorTitle", "PKI Engines");
  const summary = totalMounts === 0
    ? formatMessage("noData", "No data")
    : selectedCount === 0
      ? formatMessage("deselectAll", "Deselect All")
      : selectedCount === totalMounts
        ? formatMessage("selectAll", "Select All")
        : `${selectedCount}/${totalMounts}`;

  container.innerHTML = `
    <button type="button" class="vcv-button vcv-button-ghost vcv-mount-trigger" onclick="openMountModal()">
      <span class="vcv-mount-trigger-label">${label}</span>
      <span class="vcv-badge vcv-badge-neutral">${summary}</span>
    </button>
  `;
}

function renderMountModalList() {
  const listContainer = document.getElementById('mount-modal-list');
  if (!listContainer) return;

  const isAllSelected = state.selectedMounts.length === state.availableMounts.length;
  const isNoneSelected = state.selectedMounts.length === 0;

  const title = document.getElementById('mount-modal-title');
  if (title) {
    title.textContent = formatMessage("mountSelectorTitle", "PKI Engines");
  }

  const closeBtn = document.getElementById("mount-close");
  if (closeBtn) {
    closeBtn.textContent = formatMessage("buttonClose", "Close");
  }

  const actions = document.querySelectorAll('.vcv-modal-actions .vcv-button');
  actions.forEach((btn) => {
    if (btn.id === "mount-select-all") {
      btn.disabled = isAllSelected;
      btn.textContent = formatMessage("selectAll", "Select All");
    }
    if (btn.id === "mount-deselect-all") {
      btn.disabled = isNoneSelected;
      btn.textContent = formatMessage("deselectAll", "Deselect All");
    }
  });

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

  listContainer.innerHTML = items.join("") || `<p class="vcv-empty">${formatMessage("noData", "No data")}</p>`;
}

function openMountModal() {
  const modal = document.getElementById("mount-modal");
  if (!modal) return;
  renderMountModalList();
  modal.classList.remove("vcv-hidden");
}

function closeMountModal() {
  const modal = document.getElementById("mount-modal");
  if (!modal) return;
  modal.classList.add("vcv-hidden");
}

async function loadCertificates() {
  try {
    // Build URL with mount filter if any mounts are selected
    let url = `${API_BASE_URL}/api/certs`;
    if (state.selectedMounts.length > 0 && state.selectedMounts.length < state.availableMounts.length) {
      const mountParams = state.selectedMounts.join(',');
      url += `?mounts=${encodeURIComponent(mountParams)}`;
    }
    
    const response = await fetchWithRetry(url);
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    const data = await response.json();
    if (!Array.isArray(data)) {
      showToast(formatMessage("loadUnexpectedFormat", "Unexpected response format from server"), "error");
      return;
    }
    state.certificates = data;
    showToast(formatMessage("loadSuccess", "Certificates loaded successfully"), "success");
  } catch (error) {
    if (error.message.includes('4')) {
      showToast(formatMessage("loadFailed", "Failed to load certificates ({{status}})", { status: error.message }), "error");
    } else {
      showToast(formatMessage("loadNetworkError", "Network error loading certificates. Please try again."), "error");
    }
    return;
  }
  applyFiltersAndRender();
}

async function loadCertificateDetails(serialNumber) {
  try {
    const response = await fetchWithRetry(`${API_BASE_URL}/api/certs/${serialNumber}/details`);
    if (!response.ok) {
      showToast(
        formatMessage(
          "loadDetailsFailed",
          `Failed to load certificate details (${response.status})`,
          { status: response.status },
        ),
        'error'
      );
      return null;
    }
    const details = await response.json();
    return details;
  } catch (error) {
    showToast(
      formatMessage(
        "loadDetailsNetworkError",
        "Network error loading certificate details. Please try again.",
      ),
      'error'
    );
    return null;
  } finally {
    state.loadingDetails = false;
    updateDetailsLoadingUI();
  }
}

async function downloadCertificatePEM(serialNumber) {
  try {
    const response = await fetchWithRetry(`${API_BASE_URL}/api/certs/${serialNumber}/pem`);
    if (!response.ok) {
      showToast(
        formatMessage(
          "downloadPEMFailed",
          `Failed to download certificate PEM (${response.status})`,
          { status: response.status },
        ),
        'error'
      );
      return;
    }
    const pemData = await response.json();
    
    // Create download
    const blob = new Blob([pemData.pem], { type: 'application/x-pem-file' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `certificate-${serialNumber}.pem`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
    
    showToast(
      formatMessage("downloadPEMSuccess", "Certificate PEM downloaded successfully"),
      'success',
      3000
    );
  } catch (error) {
    showToast(
      formatMessage(
        "downloadPEMNetworkError",
        "Network error downloading certificate PEM. Please try again.",
      ),
      'error'
    );
  }
}

function updateDetailsLoadingUI() {
  const modal = document.getElementById('certificate-modal');
  if (!modal) return;
  
  const content = modal.querySelector('.vcv-modal-content');
  if (state.loadingDetails) {
    content.classList.add('vcv-modal-content--loading');
  } else {
    content.classList.remove('vcv-modal-content--loading');
  }
}

async function invalidateCacheAndRefresh() {
  try {
    await fetchWithRetry(`${API_BASE_URL}/api/cache/invalidate`, { method: 'POST' });
    await loadCertificates();
    showToast(
      formatMessage("cacheInvalidated", "Cache cleared and data refreshed"),
      'success',
      3000
    );
  } catch (error) {
    showToast(
      formatMessage("cacheInvalidateFailed", "Failed to clear cache"),
      'error'
    );
  }
}

async function rotateCRL() {
  try {
    const response = await fetchWithRetry(`${API_BASE_URL}/api/crl/rotate`, { method: 'POST' });
    if (!response.ok) {
      showToast(
        formatMessage(
          "rotateCRLFailed",
          `Failed to rotate CRL (${response.status})`,
          { status: response.status },
        ),
        'error'
      );
      return;
    }
    showToast(
      formatMessage("rotateCRLSuccess", "CRL rotated successfully"),
      'success',
      3000
    );
  } catch {
    showToast(
      formatMessage(
        "rotateCRLNetworkError",
        "Network error rotating CRL. Please try again.",
      ),
      'error'
    );
  }
}

async function downloadCRL() {
  try {
    const response = await fetchWithRetry(`${API_BASE_URL}/api/crl/download`);
    if (!response.ok) {
      showToast(
        formatMessage(
          "downloadCRLFailed",
          `Failed to download CRL (${response.status})`,
          { status: response.status },
        ),
        'error'
      );
      return;
    }
    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'crl.pem';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  } catch {
    showToast(
      formatMessage(
        "downloadCRLNetworkError",
        "Network error downloading CRL. Please try again.",
      ),
      'error'
    );
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
    const statuses = getStatus(certificate);
    if (state.statusFilter !== "all" && !statuses.includes(state.statusFilter)) {
      return false;
    }
    // Expiry filter
    if (state.expiryFilter !== "all") {
      const days = calculateDaysUntilExpiry(certificate.expiresAt);
      const maxDays = parseInt(state.expiryFilter, 10);
      if (days === null || days < 0 || days > maxDays) {
        return false;
      }
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
  state.pageIndex = 0;
  paginateAndRender();
  updateSummary();
  updateSortIndicators();
  updateDashboard();
  checkExpirationNotifications();
}

function paginateAndRender() {
  const pageSizeValue = state.pageSize;
  const info = document.getElementById("vcv-page-info");
  const prevBtn = document.getElementById("vcv-page-prev");
  const nextBtn = document.getElementById("vcv-page-next");
  const countBadge = document.getElementById("vcv-page-count");

  if (pageSizeValue === "all") {
    state.pageVisible = state.visible;
    state.pageIndex = 0;
    renderTableRows(state.pageVisible);
    updateSummary();
    if (info) {
      info.textContent = formatMessage("paginationAll", "All results");
    }
    if (countBadge) {
      countBadge.textContent = `${state.visible.length}`;
      countBadge.classList.toggle("vcv-hidden", state.visible.length === 0);
    }
    if (prevBtn) prevBtn.disabled = true;
    if (nextBtn) nextBtn.disabled = true;
    return;
  }

  const size = parseInt(pageSizeValue, 10) || 25;
  const totalPages = Math.max(1, Math.ceil(state.visible.length / size));
  state.pageIndex = Math.min(state.pageIndex, totalPages - 1);
  const start = state.pageIndex * size;
  const end = start + size;
  state.pageVisible = state.visible.slice(start, end);

  renderTableRows(state.pageVisible);
  updateSummary();

  if (info) {
    info.textContent = formatMessage(
      "paginationInfo",
      `Page ${state.pageIndex + 1} of ${totalPages}`,
      { current: state.pageIndex + 1, total: totalPages },
    );
  }
  if (prevBtn) {
    prevBtn.disabled = state.pageIndex === 0;
  }
  if (nextBtn) {
    nextBtn.disabled = state.pageIndex >= totalPages - 1;
  }
  if (countBadge) {
    countBadge.textContent = `${state.visible.length}`;
    countBadge.classList.toggle("vcv-hidden", state.visible.length === 0);
  }
}

function renderTableRows(items) {
  const tbody = document.getElementById("vcv-certs-body");
  if (!tbody) {
    return;
  }
  tbody.textContent = "";
  const searchTerm = state.searchTerm.trim();
  items.forEach((certificate) => {
    const row = document.createElement("tr");

    const cnCell = document.createElement("td");
    cnCell.innerHTML = highlightText(certificate.commonName || "", searchTerm);
    row.appendChild(cnCell);

    const sanCell = document.createElement("td");
    sanCell.innerHTML = highlightText((certificate.sans || []).join(", "), searchTerm);
    row.appendChild(sanCell);

    const createdCell = document.createElement("td");
    createdCell.textContent = formatDate(certificate.createdAt || "");
    row.appendChild(createdCell);

    const expiresCell = document.createElement("td");
    expiresCell.className = "vcv-expires-cell";
    const daysRemaining = calculateDaysUntilExpiry(certificate.expiresAt);
    const expiresText = formatDate(certificate.expiresAt || "");
    
    const expiresSpan = document.createElement("div");
    expiresSpan.className = "vcv-expires-date";
    expiresSpan.textContent = expiresText;
    expiresCell.appendChild(expiresSpan);
    
    const warningThreshold = state.expirationThresholds.warning;
    const criticalThreshold = state.expirationThresholds.critical;
    if (daysRemaining !== null && daysRemaining <= warningThreshold) {
      const daysSpan = document.createElement("div");
      if (daysRemaining <= criticalThreshold) {
        daysSpan.className = "vcv-days-remaining vcv-days-critical";
      } else {
        daysSpan.className = "vcv-days-remaining vcv-days-warning";
      }
      daysSpan.textContent = daysRemaining <= 1
        ? formatMessage("daysRemainingSingular", "1 day remaining", { days: daysRemaining })
        : formatMessage("daysRemaining", `${daysRemaining} days remaining`, { days: daysRemaining });
      expiresCell.appendChild(daysSpan);
    }
    row.appendChild(expiresCell);

    const statusCell = document.createElement("td");
    const statuses = getStatus(certificate);
    
    statuses.forEach(status => {
      row.classList.add(`vcv-row-${status}`);
    });
    
    statuses.forEach(status => {
      const badge = document.createElement("span");
      badge.className = `vcv-badge vcv-badge-${status}`;
      badge.textContent = statusLabel(status);
      statusCell.appendChild(badge);
    });
    
    row.appendChild(statusCell);

    const actionsCell = document.createElement("td");
    
    const detailsButton = document.createElement("button");
    detailsButton.className = "vcv-button vcv-button-small";
    detailsButton.textContent = formatMessage("buttonDetails", "Details");
    detailsButton.onclick = () => showCertificateDetails(certificate.id);
    actionsCell.appendChild(detailsButton);
    
    const downloadButton = document.createElement("button");
    downloadButton.className = "vcv-button vcv-button-small vcv-button-primary";
    downloadButton.textContent = formatMessage("buttonDownloadPEM", "Download PEM");
    downloadButton.onclick = () => downloadCertificatePEM(certificate.id);
    actionsCell.appendChild(downloadButton);
    
    row.appendChild(actionsCell);

    tbody.appendChild(row);
  });
}

function renderTable() {
  paginateAndRender();
}

function calculateDaysUntilExpiry(expiresAt) {
  if (!expiresAt) return null;
  
  const now = new Date();
  const expiry = new Date(expiresAt);
  const diffTime = expiry - now;
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  
  return diffDays;
}

function buildCertificateDetailsUiUrl(certificateId) {
  const params = new URLSearchParams(window.location.search || "");
  const lang = params.get("lang");
  let url = `${API_BASE_URL}/ui/certs/${encodeURIComponent(certificateId)}/details`;
  if (lang) {
    url += `?lang=${encodeURIComponent(lang)}`;
  }
  return url;
}

async function showCertificateDetails(certificateId) {
  const modal = document.getElementById('certificate-modal');
  const loadingDiv = document.getElementById('details-loading');
  const contentDiv = document.getElementById('details-content');
  const downloadBtn = document.getElementById('download-pem-btn');
  
  // Reset modal
  loadingDiv.classList.remove('vcv-hidden');
  contentDiv.innerHTML = '';
  
  // Update modal title
  const modalTitle = modal.querySelector('.vcv-modal-title');
  modalTitle.textContent = formatMessage('modalDetailsTitle', 'Certificate Details');
  
  // Show modal
  modal.classList.remove('vcv-hidden');
  state.selectedCertificate = { id: certificateId };
  
  downloadBtn.onclick = () => downloadCertificatePEM(certificateId);

  const htmxClient = window.htmx;
  if (htmxClient && typeof htmxClient.ajax === "function") {
    try {
      const url = buildCertificateDetailsUiUrl(certificateId);
      await htmxClient.ajax("GET", url, "#details-content");
      if (!contentDiv.innerHTML.trim()) {
        throw new Error("empty response");
      }
      loadingDiv.classList.add('vcv-hidden');
      return;
    } catch (error) {
      loadingDiv.classList.add('vcv-hidden');
      modal.classList.add('vcv-hidden');
      showToast(formatMessage("loadDetailsNetworkError", "Network error loading certificate details. Please try again."), 'error');
      return;
    }
  }

  const details = await loadCertificateDetails(certificateId);
  if (details) {
    renderCertificateDetails(details);
    loadingDiv.classList.add('vcv-hidden');
    return;
  }
  modal.classList.add('vcv-hidden');
}

function renderCertificateDetails(details) {
  const contentDiv = document.getElementById('details-content');
  
  const detailsHTML = `
    <div class="vcv-detail-section">
      <div class="vcv-detail-section-header">Certificate Information</div>
      <div class="vcv-detail-section-content">
        <div class="vcv-detail-row">
          <div class="vcv-detail-label">${formatMessage('labelSerialNumber', 'Serial Number')}</div>
          <div class="vcv-detail-value"><code>${details.serialNumber}</code></div>
        </div>
        <div class="vcv-detail-row">
          <div class="vcv-detail-label">${formatMessage('labelSubject', 'Subject')}</div>
          <div class="vcv-detail-value">${details.subject}</div>
        </div>
        <div class="vcv-detail-row">
          <div class="vcv-detail-label">${formatMessage('labelIssuer', 'Issuer')}</div>
          <div class="vcv-detail-value">${details.issuer}</div>
        </div>
      </div>
    </div>
    
    <div class="vcv-detail-section">
      <div class="vcv-detail-section-header">Technical Details</div>
      <div class="vcv-detail-section-content">
        <div class="vcv-detail-row">
          <div class="vcv-detail-label">${formatMessage('labelKeyAlgorithm', 'Key Algorithm')}</div>
          <div class="vcv-detail-value">${details.keyAlgorithm}</div>
        </div>
        <div class="vcv-detail-row">
          <div class="vcv-detail-label">${formatMessage('labelFingerprintSHA1', 'SHA-1 Fingerprint')}</div>
          <div class="vcv-detail-value"><code>${details.fingerprintSHA1}</code></div>
        </div>
        <div class="vcv-detail-row">
          <div class="vcv-detail-label">${formatMessage('labelFingerprintSHA256', 'SHA-256 Fingerprint')}</div>
          <div class="vcv-detail-value"><code>${details.fingerprintSHA256}</code></div>
        </div>
        <div class="vcv-detail-row">
          <div class="vcv-detail-label">${formatMessage('labelUsage', 'Usage')}</div>
          <div class="vcv-detail-value">${details.usage.join(', ') || 'N/A'}</div>
        </div>
      </div>
    </div>
    
    <div class="vcv-detail-section">
      <div class="vcv-detail-section-header">${formatMessage('labelPEM', 'PEM Certificate')}</div>
      <div class="vcv-detail-section-content">
        <div class="vcv-detail-row">
          <div class="vcv-detail-value">
            <pre>${details.pem}</pre>
          </div>
        </div>
      </div>
    </div>
  `;
  
  contentDiv.innerHTML = detailsHTML;
}

function handleSearchChange(value) {
  state.searchTerm = value;
  applyFiltersAndRender();
}

function handleStatusFilterChange(value) {
  state.statusFilter = value;
  applyFiltersAndRender();
}

function handleExpiryFilterChange(value) {
  state.expiryFilter = value;
  applyFiltersAndRender();
}

function handlePageSizeChange(value) {
  state.pageSize = value;
  state.pageIndex = 0;
  paginateAndRender();
}

function handlePreviousPage() {
  if (state.pageIndex <= 0) {
    return;
  }
  state.pageIndex -= 1;
  paginateAndRender();
}

function handleNextPage() {
  const size = state.pageSize === "all" ? state.visible.length : parseInt(state.pageSize, 10) || 25;
  const totalPages = state.pageSize === "all" ? 1 : Math.max(1, Math.ceil(state.visible.length / size));
  if (state.pageIndex >= totalPages - 1) {
    return;
  }
  state.pageIndex += 1;
  paginateAndRender();
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

function closeModal() {
  const modal = document.getElementById('certificate-modal');
  if (modal) {
    modal.classList.add('vcv-hidden');
    state.selectedCertificate = null;
  }
}

// Initialize event handlers
function initEventHandlers() {
  // Search input
  const searchInput = document.getElementById("vcv-search");
  if (searchInput) {
    searchInput.addEventListener("input", (e) => handleSearchChange(e.target.value));
  }

  // Status filter
  const statusFilter = document.getElementById("vcv-status-filter");
  if (statusFilter) {
    statusFilter.addEventListener("change", (e) => handleStatusFilterChange(e.target.value));
  }

  // Expiry filter
  const expiryFilter = document.getElementById("vcv-expiry-filter");
  if (expiryFilter) {
    expiryFilter.addEventListener("change", (e) => handleExpiryFilterChange(e.target.value));
  }

  // Page size
  const pageSizeSelect = document.getElementById("vcv-page-size");
  if (pageSizeSelect) {
    pageSizeSelect.addEventListener("change", (e) => handlePageSizeChange(e.target.value));
  }

  // Pagination buttons
  const prevBtn = document.getElementById("vcv-page-prev");
  if (prevBtn) {
    prevBtn.addEventListener("click", handlePreviousPage);
  }
  const nextBtn = document.getElementById("vcv-page-next");
  if (nextBtn) {
    nextBtn.addEventListener("click", handleNextPage);
  }

  // Mount modal backdrop close
  const mountModal = document.getElementById("mount-modal");
  if (mountModal) {
    mountModal.addEventListener("click", (e) => {
      if (e.target === mountModal) {
        closeMountModal();
      }
    });
  }

  // Sort buttons
  document.querySelectorAll(".vcv-sort").forEach((button) => {
    button.addEventListener("click", () => {
      const key = button.getAttribute("data-sort-key");
      handleSortClick(key);
    });
  });

  // Refresh button
  const refreshBtn = document.getElementById("refresh-btn");
  if (refreshBtn) {
    refreshBtn.addEventListener("click", invalidateCacheAndRefresh);
  }

  const rotateCrlBtn = document.getElementById("rotate-crl-btn");
  if (rotateCrlBtn) {
    rotateCrlBtn.addEventListener("click", rotateCRL);
  }

  const downloadCrlBtn = document.getElementById("download-crl-btn");
  if (downloadCrlBtn) {
    downloadCrlBtn.addEventListener("click", downloadCRL);
  }

  // Language selector
  const langSelect = document.getElementById("vcv-lang-select");
  if (langSelect) {
    langSelect.addEventListener("change", (event) => {
      const value = event.target.value;
      const url = new URL(window.location.href);
      if (value) {
        url.searchParams.set("lang", value);
      } else {
        url.searchParams.delete("lang");
      }
      window.location.href = url.toString();
    });
  }

  // Modal close on backdrop click
  const modal = document.getElementById("certificate-modal");
  if (modal) {
    modal.addEventListener("click", (e) => {
      if (e.target === modal) {
        closeModal();
      }
    });
  }
}

// Main initialization
async function main() {
  initTheme();
  initKeyboardShortcuts();
  initEventHandlers();
  await loadTranslations();
  await loadConfig();
  renderMountSelector(); // Initialize mount selector after config is loaded
  // Load certificates in the background so the UI renders immediately
  const footer = document.querySelector(".vcv-footer");
  const usesHtmxStatus = Boolean(window.htmx && footer && footer.getAttribute("hx-get") === "/ui/status");
  if (!usesHtmxStatus) {
    await loadStatus();
  }
  loadCertificates();
}

// Start the application
main();

// Expose functions globally for HTML onclick handlers
window.closeModal = closeModal;
window.showCertificateDetails = showCertificateDetails;
window.downloadCertificatePEM = downloadCertificatePEM;
window.toggleMount = toggleMount;
window.selectAllMounts = selectAllMounts;
window.deselectAllMounts = deselectAllMounts;
window.hideToast = hideToast;
window.dismissNotifications = dismissNotifications;
window.toggleTheme = toggleTheme;
