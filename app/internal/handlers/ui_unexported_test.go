package handlers

import (
	"fmt"
	"testing"
	"time"

	"vcv/internal/certs"
	"vcv/internal/i18n"
)

func TestBuildKeySummary(t *testing.T) {
	tests := []struct {
		name     string
		details  certs.DetailedCertificate
		expected string
	}{
		{
			name: "no key algorithm and no key size",
			details: certs.DetailedCertificate{
				Certificate: certs.Certificate{},
			},
			expected: "—",
		},
		{
			name: "key algorithm only",
			details: certs.DetailedCertificate{
				KeyAlgorithm: "RSA",
			},
			expected: "RSA",
		},
		{
			name: "key size only",
			details: certs.DetailedCertificate{
				KeySize: 4096,
			},
			expected: "4096",
		},
		{
			name: "key algorithm and key size",
			details: certs.DetailedCertificate{
				KeyAlgorithm: "ECDSA",
				KeySize:      256,
			},
			expected: "ECDSA 256",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildKeySummary(tt.details)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBuildUsageSummary(t *testing.T) {
	tests := []struct {
		name     string
		usages   []string
		expected string
	}{
		{
			name:     "multiple usages",
			usages:   []string{"Digital Signature", "Key Encipherment", "Server Auth"},
			expected: "Digital Signature, Key Encipherment, Server Auth",
		},
		{
			name:     "single usage",
			usages:   []string{"Server Auth"},
			expected: "Server Auth",
		},
		{
			name:     "empty usages",
			usages:   []string{},
			expected: "—",
		},
		{
			name:     "nil usages",
			usages:   nil,
			expected: "—",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildUsageSummary(tt.usages)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestInterpolatePlaceholder(t *testing.T) {
	tests := []struct {
		name     string
		template string
		key      string
		value    string
		expected string
	}{
		{
			name:     "simple interpolation",
			template: "Hello {{name}}",
			key:      "name",
			value:    "World",
			expected: "Hello World",
		},
		{
			name:     "multiple occurrences",
			template: "{{key}} and {{key}}",
			key:      "key",
			value:    "value",
			expected: "value and value",
		},
		{
			name:     "no placeholder",
			template: "No placeholders here",
			key:      "missing",
			value:    "value",
			expected: "No placeholders here",
		},
		{
			name:     "empty template",
			template: "",
			key:      "key",
			value:    "value",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpolatePlaceholder(tt.template, tt.key, tt.value)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		fallback int
		expected int
	}{
		{
			name:     "valid integer",
			value:    "42",
			fallback: 0,
			expected: 42,
		},
		{
			name:     "negative integer",
			value:    "-10",
			fallback: 0,
			expected: -10,
		},
		{
			name:     "invalid string",
			value:    "not a number",
			fallback: 10,
			expected: 10,
		},
		{
			name:     "empty string",
			value:    "",
			fallback: 5,
			expected: 5,
		},
		{
			name:     "string with spaces",
			value:    "  25  ",
			fallback: 0,
			expected: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseInt(tt.value, tt.fallback)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestResolveSortState(t *testing.T) {
	tests := []struct {
		name        string
		state       certsQueryState
		expectedKey string
		expectedDir string
	}{
		{
			name: "sort request overrides current",
			state: certsQueryState{
				SortKey:       "commonName",
				SortDirection: "asc",
				SortRequest:   "expiresAt",
			},
			expectedKey: "expiresAt",
			expectedDir: "asc",
		},
		{
			name: "same key toggles direction",
			state: certsQueryState{
				SortKey:       "commonName",
				SortDirection: "asc",
				SortRequest:   "commonName",
			},
			expectedKey: "commonName",
			expectedDir: "desc",
		},
		{
			name: "no sort request keeps current",
			state: certsQueryState{
				SortKey:       "commonName",
				SortDirection: "asc",
				SortRequest:   "",
			},
			expectedKey: "commonName",
			expectedDir: "asc",
		},
		{
			name: "empty state defaults",
			state: certsQueryState{
				SortRequest: "expiresAt",
			},
			expectedKey: "expiresAt",
			expectedDir: "asc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, dir := resolveSortState(tt.state)
			if key != tt.expectedKey {
				t.Errorf("expected key %q, got %q", tt.expectedKey, key)
			}
			if dir != tt.expectedDir {
				t.Errorf("expected dir %q, got %q", tt.expectedDir, dir)
			}
		})
	}
}

func TestShouldResetPageIndex(t *testing.T) {
	tests := []struct {
		name       string
		triggerID  string
		pageAction string
		expected   bool
	}{
		{
			name:       "search trigger resets page",
			triggerID:  "vcv-search",
			pageAction: "",
			expected:   true,
		},
		{
			name:       "status filter trigger resets page",
			triggerID:  "vcv-status-filter",
			pageAction: "",
			expected:   true,
		},
		{
			name:       "expiry filter trigger resets page",
			triggerID:  "vcv-expiry-filter",
			pageAction: "",
			expected:   true,
		},
		{
			name:       "page size trigger resets page",
			triggerID:  "vcv-page-size",
			pageAction: "",
			expected:   true,
		},
		{
			name:       "sort trigger resets page",
			triggerID:  "vcv-sort-commonName",
			pageAction: "",
			expected:   true,
		},
		{
			name:       "pagination trigger does not reset page",
			triggerID:  "vcv-page-next",
			pageAction: "next",
			expected:   false,
		},
		{
			name:       "refresh trigger does not reset page",
			triggerID:  "vcv-refresh",
			pageAction: "",
			expected:   false,
		},
		{
			name:       "unknown trigger does not reset page",
			triggerID:  "unknown",
			pageAction: "",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldResetPageIndex(tt.triggerID, tt.pageAction)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestResolvePageIndex(t *testing.T) {
	tests := []struct {
		name     string
		state    certsQueryState
		total    int
		pageSize string
		expected int
	}{
		{
			name: "valid page index",
			state: certsQueryState{
				PageIndex: 2,
			},
			total:    100,
			pageSize: "25",
			expected: 2,
		},
		{
			name: "page index too high",
			state: certsQueryState{
				PageIndex: 10,
			},
			total:    100,
			pageSize: "25",
			expected: 3, // 100/25 = 4 pages, max index is 3
		},
		{
			name: "negative page index",
			state: certsQueryState{
				PageIndex: -1,
			},
			total:    100,
			pageSize: "25",
			expected: 0,
		},
		{
			name: "empty total",
			state: certsQueryState{
				PageIndex: 1,
			},
			total:    0,
			pageSize: "25",
			expected: 0,
		},
		{
			name: "invalid page size",
			state: certsQueryState{
				PageIndex: 1,
			},
			total:    100,
			pageSize: "invalid",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolvePageIndex(tt.state, tt.total, tt.pageSize)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestApplyPageAction(t *testing.T) {
	tests := []struct {
		name       string
		action     string
		pageIndex  int
		totalPages int
		expected   int
	}{
		{
			name:       "next action",
			action:     "next",
			pageIndex:  0,
			totalPages: 5,
			expected:   1,
		},
		{
			name:       "next action at last page",
			action:     "next",
			pageIndex:  4,
			totalPages: 5,
			expected:   4, // stays at last page
		},
		{
			name:       "prev action",
			action:     "prev",
			pageIndex:  3,
			totalPages: 5,
			expected:   2,
		},
		{
			name:       "prev action at first page",
			action:     "prev",
			pageIndex:  0,
			totalPages: 5,
			expected:   0, // stays at first page
		},
		{
			name:       "first action",
			action:     "first",
			pageIndex:  3,
			totalPages: 5,
			expected:   3,
		},
		{
			name:       "last action",
			action:     "last",
			pageIndex:  1,
			totalPages: 5,
			expected:   1,
		},
		{
			name:       "unknown action",
			action:     "unknown",
			pageIndex:  2,
			totalPages: 5,
			expected:   2, // no change
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyPageAction(tt.action, tt.pageIndex, tt.totalPages)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestPaginateCertificates(t *testing.T) {
	// Create test certificates
	testCerts := make([]certs.Certificate, 100)
	for i := 0; i < 100; i++ {
		testCerts[i] = certs.Certificate{
			ID:         fmt.Sprintf("cert-%d", i),
			CommonName: fmt.Sprintf("test%d.example.com", i),
		}
	}

	tests := []struct {
		name          string
		pageIndex     int
		pageSize      string
		expectedLen   int
		expectedFirst string
		expectedLast  string
	}{
		{
			name:          "first page of 25",
			pageIndex:     0,
			pageSize:      "25",
			expectedLen:   25,
			expectedFirst: "cert-0",
			expectedLast:  "cert-24",
		},
		{
			name:          "second page of 25",
			pageIndex:     1,
			pageSize:      "25",
			expectedLen:   25,
			expectedFirst: "cert-25",
			expectedLast:  "cert-49",
		},
		{
			name:          "last page of 30",
			pageIndex:     3,
			pageSize:      "30",
			expectedLen:   10,
			expectedFirst: "cert-90",
			expectedLast:  "cert-99",
		},
		{
			name:          "invalid page size",
			pageIndex:     0,
			pageSize:      "invalid",
			expectedLen:   25,
			expectedFirst: "cert-0",
			expectedLast:  "cert-24",
		},
		{
			name:          "page size all",
			pageIndex:     0,
			pageSize:      "all",
			expectedLen:   100,
			expectedFirst: "cert-0",
			expectedLast:  "cert-99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := paginateCertificates(testCerts, tt.pageIndex, tt.pageSize)
			if len(result) != tt.expectedLen {
				t.Errorf("expected length %d, got %d", tt.expectedLen, len(result))
			}
			if len(result) > 0 {
				if result[0].ID != tt.expectedFirst {
					t.Errorf("expected first cert %q, got %q", tt.expectedFirst, result[0].ID)
				}
				if result[len(result)-1].ID != tt.expectedLast {
					t.Errorf("expected last cert %q, got %q", tt.expectedLast, result[len(result)-1].ID)
				}
			}
		})
	}
}

func TestDaysUntil(t *testing.T) {
	now := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		expiresAt time.Time
		now       time.Time
		expected  int
	}{
		{
			name:      "expires tomorrow",
			expiresAt: now.Add(24 * time.Hour),
			now:       now,
			expected:  1,
		},
		{
			name:      "expires in 7 days",
			expiresAt: now.Add(7 * 24 * time.Hour),
			now:       now,
			expected:  7,
		},
		{
			name:      "expired yesterday",
			expiresAt: now.Add(-24 * time.Hour),
			now:       now,
			expected:  -1,
		},
		{
			name:      "expires today",
			expiresAt: now.Add(2 * time.Hour),
			now:       now,
			expected:  0,
		},
		{
			name:      "expires in 30 days",
			expiresAt: now.Add(30 * 24 * time.Hour),
			now:       now,
			expected:  30,
		},
		{
			name:      "expired 2 hours ago floors to -1",
			expiresAt: now.Add(-2 * time.Hour),
			now:       now,
			expected:  -1,
		},
		{
			name:      "expired 36 hours ago floors to -2",
			expiresAt: now.Add(-36 * time.Hour),
			now:       now,
			expected:  -2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := daysUntil(tt.expiresAt, tt.now)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		needle   string
		expected bool
	}{
		{
			name:     "string present",
			values:   []string{"apple", "banana", "cherry"},
			needle:   "banana",
			expected: true,
		},
		{
			name:     "string absent",
			values:   []string{"apple", "banana", "cherry"},
			needle:   "orange",
			expected: false,
		},
		{
			name:     "empty slice",
			values:   []string{},
			needle:   "apple",
			expected: false,
		},
		{
			name:     "nil slice",
			values:   nil,
			needle:   "apple",
			expected: false,
		},
		{
			name:     "empty needle",
			values:   []string{"apple", "banana"},
			needle:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsString(tt.values, tt.needle)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestClampInt(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		min      int
		max      int
		expected int
	}{
		{
			name:     "value within range",
			value:    5,
			min:      1,
			max:      10,
			expected: 5,
		},
		{
			name:     "value below min",
			value:    0,
			min:      1,
			max:      10,
			expected: 1,
		},
		{
			name:     "value above max",
			value:    15,
			min:      1,
			max:      10,
			expected: 10,
		},
		{
			name:     "value equals min",
			value:    1,
			min:      1,
			max:      10,
			expected: 1,
		},
		{
			name:     "value equals max",
			value:    10,
			min:      1,
			max:      10,
			expected: 10,
		},
		{
			name:     "negative range",
			value:    -5,
			min:      -10,
			max:      -1,
			expected: -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := clampInt(tt.value, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestMaxInt(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "a greater than b",
			a:        10,
			b:        5,
			expected: 10,
		},
		{
			name:     "b greater than a",
			a:        5,
			b:        10,
			expected: 10,
		},
		{
			name:     "a equals b",
			a:        5,
			b:        5,
			expected: 5,
		},
		{
			name:     "negative values",
			a:        -5,
			b:        -10,
			expected: -5,
		},
		{
			name:     "zero and negative",
			a:        0,
			b:        -5,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maxInt(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestStatusLabelForMessages(t *testing.T) {
	messages := i18n.MessagesForLanguage(i18n.LanguageEnglish)

	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{
			name:     "valid status",
			status:   "valid",
			expected: messages.StatusLabelValid,
		},
		{
			name:     "expired status",
			status:   "expired",
			expected: messages.StatusLabelExpired,
		},
		{
			name:     "revoked status",
			status:   "revoked",
			expected: messages.StatusLabelRevoked,
		},
		{
			name:     "unknown status",
			status:   "unknown",
			expected: messages.StatusLabelRevoked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := statusLabelForMessages(tt.status, messages)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
