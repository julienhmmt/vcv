# Prometheus metrics reference

VCV exposes comprehensive Prometheus metrics at the `/metrics` endpoint for monitoring certificate inventory, expiration status, and Vault connectivity.

## Configuration

### Threshold configuration

Expiration thresholds are configurable via `settings.json`:

```json
{
  "certificates": {
    "expiration_thresholds": {
      "critical": 2,
      "warning": 10
    }
  }
}
```

Or via environment variables (legacy):

- `VCV_EXPIRE_CRITICAL` (default: 7 days)
- `VCV_EXPIRE_WARNING` (default: 30 days)

### Metrics flags

- `VCV_METRICS_PER_CERTIFICATE=true` - Enable high-cardinality per-certificate metrics (disabled by default)
- `VCV_METRICS_ENHANCED=true` - Enable enhanced metrics like expiry buckets (enabled by default)

## Core metrics

### Certificate inventory

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `vcv_certificates_total` | Gauge | `vault_id`, `pki`, `status` | Total certificates by status (valid/expired/revoked) |
| `vcv_certificates_expired_count` | Gauge | - | Total number of expired certificates |
| `vcv_certificates_expiring_soon_count` | Gauge | `vault_id`, `pki`, `level` | Certificates expiring within threshold window (level: warning/critical) |
| `vcv_certificates_last_fetch_timestamp_seconds` | Gauge | - | Unix timestamp of last successful certificate fetch |
| `vcv_cache_size` | Gauge | - | Number of items currently cached |

### Expiration thresholds

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `vcv_expiration_threshold_critical_days` | Gauge | - | Configured critical threshold in days |
| `vcv_expiration_threshold_warning_days` | Gauge | - | Configured warning threshold in days |

**Use Case**: These metrics expose the configured thresholds so you can validate alert rules match your configuration.

### Expiry Time Buckets (Enhanced Metrics)

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `vcv_certificates_expiry_bucket` | Gauge | `vault_id`, `pki`, `bucket` | Certificate count by expiration time range |

**Buckets**:

- `0-7d` - Expiring in 0-7 days
- `7-30d` - Expiring in 7-30 days
- `30-90d` - Expiring in 30-90 days
- `90d+` - Expiring in 90+ days
- `expired` - Already expired
- `revoked` - Revoked certificates

**Use Case**: Trend analysis, capacity planning, and understanding certificate lifecycle distribution.

### Vault connectivity

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `vcv_vault_connected` | Gauge | `vault_id` | Vault connection status (1=connected, 0=disconnected) |
| `vcv_vault_list_certificates_success` | Gauge | `vault_id` | Whether last certificate listing succeeded (1=success, 0=failure) |
| `vcv_vault_list_certificates_error` | Gauge | `vault_id` | Whether last certificate listing errored (1=error, 0=no error) |
| `vcv_vault_list_certificates_duration_seconds` | Gauge | `vault_id` | Duration of last certificate listing operation |
| `vcv_certificates_partial_scrape` | Gauge | `vault_id` | Whether last scrape was partial due to vault errors (1=partial, 0=complete) |

### Configuration metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `vcv_vaults_configured` | Gauge | - | Number of Vault instances configured |
| `vcv_pki_mounts_configured` | Gauge | `vault_id` | Number of PKI mounts configured per vault |

### Exporter Health

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `vcv_certificate_exporter_last_scrape_success` | Gauge | - | Whether last scrape succeeded (1=success, 0=failure) |
| `vcv_certificate_exporter_last_scrape_duration_seconds` | Gauge | - | Duration of last certificate scrape |

## Per-Certificate Metrics (High Cardinality)

**⚠️ Warning**: These metrics are disabled by default due to high cardinality. Enable with `VCV_METRICS_PER_CERTIFICATE=true`.

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `vcv_certificate_expiry_timestamp_seconds` | Gauge | `certificate_id`, `common_name`, `status`, `vault_id`, `pki` | Certificate expiration timestamp (Unix epoch) |
| `vcv_certificate_days_until_expiry` | Gauge | `certificate_id`, `common_name`, `status`, `vault_id`, `pki` | Days remaining until expiration (negative if expired) |

**Use Case**: Debugging specific certificates, drill-down analysis. Not recommended for large deployments (>1000 certificates).

## Label Values

### Special Label Values

- `vault_id="__all__"` - Aggregated across all vaults
- `pki="__all__"` - Aggregated across all PKI mounts

### Status Values

- `valid` - Certificate is valid and not expired
- `expired` - Certificate has expired
- `revoked` - Certificate has been revoked

### Level Values

- `critical` - Within critical threshold (default: 7 days)
- `warning` - Within warning threshold (default: 30 days)

## Example Queries

### Basic Monitoring

```promql
# Total valid certificates
sum(vcv_certificates_total{status="valid"})

# Certificates expiring in critical window
sum(vcv_certificates_expiring_soon_count{level="critical"})

# Certificates expiring in warning window
sum(vcv_certificates_expiring_soon_count{level="warning"})

# Expired certificates
vcv_certificates_expired_count

# Vault connectivity status
vcv_vault_connected{vault_id!="__all__"}
```

### Trend Analysis

```promql
# Certificate distribution by expiry bucket
sum by (bucket) (vcv_certificates_expiry_bucket{vault_id="__all__", pki="__all__"})

# Certificates expiring in next 7 days per vault
sum by (vault_id) (vcv_certificates_expiry_bucket{bucket="0-7d"})

# Revocation rate (calculated)
sum(vcv_certificates_total{status="revoked"}) / sum(vcv_certificates_total) * 100
```

### Per-Vault Analysis

```promql
# Certificates per vault
sum by (vault_id) (vcv_certificates_total{vault_id!="__all__"})

# Critical certificates per vault and PKI
vcv_certificates_expiring_soon_count{level="critical", vault_id!="__all__"}

# Vault listing duration
vcv_vault_list_certificates_duration_seconds{vault_id!="__all__"}
```

## Alert Rules

### Critical Alerts

```yaml
groups:
  - name: vcv_critical
    interval: 1m
    rules:
      - alert: VCVExporterDown
        expr: vcv_certificate_exporter_last_scrape_success == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "VCV exporter scrape failed"
          description: "The exporter could not list certificates on the last scrape."

      - alert: VCVVaultDown
        expr: vcv_vault_connected{vault_id!="__all__"} == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Vault down ({{ $labels.vault_id }})"
          description: "Cannot connect to Vault '{{ $labels.vault_id }}'."

      - alert: VCVCertificatesExpiringSoonCritical
        expr: sum by (vault_id, pki) (vcv_certificates_expiring_soon_count{level="critical"}) > 0
        labels:
          severity: critical
        annotations:
          summary: "Certificates expiring soon (critical)"
          description: "{{ $value }} certificates are expiring within {{ ALERT_VALUE vcv_expiration_threshold_critical_days }} days (vault={{ $labels.vault_id }}, pki={{ $labels.pki }})."
```

### Warning Alerts

```yaml
      - alert: VCVCertificatesExpiringSoonWarning
        expr: sum by (vault_id, pki) (vcv_certificates_expiring_soon_count{level="warning"}) > 0
        labels:
          severity: warning
        annotations:
          summary: "Certificates expiring soon (warning)"
          description: "{{ $value }} certificates are expiring within {{ ALERT_VALUE vcv_expiration_threshold_warning_days }} days (vault={{ $labels.vault_id }}, pki={{ $labels.pki }})."

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

      - alert: VCVHighExpiryRate
        expr: sum(vcv_certificates_expiry_bucket{bucket="0-7d"}) > 10
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "High certificate expiry rate"
          description: "{{ $value }} certificates are expiring in the next 7 days."
```

## Grafana Dashboard Queries

### Certificate Status Overview

```promql
# Donut chart - Certificate status distribution
sum by (status) (vcv_certificates_total{vault_id="__all__", pki="__all__"})
```

### Expiration Timeline

```promql
# Bar chart - Certificates by expiry bucket
sum by (bucket) (vcv_certificates_expiry_bucket{vault_id="__all__", pki="__all__"})
```

### Vault Health Matrix

```promql
# Table - Vault connectivity and certificate counts
sum by (vault_id) (vcv_certificates_total{vault_id!="__all__"})
```

### Threshold Compliance

```promql
# Gauge - Critical threshold
vcv_expiration_threshold_critical_days

# Gauge - Warning threshold
vcv_expiration_threshold_warning_days

# Stat - Certificates in critical window
sum(vcv_certificates_expiring_soon_count{level="critical"})

# Stat - Certificates in warning window
sum(vcv_certificates_expiring_soon_count{level="warning"})
```

## Best Practices

1. **Use aggregated metrics for alerting** - Prefer `vcv_certificates_expiring_soon_count` over per-certificate metrics
2. **Monitor threshold configuration** - Use `vcv_expiration_threshold_*` metrics to validate alert rules
3. **Track trends with buckets** - Use `vcv_certificates_expiry_bucket` for capacity planning
4. **Avoid high-cardinality metrics in production** - Only enable per-certificate metrics for debugging
5. **Set appropriate alert thresholds** - Align Prometheus alerts with your configured expiration thresholds
6. **Monitor vault connectivity** - Alert on `vcv_vault_connected` to catch infrastructure issues early
7. **Track scrape health** - Monitor `vcv_certificate_exporter_last_scrape_success` and scrape duration

## Troubleshooting

### No metrics appearing

1. Check `/metrics` endpoint is accessible
2. Verify Prometheus scrape configuration
3. Check `vcv_certificate_exporter_last_scrape_success` metric

### Metrics show 0 certificates

1. Check `vcv_vault_connected` - ensure vault is reachable
2. Check `vcv_vault_list_certificates_error` - look for listing errors
3. Verify Vault token has correct permissions
4. Check logs for authentication errors

### Threshold metrics don't match alerts

1. Verify `settings.json` configuration is loaded correctly
2. Check `vcv_expiration_threshold_critical_days` and `vcv_expiration_threshold_warning_days` values
3. Ensure Prometheus alert rules reference the correct thresholds

### High cardinality issues

1. Disable per-certificate metrics: `VCV_METRICS_PER_CERTIFICATE=false`
2. Use aggregated metrics instead
3. Consider increasing Prometheus memory limits if needed
