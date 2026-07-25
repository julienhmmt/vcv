// Package notify delivers outbound webhook alerts when certificate expiry
// crosses into the warning or critical threshold, independent of anyone
// having the dashboard open. See internal/certs.CountExpiring for the
// threshold math (shared with the Prometheus collector).
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"vcv/internal/certs"
	"vcv/internal/config"
	"vcv/internal/logger"
)

// CertLister lists all certificates the notifier should evaluate. Satisfied
// by vault.Client.
type CertLister interface {
	ListCertificates(ctx context.Context) ([]certs.Certificate, error)
}

// SettingsLoader returns the current app configuration, read fresh from
// disk. Satisfied by config.Load - passing it directly means a webhook URL
// or threshold edit via the admin panel takes effect on the next Check
// without a restart.
type SettingsLoader func() (config.Config, error)

type tier int

const (
	tierNone tier = iota
	tierWarning
	tierCritical
)

func (t tier) String() string {
	switch t {
	case tierCritical:
		return "critical"
	case tierWarning:
		return "warning"
	default:
		return "none"
	}
}

func currentTier(warning, critical int) tier {
	if critical > 0 {
		return tierCritical
	}
	if warning > 0 {
		return tierWarning
	}
	return tierNone
}

const httpTimeout = 10 * time.Second

// errWebhookDeliveryFailed is logged verbatim - it must never wrap the raw
// transport error, since Go's http.Client embeds the full request URL
// (which may carry a secret, e.g. a Slack webhook path) in that error text.
var errWebhookDeliveryFailed = errors.New("webhook request failed")

// Notifier polls certificate expiry state and POSTs a webhook when the
// overall tier increases. Semantics mirror the frontend's
// expiry-notify.ts: escalate-only, no repeat alerts at the same tier, reset
// once expiry clears back to none. The zero value's lastTier (tierNone)
// means the first Check after startup always notifies if a tier is active,
// matching the frontend's "always notify on initial load" behavior.
type Notifier struct {
	certs    CertLister
	settings SettingsLoader
	client   *http.Client
	now      func() time.Time

	mu       sync.Mutex
	lastTier tier
}

// New builds a Notifier. certLister and settingsLoader are read on every
// Check call so admin edits (webhook URL, thresholds) apply live.
func New(certLister CertLister, settingsLoader SettingsLoader) *Notifier {
	return &Notifier{
		certs:    certLister,
		settings: settingsLoader,
		client:   &http.Client{Timeout: httpTimeout},
		now:      time.Now,
	}
}

// Check lists certificates, computes the expiry tier, and delivers a
// webhook when the tier has increased since the last check. All failures
// (settings load, cert list, delivery) are logged and swallowed - a broken
// webhook must never affect the rest of the app.
func (n *Notifier) Check(ctx context.Context) {
	settings, err := n.settings()
	if err != nil {
		logger.Get().Warn().Err(err).Msg("notify: failed to load settings")
		return
	}
	webhookURL := strings.TrimSpace(settings.Notifications.WebhookURL)
	if webhookURL == "" {
		return
	}

	certificates, err := n.certs.ListCertificates(ctx)
	if err != nil {
		logger.Get().Warn().Err(err).Msg("notify: failed to list certificates")
		return
	}

	thresholds := settings.ExpirationThresholds
	warning, critical := certs.CountExpiring(certificates, thresholds.Warning, thresholds.Critical, n.now())
	current := currentTier(warning, critical)

	n.mu.Lock()
	defer n.mu.Unlock()

	if current == tierNone {
		n.lastTier = tierNone
		return
	}
	if current <= n.lastTier {
		return
	}

	if deliverErr := n.deliver(ctx, webhookURL, current, warning, critical, thresholds); deliverErr != nil {
		logger.Get().Warn().Err(deliverErr).Str("tier", current.String()).Msg("notify: webhook delivery failed, will retry next check")
		return
	}
	n.lastTier = current
}

type webhookPayload struct {
	// Text is a plain-language summary; Slack, Discord, and Mattermost all
	// render a top-level "text" field out of the box with no extra config.
	Text          string            `json:"text"`
	Tier          string            `json:"tier"`
	WarningCount  int               `json:"warning_count"`
	CriticalCount int               `json:"critical_count"`
	Thresholds    webhookThresholds `json:"thresholds"`
}

type webhookThresholds struct {
	WarningDays  int `json:"warning_days"`
	CriticalDays int `json:"critical_days"`
}

func (n *Notifier) deliver(ctx context.Context, webhookURL string, current tier, warning, critical int, thresholds config.ExpirationThresholds) error {
	count, days := warning, thresholds.Warning
	if current == tierCritical {
		count, days = critical, thresholds.Critical
	}
	payload := webhookPayload{
		Text:          fmt.Sprintf("%d certificate(s) expiring within %d days or fewer (%s)", count, days, current),
		Tier:          current.String(),
		WarningCount:  warning,
		CriticalCount: critical,
		Thresholds:    webhookThresholds{WarningDays: thresholds.Warning, CriticalDays: thresholds.Critical},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return errWebhookDeliveryFailed
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return errWebhookDeliveryFailed
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}
