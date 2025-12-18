# VaultCertsViewer — Technical overview

This document describes the technical structure of VaultCertsViewer (vcv), a single Go binary that embeds a static HTML/CSS/JS UI to browse and manage certificates from one or more HashiCorp Vault PKI mounts.

## Architecture

- **Backend**: Go + chi router, Vault client (`github.com/hashicorp/vault/api`), zerolog-based logging.
- **Frontend**: Plain `index.html`, `styles.css`, `app-htmx.js` served from the embedded filesystem (no Node/bundler).
- **Binary layout**: `app/cmd/server` embeds `/web` assets via Go `embed`; a single executable serves both API and UI.
- **HTMX Integration**: Certificate UI fragments under `/ui/*` and optional Admin panel under `/admin/*`.

## Directory layout (app/)

- `cmd/server/main.go` — entrypoint, router, middleware, static file serving, graceful shutdown.
- `cmd/server/web/` — `index.html`, `assets/app-htmx.js`, `assets/styles.css`, `templates/` (UI fragments + Admin templates).
- `config/` — environment-backed configuration loading with expiration threshold support.
- `internal/cache/` — simple in-memory TTL cache (used by Vault client).
- `internal/handlers/` — HTTP handlers (`certs`, `i18n`, `health`, `ready`, `ui` routes).
- `internal/metrics/` — Prometheus collectors.
- `internal/logger/` — zerolog initialization and structured helpers (HTTP events, panic).
- `internal/vault/` — Vault client implementations with graceful shutdown support.
- `internal/version/` — build version info (injected via ldflags).
- `middleware/` — request ID, HTTP logging, panic recovery, CORS, security headers.

## API surface

| Endpoint | Method | Description |
| ---------- | -------- | ------------- |
| `/` | GET | Embedded UI |
| `/api/cache/invalidate` | POST | Clear Vault cache |
| `/api/certs/{id}/details` | GET | Detailed certificate view |
| `/api/certs/{id}/pem` | GET | PEM content |
| `/api/certs/{id}/pem/download` | GET | Download PEM content |
| `/api/certs` | GET | List certificates |
| `/api/config` | GET | Application configuration (thresholds) |
| `/api/health` | GET | Liveness probe |
| `/api/i18n` | GET | UI translations (lang via query param) |
| `/api/ready` | GET | Readiness probe |
| `/api/status` | GET | Vault connection status (per vault) |
| `/api/version` | GET | Application version info |
| `/metrics` | GET | Prometheus metrics |
| `/ui/certs` | GET | HTMX fragment: certificates table + dashboard |
| `/ui/certs/refresh` | POST | HTMX fragment: refresh certificates |
| `/ui/certs/{id}/details` | GET | HTMX fragment: certificate details |
| `/ui/theme/toggle` | POST | Toggle dark/light theme |
| `/ui/status` | GET | Real-time Vault connection status |
| `/admin` | GET | Admin page (enabled only if `VCV_ADMIN_PASSWORD` is set to a bcrypt hash) |
| `/admin/panel` | GET | Admin panel fragment (HTMX) |
| `/admin/login` | POST | Admin login (HTMX) |
| `/admin/logout` | POST | Admin logout (HTMX) |
| `/api/admin/login` | POST | Admin login (JSON) |
| `/api/admin/logout` | POST | Admin logout (JSON) |
| `/api/admin/settings` | GET/PUT | Admin settings (JSON, requires auth) |

## Configuration (settings.json)

The primary configuration source is a JSON file.

Recommended deployment pattern:

- Mount a `settings.json` file into the container under `/app/settings.json` (the image `WORKDIR` is `/app`).
- The application will automatically discover the settings file without requiring `SETTINGS_PATH`.

If you enable the Admin panel, the settings file must be writable so changes can be persisted.

### Resolution order

Configuration is loaded in this order:

- `SETTINGS_PATH` if set
- `settings.<APP_ENV>.json` (default `APP_ENV=dev`)
- `settings.json`
- `./settings.json`
- `/etc/vcv/settings.json`

### Schema overview

- `app.env`, `app.port`
- `app.logging.level`, `app.logging.format`, `app.logging.output`, `app.logging.file_path`
- `cors.allowed_origins`, `cors.allow_credentials`
- `certificates.expiration_thresholds.critical`, `certificates.expiration_thresholds.warning`
- `vaults[]`: list of Vault instances
  - `address`, `token`, `tls_insecure`
  - `pki_mounts` (recommended)
  - `pki_mount` (backward-compatible alias)
  - `tls_ca_cert_base64` (preferred; base64-encoded PEM CA bundle)
  - `tls_ca_cert` (file path to a PEM CA bundle)
  - `tls_ca_path` (directory containing CA certs)
  - `tls_server_name` (SNI override)

### Multi-Vault model

The configuration supports defining multiple Vault instances (`vaults[]`).

- The first enabled Vault instance becomes the primary Vault.

### Legacy env fallback

For backward compatibility, configuration can still be provided via environment variables if no settings file is found.

Key legacy variables include `VAULT_ADDR`, `VAULT_READ_TOKEN`, `VAULT_PKI_MOUNTS`, `VCV_EXPIRE_CRITICAL`, `VCV_EXPIRE_WARNING`.

TLS-related env variables include `VAULT_TLS_INSECURE` (or `VAULT_SKIP_VERIFY`), `VAULT_TLS_CA_CERT_BASE64` (or `VAULT_CACERT_BYTES`), `VAULT_TLS_CA_CERT` (or `VAULT_CACERT`), `VAULT_TLS_CA_PATH` (or `VAULT_CAPATH`), and `VAULT_TLS_SERVER_NAME`.

Precedence rules:

- If `tls_ca_cert_base64` is set, it is used and `tls_ca_cert` / `tls_ca_path` are ignored.
- Otherwise, `tls_ca_cert` / `tls_ca_path` are used (if set).

### Logging initialization note

`internal/logger.Init` reads output settings from environment variables (`LOG_OUTPUT`, `LOG_FORMAT`, `LOG_FILE_PATH`). When loading from `settings.json`, `config.Load` sets those env vars to ensure logging matches file configuration.

## Security

- **Container hardening**: read-only filesystem, no-new-privileges, dropped capabilities.
- **Graceful shutdown**: proper cleanup of HTTP server and background goroutines.

## Logging

- Initialized in `cmd/server/main.go` via `internal/logger.Init`.
- Middlewares emit structured HTTP events with `request_id`, status, duration.
- Handlers use `HTTPEvent`/`HTTPError` helpers; panic recovery logs stack traces.
- Version info logged at startup.

## Internationalization

- Languages: en, fr, es, de, it.
- `/api/i18n` returns messages; the UI selects language via header dropdown or `?lang=xx`.
- Short day labels (`daysRemainingShort`) and expiry filters are translated.
- Toast notifications for Vault connection status are fully translated.

## Frontend Features

### User Experience

- Real-time search with debouncing (300ms)
- Visual loading indicators on refresh button
- Certificate status badges (valid/expired/revoked)
- Vault connection monitoring with toast notifications
- Responsive design with sticky header
- Dark/light theme persistence
- Modal mount selector for multi-PKI support
- Configurable pagination (25/50/75/100/all)
- Sortable columns with visual indicators

## Metrics

Metrics are exposed at `/metrics`.

### Cardinality guidance

- Aggregated metrics are safe for large inventories (multi-vault, multi-PKI).
- The per-certificate expiry metric is disabled by default to avoid high cardinality. Enable it only when you really need it.

### Exported metrics

- `vcv_certificate_exporter_last_scrape_success`
- `vcv_certificate_exporter_last_scrape_duration_seconds`
- `vcv_certificates_last_fetch_timestamp_seconds`
- `vcv_cache_size`
- `vcv_vault_connected{vault_id}`
  - `vault_id="__all__"` reflects the overall connection status.
  - `vault_id="<id>"` reflects per-vault status when the app is configured with multiple vaults.
- `vcv_vault_list_certificates_success{vault_id}`
- `vcv_vault_list_certificates_error{vault_id}`
- `vcv_vault_list_certificates_duration_seconds{vault_id}`
- `vcv_certificates_partial_scrape{vault_id}`
- `vcv_certificates_total{vault_id, pki, status}`
  - `status`: `valid`, `expired`, `revoked`
  - `vault_id`, `pki` support `__all__` for global totals.
- `vcv_certificates_expiring_soon_count{vault_id, pki, level}`
  - `level`: `warning`, `critical`
  - Uses the configured expiration thresholds.
- `vcv_vaults_configured`
- `vcv_pki_mounts_configured{vault_id}`

### Optional metric (disabled by default)

- `vcv_certificate_expiry_timestamp_seconds{certificate_id, common_name, status, vault_id, pki}`

Enable it with:

```bash
VCV_METRICS_PER_CERTIFICATE=true
```

Example:

```bash
# HELP vcv_cache_size Number of items currently cached
# TYPE vcv_cache_size gauge
vcv_cache_size 0
# HELP vcv_certificate_exporter_last_scrape_duration_seconds Duration of the last certificate scrape in seconds
# TYPE vcv_certificate_exporter_last_scrape_duration_seconds gauge
vcv_certificate_exporter_last_scrape_duration_seconds 0.000118208
# HELP vcv_certificate_exporter_last_scrape_success Whether the last scrape succeeded (1) or failed (0)
# TYPE vcv_certificate_exporter_last_scrape_success gauge
vcv_certificate_exporter_last_scrape_success 1
# HELP vcv_certificates_expired_count Number of expired certificates
# TYPE vcv_certificates_expired_count gauge
vcv_certificates_expired_count 30
# HELP vcv_certificates_expiring_soon_count Number of certificates expiring soon within threshold window
# TYPE vcv_certificates_expiring_soon_count gauge
vcv_certificates_expiring_soon_count{level="critical",pki="__all__",vault_id="__all__"} 17
vcv_certificates_expiring_soon_count{level="critical",pki="pki",vault_id="vault-main"} 3
vcv_certificates_expiring_soon_count{level="critical",pki="pki_blockchain",vault_id="vault-dev-3"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_cloud",vault_id="vault-dev-3"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_corporate",vault_id="vault-dev-2"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_dev",vault_id="vault-main"} 1
vcv_certificates_expiring_soon_count{level="critical",pki="pki_dmz",vault_id="vault-dev-5"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_edge",vault_id="vault-dev-3"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_external",vault_id="vault-dev-2"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_internal",vault_id="vault-dev-5"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_iot",vault_id="vault-dev-3"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_lab",vault_id="vault-dev-4"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_partners",vault_id="vault-dev-2"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_perf",vault_id="vault-dev-4"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_production",vault_id="vault-main"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_qa",vault_id="vault-dev-4"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_shared",vault_id="vault-dev-5"} 0
vcv_certificates_expiring_soon_count{level="critical",pki="pki_stage",vault_id="vault-main"} 1
vcv_certificates_expiring_soon_count{level="critical",pki="pki_vault2",vault_id="vault-dev-2"} 2
vcv_certificates_expiring_soon_count{level="critical",pki="pki_vault3",vault_id="vault-dev-3"} 2
vcv_certificates_expiring_soon_count{level="critical",pki="pki_vault4",vault_id="vault-dev-4"} 4
vcv_certificates_expiring_soon_count{level="critical",pki="pki_vault5",vault_id="vault-dev-5"} 4
vcv_certificates_expiring_soon_count{level="warning",pki="__all__",vault_id="__all__"} 45
vcv_certificates_expiring_soon_count{level="warning",pki="pki",vault_id="vault-main"} 7
vcv_certificates_expiring_soon_count{level="warning",pki="pki_blockchain",vault_id="vault-dev-3"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_cloud",vault_id="vault-dev-3"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_corporate",vault_id="vault-dev-2"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_dev",vault_id="vault-main"} 2
vcv_certificates_expiring_soon_count{level="warning",pki="pki_dmz",vault_id="vault-dev-5"} 5
vcv_certificates_expiring_soon_count{level="warning",pki="pki_edge",vault_id="vault-dev-3"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_external",vault_id="vault-dev-2"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_internal",vault_id="vault-dev-5"} 5
vcv_certificates_expiring_soon_count{level="warning",pki="pki_iot",vault_id="vault-dev-3"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_lab",vault_id="vault-dev-4"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_partners",vault_id="vault-dev-2"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_perf",vault_id="vault-dev-4"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_production",vault_id="vault-main"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_qa",vault_id="vault-dev-4"} 6
vcv_certificates_expiring_soon_count{level="warning",pki="pki_shared",vault_id="vault-dev-5"} 0
vcv_certificates_expiring_soon_count{level="warning",pki="pki_stage",vault_id="vault-main"} 2
vcv_certificates_expiring_soon_count{level="warning",pki="pki_vault2",vault_id="vault-dev-2"} 5
vcv_certificates_expiring_soon_count{level="warning",pki="pki_vault3",vault_id="vault-dev-3"} 5
vcv_certificates_expiring_soon_count{level="warning",pki="pki_vault4",vault_id="vault-dev-4"} 4
vcv_certificates_expiring_soon_count{level="warning",pki="pki_vault5",vault_id="vault-dev-5"} 4
# HELP vcv_certificates_last_fetch_timestamp_seconds Timestamp of last successful certificates fetch
# TYPE vcv_certificates_last_fetch_timestamp_seconds gauge
vcv_certificates_last_fetch_timestamp_seconds 1.765985686e+09
# HELP vcv_certificates_total Total certificates grouped by status
# TYPE vcv_certificates_total gauge
vcv_certificates_total{pki="__all__",status="expired",vault_id="__all__"} 30
vcv_certificates_total{pki="__all__",status="revoked",vault_id="__all__"} 14
vcv_certificates_total{pki="__all__",status="valid",vault_id="__all__"} 85
vcv_certificates_total{pki="pki",status="expired",vault_id="vault-main"} 3
vcv_certificates_total{pki="pki",status="revoked",vault_id="vault-main"} 0
vcv_certificates_total{pki="pki",status="valid",vault_id="vault-main"} 12
vcv_certificates_total{pki="pki_blockchain",status="expired",vault_id="vault-dev-3"} 0
vcv_certificates_total{pki="pki_blockchain",status="revoked",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_blockchain",status="valid",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_cloud",status="expired",vault_id="vault-dev-3"} 0
vcv_certificates_total{pki="pki_cloud",status="revoked",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_cloud",status="valid",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_corporate",status="expired",vault_id="vault-dev-2"} 0
vcv_certificates_total{pki="pki_corporate",status="revoked",vault_id="vault-dev-2"} 1
vcv_certificates_total{pki="pki_corporate",status="valid",vault_id="vault-dev-2"} 1
vcv_certificates_total{pki="pki_dev",status="expired",vault_id="vault-main"} 1
vcv_certificates_total{pki="pki_dev",status="revoked",vault_id="vault-main"} 2
vcv_certificates_total{pki="pki_dev",status="valid",vault_id="vault-main"} 5
vcv_certificates_total{pki="pki_dmz",status="expired",vault_id="vault-dev-5"} 0
vcv_certificates_total{pki="pki_dmz",status="revoked",vault_id="vault-dev-5"} 0
vcv_certificates_total{pki="pki_dmz",status="valid",vault_id="vault-dev-5"} 6
vcv_certificates_total{pki="pki_edge",status="expired",vault_id="vault-dev-3"} 0
vcv_certificates_total{pki="pki_edge",status="revoked",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_edge",status="valid",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_external",status="expired",vault_id="vault-dev-2"} 0
vcv_certificates_total{pki="pki_external",status="revoked",vault_id="vault-dev-2"} 1
vcv_certificates_total{pki="pki_external",status="valid",vault_id="vault-dev-2"} 1
vcv_certificates_total{pki="pki_internal",status="expired",vault_id="vault-dev-5"} 0
vcv_certificates_total{pki="pki_internal",status="revoked",vault_id="vault-dev-5"} 1
vcv_certificates_total{pki="pki_internal",status="valid",vault_id="vault-dev-5"} 6
vcv_certificates_total{pki="pki_iot",status="expired",vault_id="vault-dev-3"} 0
vcv_certificates_total{pki="pki_iot",status="revoked",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_iot",status="valid",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_lab",status="expired",vault_id="vault-dev-4"} 0
vcv_certificates_total{pki="pki_lab",status="revoked",vault_id="vault-dev-4"} 0
vcv_certificates_total{pki="pki_lab",status="valid",vault_id="vault-dev-4"} 7
vcv_certificates_total{pki="pki_partners",status="expired",vault_id="vault-dev-2"} 0
vcv_certificates_total{pki="pki_partners",status="revoked",vault_id="vault-dev-2"} 1
vcv_certificates_total{pki="pki_partners",status="valid",vault_id="vault-dev-2"} 1
vcv_certificates_total{pki="pki_perf",status="expired",vault_id="vault-dev-4"} 0
vcv_certificates_total{pki="pki_perf",status="revoked",vault_id="vault-dev-4"} 0
vcv_certificates_total{pki="pki_perf",status="valid",vault_id="vault-dev-4"} 1
vcv_certificates_total{pki="pki_production",status="expired",vault_id="vault-main"} 0
vcv_certificates_total{pki="pki_production",status="revoked",vault_id="vault-main"} 0
vcv_certificates_total{pki="pki_production",status="valid",vault_id="vault-main"} 1
vcv_certificates_total{pki="pki_qa",status="expired",vault_id="vault-dev-4"} 0
vcv_certificates_total{pki="pki_qa",status="revoked",vault_id="vault-dev-4"} 0
vcv_certificates_total{pki="pki_qa",status="valid",vault_id="vault-dev-4"} 7
vcv_certificates_total{pki="pki_shared",status="expired",vault_id="vault-dev-5"} 0
vcv_certificates_total{pki="pki_shared",status="revoked",vault_id="vault-dev-5"} 0
vcv_certificates_total{pki="pki_shared",status="valid",vault_id="vault-dev-5"} 6
vcv_certificates_total{pki="pki_stage",status="expired",vault_id="vault-main"} 1
vcv_certificates_total{pki="pki_stage",status="revoked",vault_id="vault-main"} 0
vcv_certificates_total{pki="pki_stage",status="valid",vault_id="vault-main"} 5
vcv_certificates_total{pki="pki_vault2",status="expired",vault_id="vault-dev-2"} 5
vcv_certificates_total{pki="pki_vault2",status="revoked",vault_id="vault-dev-2"} 1
vcv_certificates_total{pki="pki_vault2",status="valid",vault_id="vault-dev-2"} 6
vcv_certificates_total{pki="pki_vault3",status="expired",vault_id="vault-dev-3"} 5
vcv_certificates_total{pki="pki_vault3",status="revoked",vault_id="vault-dev-3"} 1
vcv_certificates_total{pki="pki_vault3",status="valid",vault_id="vault-dev-3"} 6
vcv_certificates_total{pki="pki_vault4",status="expired",vault_id="vault-dev-4"} 7
vcv_certificates_total{pki="pki_vault4",status="revoked",vault_id="vault-dev-4"} 1
vcv_certificates_total{pki="pki_vault4",status="valid",vault_id="vault-dev-4"} 5
vcv_certificates_total{pki="pki_vault5",status="expired",vault_id="vault-dev-5"} 8
vcv_certificates_total{pki="pki_vault5",status="revoked",vault_id="vault-dev-5"} 1
vcv_certificates_total{pki="pki_vault5",status="valid",vault_id="vault-dev-5"} 5
# HELP vcv_vault_connected Vault connection status (1=connected,0=disconnected)
# TYPE vcv_vault_connected gauge
vcv_vault_connected{vault_id="__all__"} 0
vcv_vault_connected{vault_id="vault-dev-2"} 1
vcv_vault_connected{vault_id="vault-dev-3"} 1
vcv_vault_connected{vault_id="vault-dev-4"} 1
vcv_vault_connected{vault_id="vault-dev-5"} 1
vcv_vault_connected{vault_id="vault-dev-6"} 0
vcv_vault_connected{vault_id="vault-main"} 1
```

## Build & run

### Production

See README.md on the root path for production deployment instructions.

### Development

A HashiCorp Vault server is required to run the application in development mode. Thus, a container with an init script is provided in `docker-compose.dev.yml`. It will initialize a Vault server with a PKI mount and some certs.

```bash
make dev
```

Binary serves UI and API at `http://localhost:52000`.

## Testing

```bash
cd app && go test ./...
```

### Test Coverage

- Unit tests for all major packages
- Mock Vault client for offline testing
- HTTP handler tests with httptest.Server
- Configuration validation tests
- Internationalization tests

Test targets:

- `make test-offline`: Run tests without Vault dependency
- `make test-dev`: Run tests against dev Vault instance

## Development notes

- No external frontend toolchain; edit `app-htmx.js`/`styles.css` directly.
- Request IDs are added to all responses; include them when correlating logs.
- Use `VAULT_TLS_INSECURE=true` only in development environments.
- HTMX partial templates are in `cmd/server/web/templates/`.
- JavaScript uses modern ES6+ features with browser-native APIs.
- CSS uses custom properties for theming and responsive design.

## Performance Considerations

- In-memory caching with configurable TTL (default 5 minutes)
- Request deduplication for concurrent identical requests
- Efficient DOM updates via HTMX partial swapping
- Lazy loading of certificate details
- Optimized search with client-side filtering
