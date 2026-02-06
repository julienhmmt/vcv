package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yuin/goldmark"

	"vcv/config"
	"vcv/internal/certs"
	"vcv/internal/i18n"
	"vcv/internal/logger"
	"vcv/internal/vault"
	"vcv/internal/version"
	"vcv/middleware"
)

type certDetailsTemplateData struct {
	Badges        []certStatusBadgeTemplateData
	Certificate   certs.DetailedCertificate
	CertificateID string
	CreatedAtDate string
	CreatedAtText string
	CreatedAtTime string
	DaysLabel     string
	ExpiresAtDate string
	ExpiresAtText string
	ExpiresAtTime string
	ExpiryHint    string
	ExpiryState   string
	ExpiryTone    string
	KeySummary    string
	Language      i18n.Language
	Messages      i18n.Messages
	UsageSummary  string
}

type statusIndicatorTemplateData struct {
	Messages    i18n.Messages
	VersionText string
	Items       []vaultStatusItem
	Summary     *vaultStatusItem
}

type vaultStatusItem struct {
	Text      string
	Class     string
	Title     string
	Connected bool
}

type vaultHealthCheckResult struct {
	index    int
	instance config.VaultInstance
	entry    vaultHealthCacheEntry
}

type vaultHealthCache struct {
	ttl    time.Duration
	mu     sync.Mutex
	values map[string]vaultHealthCacheEntry
}

type vaultHealthCacheEntry struct {
	checkedAt       time.Time
	connected       bool
	errText         string
	isNotConfigured bool
}

func newVaultHealthCache(ttl time.Duration) *vaultHealthCache {
	return &vaultHealthCache{ttl: ttl, values: make(map[string]vaultHealthCacheEntry)}
}

func (c *vaultHealthCache) get(vaultID string) (vaultHealthCacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.values[vaultID]
	if !ok {
		return vaultHealthCacheEntry{}, false
	}
	if time.Since(entry.checkedAt) > c.ttl {
		return vaultHealthCacheEntry{}, false
	}
	return entry, true
}

func (c *vaultHealthCache) set(vaultID string, entry vaultHealthCacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[vaultID] = entry
}

func (c *vaultHealthCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values = make(map[string]vaultHealthCacheEntry)
}

type themeToggleTemplateData struct {
	Theme string
	Icon  string
}

type indexTemplateData struct {
	Language       string
	Messages       i18n.Messages
	Certs          certsFragmentTemplateData
	AppVersionText string
}

type certsFragmentTemplateData struct {
	Rows                []certRowTemplateData
	Messages            i18n.Messages
	ShowVaultMount      bool
	PageInfoText        string
	PageCountText       string
	PageCountHidden     bool
	PagePrevDisabled    bool
	PageNextDisabled    bool
	PageIndex           int
	SortKey             string
	SortDirection       string
	SortCommonActive    bool
	SortCreatedActive   bool
	SortExpiresActive   bool
	SortVaultActive     bool
	SortPkiActive       bool
	SortCommonDir       string
	SortCreatedDir      string
	SortExpiresDir      string
	SortVaultDir        string
	SortPkiDir          string
	PaginationPrevText  string
	PaginationNextText  string
	DashboardTotal      int
	DashboardValid      int
	DashboardExpiring   int
	DashboardExpired    int
	DashboardRevoked    int
	DashboardCertsLabel string
	AdminDocsTitle      string `json:"adminDocsTitle"`
}

type dashboardStatsTemplateData struct {
	Total    int
	Valid    int
	Expiring int
	Expired  int
	Revoked  int
}

type certRowTemplateData struct {
	ID                 string
	CommonName         string
	VaultName          string
	MountName          string
	ShowVaultMount     bool
	Sans               string
	CreatedAt          string
	CreatedAtDate      string
	CreatedAtTime      string
	ExpiresAt          string
	ExpiresAtDate      string
	ExpiresAtTime      string
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
	VaultFilter    string
	PKIFilter      string
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
	templates := template.New("").Funcs(templateFuncMap())
	if t, err := templates.ParseFS(webFS, "templates/*.html"); err == nil {
		templates = t
	} else {
		logger.Get().Error().Err(err).Msg("failed to parse templates")
	}
	// Try to add index.html if it exists
	if indexData, err := fs.ReadFile(webFS, "index.html"); err == nil {
		if _, err := templates.New("index.html").Parse(string(indexData)); err != nil {
			logger.Get().Error().Err(err).Msg("failed to parse index.html")
		}
	}
	vaultHealthCache := newVaultHealthCache(30 * time.Second)
	vaultDisplayNames := buildVaultDisplayNames(vaultInstances)
	showVaultMount := shouldShowVaultMount(vaultInstances)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if templates == nil || templates.Lookup("index.html") == nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		language := i18n.ResolveLanguage(r)
		messages := i18n.MessagesForLanguage(language)
		queryState := parseCertsQueryState(r)
		var certificates []certs.Certificate
		var listErr error
		certificates, listErr = vaultClient.ListCertificates(r.Context())
		if listErr != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, listErr).
				Str("request_id", requestID).
				Msg("failed to list certificates for index")
			certificates = []certs.Certificate{}
		}
		certsData := buildCertsFragmentData(certificates, expirationThresholds, messages, queryState, vaultDisplayNames, showVaultMount)
		data := indexTemplateData{Language: string(language), Messages: messages, Certs: certsData, AppVersionText: version.Version}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to render index template")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	})

	renderCerts := func(w http.ResponseWriter, r *http.Request) bool {
		language := i18n.ResolveLanguage(r)
		messages := i18n.MessagesForLanguage(language)
		queryState := parseCertsQueryState(r)
		certificates, listErr := vaultClient.ListCertificates(r.Context())
		if listErr != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, listErr).
				Str("request_id", requestID).
				Msg("failed to list certificates")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return false
		}
		if renderErr := renderCertsFragment(w, templates, certificates, expirationThresholds, messages, queryState, vaultDisplayNames, showVaultMount); renderErr != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, renderErr).
				Str("request_id", requestID).
				Msg("failed to render certs fragment template")
			return false
		}
		return true
	}
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
		if !renderCerts(w, r) {
			return
		}
		requestID := middleware.GetRequestID(r.Context())
		logger.HTTPEvent(r.Method, r.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Msg("rendered certs fragment")
	})
	router.Post("/ui/certs/refresh", func(w http.ResponseWriter, r *http.Request) {
		vaultClient.InvalidateCache()
		if !renderCerts(w, r) {
			return
		}
		requestID := middleware.GetRequestID(r.Context())
		logger.HTTPEvent(r.Method, r.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Msg("refreshed certs fragment")
	})
	router.Get("/ui/status", func(w http.ResponseWriter, r *http.Request) {
		if templates == nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		language := i18n.ResolveLanguage(r)
		messages := i18n.MessagesForLanguage(language)

		results := checkVaultsHealth(r.Context(), vaultInstances, vaultStatusClients, vaultHealthCache)
		connectedCount := 0
		for _, res := range results {
			if res.entry.connected {
				connectedCount++
			}
		}
		totalCount := len(vaultInstances)

		var summary *vaultStatusItem
		if totalCount == 0 {
			summary = &vaultStatusItem{Text: messages.FooterVaultNotConfigured, Class: "vcv-status-state-neutral", Title: vault.ErrVaultNotConfigured.Error()}
		} else {
			text := ""
			if totalCount > 1 {
				summaryValue := interpolatePlaceholder(messages.FooterVaultSummary, "up", fmt.Sprintf("%d", connectedCount))
				text = interpolatePlaceholder(summaryValue, "total", fmt.Sprintf("%d", totalCount))
			} else {
				// Single vault: use name
				name := strings.TrimSpace(results[0].instance.DisplayName)
				if name == "" {
					name = strings.TrimSpace(results[0].instance.ID)
				}
				if name == "" {
					name = "Vault"
				}
				text = name
			}

			var class string
			if connectedCount == totalCount {
				class = "vcv-status-state-ok"
			} else {
				class = "vcv-status-state-error"
			}

			summary = &vaultStatusItem{Text: text, Class: class, Title: text}
		}

		data := statusIndicatorTemplateData{
			Messages:    messages,
			VersionText: interpolatePlaceholder(messages.FooterVersion, "version", version.Version),
			Summary:     summary,
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if err := templates.ExecuteTemplate(w, "status-indicator.html", data); err != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to render status indicator template")
			return
		}
		requestID := middleware.GetRequestID(r.Context())
		logger.HTTPEvent(r.Method, r.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Msg("rendered status indicator")
	})
	renderVaultStatus := func(w http.ResponseWriter, r *http.Request) bool {
		if templates == nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return false
		}
		language := i18n.ResolveLanguage(r)
		messages := i18n.MessagesForLanguage(language)

		results := checkVaultsHealth(r.Context(), vaultInstances, vaultStatusClients, vaultHealthCache)

		items := make([]vaultStatusItem, 0, len(results))
		if len(results) == 0 {
			items = append(items, vaultStatusItem{Text: messages.FooterVaultNotConfigured, Title: vault.ErrVaultNotConfigured.Error()})
		}
		for _, res := range results {
			title := ""
			cssClass := ""
			if !res.entry.connected {
				if res.entry.isNotConfigured {
					title = messages.FooterVaultNotConfigured
				} else {
					cssClass = "vcv-status-state-error"
					title = res.entry.errText
				}
			} else {
				cssClass = "vcv-status-state-ok"
				title = messages.FooterVaultConnected
			}
			name := strings.TrimSpace(res.instance.DisplayName)
			if name == "" {
				name = strings.TrimSpace(res.instance.ID)
			}
			if name == "" {
				name = "Vault"
			}
			items = append(items, vaultStatusItem{
				Text:      name,
				Class:     cssClass,
				Title:     title,
				Connected: res.entry.connected,
			})
		}

		data := statusIndicatorTemplateData{
			Messages: messages,
			Items:    items,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if err := templates.ExecuteTemplate(w, "vault-status-fragment.html", data); err != nil {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, http.StatusInternalServerError, err).
				Str("request_id", requestID).
				Msg("failed to render vault status fragment template")
			return false
		}
		return true
	}
	router.Get("/ui/vaults/status", func(w http.ResponseWriter, r *http.Request) {
		if !renderVaultStatus(w, r) {
			return
		}
		requestID := middleware.GetRequestID(r.Context())
		logger.HTTPEvent(r.Method, r.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Msg("rendered vault status fragment")
	})
	router.Post("/ui/vaults/refresh", func(w http.ResponseWriter, r *http.Request) {
		vaultHealthCache.clear()
		if !renderVaultStatus(w, r) {
			return
		}
		requestID := middleware.GetRequestID(r.Context())
		logger.HTTPEvent(r.Method, r.URL.Path, http.StatusOK, 0).
			Str("request_id", requestID).
			Msg("refreshed vault status")
	})
	router.Get("/ui/certs/{id:[^/]*}/details", func(w http.ResponseWriter, r *http.Request) {
		if templates == nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		certificateID, statusCode, decodeErr := decodeCertificateIDParam(r)
		if statusCode != http.StatusOK {
			requestID := middleware.GetRequestID(r.Context())
			logger.HTTPError(r.Method, r.URL.Path, statusCode, decodeErr).
				Str("request_id", requestID).
				Msg("missing certificate id in path")
			http.Error(w, http.StatusText(statusCode), statusCode)
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
		language := i18n.ResolveLanguage(r)
		messages := i18n.MessagesForLanguage(language)
		now := time.Now()
		statuses := certificateStatuses(details.Certificate, now)
		badgeViews := make([]certStatusBadgeTemplateData, 0, len(statuses))
		for _, status := range statuses {
			badgeViews = append(badgeViews, certStatusBadgeTemplateData{Class: "vcv-badge vcv-badge-" + status, Label: statusLabelForMessages(status, messages)})
		}
		createdAtText := formatTime(details.CreatedAt)
		createdAtDate := formatDate(details.CreatedAt)
		createdAtTime := formatClock(details.CreatedAt)
		expiresAtText := formatTime(details.ExpiresAt)
		expiresAtDate := formatDate(details.ExpiresAt)
		expiresAtTime := formatClock(details.ExpiresAt)
		daysLabel := ""
		expiryTone := "neutral"
		expiryHint := ""
		expiryState := "scheduled"
		daysRemaining := daysUntil(details.ExpiresAt.UTC(), now)
		hasExpired := !details.ExpiresAt.IsZero() && !details.ExpiresAt.After(now)
		if hasExpired {
			expiryTone = "critical"
			expiryState = "expired"
			daysLabel = interpolatePlaceholder(messages.ExpiredSince, "date", details.ExpiresAt.UTC().Format("2006-01-02"))
		} else if daysRemaining >= 0 {
			if daysRemaining == 0 || daysRemaining == 1 {
				daysLabel = interpolatePlaceholder(messages.DaysRemainingSingular, "days", fmt.Sprintf("%d", daysRemaining))
			} else {
				daysLabel = interpolatePlaceholder(messages.DaysRemaining, "days", fmt.Sprintf("%d", daysRemaining))
			}
			if daysRemaining <= expirationThresholds.Critical {
				expiryTone = "critical"
			} else if daysRemaining <= expirationThresholds.Warning {
				expiryTone = "warning"
			} else {
				expiryTone = "ok"
			}
		}
		if expiresAtText != "" {
			expiryHint = fmt.Sprintf("%s: %s", messages.ColumnExpiresAt, expiresAtText)
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
			CreatedAtText: createdAtText,
			CreatedAtDate: createdAtDate,
			CreatedAtTime: createdAtTime,
			ExpiresAtText: expiresAtText,
			ExpiresAtDate: expiresAtDate,
			ExpiresAtTime: expiresAtTime,
			ExpiryState:   expiryState,
			ExpiryTone:    expiryTone,
			ExpiryHint:    expiryHint,
			DaysLabel:     daysLabel,
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

	router.Get("/ui/docs/user-guide", func(w http.ResponseWriter, r *http.Request) {
		language := i18n.ResolveLanguage(r)
		filename := fmt.Sprintf("docs/user-guide.%s.md", language)

		mdContent, err := fs.ReadFile(webFS, filename)
		if err != nil {
			filename = "docs/user-guide.en.md"
			mdContent, err = fs.ReadFile(webFS, filename)
			if err != nil {
				http.Error(w, "Documentation not found", http.StatusNotFound)
				return
			}
		}

		var buf bytes.Buffer
		if err := goldmark.Convert(mdContent, &buf); err != nil {
			http.Error(w, "Failed to render documentation", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(buf.Bytes()); err != nil {
			logger.Get().Error().Err(err).Msg("failed to write documentation response")
		}
	})

	router.Get("/ui/docs/configuration", func(w http.ResponseWriter, r *http.Request) {
		language := i18n.ResolveLanguage(r)
		filename := fmt.Sprintf("docs/configuration.%s.md", language)

		mdContent, err := fs.ReadFile(webFS, filename)
		if err != nil {
			filename = "docs/configuration.en.md"
			mdContent, err = fs.ReadFile(webFS, filename)
			if err != nil {
				http.Error(w, "Documentation not found", http.StatusNotFound)
				return
			}
		}

		var buf bytes.Buffer
		if err := goldmark.Convert(mdContent, &buf); err != nil {
			http.Error(w, "Failed to render documentation", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(buf.Bytes()); err != nil {
			logger.Get().Error().Err(err).Msg("failed to write documentation response")
		}
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
		VaultFilter:    strings.TrimSpace(query.Get("vault")),
		PKIFilter:      strings.TrimSpace(query.Get("pki")),
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
	if state.VaultFilter == "" {
		state.VaultFilter = "all"
	}
	if state.PKIFilter == "" {
		state.PKIFilter = "all"
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

func shouldResetPageIndex(triggerID string, pageAction string) bool {
	if pageAction == "prev" || pageAction == "next" {
		return false
	}
	switch triggerID {
	case "vcv-search", "vcv-status-filter", "vcv-expiry-filter", "vcv-vault-filter", "vcv-pki-filter", "vcv-page-size", "vcv-mounts", "mount-selector", "vcv-sort-commonName", "vcv-sort-createdAt", "vcv-sort-expiresAt":
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

func applyCertificateFilters(items []certs.Certificate, state certsQueryState, sortKey string, sortDirection string, thresholds config.ExpirationThresholds) []certs.Certificate {
	loweredTerm := strings.ToLower(strings.TrimSpace(state.SearchTerm))
	vaultFilter := strings.ToLower(strings.TrimSpace(state.VaultFilter))
	pkiFilter := strings.ToLower(strings.TrimSpace(state.PKIFilter))
	now := time.Now().UTC()
	maxDays := -1
	if state.ExpiryFilter != "" && state.ExpiryFilter != "all" {
		maxDays = parseInt(state.ExpiryFilter, -1)
	}
	filtered := make([]certs.Certificate, 0, len(items))
	for _, certificate := range items {
		vaultID, mountName := extractVaultIDAndMountName(certificate.ID)
		if vaultFilter != "" && vaultFilter != "all" {
			if strings.ToLower(vaultID) != vaultFilter {
				continue
			}
		}
		if pkiFilter != "" && pkiFilter != "all" {
			if strings.ToLower(mountName) != pkiFilter {
				continue
			}
		}
		statuses := certificateStatuses(certificate, now)
		if state.StatusFilter == "expiring" {
			if !containsString(statuses, "valid") {
				continue
			}
			days := daysUntil(certificate.ExpiresAt.UTC(), now)
			if thresholds.Warning <= 0 || days < 0 || days > thresholds.Warning {
				continue
			}
		} else if state.StatusFilter != "all" && !containsString(statuses, state.StatusFilter) {
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
	return int(math.Floor(diff.Hours() / 24))
}

func sortCertificates(items []certs.Certificate, sortKey string, sortDirection string) []certs.Certificate {
	sorted := make([]certs.Certificate, len(items))
	copy(sorted, items)
	sort.SliceStable(sorted, func(left int, right int) bool {
		leftCert := sorted[left]
		rightCert := sorted[right]
		ascending := sortDirection != "desc"
		if sortKey == "vault" {
			leftVault, leftMount := extractVaultIDAndMountName(leftCert.ID)
			rightVault, rightMount := extractVaultIDAndMountName(rightCert.ID)
			leftValue := strings.ToLower(leftVault)
			rightValue := strings.ToLower(rightVault)
			if leftValue == rightValue {
				if ascending {
					return strings.ToLower(leftMount) < strings.ToLower(rightMount)
				}
				return strings.ToLower(rightMount) < strings.ToLower(leftMount)
			}
			if ascending {
				return leftValue < rightValue
			}
			return rightValue < leftValue
		}
		if sortKey == "pki" {
			leftVault, leftMount := extractVaultIDAndMountName(leftCert.ID)
			rightVault, rightMount := extractVaultIDAndMountName(rightCert.ID)
			leftValue := strings.ToLower(leftMount)
			rightValue := strings.ToLower(rightMount)
			if leftValue == rightValue {
				if ascending {
					return strings.ToLower(leftVault) < strings.ToLower(rightVault)
				}
				return strings.ToLower(rightVault) < strings.ToLower(leftVault)
			}
			if ascending {
				return leftValue < rightValue
			}
			return rightValue < leftValue
		}
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

func buildVaultDisplayNames(instances []config.VaultInstance) map[string]string {
	values := make(map[string]string, len(instances))
	for _, instance := range instances {
		vaultID := strings.TrimSpace(instance.ID)
		if vaultID == "" {
			continue
		}
		displayName := strings.TrimSpace(instance.DisplayName)
		if displayName == "" {
			displayName = vaultID
		}
		values[vaultID] = displayName
	}
	return values
}

func countUniqueMounts(instances []config.VaultInstance) int {
	uniqueMounts := make(map[string]struct{}, 4)
	for _, instance := range instances {
		for _, mount := range instance.PKIMounts {
			trimmed := strings.TrimSpace(mount)
			if trimmed == "" {
				continue
			}
			uniqueMounts[trimmed] = struct{}{}
		}
	}
	return len(uniqueMounts)
}

func shouldShowVaultMount(instances []config.VaultInstance) bool {
	if len(instances) > 1 {
		return true
	}
	return countUniqueMounts(instances) > 1
}

func extractVaultIDAndMountName(certificateID string) (string, string) {
	trimmed := strings.TrimSpace(certificateID)
	if trimmed == "" {
		return "", ""
	}
	vaultID := ""
	mountSerial := trimmed
	if parts := strings.SplitN(trimmed, "|", 2); len(parts) == 2 {
		vaultID = strings.TrimSpace(parts[0])
		mountSerial = strings.TrimSpace(parts[1])
	}
	parts := strings.SplitN(mountSerial, ":", 2)
	if len(parts) < 2 {
		return vaultID, ""
	}
	mountName := strings.TrimSpace(parts[0])
	return vaultID, mountName
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

func resolveExpirationLevel(daysRemaining int, thresholds config.ExpirationThresholds) string {
	if thresholds.Critical > 0 && daysRemaining <= thresholds.Critical {
		return "critical"
	}
	if thresholds.Warning > 0 && daysRemaining <= thresholds.Warning {
		return "warning"
	}
	return "ok"
}

func buildCertRows(items []certs.Certificate, messages i18n.Messages, thresholds config.ExpirationThresholds, vaultDisplayNames map[string]string, showVaultMount bool, now time.Time) []certRowTemplateData {
	rows := make([]certRowTemplateData, 0, len(items))
	for _, certificate := range items {
		vaultID, mountName := extractVaultIDAndMountName(certificate.ID)
		vaultName := vaultID
		if vaultDisplayNames != nil {
			if displayName, ok := vaultDisplayNames[vaultID]; ok {
				vaultName = displayName
			}
		}
		statuses := certificateStatuses(certificate, now)
		badgeViews := make([]certStatusBadgeTemplateData, 0, len(statuses))
		rowClasses := make([]string, 0, len(statuses))
		daysRemaining := daysUntil(certificate.ExpiresAt.UTC(), now)
		isExpiringSoon := thresholds.Warning > 0 && daysRemaining >= 0 && daysRemaining <= thresholds.Warning
		isCritical := thresholds.Critical > 0 && daysRemaining >= 0 && daysRemaining <= thresholds.Critical
		for _, status := range statuses {
			rowClasses = append(rowClasses, "vcv-row-"+status)
			badgeClass := "vcv-badge vcv-badge-" + status
			if status == "valid" {
				if isCritical {
					badgeClass = "vcv-badge vcv-badge-critical"
				} else if isExpiringSoon {
					badgeClass = "vcv-badge vcv-badge-warning"
				}
			}
			badgeViews = append(badgeViews, certStatusBadgeTemplateData{Class: badgeClass, Label: statusLabelForMessages(status, messages)})
		}
		daysRemainingText := ""
		daysRemainingClass := ""
		expiresCellClass := "vcv-expires-cell"
		expiresDateClass := "vcv-expires-date"
		hasExpired := !certificate.ExpiresAt.IsZero() && !certificate.ExpiresAt.After(now)
		if hasExpired {
			daysSinceExpiry := int(math.Abs(float64(daysRemaining)))
			switch daysSinceExpiry {
			case 0:
				daysRemainingText = messages.ExpiredToday
			case 1:
				daysRemainingText = interpolatePlaceholder(messages.ExpiredDaysSingular, "days", "1")
			default:
				daysRemainingText = interpolatePlaceholder(messages.ExpiredDays, "days", fmt.Sprintf("%d", daysSinceExpiry))
			}
			daysRemainingClass = "vcv-days-remaining vcv-days-critical"
			expiresCellClass = "vcv-expires-cell vcv-expires-cell-critical"
			expiresDateClass = "vcv-expires-date vcv-expires-date-critical"
		} else if isExpiringSoon {
			switch daysRemaining {
			case 0:
				daysRemainingText = messages.ExpiringToday
			case 1:
				daysRemainingText = interpolatePlaceholder(messages.DaysRemainingSingular, "days", "1")
			default:
				daysRemainingText = interpolatePlaceholder(messages.DaysRemaining, "days", fmt.Sprintf("%d", daysRemaining))
			}
			level := resolveExpirationLevel(daysRemaining, thresholds)
			daysRemainingClass = "vcv-days-remaining vcv-days-" + level
			expiresCellClass = "vcv-expires-cell vcv-expires-cell-" + level
			expiresDateClass = "vcv-expires-date vcv-expires-date-" + level
		}
		rows = append(rows, certRowTemplateData{
			ID:                 certificate.ID,
			CommonName:         certificate.CommonName,
			VaultName:          vaultName,
			MountName:          mountName,
			ShowVaultMount:     showVaultMount,
			Sans:               strings.Join(certificate.Sans, ", "),
			CreatedAt:          formatTimeReadable(certificate.CreatedAt),
			CreatedAtDate:      formatDateCompact(certificate.CreatedAt),
			CreatedAtTime:      formatClock(certificate.CreatedAt),
			ExpiresAt:          formatTimeReadable(certificate.ExpiresAt),
			ExpiresAtDate:      formatDateCompact(certificate.ExpiresAt),
			ExpiresAtTime:      formatClock(certificate.ExpiresAt),
			ExpiresCellClass:   expiresCellClass,
			ExpiresDateClass:   expiresDateClass,
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

func checkVaultsHealth(ctx context.Context, instances []config.VaultInstance, clients map[string]vault.Client, cache *vaultHealthCache) []vaultHealthCheckResult {
	if len(instances) == 0 || len(clients) == 0 {
		return []vaultHealthCheckResult{}
	}

	resultChan := make(chan vaultHealthCheckResult, len(instances))
	var wg sync.WaitGroup

	for idx, instance := range instances {
		entry, found := cache.get(instance.ID)
		if found {
			resultChan <- vaultHealthCheckResult{index: idx, instance: instance, entry: entry}
			continue
		}

		client, ok := clients[instance.ID]
		if !ok || client == nil {
			entry := vaultHealthCacheEntry{checkedAt: time.Now(), connected: false, errText: "missing vault status client"}
			resultChan <- vaultHealthCheckResult{index: idx, instance: instance, entry: entry}
			continue
		}

		wg.Add(1)
		go func(i int, inst config.VaultInstance, cl vault.Client) {
			defer wg.Done()
			// Create a context with timeout for health check
			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			vaultErr := cl.CheckConnection(checkCtx)
			e := vaultHealthCacheEntry{checkedAt: time.Now(), connected: vaultErr == nil}
			if vaultErr != nil {
				e.errText = vaultErr.Error()
				e.isNotConfigured = errors.Is(vaultErr, vault.ErrVaultNotConfigured)
			}
			cache.set(inst.ID, e)
			resultChan <- vaultHealthCheckResult{index: i, instance: inst, entry: e}
		}(idx, instance, client)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	results := make([]vaultHealthCheckResult, 0, len(instances))
	for res := range resultChan {
		results = append(results, res)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].index < results[j].index
	})

	return results
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format("2006-01-02 15:04:05")
}

func formatTimeReadable(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format("Jan 02, 2006 15:04")
}

func formatDateCompact(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format("02 Jan 2006")
}

func formatDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format("2006-01-02")
}

func formatClock(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format("15:04:05")
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

func renderCertsFragment(w http.ResponseWriter, templates *template.Template, certificates []certs.Certificate, expirationThresholds config.ExpirationThresholds, messages i18n.Messages, queryState certsQueryState, vaultDisplayNames map[string]string, showVaultMount bool) error {
	if templates == nil {
		return fmt.Errorf("templates not available")
	}
	data := buildCertsFragmentData(certificates, expirationThresholds, messages, queryState, vaultDisplayNames, showVaultMount)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return templates.ExecuteTemplate(w, "certs-fragment.html", data)
}

func buildCertsFragmentData(certificates []certs.Certificate, expirationThresholds config.ExpirationThresholds, messages i18n.Messages, queryState certsQueryState, vaultDisplayNames map[string]string, showVaultMount bool) certsFragmentTemplateData {
	filteredByMount := filterCertificatesByMounts(certificates, queryState.SelectedMounts)
	dashboardStats := computeDashboardStats(filteredByMount, expirationThresholds)
	sortKey, sortDirection := resolveSortState(queryState)
	visible := applyCertificateFilters(filteredByMount, queryState, sortKey, sortDirection, expirationThresholds)
	pageIndex := resolvePageIndex(queryState, len(visible), queryState.PageSize)
	_, totalPages := paginateCertificates(visible, pageIndex, queryState.PageSize)
	if shouldResetPageIndex(queryState.TriggerID, queryState.PageAction) {
		pageIndex = 0
		_, totalPages = paginateCertificates(visible, pageIndex, queryState.PageSize)
	}
	pageIndex = applyPageAction(queryState.PageAction, pageIndex, totalPages)
	pageVisible, _ := paginateCertificates(visible, pageIndex, queryState.PageSize)

	rows := buildCertRows(pageVisible, messages, expirationThresholds, vaultDisplayNames, showVaultMount, time.Now().UTC())

	return certsFragmentTemplateData{
		Rows:                rows,
		Messages:            messages,
		ShowVaultMount:      showVaultMount,
		PageInfoText:        buildPaginationInfo(messages, queryState.PageSize, pageIndex, totalPages),
		PageCountText:       fmt.Sprintf("%d", len(visible)),
		PageCountHidden:     len(visible) == 0,
		PagePrevDisabled:    pageIndex <= 0,
		PageNextDisabled:    pageIndex >= totalPages-1,
		PageIndex:           pageIndex,
		SortKey:             sortKey,
		SortDirection:       sortDirection,
		SortCommonActive:    sortKey == "commonName",
		SortCreatedActive:   sortKey == "createdAt",
		SortExpiresActive:   sortKey == "expiresAt",
		SortVaultActive:     sortKey == "vault",
		SortPkiActive:       sortKey == "pki",
		SortCommonDir:       nextSortDirection(sortKey, sortDirection, "commonName"),
		SortCreatedDir:      nextSortDirection(sortKey, sortDirection, "createdAt"),
		SortExpiresDir:      nextSortDirection(sortKey, sortDirection, "expiresAt"),
		SortVaultDir:        nextSortDirection(sortKey, sortDirection, "vault"),
		SortPkiDir:          nextSortDirection(sortKey, sortDirection, "pki"),
		PaginationPrevText:  messages.PaginationPrev,
		PaginationNextText:  messages.PaginationNext,
		DashboardTotal:      dashboardStats.Total,
		DashboardValid:      dashboardStats.Valid,
		DashboardExpiring:   dashboardStats.Expiring,
		DashboardExpired:    dashboardStats.Expired,
		DashboardRevoked:    dashboardStats.Revoked,
		DashboardCertsLabel: messages.DashboardCertsLabel,
		AdminDocsTitle:      messages.AdminDocsTitle,
	}
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func computeDashboardStats(certificates []certs.Certificate, thresholds config.ExpirationThresholds) dashboardStatsTemplateData {
	stats := dashboardStatsTemplateData{}
	now := time.Now().UTC()
	stats.Total = len(certificates)
	for _, cert := range certificates {
		if cert.Revoked {
			stats.Revoked++
			continue
		}
		if !cert.ExpiresAt.IsZero() && !cert.ExpiresAt.After(now) {
			stats.Expired++
			continue
		}
		days := daysUntil(cert.ExpiresAt.UTC(), now)
		if thresholds.Warning > 0 && days >= 0 && days <= thresholds.Warning {
			stats.Expiring++
			continue
		}
		stats.Valid++
	}
	return stats
}

func nextSortDirection(currentKey, currentDir, targetKey string) string {
	if currentKey != targetKey {
		return "asc"
	}
	if currentDir == "asc" {
		return "desc"
	}
	return "asc"
}
