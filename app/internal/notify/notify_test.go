package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vcv/internal/certs"
	"vcv/internal/config"
	"vcv/internal/vault"
)

type fakeCertLister struct {
	certificates []certs.Certificate
	err          error
}

func (f fakeCertLister) ListCertificates(_ context.Context) ([]certs.Certificate, error) {
	return f.certificates, f.err
}

func certExpiringIn(id string, days int) certs.Certificate {
	return certs.Certificate{ID: id, ExpiresAt: time.Now().Add(time.Duration(days) * 24 * time.Hour)}
}

func settingsWithWebhook(url string) config.Config {
	return config.Config{
		Notifications:        config.NotificationsConfig{WebhookURL: url},
		ExpirationThresholds: config.ExpirationThresholds{Warning: 30, Critical: 7},
	}
}

func settingsWithTargets(targets ...config.WebhookTarget) config.Config {
	return config.Config{
		Notifications:        config.NotificationsConfig{Webhooks: targets},
		ExpirationThresholds: config.ExpirationThresholds{Warning: 30, Critical: 7},
	}
}

// noSleep stubs retry backoff so failure-path tests stay fast.
func noSleep(_ time.Duration) {}

func TestNotifier_NoWebhookConfigured_NeverCallsOut(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	lister := fakeCertLister{certificates: []certs.Certificate{certExpiringIn("a", 1)}}
	settings := func() (config.Config, error) { return settingsWithWebhook(""), nil }
	n := New(lister, settings)

	n.Check(context.Background())

	assert.Equal(t, int32(0), callCount.Load())
}

func TestNotifier_EscalatesFromNoneToWarning_Delivers(t *testing.T) {
	var received webhookPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	lister := fakeCertLister{certificates: []certs.Certificate{certExpiringIn("a", 20)}}
	settings := func() (config.Config, error) { return settingsWithWebhook(server.URL), nil }
	n := New(lister, settings)

	n.Check(context.Background())

	assert.Equal(t, "warning", received.Tier)
	assert.Equal(t, 1, received.WarningCount)
	assert.Equal(t, 0, received.CriticalCount)
	assert.Contains(t, received.Text, "1 certificate(s)")
}

func TestNotifier_SameTierTwice_DeliversOnce(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	lister := fakeCertLister{certificates: []certs.Certificate{certExpiringIn("a", 20)}}
	settings := func() (config.Config, error) { return settingsWithWebhook(server.URL), nil }
	n := New(lister, settings)

	n.Check(context.Background())
	n.Check(context.Background())
	n.Check(context.Background())

	assert.Equal(t, int32(1), callCount.Load())
}

func TestNotifier_EscalatesWarningToCritical_DeliversAgain(t *testing.T) {
	var tiers []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload webhookPayload
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		tiers = append(tiers, payload.Tier)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	lister := &mutableCertLister{certificates: []certs.Certificate{certExpiringIn("a", 20)}}
	settings := func() (config.Config, error) { return settingsWithWebhook(server.URL), nil }
	n := New(lister, settings)

	n.Check(context.Background()) // none -> warning: delivers
	lister.certificates = []certs.Certificate{certExpiringIn("a", 3)}
	n.Check(context.Background()) // warning -> critical: delivers

	assert.Equal(t, []string{"warning", "critical"}, tiers)
}

func TestNotifier_ClearsToNone_ThenReescalates(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	lister := &mutableCertLister{certificates: []certs.Certificate{certExpiringIn("a", 20)}}
	settings := func() (config.Config, error) { return settingsWithWebhook(server.URL), nil }
	n := New(lister, settings)

	n.Check(context.Background()) // none -> warning: delivers (1)
	lister.certificates = []certs.Certificate{certExpiringIn("a", 400)}
	n.Check(context.Background()) // warning -> none: resets, no delivery
	lister.certificates = []certs.Certificate{certExpiringIn("a", 20)}
	n.Check(context.Background()) // none -> warning again: delivers (2)

	assert.Equal(t, int32(2), callCount.Load())
}

func TestNotifier_DeliveryFailure_RetriesOnNextCheck(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		if int(n) <= deliveryAttempts {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	lister := fakeCertLister{certificates: []certs.Certificate{certExpiringIn("a", 20)}}
	settings := func() (config.Config, error) { return settingsWithWebhook(server.URL), nil }
	n := New(lister, settings)
	n.sleep = noSleep

	n.Check(context.Background()) // fails (500) on all attempts, tier stays none
	n.Check(context.Background()) // retries, succeeds

	assert.Equal(t, int32(deliveryAttempts+1), callCount.Load())
}

func TestNotifier_SettingsLoadError_NoOp(t *testing.T) {
	lister := fakeCertLister{certificates: []certs.Certificate{certExpiringIn("a", 1)}}
	settings := func() (config.Config, error) { return config.Config{}, assert.AnError }
	n := New(lister, settings)

	assert.NotPanics(t, func() { n.Check(context.Background()) })
}

func TestNotifier_CertListError_NoOp(t *testing.T) {
	lister := fakeCertLister{err: assert.AnError}
	settings := func() (config.Config, error) { return settingsWithWebhook("https://example.com/hook"), nil }
	n := New(lister, settings)

	assert.NotPanics(t, func() { n.Check(context.Background()) })
}

// TestNotifier_DeliveryErrorNeverLeaksURL guards the specific security property
// that matters here: a Slack-style webhook URL embeds an auth token in its
// path, and Go's http.Client wraps unreachable-host errors with the full
// request URL. deliver() must never let that raw error escape.
func TestNotifier_DeliveryErrorNeverLeaksURL(t *testing.T) {
	const secretURL = "http://127.0.0.1:1/services/T000SECRET/B111SECRET/XXXXSECRETTOKEN"
	lister := fakeCertLister{certificates: []certs.Certificate{certExpiringIn("a", 20)}}
	settings := func() (config.Config, error) { return settingsWithWebhook(secretURL), nil }
	n := New(lister, settings)

	err := n.deliver(context.Background(), secretURL, tierWarning, 1, 0, config.ExpirationThresholds{Warning: 30, Critical: 7})

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "SECRET")
	assert.NotContains(t, err.Error(), secretURL)
}

type mutableCertLister struct {
	certificates []certs.Certificate
}

func (m *mutableCertLister) ListCertificates(_ context.Context) ([]certs.Certificate, error) {
	return m.certificates, nil
}

// fakeEnvelopeLister simulates a multi-vault listing where some vaults may
// have errored: it satisfies both CertLister and vault.CertificatesEnvelopeLister.
type fakeEnvelopeLister struct {
	certificates []certs.Certificate
	vaultErrs    []vault.VaultError
}

func (f *fakeEnvelopeLister) ListCertificates(_ context.Context) ([]certs.Certificate, error) {
	return f.certificates, nil
}

func (f *fakeEnvelopeLister) ListCertificatesEnvelope(_ context.Context) ([]certs.Certificate, []vault.VaultError) {
	return f.certificates, f.vaultErrs
}

func TestNotifier_PartialListing_StillDeliversForPresentCerts(t *testing.T) {
	var received webhookPayload
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	lister := &fakeEnvelopeLister{
		certificates: []certs.Certificate{certExpiringIn("present", 3)},
		vaultErrs:    []vault.VaultError{{VaultID: "down", Message: "unreachable"}},
	}
	settings := func() (config.Config, error) { return settingsWithWebhook(server.URL), nil }
	n := New(lister, settings)

	n.Check(context.Background())

	assert.Equal(t, int32(1), callCount.Load())
	assert.Equal(t, "critical", received.Tier)
}

func TestNotifier_PartialListing_MissingCertDoesNotResetState(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	lister := &fakeEnvelopeLister{certificates: []certs.Certificate{certExpiringIn("a", 3)}}
	settings := func() (config.Config, error) { return settingsWithWebhook(server.URL), nil }
	n := New(lister, settings)

	n.Check(context.Background()) // full listing, critical: delivers (1)
	lister.certificates = nil
	lister.vaultErrs = []vault.VaultError{{VaultID: "down", Message: "unreachable"}}
	n.Check(context.Background()) // partial listing, cert "a" missing: no reset
	lister.certificates = []certs.Certificate{certExpiringIn("a", 3)}
	lister.vaultErrs = nil
	n.Check(context.Background()) // full listing, still critical: deduped

	assert.Equal(t, int32(1), callCount.Load())
}

func countingServer(t *testing.T, failFirst int) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if int(callCount.Add(1)) <= failFirst {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	return server, &callCount
}

func TestNotifier_FansOutToMultipleTargets(t *testing.T) {
	serverA, callsA := countingServer(t, 0)
	serverB, callsB := countingServer(t, 0)

	lister := fakeCertLister{certificates: []certs.Certificate{certExpiringIn("a", 20)}}
	settings := func() (config.Config, error) {
		return settingsWithTargets(
			config.WebhookTarget{URL: serverA.URL},
			config.WebhookTarget{URL: serverB.URL, Levels: []string{"warning"}},
		), nil
	}
	n := New(lister, settings)
	n.sleep = noSleep

	n.Check(context.Background())

	assert.Equal(t, int32(1), callsA.Load())
	assert.Equal(t, int32(1), callsB.Load())
}

func TestNotifier_LevelsGateDelivery(t *testing.T) {
	server, calls := countingServer(t, 0)

	lister := &mutableCertLister{certificates: []certs.Certificate{certExpiringIn("a", 20)}}
	settings := func() (config.Config, error) {
		return settingsWithTargets(config.WebhookTarget{URL: server.URL, Levels: []string{"critical"}}), nil
	}
	n := New(lister, settings)
	n.sleep = noSleep

	n.Check(context.Background()) // warning tier: critical-only target stays silent
	assert.Equal(t, int32(0), calls.Load())

	lister.certificates = []certs.Certificate{certExpiringIn("a", 3)}
	n.Check(context.Background()) // critical tier: delivers
	assert.Equal(t, int32(1), calls.Load())
}

func TestNotifier_FailedTargetRetriesAlone(t *testing.T) {
	serverA, callsA := countingServer(t, 0)
	serverB, callsB := countingServer(t, deliveryAttempts)

	lister := fakeCertLister{certificates: []certs.Certificate{certExpiringIn("a", 20)}}
	settings := func() (config.Config, error) {
		return settingsWithTargets(
			config.WebhookTarget{URL: serverA.URL},
			config.WebhookTarget{URL: serverB.URL},
		), nil
	}
	n := New(lister, settings)
	n.sleep = noSleep

	n.Check(context.Background()) // A delivers; B exhausts immediate retries
	assert.Equal(t, int32(1), callsA.Load())
	assert.Equal(t, int32(deliveryAttempts), callsB.Load())

	n.Check(context.Background()) // only B is retried
	assert.Equal(t, int32(1), callsA.Load())
	assert.Equal(t, int32(deliveryAttempts+1), callsB.Load())
}

func TestNotifier_ImmediateRetryWithinCheck(t *testing.T) {
	server, calls := countingServer(t, 1)

	lister := fakeCertLister{certificates: []certs.Certificate{certExpiringIn("a", 20)}}
	settings := func() (config.Config, error) { return settingsWithWebhook(server.URL), nil }
	n := New(lister, settings)
	n.sleep = noSleep

	n.Check(context.Background())

	assert.Equal(t, int32(2), calls.Load())
}

func TestNotifier_LegacyURLDeduped(t *testing.T) {
	server, calls := countingServer(t, 0)

	lister := fakeCertLister{certificates: []certs.Certificate{certExpiringIn("a", 20)}}
	settings := func() (config.Config, error) {
		cfg := settingsWithTargets(config.WebhookTarget{URL: server.URL})
		cfg.Notifications.WebhookURL = server.URL
		return cfg, nil
	}
	n := New(lister, settings)
	n.sleep = noSleep

	n.Check(context.Background())

	assert.Equal(t, int32(1), calls.Load())
}

func TestNotifier_UnknownLevelsDropped(t *testing.T) {
	server, calls := countingServer(t, 0)

	lister := fakeCertLister{certificates: []certs.Certificate{certExpiringIn("a", 20)}}
	settings := func() (config.Config, error) {
		return settingsWithTargets(config.WebhookTarget{URL: server.URL, Levels: []string{"bogus"}}), nil
	}
	n := New(lister, settings)
	n.sleep = noSleep

	n.Check(context.Background())

	assert.Equal(t, int32(0), calls.Load())
}
