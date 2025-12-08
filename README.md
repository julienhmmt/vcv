# VaultCertsViewer

VaultCertsViewer (vcv) is a lightweight web UI that lists and inspects certificates stored in a HashiCorp Vault PKI mount, especially their expiration dates and SANs.

Currently, VaultCertsViewer (vcv) can only view one mount at a time. If you have (for example) 4 mounts, you'll need 4 instances of vcv.

## What it does

- Discovers all certificates in a Vault PKI and shows them in a searchable, filterable table.
- Shows common names (CN) and SANs.
- Displays status distribution (valid / expired / revoked) and upcoming expirations.
- Highlights certificates expiring soon (7/30 days) and shows details (CN, SAN, fingerprints, issuer, validity).
- Lets you pick UI language (en, fr, es, de, it) and theme (light/dark).

## Why it exists

The native Vault UI is heavy and not convenient for quickly checking certificate expirations and details. VaultCertsViewer gives platform / security / ops teams a fast, **read-only** view of the Vault PKI inventory with only the essential information.

## Who should use it

- Teams operating Vault PKI who need visibility on their certificates.
- Operators who want a ready-to-use browser view alongside Vault CLI or Web UI.

## How to deploy and use

In HashiCorp Vault, create a read-only role and token for the API to reach the target PKI engine (adjust `pki` if you use another mount):

```bash
vault policy write vcv - <<'EOF'
path "pki/certs"    { capabilities = ["list"] }
path "pki/certs/*"  { capabilities = ["read","list"] }
path "sys/health"   { capabilities = ["read"] }
EOF
vault write auth/token/roles/vcv allowed_policies="vcv" orphan=true period="24h"
vault token create -role="vcv" -policy="vcv" -period="24h" -renewable=true
```

This dedicated token limits permissions to certificate listing/reading, can be renewed, and is used as `VAULT_READ_TOKEN` by the app.

### docker-compose

Grab `docker-compose.yml`, put it in a directory and create `.env` file with these variables:

```text
VAULT_ADDR=<you vault address>
VAULT_READ_TOKEN=<previously generated token>
VAULT_PKI_MOUNT=<pki engine name>
```

then launch instance:

```bash
docker compose up -d
```

No storage needed, unless you want to log to a file.

### docker run

Start the container with this command:

```bash
docker run -d \
  -e "APP_ENV=prod" \
  -e "LOG_FORMAT=json" \
  -e "LOG_OUTPUT=stdout" \
  -e "VAULT_ADDR=http://changeme:8200" \
  -e "VAULT_READ_TOKEN=changeme" \
  -e "VAULT_PKI_MOUNT=changeme" \
  -e "VAULT_TLS_INSECURE=true" \
  -e "LOG_LEVEL=info" \
  --cap-drop=ALL --read-only --security-opt no-new-privileges:true \
  -p 52000:52000 jhmmt/vcv:1.1
```

## Translations

The UI is localized in English, French, Spanish, German, and Italian. Language is selectable in the header or via `?lang=xx`.

## Export metrics to Prometheus

Metrics are exposed at `/metrics` endpoint.

- vcv_cache_size
- vcv_certificate_expiry_timestamp_seconds{serial_number, common_name, status}
- vcv_certificate_exporter_last_scrape_success
- vcv_certificates_expired_count
- vcv_certificates_last_fetch_timestamp_seconds
- vcv_certificates_total{status}
- vcv_vault_connected

To scrape metrics, add this to your Prometheus config:

```yaml
scrape_configs:
  - job_name: vcv
    static_configs:
      - targets: ['localhost:52000']
    metrics_path: /metrics
```

Example scrape output (truncated):

```bash
$ curl -v http://localhost:52000/metrics
...
# HELP vcv_cache_size Number of items currently cached
# TYPE vcv_cache_size gauge
vcv_cache_size 0
# HELP vcv_certificate_expiry_timestamp_seconds Certificate expiration timestamp in seconds since epoch
# TYPE vcv_certificate_expiry_timestamp_seconds gauge
vcv_certificate_expiry_timestamp_seconds{common_name="api.internal",serial_number="52:e3:c0:23:ba:f4:51:ae:1b:59:24:4a:d1:03:e1:a7:8a:96:a7:80",status="active"} 1.767710142e+09
vcv_certificate_expiry_timestamp_seconds{common_name="example.internal",serial_number="35:1b:ff:d3:e2:f3:53:14:b1:7f:9e:d3:77:a6:25:72:a2:63:15:99",status="active"} 1.767710142e+09
vcv_certificate_expiry_timestamp_seconds{common_name="expired.internal",serial_number="74:5a:ed:76:98:b1:c8:e3:d7:a5:bb:a2:67:7f:f6:4f:2a:31:48:18",status="active"} 1.765118144e+09
vcv_certificate_expiry_timestamp_seconds{common_name="expiring-soon.internal",serial_number="36:c6:0b:ef:2c:a5:2f:08:89:6a:13:fe:2a:9e:43:84:38:a4:a9:af",status="active"} 1.765204542e+09
vcv_certificate_expiry_timestamp_seconds{common_name="expiring-week.internal",serial_number="47:c9:8f:71:2a:d7:14:49:96:64:af:d6:15:ec:e9:86:a6:59:cf:26",status="active"} 1.765722942e+09
vcv_certificate_expiry_timestamp_seconds{common_name="revoked.internal",serial_number="2d:08:41:de:10:5a:21:0e:63:0d:5d:8e:f9:4e:ce:4b:7b:31:2e:2d",status="revoked"} 1.767710145e+09
vcv_certificate_expiry_timestamp_seconds{common_name="vcv.local",serial_number="48:88:7a:6a:65:85:85:8b:0a:2a:12:7f:a7:6f:dc:62:3a:f2:7a:ba",status="active"} 1.796654141e+09
# HELP vcv_certificate_exporter_last_scrape_success Whether the last scrape succeeded (1) or failed (0)
# TYPE vcv_certificate_exporter_last_scrape_success gauge
vcv_certificate_exporter_last_scrape_success 1
# HELP vcv_certificates_expired_count Number of expired certificates
# TYPE vcv_certificates_expired_count gauge
vcv_certificates_expired_count 1
# HELP vcv_certificates_expires_soon_count Number of certificates expiring soon within threshold window
# TYPE vcv_certificates_expires_soon_count gauge
vcv_certificates_expires_soon_count 4
# HELP vcv_certificates_last_fetch_timestamp_seconds Timestamp of last successful certificates fetch
# TYPE vcv_certificates_last_fetch_timestamp_seconds gauge
vcv_certificates_last_fetch_timestamp_seconds 1.765118171e+09
# HELP vcv_certificates_total Total certificates grouped by status
# TYPE vcv_certificates_total gauge
vcv_certificates_total{status="active"} 6
vcv_certificates_total{status="revoked"} 1
# HELP vcv_vault_connected Vault connection status (1=connected,0=disconnected)
# TYPE vcv_vault_connected gauge
vcv_vault_connected 1
```

If you are using AlertManager, you can create alerts based on these metrics. For example, using only the expiry timestamp and generic counters:

```yaml
- alert: VCVExpiredCerts
  expr: vcv_certificates_expired_count > 0

- alert: VCVExpiringSoon_14d
  expr: (vcv_certificate_expiry_timestamp_seconds - time()) / 86400 < 14

- alert: VCVStaleData
  expr: time() - vcv_certificates_last_fetch_timestamp_seconds > 3600

- alert: VCVVaultDown
  expr: vcv_vault_connected == 0
```

You can adjust the "soon" window (here 14 days) directly in PromQL without changing the exporter.

## More details

- Technical documentation: [app/README.md](app/README.md)
- French overview: [README.fr.md](README.fr.md)

## Picture of the app

<img width="3024" height="2807" alt="VaultCertsViewer" src="https://github.com/user-attachments/assets/8b097046-d921-4b1d-a270-f86e8be5fc36" />
