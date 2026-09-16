# VaultCertsViewer alerting

Two independent alerting paths: a built-in webhook for teams without a
Prometheus/Alertmanager stack, and Prometheus metrics + Alertmanager rules
for teams that already have one. Use either, or both.

## Built-in webhook notifications

Set a webhook URL (Settings → Notifications in the admin panel, or
`notifications.webhook_url` in `settings.json`) and the server POSTs a JSON
alert whenever a certificate crosses the warning or critical expiration
threshold — no browser tab needs to be open.

```json
{
  "text": "3 certificate(s) expiring within 7 days or fewer (critical)",
  "tier": "critical",
  "warning_count": 5,
  "critical_count": 3,
  "thresholds": { "warning_days": 30, "critical_days": 7 }
}
```

The top-level `text` field renders as-is in Slack, Discord, and Mattermost
incoming webhooks with no extra configuration; the structured fields are
there for anything else (n8n, a custom script, a second Slack block).

- **Cadence**: checked once on startup, then every 15 minutes.
- **Escalate-only**: one alert per tier (warning, then critical) — it doesn't
  repeat every 15 minutes at the same tier. Once expiry clears back below the
  warning threshold, the next crossing alerts again from the top.
- **Failure handling**: a failed delivery (timeout, non-2xx, DNS failure) is
  retried immediately (3 attempts with backoff), then logged and retried on
  the next check; it never affects the rest of the app.
- **Treat the URL as a secret**: many providers (Slack, Discord) embed an
  auth token in the webhook path. The admin API masks it the same way it
  masks Vault tokens — blank on every read, and a blank/masked value on save
  preserves the stored URL rather than clearing it.

## Routed webhooks (multiple endpoints, severity filter)

`notifications.webhooks` adds routed targets alongside the legacy
`notifications.webhook_url` (which behaves as an all-tiers target; a URL
listed in both places delivers once):

```json
"notifications": {
  "webhook_url": "",
  "webhooks": [
    { "url": "https://hooks.example.com/critical-only", "levels": ["critical"] },
    { "url": "https://hooks.example.com/all-tiers" }
  ]
}
```

- `levels` gates delivery to `"warning"` and/or `"critical"`; empty means
  all tiers. Unknown levels are rejected at save time.
- Escalate-only tracking and immediate retries are per target: one failing
  endpoint does not delay or suppress the others, and only the failed one
  is retried on the next check.
- Routed URLs get the same secret treatment as the legacy URL (masked on
  read, blank/masked on save preserves the stored URL positionally; send an
  explicitly empty list to clear all routed targets).

## Prometheus and Alertmanager

If you are using AlertManager, you can create alerts based on these metrics.

Recommended approach:

- Prefer the aggregated metrics (`vcv_certificates_expiring_soon_count`, `vcv_certificates_total`) for alerting.
- Use the per-certificate metric only for debugging / drill-down (it is disabled by default because it can be high-cardinality).

Example alert rules (multi-vault friendly):

```yaml
- alert: VCVExporterScrapeFailed
  expr: vcv_certificate_exporter_last_scrape_success == 0
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "VCV exporter scrape failed"
    description: "The exporter could not list certificates on the last scrape."

- alert: VCVVaultDown_Global
  expr: vcv_vault_connected{vault_id="__all__"} == 0
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "At least one Vault is down"
    description: "The exporter cannot connect to one or more Vault instances."

- alert: VCVVaultDown
  expr: vcv_vault_connected{vault_id!="__all__"} == 0
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Vault down ({{ $labels.vault_id }})"
    description: "The exporter cannot connect to Vault '{{ $labels.vault_id }}'."

- alert: VCVVaultListingError
  expr: vcv_vault_list_certificates_error{vault_id!="__all__"} == 1
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "Cannot list certificates ({{ $labels.vault_id }})"
    description: "Listing certificates failed for Vault '{{ $labels.vault_id }}'."

- alert: VCVPartialScrape
  expr: vcv_certificates_partial_scrape{vault_id="__all__"} == 1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "VCV partial scrape"
    description: "At least one Vault failed during listing; aggregated counts may be incomplete."

- alert: VCVStaleInventory
  expr: time() - vcv_certificates_last_fetch_timestamp_seconds > 3600
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "VCV inventory is stale"
    description: "The exporter has not refreshed certificates for more than 1 hour."

- alert: VCVExpiringSoonCritical
  expr: sum by (vault_id, pki) (vcv_certificates_expiring_soon_count{level="critical"}) > 0
  labels:
    severity: critical
  annotations:
    summary: "Certificates expiring soon (critical)"
    description: "{{ $value }} certificates are expiring within the critical threshold (vault={{ $labels.vault_id }}, pki={{ $labels.pki }})."

- alert: VCVExpiringSoonWarning
  expr: sum by (vault_id, pki) (vcv_certificates_expiring_soon_count{level="warning"}) > 0
  labels:
    severity: warning
  annotations:
    summary: "Certificates expiring soon (warning)"
    description: "{{ $value }} certificates are expiring within the warning threshold (vault={{ $labels.vault_id }}, pki={{ $labels.pki }})."
```
