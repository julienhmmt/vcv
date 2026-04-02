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

| Metric                                          | Type  | Labels                      | Description                                                             |
| ----------------------------------------------- | ----- | --------------------------- | ----------------------------------------------------------------------- |
| `vcv_certificates_total`                        | Gauge | `vault_id`, `pki`, `status` | Total certificates by status (valid/expired/revoked)                    |
| `vcv_certificates_expired_count`                | Gauge | -                           | Total number of expired certificates                                    |
| `vcv_certificates_expiring_soon_count`          | Gauge | `vault_id`, `pki`, `level`  | Certificates expiring within threshold window (level: warning/critical) |
| `vcv_certificates_last_fetch_timestamp_seconds` | Gauge | -                           | Unix timestamp of last successful certificate fetch                     |
| `vcv_cache_size`                                | Gauge | -                           | Number of items currently cached                                        |

### Expiration thresholds

| Metric                                   | Type  | Labels | Description                           |
| ---------------------------------------- | ----- | ------ | ------------------------------------- |
| `vcv_expiration_threshold_critical_days` | Gauge | -      | Configured critical threshold in days |
| `vcv_expiration_threshold_warning_days`  | Gauge | -      | Configured warning threshold in days  |

**Use case**: These metrics expose the configured thresholds so you can validate alert rules match your configuration.

### Expiry time buckets (enhanced metrics)

| Metric                           | Type  | Labels                      | Description                                |
| -------------------------------- | ----- | --------------------------- | ------------------------------------------ |
| `vcv_certificates_expiry_bucket` | Gauge | `vault_id`, `pki`, `bucket` | Certificate count by expiration time range |

**Buckets**:

- `0-7d` - Expiring in 0-7 days
- `7-30d` - Expiring in 7-30 days
- `30-90d` - Expiring in 30-90 days
- `90d+` - Expiring in 90+ days
- `expired` - Already expired
- `revoked` - Revoked certificates

**Use case**: Trend analysis, capacity planning, and understanding certificate lifecycle distribution.

### Vault connectivity

| Metric                                         | Type  | Labels     | Description                                                                 |
| ---------------------------------------------- | ----- | ---------- | --------------------------------------------------------------------------- |
| `vcv_vault_connected`                          | Gauge | `vault_id` | Vault connection status (1=connected, 0=disconnected)                       |
| `vcv_vault_list_certificates_success`          | Gauge | `vault_id` | Whether last certificate listing succeeded (1=success, 0=failure)           |
| `vcv_vault_list_certificates_error`            | Gauge | `vault_id` | Whether last certificate listing errored (1=error, 0=no error)              |
| `vcv_vault_list_certificates_duration_seconds` | Gauge | `vault_id` | Duration of last certificate listing operation                              |
| `vcv_certificates_partial_scrape`              | Gauge | `vault_id` | Whether last scrape was partial due to vault errors (1=partial, 0=complete) |

### Configuration metrics

| Metric                      | Type  | Labels     | Description                               |
| --------------------------- | ----- | ---------- | ----------------------------------------- |
| `vcv_vaults_configured`     | Gauge | -          | Number of Vault instances configured      |
| `vcv_pki_mounts_configured` | Gauge | `vault_id` | Number of PKI mounts configured per vault |

### Exporter health

| Metric                                                  | Type  | Labels | Description                                          |
| ------------------------------------------------------- | ----- | ------ | ---------------------------------------------------- |
| `vcv_certificate_exporter_last_scrape_success`          | Gauge | -      | Whether last scrape succeeded (1=success, 0=failure) |
| `vcv_certificate_exporter_last_scrape_duration_seconds` | Gauge | -      | Duration of last certificate scrape                  |

## Per-certificate metrics (high cardinality)

**⚠️ Warning**: These metrics are disabled by default due to high cardinality. Enable with `VCV_METRICS_PER_CERTIFICATE=true`.

| Metric                                     | Type  | Labels                                                       | Description                                           |
| ------------------------------------------ | ----- | ------------------------------------------------------------ | ----------------------------------------------------- |
| `vcv_certificate_expiry_timestamp_seconds` | Gauge | `certificate_id`, `common_name`, `status`, `vault_id`, `pki` | Certificate expiration timestamp (Unix epoch)         |
| `vcv_certificate_days_until_expiry`        | Gauge | `certificate_id`, `common_name`, `status`, `vault_id`, `pki` | Days remaining until expiration (negative if expired) |

**Use case**: Debugging specific certificates, drill-down analysis. Not recommended for large deployments (>1000 certificates).

## Pinned certificate metrics (selective monitoring)

**✅ Recommended**: Track specific critical certificates without full per-certificate metrics overhead.

| Metric                                            | Type  | Labels                                                       | Description                                        |
| ------------------------------------------------- | ----- | ------------------------------------------------------------ | -------------------------------------------------- |
| `vcv_pinned_certificate_expiry_timestamp_seconds` | Gauge | `certificate_id`, `common_name`, `status`, `vault_id`, `pki` | Expiration timestamp for pinned certificates       |
| `vcv_pinned_certificate_days_until_expiry`        | Gauge | `certificate_id`, `common_name`, `status`, `vault_id`, `pki` | Days until expiry for pinned certificates          |

**Configuration**: Add certificate identifiers (CN, ID, or SAN) to `settings.json`. Supports wildcard patterns:

```json
{
  "metrics": {
    "pinned_certificates": [
      "api.production.local",
      "*.critical-service.local",
      "loadbalancer.production.local",
      "vault-main|pki:13:71:e8:49:c4:54:b5:3b"
    ]
  }
}
```

**Wildcard Support**: Use `*` for pattern matching (e.g., `*.production.local` matches any subdomain).

**Use case**: Monitor 10-50 critical certificates (API gateways, load balancers, etc.) without enabling full per-certificate metrics.

## Certificate issuer metrics (enhanced)

> ⚠️ **PLACEHOLDER DATA**: Currently uses domain-based heuristics. Requires PEM parsing for accurate issuer CN.

| Metric                              | Type  | Labels                       | Description                                |
| ----------------------------------- | ----- | ---------------------------- | ------------------------------------------ |
| `vcv_certificates_by_issuer_total`  | Gauge | `vault_id`, `pki`, `issuer_cn` | Total certificates grouped by issuer CN    |

**Use case**: Identify certificate sources, track CA distribution, detect unauthorized issuers (after PEM parsing implementation).

## Cryptographic strength metrics (enhanced)

> ⚠️ **PLACEHOLDER DATA**: Returns hardcoded "RSA 2048" for all certificates. Weak key detection **non-functional** until PEM parsing implemented. Do not use for security decisions.

| Metric                                | Type  | Labels                                   | Description                                        |
| ------------------------------------- | ----- | ---------------------------------------- | -------------------------------------------------- |
| `vcv_certificates_by_key_type_total`  | Gauge | `vault_id`, `pki`, `algorithm`, `key_size` | Total certificates by key algorithm and size       |
| `vcv_certificates_weak_keys_total`    | Gauge | `vault_id`, `pki`                        | Number of certificates with weak keys ⚠️ **Always 0** |

**Weak key criteria** (not yet functional):
- RSA keys < 2048 bits (requires PEM parsing)
- DSA keys (any size) (requires PEM parsing)

**Use case**: Security auditing, compliance verification, identify certificates requiring rotation (after PEM parsing implementation).

## Subject Alternative Names metrics (enhanced)

| Metric                                  | Type  | Labels                       | Description                                        |
| --------------------------------------- | ----- | ---------------------------- | -------------------------------------------------- |
| `vcv_certificates_with_sans_total`      | Gauge | `vault_id`, `pki`            | Number of certificates with SANs                   |
| `vcv_certificates_san_count_bucket`     | Gauge | `vault_id`, `pki`, `bucket`  | Certificates grouped by SAN count range            |

**SAN count buckets**:
- `0` - No SANs
- `1-5` - 1 to 5 SANs
- `6-10` - 6 to 10 SANs
- `11+` - More than 10 SANs

**Use case**: Track multi-domain certificates, identify wildcard usage patterns.

## Certificate age metrics (enhanced)

| Metric                           | Type  | Labels                       | Description                                        |
| -------------------------------- | ----- | ---------------------------- | -------------------------------------------------- |
| `vcv_certificates_age_bucket`    | Gauge | `vault_id`, `pki`, `bucket`  | Certificates grouped by age since issuance         |

**Age buckets**:
- `0-30d` - Issued in last 30 days
- `30-90d` - Issued 30-90 days ago
- `90-180d` - Issued 90-180 days ago
- `180-365d` - Issued 180-365 days ago
- `1y+` - Issued over 1 year ago

**Use case**: Understand certificate lifecycle, identify stale certificates, track rotation patterns.

## Certificate renewal metrics (enhanced)

| Metric                               | Type  | Labels                | Description                                        |
| ------------------------------------ | ----- | --------------------- | -------------------------------------------------- |
| `vcv_certificates_issued_last_24h`   | Gauge | `vault_id`, `pki`     | Certificates issued in last 24 hours               |
| `vcv_certificates_issued_last_7d`    | Gauge | `vault_id`, `pki`     | Certificates issued in last 7 days                 |
| `vcv_certificates_issued_last_30d`   | Gauge | `vault_id`, `pki`     | Certificates issued in last 30 days                |

**Use case**: Track renewal activity, detect automation issues, capacity planning, anomaly detection.

## Label values

### Special label values

- `vault_id="__all__"` - Aggregated across all vaults
- `pki="__all__"` - Aggregated across all PKI mounts

### Status values

- `valid` - Certificate is valid and not expired
- `expired` - Certificate has expired
- `revoked` - Certificate has been revoked

### Level values

- `critical` - Within critical threshold (default: 7 days)
- `warning` - Within warning threshold (default: 30 days)

## Example queries

### Basic monitoring

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

### Trend analysis

```promql
# Certificate distribution by expiry bucket
sum by (bucket) (vcv_certificates_expiry_bucket{vault_id="__all__", pki="__all__"})

# Certificates expiring in next 7 days per vault
sum by (vault_id) (vcv_certificates_expiry_bucket{bucket="0-7d"})

# Revocation rate (calculated)
sum(vcv_certificates_total{status="revoked"}) / sum(vcv_certificates_total) * 100
```

### Per-vault analysis

```promql
# Certificates per vault
sum by (vault_id) (vcv_certificates_total{vault_id!="__all__"})

# Critical certificates per vault and PKI
vcv_certificates_expiring_soon_count{level="critical", vault_id!="__all__"}

# Vault listing duration
vcv_vault_list_certificates_duration_seconds{vault_id!="__all__"}
```

### Pinned certificate monitoring

```promql
# Days until expiry for specific pinned certificate
vcv_pinned_certificate_days_until_expiry{common_name="api.production.local"}

# All pinned certificates expiring in next 30 days
vcv_pinned_certificate_days_until_expiry < 30

# Pinned certificate status
vcv_pinned_certificate_expiry_timestamp_seconds{status="valid"}
```

### Security and compliance

```promql
# Total weak keys across all vaults
sum(vcv_certificates_weak_keys_total)

# Weak keys per vault
vcv_certificates_weak_keys_total{vault_id!="__all__"}

# RSA key size distribution
sum by (key_size) (vcv_certificates_by_key_type_total{algorithm="RSA"})

# Certificates by issuer
topk(10, sum by (issuer_cn) (vcv_certificates_by_issuer_total))

# Percentage of certificates with weak keys
sum(vcv_certificates_weak_keys_total) / sum(vcv_certificates_total{status="valid"}) * 100
```

### Certificate lifecycle analysis

```promql
# Certificate age distribution
sum by (bucket) (vcv_certificates_age_bucket{vault_id="__all__", pki="__all__"})

# Certificates older than 1 year
sum(vcv_certificates_age_bucket{bucket="1y+"})

# Recent issuance activity (last 24h)
sum(vcv_certificates_issued_last_24h)

# Weekly renewal rate
sum(vcv_certificates_issued_last_7d) / 7

# Monthly renewal trend
sum(vcv_certificates_issued_last_30d)

# Renewal rate per vault
sum by (vault_id) (vcv_certificates_issued_last_7d{vault_id!="__all__"})
```

### SAN analysis

```promql
# Certificates with SANs
sum(vcv_certificates_with_sans_total)

# SAN count distribution
sum by (bucket) (vcv_certificates_san_count_bucket{vault_id="__all__", pki="__all__"})

# Multi-domain certificates (6+ SANs)
sum(vcv_certificates_san_count_bucket{bucket=~"6-10|11+"})

# Percentage of certificates with SANs
sum(vcv_certificates_with_sans_total) / sum(vcv_certificates_total{status="valid"}) * 100
```

## Alert rules

### Critical alerts

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

### Warning alerts

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

- alert: VCVWeakKeysDetected
  expr: sum(vcv_certificates_weak_keys_total) > 0
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Weak cryptographic keys detected"
    description: "{{ $value }} certificates with weak keys detected. Review and rotate immediately."

- alert: VCVPinnedCertificateExpiring
  expr: vcv_pinned_certificate_days_until_expiry < 30
  labels:
    severity: warning
  annotations:
    summary: "Pinned certificate expiring soon"
    description: "Critical certificate '{{ $labels.common_name }}' expires in {{ $value }} days."

- alert: VCVPinnedCertificateCritical
  expr: vcv_pinned_certificate_days_until_expiry < 7
  labels:
    severity: critical
  annotations:
    summary: "Pinned certificate expiring critically soon"
    description: "Critical certificate '{{ $labels.common_name }}' expires in {{ $value }} days!"

- alert: VCVRenewalAnomalyDetected
  expr: rate(vcv_certificates_issued_last_24h[1h]) < 0.5 * rate(vcv_certificates_issued_last_24h[24h] offset 7d)
  for: 2h
  labels:
    severity: warning
  annotations:
    summary: "Certificate renewal rate anomaly"
    description: "Certificate issuance rate has dropped significantly. Check automation."

- alert: VCVStaleCertificates
  expr: sum(vcv_certificates_age_bucket{bucket="1y+"}) > 100
  for: 1h
  labels:
    severity: info
  annotations:
    summary: "High number of stale certificates"
    description: "{{ $value }} certificates are older than 1 year. Consider cleanup."
```

## Grafana dashboard queries

### Certificate status overview

```promql
# Donut chart - Certificate status distribution
sum by (status) (vcv_certificates_total{vault_id="__all__", pki="__all__"})
```

### Expiration timeline

```promql
# Bar chart - Certificates by expiry bucket
sum by (bucket) (vcv_certificates_expiry_bucket{vault_id="__all__", pki="__all__"})
```

### Vault health matrix

```promql
# Table - Vault connectivity and certificate counts
sum by (vault_id) (vcv_certificates_total{vault_id!="__all__"})
```

### Threshold compliance

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

## Best practices

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
