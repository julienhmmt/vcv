# VaultCertsViewer alerting with Prometheus and Alertmanager

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
