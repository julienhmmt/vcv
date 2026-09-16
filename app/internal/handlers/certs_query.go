package handlers

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"vcv/internal/certs"
	"vcv/internal/config"
)

// maxCertsPageSize caps numeric page_size values. Explicit page_size=all
// stays uncapped (same as today's default full-list behavior).
const maxCertsPageSize = 1000

// Sort keys accepted by GET /api/certs, mirroring the web UI options.
const (
	certsSortCommonName = "commonName"
	certsSortExpiresAt  = "expiresAt"
	certsSortVault      = "vault"
	certsSortPKI        = "pki"
)

// certsQuery is the parsed GET /api/certs query string.
type certsQuery struct {
	search    string
	statuses  map[string]struct{} // empty = all statuses
	certType  string              // "all" = all types
	sortKey   string              // "" keeps backend default order
	ascending bool
	page      int  // 1-based
	pageSize  int  // <= 0 = all
	paged     bool // pagination explicitly requested
}

// parseCertsQuery parses and validates the list query, failing fast with a
// client-facing message on invalid values.
func parseCertsQuery(query url.Values) (certsQuery, error) {
	parsed := certsQuery{certType: "all", ascending: true, page: 1, pageSize: -1}
	if raw := strings.TrimSpace(query.Get("search")); raw != "" {
		parsed.search = strings.ToLower(raw)
	}
	if raw := strings.TrimSpace(query.Get("status")); raw != "" {
		parsed.statuses = make(map[string]struct{})
		for _, status := range strings.Split(raw, ",") {
			status = strings.ToLower(strings.TrimSpace(status))
			switch status {
			case certs.StatusValid,
				certs.StatusWarning,
				certs.StatusCritical,
				certs.StatusExpired,
				certs.StatusRevoked:
				parsed.statuses[status] = struct{}{}
			case "":
			default:
				return certsQuery{}, fmt.Errorf("invalid status %q: want valid,warning,critical,expired,revoked", status)
			}
		}
	}
	if raw := strings.TrimSpace(query.Get("cert_type")); raw != "" {
		switch raw {
		case "all", "machine", "user", "both", "unknown":
			parsed.certType = raw
		default:
			return certsQuery{}, fmt.Errorf("invalid cert_type %q: want all,machine,user,both,unknown", raw)
		}
	}
	if raw := strings.TrimSpace(query.Get("sort")); raw != "" {
		switch raw {
		case certsSortCommonName, certsSortExpiresAt, certsSortVault, certsSortPKI:
			parsed.sortKey = raw
		default:
			return certsQuery{}, fmt.Errorf("invalid sort %q: want commonName,expiresAt,vault,pki", raw)
		}
	}
	if raw := strings.TrimSpace(query.Get("order")); raw != "" {
		switch strings.ToLower(raw) {
		case "asc":
			parsed.ascending = true
		case "desc":
			parsed.ascending = false
		default:
			return certsQuery{}, fmt.Errorf("invalid order %q: want asc,desc", raw)
		}
	}
	if raw := strings.TrimSpace(query.Get("page")); raw != "" {
		page, err := strconv.Atoi(raw)
		if err != nil || page < 1 {
			return certsQuery{}, fmt.Errorf("invalid page %q: want an integer >= 1", raw)
		}
		parsed.page = page
		parsed.paged = true
	}
	if raw := strings.TrimSpace(query.Get("page_size")); raw != "" {
		if strings.EqualFold(raw, "all") {
			parsed.pageSize = -1
		} else {
			size, err := strconv.Atoi(raw)
			if err != nil || size < 1 {
				return certsQuery{}, fmt.Errorf("invalid page_size %q: want an integer >= 1 or all", raw)
			}
			if size > maxCertsPageSize {
				size = maxCertsPageSize
			}
			parsed.pageSize = size
		}
		parsed.paged = true
	}
	return parsed, nil
}

// resolveCertThresholds applies the 7/30 day defaults like the metrics
// collector does when settings carry non-positive values.
func resolveCertThresholds(thresholds config.ExpirationThresholds) (int, int) {
	critical, warning := thresholds.Critical, thresholds.Warning
	if critical <= 0 {
		critical = 7
	}
	if warning <= 0 {
		warning = 30
	}
	return critical, warning
}

// certsStatusCounts is the status breakdown of the scoped (mount, search,
// type) result set, ignoring the status filter itself so the UI can render
// per-status facets.
type certsStatusCounts struct {
	Valid    int `json:"valid"`
	Warning  int `json:"warning"`
	Critical int `json:"critical"`
	Expired  int `json:"expired"`
	Revoked  int `json:"revoked"`
	Total    int `json:"total"`
}

// matchCertsSearch mirrors the web UI search: case-insensitive substring
// over common name, serial number, and SANs.
func matchCertsSearch(certificate certs.Certificate, loweredQuery string) bool {
	if loweredQuery == "" {
		return true
	}
	if strings.Contains(strings.ToLower(certificate.CommonName), loweredQuery) {
		return true
	}
	if strings.Contains(strings.ToLower(certificate.SerialNumber), loweredQuery) {
		return true
	}
	for _, san := range certificate.Sans {
		if strings.Contains(strings.ToLower(san), loweredQuery) {
			return true
		}
	}
	return false
}

// applyCertsQuery filters, sorts, and paginates the mount-scoped inventory.
// It returns the page slice (never nil), the filtered total, total pages,
// and status counts over the scoped set.
func applyCertsQuery(all []certs.Certificate, query certsQuery, criticalDays, warningDays int, now time.Time) ([]certs.Certificate, int, int, certsStatusCounts) {
	scoped := make([]certs.Certificate, 0, len(all))
	statuses := make([]string, 0, len(all))
	counts := certsStatusCounts{}
	for _, certificate := range all {
		if query.certType != "all" && certificate.CertType != query.certType {
			continue
		}
		if !matchCertsSearch(certificate, query.search) {
			continue
		}
		status := certs.StatusAt(certificate, criticalDays, warningDays, now)
		scoped = append(scoped, certificate)
		statuses = append(statuses, status)
		counts.Total++
		switch status {
		case certs.StatusValid:
			counts.Valid++
		case certs.StatusWarning:
			counts.Warning++
		case certs.StatusCritical:
			counts.Critical++
		case certs.StatusExpired:
			counts.Expired++
		case certs.StatusRevoked:
			counts.Revoked++
		}
	}
	filtered := make([]certs.Certificate, 0, len(scoped))
	for i, certificate := range scoped {
		if len(query.statuses) > 0 {
			if _, ok := query.statuses[statuses[i]]; !ok {
				continue
			}
		}
		filtered = append(filtered, certificate)
	}
	sortCertsForQuery(filtered, query.sortKey, query.ascending)
	pageItems, totalPages := paginateCerts(filtered, query.page, query.pageSize)
	return pageItems, len(filtered), totalPages, counts
}

// certVaultAndMount splits a composite certificate ID ("vault|mount:serial")
// the same way the web UI does; unknown shapes yield empty parts.
func certVaultAndMount(id string) (string, string) {
	rest := id
	vaultID := ""
	if head, tail, ok := strings.Cut(id, "|"); ok {
		vaultID, rest = strings.TrimSpace(head), tail
	}
	mount, _, _ := strings.Cut(rest, ":")
	return vaultID, strings.TrimSpace(mount)
}

// sortCertsForQuery orders the slice in place; empty sortKey keeps backend
// default order. Ties break on ID so pages are deterministic.
func sortCertsForQuery(list []certs.Certificate, sortKey string, ascending bool) {
	if sortKey == "" {
		return
	}
	less := func(a, b certs.Certificate) bool {
		switch sortKey {
		case certsSortExpiresAt:
			return a.ExpiresAt.Before(b.ExpiresAt)
		case certsSortVault:
			aVault, _ := certVaultAndMount(a.ID)
			bVault, _ := certVaultAndMount(b.ID)
			return strings.ToLower(aVault) < strings.ToLower(bVault)
		case certsSortPKI:
			_, aMount := certVaultAndMount(a.ID)
			_, bMount := certVaultAndMount(b.ID)
			return strings.ToLower(aMount) < strings.ToLower(bMount)
		default: // certsSortCommonName
			return strings.ToLower(a.CommonName) < strings.ToLower(b.CommonName)
		}
	}
	sort.SliceStable(list, func(i, j int) bool {
		a, b := list[i], list[j]
		if less(a, b) {
			return ascending
		}
		if less(b, a) {
			return !ascending
		}
		return a.ID < b.ID
	})
}

// paginateCerts slices the filtered list into the requested 1-based page.
// pageSize <= 0 returns everything with a single page. The slice is never nil.
func paginateCerts(list []certs.Certificate, page, pageSize int) ([]certs.Certificate, int) {
	if pageSize <= 0 {
		if list == nil {
			return []certs.Certificate{}, 1
		}
		return list, 1
	}
	totalPages := (len(list) + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}
	start := (page - 1) * pageSize
	if start >= len(list) {
		return []certs.Certificate{}, totalPages
	}
	end := start + pageSize
	if end > len(list) {
		end = len(list)
	}
	return list[start:end], totalPages
}
