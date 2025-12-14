package handlers

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"vcv/config"
	"vcv/internal/certs"
	"vcv/internal/i18n"
	"vcv/internal/logger"
	"vcv/internal/vault"
	"vcv/internal/version"
	"vcv/middleware"
)

const footerVaultPreviewMaxCount int = 3

type certDetailsTemplateData struct {
	Certificate   certs.DetailedCertificate
	Messages      i18n.Messages
	Badges        []certStatusBadgeTemplateData
	KeySummary    string
	UsageSummary  string
	Language      i18n.Language
	CertificateID string
}

type footerStatusTemplateData struct {
	VersionText string
	VaultPills  []footerVaultStatusTemplateData

	VaultSummaryPill  *footerVaultStatusTemplateData
	VaultPreviewPills []footerVaultStatusTemplateData
	VaultAllPills     []footerVaultStatusTemplateData
	VaultHiddenCount  int
}

type footerVaultStatusTemplateData struct {
	Text  string
	Class string
	Title string
}

type footerVaultHealthCache struct {
	ttl    time.Duration
	mu     sync.Mutex
	values map[string]footerVaultHealthCacheEntry
}

type footerVaultHealthCacheEntry struct {
	checkedAt       time.Time
	connected       bool
	errText         string
	isNotConfigured bool
}

func newFooterVaultHealthCache(ttl time.Duration) *footerVaultHealthCache {
	return &footerVaultHealthCache{ttl: ttl, values: make(map[string]footerVaultHealthCacheEntry)}
}

func (c *footerVaultHealthCache) get(vaultID string) (footerVaultHealthCacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.values[vaultID]
	if !ok {
		return footerVaultHealthCacheEntry{}, false
	}
	if time.Since(entry.checkedAt) > c.ttl {
		return footerVaultHealthCacheEntry{}, false
	}
	return entry, true
}

func (c *footerVaultHealthCache) set(vaultID string, entry footerVaultHealthCacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[vaultID] = entry
}

type themeToggleTemplateData struct {
	Theme string
	Icon  string
}

type certsFragmentTemplateData struct {
	Rows                  []certRowTemplateData
	Messages              i18n.Messages
	PageInfoText          string
	PageCountText         string
	PageCountHidden       bool
	PagePrevDisabled      bool
	PageNextDisabled      bool
	PageIndex             int
	SortKey               string
	SortDirection         string
	SortCommonActive      bool
	SortCreatedActive     bool
	SortExpiresActive     bool
	SortCommonDir         string
	SortCreatedDir        string
	SortExpiresDir        string
	PaginationPrevText    string
	PaginationNextText    string
	DashboardTotal        int
	DashboardValid        int
	DashboardExpiring     int
	DashboardExpired      int
	ChartTotal            int
	ChartValid            int
	ChartExpired          int
	ChartRevoked          int
	ChartHasData          bool
	DonutCircumference    string
	DonutHasValid         bool
	DonutHasExpired       bool
	DonutHasRevoked       bool
	DonutValidDash        string
	DonutExpiredDash      string
	DonutRevokedDash      string
	DonutValidDashArray   string
	DonutExpiredDashArray string
	DonutRevokedDashArray string
	DonutValidOffset      string
	DonutExpiredOffset    string
	DonutRevokedOffset    string
	DualStatusCount       int
	DualStatusNoteText    string
	TimelineItems         []expiryTimelineItemTemplateData
}

type expiryTimelineItemTemplateData struct {
	ID        string
	Name      string
	DotClass  string
	Days      int
	DaysLabel string
}

type certRowTemplateData struct {
	ID                 string
	CommonName         string
	Sans               string
	CreatedAt          string
	ExpiresAt          string
	ExpiresCellClass   string
	ExpiresDateClass   string
	DaysRemainingText  string
	DaysRemainingClass string
	RowClass           string
	Badges             []certStatusBadgeTemplateData
	ButtonDetailsText  string
	ButtonDownloadPEM  string
}

type certStatusBadgeTemplateData struct {
	Class string
	Label string
}

type certsQueryState struct {
	SearchTerm     string
	StatusFilter   string
	ExpiryFilter   string
	PageSize       string
	PageIndex      int
	SortKey        string
	SortDirection  string
	SelectedMounts []string
	PageAction     string
	SortRequest    string
	TriggerID      string
}

func RegisterUIRoutes(router chi.Router, vaultClient vault.Client, vaultInstances []config.VaultInstance, vaultStatusClients map[string]vault.Client, webFS fs.FS, expirationThresholds config.ExpirationThresholds) {
	templates, err := template.ParseFS(webFS, "templates/*.html")
	if err != nil {
		panic(err)
	}
	vaultHealthCache := newFooterVaultHealthCache(5 * time.Second)
	router.Post("/ui/theme/toggle", func(w http.ResponseWriter, r *http.Request) {
		if parseErr := r.ParseForm(); parseErr != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		currentTheme := strings.TrimSpace(r.Form.Get("theme"))
		if currentTheme != "dark" {
			currentTheme = "light"
		}
		nextTheme := "dark"
		if currentTheme == "dark" {
			nextTheme = "light"
		}
		icon := "🌙"
		if nextTheme == "dark" {
			icon = "☀️"
		}
		data := themeToggleTemplateData{Theme: nextTheme, Icon: icon}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if execErr := templates.ExecuteTemplate(w, "theme-toggle-fragment.html", data); execErr != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, execErr).
				Str("request_id", requestID).
				Msg("failed to render theme toggle fragment template")
			return
		}
		requestID := middleware.GetRequestID(r.Context())
		logger.HTTPEvent(r.Method, r.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Str("theme", nextTheme).
			Msg("toggled theme")
	})
	router.Get("/ui/certs", func(w http.ResponseWriter, r *http.Request) {
		language := resolveLanguage(r)
		messages := i18n.MessagesForLanguage(language)
		queryState := parseCertsQueryState(r)
		certificates, listErr := vaultClient.ListCertificates(r.Context())
		if listErr != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, listErr).
				Str("request_id", requestID).
				Msg("failed to list certificates")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if err := renderCertsFragment(w, templates, certificates, expirationThresholds, messages, queryState); err != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to render certs fragment template")
			return
		}
		requestID := middleware.GetRequestID(r.Context())
		logger.HTTPEvent(r.Method, r.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Msg("rendered certs fragment")
	})
	router.Post("/ui/certs/refresh", func(w http.ResponseWriter, r *http.Request) {
		vaultClient.InvalidateCache()
		language := resolveLanguage(r)
		messages := i18n.MessagesForLanguage(language)
		queryState := parseCertsQueryState(r)
		certificates, listErr := vaultClient.ListCertificates(r.Context())
		if listErr != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, listErr).
				Str("request_id", requestID).
				Msg("failed to list certificates")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if err := renderCertsFragment(w, templates, certificates, expirationThresholds, messages, queryState); err != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to render certs fragment template")
			return
		}
		requestID := middleware.GetRequestID(r.Context())
		logger.HTTPEvent(r.Method, r.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Msg("refreshed certs fragment")
	})
	router.Get("/ui/status", func(w http.ResponseWriter, r *http.Request) {
		language := resolveLanguage(r)
		messages := i18n.MessagesForLanguage(language)
		vaultPills := make([]footerVaultStatusTemplateData, 0, len(vaultInstances))
		connectedCount := 0
		totalCount := len(vaultInstances)
		if len(vaultInstances) == 0 || len(vaultStatusClients) == 0 {
			vaultPills = append(vaultPills, footerVaultStatusTemplateData{Text: messages.FooterVaultNotConfigured, Class: "vcv-footer-pill", Title: vault.ErrVaultNotConfigured.Error()})
		} else {
			for _, instance := range vaultInstances {
				name := strings.TrimSpace(instance.DisplayName)
				if name == "" {
					name = strings.TrimSpace(instance.ID)
				}
				if name == "" {
					name = "Vault"
				}
				client, ok := vaultStatusClients[instance.ID]
				if !ok || client == nil {
					vaultPills = append(vaultPills, footerVaultStatusTemplateData{Text: name, Class: "vcv-footer-pill vcv-footer-pill-error", Title: "missing vault status client"})
					continue
				}
				title := ""
				cssClass := "vcv-footer-pill"
				entry, found := vaultHealthCache.get(instance.ID)
				if !found {
					vaultErr := client.CheckConnection(r.Context())
					entry = footerVaultHealthCacheEntry{checkedAt: time.Now(), connected: vaultErr == nil}
					if vaultErr != nil {
						entry.errText = vaultErr.Error()
						entry.isNotConfigured = errors.Is(vaultErr, vault.ErrVaultNotConfigured)
					}
					vaultHealthCache.set(instance.ID, entry)
				}
				if !entry.connected {
					if entry.isNotConfigured {
						title = messages.FooterVaultNotConfigured
					} else {
						cssClass = "vcv-footer-pill vcv-footer-pill-error"
						title = entry.errText
					}
				} else {
					cssClass = "vcv-footer-pill vcv-footer-pill-ok"
					title = messages.FooterVaultConnected
					connectedCount++
				}
				vaultPills = append(vaultPills, footerVaultStatusTemplateData{Text: name, Class: cssClass, Title: title})
			}
		}
		data := footerStatusTemplateData{VersionText: interpolatePlaceholder(messages.FooterVersion, "version", version.Version), VaultPills: vaultPills}
		if totalCount > 1 {
			summaryValue := interpolatePlaceholder(messages.FooterVaultSummary, "up", fmt.Sprintf("%d", connectedCount))
			summaryText := interpolatePlaceholder(summaryValue, "total", fmt.Sprintf("%d", totalCount))
			summaryClass := "vcv-footer-pill vcv-footer-pill-summary"
			if connectedCount == totalCount {
				summaryClass = summaryClass + " vcv-footer-pill-ok"
			} else {
				summaryClass = summaryClass + " vcv-footer-pill-error"
			}
			summary := footerVaultStatusTemplateData{Text: summaryText, Class: summaryClass, Title: summaryText}
			previewEnd := footerVaultPreviewMaxCount
			if previewEnd > len(vaultPills) {
				previewEnd = len(vaultPills)
			}
			preview := make([]footerVaultStatusTemplateData, 0, previewEnd)
			preview = append(preview, vaultPills[:previewEnd]...)
			data.VaultSummaryPill = &summary
			data.VaultPreviewPills = preview
			data.VaultAllPills = vaultPills
			data.VaultHiddenCount = len(vaultPills) - previewEnd
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if err := templates.ExecuteTemplate(w, "footer-status.html", data); err != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to render footer status template")
			return
		}
		requestID := middleware.GetRequestID(r.Context())
		logger.HTTPEvent(r.Method, r.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Msg("rendered footer status")
	})
	router.Get("/ui/certs/{id:[^/]*}/details", func(w http.ResponseWriter, r *http.Request) {
		certificateIDParam := chi.URLParam(r, "id")
		if certificateIDParam == "" {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusBadRequest, nil).
				Str("request_id", requestID).
				Msg("missing certificate id in path")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		certificateID, err := url.PathUnescape(certificateIDParam)
		if err != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusBadRequest, err).
				Str("request_id", requestID).
				Msg("failed to decode certificate id")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		details, err := vaultClient.GetCertificateDetails(r.Context(), certificateID)
		if err != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Str("serial_number", certificateID).
				Msg("failed to get certificate details")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		language := resolveLanguage(r)
		messages := i18n.MessagesForLanguage(language)
		statuses := certificateStatuses(details.Certificate, time.Now())
		badgeViews := make([]certStatusBadgeTemplateData, 0, len(statuses))
		for _, status := range statuses {
			badgeViews = append(badgeViews, certStatusBadgeTemplateData{Class: "vcv-badge vcv-badge-" + status, Label: statusLabelForMessages(status, messages)})
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		data := certDetailsTemplateData{
			Certificate:   details,
			Messages:      messages,
			Badges:        badgeViews,
			KeySummary:    buildKeySummary(details),
			UsageSummary:  buildUsageSummary(details.Usage),
			Language:      language,
			CertificateID: certificateID,
		}
		if err := templates.ExecuteTemplate(w, "cert-details.html", data); err != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to render certificate details template")
			return
		}
		requestID := middleware.GetRequestID(r.Context())
		logger.HTTPEvent(r.Method, r.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Str("serial_number", certificateID).
			Msg("rendered certificate details")
	})
}

func buildKeySummary(details certs.DetailedCertificate) string {
	if details.KeyAlgorithm == "" && details.KeySize == 0 {
		return "—"
	}
	if details.KeySize == 0 {
		return details.KeyAlgorithm
	}
	if details.KeyAlgorithm == "" {
		return fmt.Sprintf("%d", details.KeySize)
	}
	return fmt.Sprintf("%s %d", details.KeyAlgorithm, details.KeySize)
}

func buildUsageSummary(usages []string) string {
	trimmed := make([]string, 0, len(usages))
	for _, usage := range usages {
		value := strings.TrimSpace(usage)
		if value == "" {
			continue
		}
		trimmed = append(trimmed, value)
	}
	if len(trimmed) == 0 {
		return "—"
	}
	return strings.Join(trimmed, ", ")
}

func interpolatePlaceholder(templateValue, key, value string) string {
	replaced := strings.ReplaceAll(templateValue, "{{"+key+"}}", value)
	return strings.ReplaceAll(replaced, "{{ "+key+" }}", value)
}

func parseCertsQueryState(r *http.Request) certsQueryState {
	query := r.URL.Query()
	if parseErr := r.ParseForm(); parseErr == nil {
		query = r.Form
	}
	pageIndex := parseInt(query.Get("page"), 0)
	state := certsQueryState{
		SearchTerm:     strings.TrimSpace(query.Get("search")),
		StatusFilter:   strings.TrimSpace(query.Get("status")),
		ExpiryFilter:   strings.TrimSpace(query.Get("expiry")),
		PageSize:       strings.TrimSpace(query.Get("pageSize")),
		PageIndex:      pageIndex,
		SortKey:        strings.TrimSpace(query.Get("sortKey")),
		SortDirection:  strings.TrimSpace(query.Get("sortDir")),
		SelectedMounts: parseMountsQueryParam(query),
		PageAction:     strings.TrimSpace(query.Get("pageAction")),
		SortRequest:    strings.TrimSpace(query.Get("sort")),
		TriggerID:      strings.TrimSpace(r.Header.Get("HX-Trigger")),
	}
	if state.StatusFilter == "" {
		state.StatusFilter = "all"
	}
	if state.ExpiryFilter == "" {
		state.ExpiryFilter = "all"
	}
	if state.PageSize == "" {
		state.PageSize = "25"
	}
	if state.SortKey == "" {
		state.SortKey = "commonName"
	}
	if state.SortDirection == "" {
		state.SortDirection = "asc"
	}
	return state
}

func parseInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return parsed
}

func resolveSortState(state certsQueryState) (string, string) {
	key := state.SortKey
	direction := state.SortDirection
	requested := state.SortRequest
	if requested == "" {
		return key, direction
	}
	if requested == key {
		if direction == "asc" {
			return key, "desc"
		}
		return key, "asc"
	}
	return requested, "asc"
}

func resolveSortDirAttribute(activeKey, activeDir, buttonKey string) string {
	if activeKey != buttonKey {
		return ""
	}
	return activeDir
}

func shouldResetPageIndex(triggerID string, pageAction string) bool {
	if pageAction == "prev" || pageAction == "next" {
		return false
	}
	switch triggerID {
	case "vcv-search", "vcv-status-filter", "vcv-expiry-filter", "vcv-page-size", "vcv-mounts", "mount-selector", "vcv-sort-commonName", "vcv-sort-createdAt", "vcv-sort-expiresAt":
		return true
	default:
		return false
	}
}

func resolvePageIndex(state certsQueryState, total int, pageSize string) int {
	if pageSize == "all" {
		return 0
	}
	size := parseInt(pageSize, 25)
	if size <= 0 {
		size = 25
	}
	totalPages := maxInt(1, int((total+size-1)/size))
	return clampInt(state.PageIndex, 0, totalPages-1)
}

func applyPageAction(action string, pageIndex int, totalPages int) int {
	if totalPages <= 0 {
		return 0
	}
	switch action {
	case "prev":
		return clampInt(pageIndex-1, 0, totalPages-1)
	case "next":
		return clampInt(pageIndex+1, 0, totalPages-1)
	default:
		return clampInt(pageIndex, 0, totalPages-1)
	}
}

func paginateCertificates(items []certs.Certificate, pageIndex int, pageSize string) ([]certs.Certificate, int) {
	if pageSize == "all" {
		return items, 1
	}
	size := parseInt(pageSize, 25)
	if size <= 0 {
		size = 25
	}
	totalPages := maxInt(1, int((len(items)+size-1)/size))
	pageIndex = clampInt(pageIndex, 0, totalPages-1)
	start := pageIndex * size
	end := start + size
	if start > len(items) {
		return []certs.Certificate{}, totalPages
	}
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], totalPages
}

func applyCertificateFilters(items []certs.Certificate, state certsQueryState, sortKey string, sortDirection string) []certs.Certificate {
	loweredTerm := strings.ToLower(strings.TrimSpace(state.SearchTerm))
	now := time.Now().UTC()
	maxDays := -1
	if state.ExpiryFilter != "" && state.ExpiryFilter != "all" {
		maxDays = parseInt(state.ExpiryFilter, -1)
	}
	filtered := make([]certs.Certificate, 0, len(items))
	for _, certificate := range items {
		statuses := certificateStatuses(certificate, now)
		if state.StatusFilter != "all" && !containsString(statuses, state.StatusFilter) {
			continue
		}
		if maxDays >= 0 {
			days := daysUntil(certificate.ExpiresAt, now)
			if days < 0 || days > maxDays {
				continue
			}
		}
		if loweredTerm != "" {
			cn := strings.ToLower(certificate.CommonName)
			sans := strings.ToLower(strings.Join(certificate.Sans, " "))
			if !strings.Contains(cn, loweredTerm) && !strings.Contains(sans, loweredTerm) {
				continue
			}
		}
		filtered = append(filtered, certificate)
	}
	sorted := sortCertificates(filtered, sortKey, sortDirection)
	return sorted
}

func certificateStatuses(certificate certs.Certificate, now time.Time) []string {
	statuses := make([]string, 0, 2)
	if certificate.Revoked {
		statuses = append(statuses, "revoked")
	}
	if !certificate.ExpiresAt.IsZero() && !certificate.ExpiresAt.After(now) {
		statuses = append(statuses, "expired")
	}
	if len(statuses) == 0 {
		statuses = append(statuses, "valid")
	}
	return statuses
}

func daysUntil(expiresAt time.Time, now time.Time) int {
	if expiresAt.IsZero() {
		return -1
	}
	diff := expiresAt.Sub(now)
	return int(math.Ceil(diff.Hours() / 24))
}

func sortCertificates(items []certs.Certificate, sortKey string, sortDirection string) []certs.Certificate {
	sorted := make([]certs.Certificate, len(items))
	copy(sorted, items)
	sort.SliceStable(sorted, func(left int, right int) bool {
		leftCert := sorted[left]
		rightCert := sorted[right]
		ascending := sortDirection != "desc"
		if sortKey == "createdAt" {
			if ascending {
				return leftCert.CreatedAt.Before(rightCert.CreatedAt)
			}
			return rightCert.CreatedAt.Before(leftCert.CreatedAt)
		}
		if sortKey == "expiresAt" {
			if ascending {
				return leftCert.ExpiresAt.Before(rightCert.ExpiresAt)
			}
			return rightCert.ExpiresAt.Before(leftCert.ExpiresAt)
		}
		leftName := strings.ToLower(leftCert.CommonName)
		rightName := strings.ToLower(rightCert.CommonName)
		if ascending {
			return leftName < rightName
		}
		return rightName < leftName
	})
	return sorted
}

func buildPaginationInfo(messages i18n.Messages, pageSize string, pageIndex int, totalPages int) string {
	if pageSize == "all" {
		return messages.PaginationAll
	}
	if totalPages <= 0 {
		totalPages = 1
	}
	current := pageIndex + 1
	value := interpolatePlaceholder(messages.PaginationInfo, "current", fmt.Sprintf("%d", current))
	return interpolatePlaceholder(value, "total", fmt.Sprintf("%d", totalPages))
}

func buildCertRows(items []certs.Certificate, messages i18n.Messages, thresholds config.ExpirationThresholds) []certRowTemplateData {
	now := time.Now().UTC()
	rows := make([]certRowTemplateData, 0, len(items))
	for _, certificate := range items {
		statuses := certificateStatuses(certificate, now)
		badgeViews := make([]certStatusBadgeTemplateData, 0, len(statuses))
		rowClasses := make([]string, 0, len(statuses))
		for _, status := range statuses {
			rowClasses = append(rowClasses, "vcv-row-"+status)
			badgeViews = append(badgeViews, certStatusBadgeTemplateData{Class: "vcv-badge vcv-badge-" + status, Label: statusLabelForMessages(status, messages)})
		}
		daysRemainingText := ""
		daysRemainingClass := ""
		daysRemaining := daysUntil(certificate.ExpiresAt.UTC(), now)
		if thresholds.Warning > 0 && daysRemaining >= 0 && daysRemaining <= thresholds.Warning {
			if thresholds.Critical > 0 && daysRemaining <= thresholds.Critical {
				daysRemainingClass = "vcv-days-remaining vcv-days-critical"
			} else {
				daysRemainingClass = "vcv-days-remaining vcv-days-warning"
			}
			if daysRemaining == 1 {
				daysRemainingText = interpolatePlaceholder(messages.DaysRemainingSingular, "days", "1")
			} else {
				daysRemainingText = interpolatePlaceholder(messages.DaysRemaining, "days", fmt.Sprintf("%d", daysRemaining))
			}
		}
		rows = append(rows, certRowTemplateData{
			ID:                 certificate.ID,
			CommonName:         certificate.CommonName,
			Sans:               strings.Join(certificate.Sans, ", "),
			CreatedAt:          formatTime(certificate.CreatedAt),
			ExpiresAt:          formatTime(certificate.ExpiresAt),
			ExpiresCellClass:   "vcv-expires-cell",
			ExpiresDateClass:   "vcv-expires-date",
			DaysRemainingText:  daysRemainingText,
			DaysRemainingClass: daysRemainingClass,
			RowClass:           strings.Join(rowClasses, " "),
			Badges:             badgeViews,
			ButtonDetailsText:  messages.ButtonDetails,
			ButtonDownloadPEM:  messages.ButtonDownloadPEM,
		})
	}
	return rows
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format("2006-01-02 15:04:05")
}

func statusLabelForMessages(status string, messages i18n.Messages) string {
	switch status {
	case "valid":
		return messages.StatusLabelValid
	case "expired":
		return messages.StatusLabelExpired
	default:
		return messages.StatusLabelRevoked
	}
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func renderCertsFragment(w http.ResponseWriter, templates *template.Template, certificates []certs.Certificate, expirationThresholds config.ExpirationThresholds, messages i18n.Messages, queryState certsQueryState) error {
	filteredByMount := filterCertificatesByMounts(certificates, queryState.SelectedMounts)
	dashboardStats := computeDashboardStats(filteredByMount, expirationThresholds)
	chartData := computeStatusChartData(filteredByMount, messages)
	timelineItems := computeExpiryTimelineItems(filteredByMount, expirationThresholds, messages)
	sortKey, sortDirection := resolveSortState(queryState)
	visible := applyCertificateFilters(filteredByMount, queryState, sortKey, sortDirection)
	pageIndex := resolvePageIndex(queryState, len(visible), queryState.PageSize)
	_, totalPages := paginateCertificates(visible, pageIndex, queryState.PageSize)
	if shouldResetPageIndex(queryState.TriggerID, queryState.PageAction) {
		pageIndex = 0
		_, totalPages = paginateCertificates(visible, pageIndex, queryState.PageSize)
	}
	pageIndex = applyPageAction(queryState.PageAction, pageIndex, totalPages)
	pageVisible, _ := paginateCertificates(visible, pageIndex, queryState.PageSize)
	data := certsFragmentTemplateData{
		ChartExpired:          chartData.Expired,
		ChartHasData:          chartData.Total > 0,
		ChartRevoked:          chartData.Revoked,
		ChartTotal:            chartData.Total,
		ChartValid:            chartData.Valid,
		DashboardExpired:      dashboardStats.Expired,
		DashboardExpiring:     dashboardStats.ExpiringSoon,
		DashboardTotal:        dashboardStats.Total,
		DashboardValid:        dashboardStats.Valid,
		DonutCircumference:    chartData.Circumference,
		DonutExpiredDash:      chartData.ExpiredDash,
		DonutExpiredDashArray: chartData.ExpiredDashArray,
		DonutExpiredOffset:    chartData.ExpiredOffset,
		DonutHasExpired:       chartData.Expired > 0,
		DonutHasRevoked:       chartData.Revoked > 0,
		DonutHasValid:         chartData.Valid > 0,
		DonutRevokedDash:      chartData.RevokedDash,
		DonutRevokedDashArray: chartData.RevokedDashArray,
		DonutRevokedOffset:    chartData.RevokedOffset,
		DonutValidDash:        chartData.ValidDash,
		DonutValidDashArray:   chartData.ValidDashArray,
		DonutValidOffset:      chartData.ValidOffset,
		DualStatusCount:       chartData.DualStatusCount,
		DualStatusNoteText:    chartData.DualStatusNoteText,
		Messages:              messages,
		PageCountHidden:       len(visible) == 0,
		PageCountText:         fmt.Sprintf("%d", len(visible)),
		PageIndex:             pageIndex,
		PageInfoText:          buildPaginationInfo(messages, queryState.PageSize, pageIndex, totalPages),
		PageNextDisabled:      queryState.PageSize == "all" || pageIndex >= totalPages-1,
		PagePrevDisabled:      queryState.PageSize == "all" || pageIndex <= 0,
		PaginationNextText:    messages.PaginationNext,
		PaginationPrevText:    messages.PaginationPrev,
		Rows:                  buildCertRows(pageVisible, messages, expirationThresholds),
		SortCommonActive:      sortKey == "commonName",
		SortCommonDir:         resolveSortDirAttribute(sortKey, sortDirection, "commonName"),
		SortCreatedActive:     sortKey == "createdAt",
		SortCreatedDir:        resolveSortDirAttribute(sortKey, sortDirection, "createdAt"),
		SortDirection:         sortDirection,
		SortExpiresActive:     sortKey == "expiresAt",
		SortExpiresDir:        resolveSortDirAttribute(sortKey, sortDirection, "expiresAt"),
		SortKey:               sortKey,
		TimelineItems:         timelineItems,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return templates.ExecuteTemplate(w, "certs-fragment.html", data)
}

func clampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

type dashboardStats struct {
	Total        int
	Valid        int
	Expired      int
	ExpiringSoon int
}

type statusChartData struct {
	Total              int
	Valid              int
	Expired            int
	Revoked            int
	DualStatusCount    int
	DualStatusNoteText string
	Circumference      string
	ValidDash          string
	ExpiredDash        string
	RevokedDash        string
	ValidDashArray     string
	ExpiredDashArray   string
	RevokedDashArray   string
	ValidOffset        string
	ExpiredOffset      string
	RevokedOffset      string
}

func computeDashboardStats(certificates []certs.Certificate, thresholds config.ExpirationThresholds) dashboardStats {
	now := time.Now().UTC()
	stats := dashboardStats{Total: len(certificates)}
	for _, certificate := range certificates {
		statuses := certificateStatuses(certificate, now)
		if containsString(statuses, "valid") {
			stats.Valid += 1
		}
		if containsString(statuses, "expired") {
			stats.Expired += 1
		}
		if thresholds.Warning > 0 {
			days := daysUntil(certificate.ExpiresAt.UTC(), now)
			if days > 0 && days <= thresholds.Warning {
				stats.ExpiringSoon += 1
			}
		}
	}
	return stats
}

func computeStatusChartData(certificates []certs.Certificate, messages i18n.Messages) statusChartData {
	now := time.Now().UTC()
	chart := statusChartData{}
	for _, certificate := range certificates {
		statuses := certificateStatuses(certificate, now)
		hasRevoked := containsString(statuses, "revoked")
		hasExpired := containsString(statuses, "expired")
		if hasRevoked && hasExpired {
			chart.DualStatusCount += 1
		}
		if hasRevoked {
			chart.Revoked += 1
			continue
		}
		if hasExpired {
			chart.Expired += 1
			continue
		}
		chart.Valid += 1
	}
	chart.Total = chart.Valid + chart.Expired + chart.Revoked
	if chart.Total == 0 {
		return chart
	}
	circumference := 2 * math.Pi * 50
	validDash := (float64(chart.Valid) / float64(chart.Total)) * circumference
	expiredDash := (float64(chart.Expired) / float64(chart.Total)) * circumference
	revokedDash := (float64(chart.Revoked) / float64(chart.Total)) * circumference
	startOffset := circumference / 4
	chart.Circumference = fmt.Sprintf("%.3f", circumference)
	chart.ValidDash = fmt.Sprintf("%.3f", validDash)
	chart.ExpiredDash = fmt.Sprintf("%.3f", expiredDash)
	chart.RevokedDash = fmt.Sprintf("%.3f", revokedDash)
	chart.ValidDashArray = fmt.Sprintf("%.3f %.3f", validDash, circumference-validDash)
	chart.ExpiredDashArray = fmt.Sprintf("%.3f %.3f", expiredDash, circumference-expiredDash)
	chart.RevokedDashArray = fmt.Sprintf("%.3f %.3f", revokedDash, circumference-revokedDash)
	chart.ValidOffset = fmt.Sprintf("%.3f", startOffset)
	chart.ExpiredOffset = fmt.Sprintf("%.3f", startOffset-validDash)
	chart.RevokedOffset = fmt.Sprintf("%.3f", startOffset-validDash-expiredDash)
	if chart.DualStatusCount > 0 {
		note := interpolatePlaceholder(messages.DualStatusNote, "count", fmt.Sprintf("%d", chart.DualStatusCount))
		chart.DualStatusNoteText = note
	}
	return chart
}

func computeExpiryTimelineItems(certificates []certs.Certificate, thresholds config.ExpirationThresholds, messages i18n.Messages) []expiryTimelineItemTemplateData {
	if thresholds.Warning <= 0 {
		return []expiryTimelineItemTemplateData{}
	}
	now := time.Now().UTC()
	items := make([]expiryTimelineItemTemplateData, 0, len(certificates))
	for _, certificate := range certificates {
		days := daysUntil(certificate.ExpiresAt.UTC(), now)
		if days <= 0 || days > thresholds.Warning {
			continue
		}
		dotClass := "vcv-timeline-dot-warning"
		if thresholds.Critical > 0 && days <= thresholds.Critical {
			dotClass = "vcv-timeline-dot-critical"
		}
		label := interpolatePlaceholder(messages.DaysRemainingShort, "days", fmt.Sprintf("%d", days))
		items = append(items, expiryTimelineItemTemplateData{ID: certificate.ID, Name: certificate.CommonName, DotClass: dotClass, Days: days, DaysLabel: label})
	}
	sort.SliceStable(items, func(left int, right int) bool {
		return items[left].Days < items[right].Days
	})
	if len(items) > 10 {
		return items[:10]
	}
	return items
}
