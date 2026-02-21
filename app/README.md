# VaultCertsViewer — Technical overview

This document describes the technical structure of VaultCertsViewer (vcv), a single Go binary that embeds a static HTML/CSS/JS UI to browse and manage certificates from one or more HashiCorp Vault or OpenBao PKI mounts.

## Architecture

- **Backend**: Go + chi router, Vault/OpenBao client (`github.com/hashicorp/vault/api`), zerolog-based logging.
- **Frontend**: Plain `index.html`, `styles.css`, `app-htmx.js` served from the embedded filesystem (no Node/bundler).
- **Binary layout**: `app/cmd/server` embeds `/web` assets via Go `embed`; a single executable serves both API and UI.
- **HTMX Integration**: Certificate UI fragments under `/ui/*` and optional Admin panel under `/admin/*`.
- **Security**: Rate limiting (prod only), CSRF protection, security headers, request ID tracking, body size limits.
- **OpenBao compatibility**: Uses the same Vault client library which works with both HashiCorp Vault and OpenBao due to API compatibility.

## Directory layout (app/)

- `cmd/server/main.go` — entrypoint, router, middleware, static file serving, graceful shutdown.
- `internal/cache/` — simple in-memory TTL cache (used by Vault client).
- `internal/config/` — environment-backed configuration loading with expiration threshold support.
- `internal/errors/` — custom error types and helpers.
- `internal/handlers/` — HTTP handlers (`certs`, `i18n`, `health`, `ready`, `ui`, `admin` routes).
- `internal/httputil/` - HTTP utility functions for client IP extraction, used by rate limiting middleware.
- `internal/i18n/` - Internationalization support.
- `internal/logger/` — zerolog initialization and structured helpers (HTTP events, panic).
- `internal/metrics/` — Prometheus collectors.
- `internal/middleware/` — request ID, HTTP logging, panic recovery, CORS, security headers, rate limiting, CSRF protection, body limit.
- `internal/validation/` - Validation helpers.
- `internal/vault/` — Vault client implementations with graceful shutdown support.
- `internal/version/` — build version info (injected via ldflags).
- `web/` — `index.html`, `assets/app-htmx.js`, `assets/styles.css`, `templates/` (UI fragments + Admin templates).

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
| `/ui/vaults/status` | GET | HTMX fragment: vault status |
| `/ui/vaults/refresh` | POST | HTMX fragment: refresh vaults |
| `/admin` | GET | Admin page (enabled only if admin password is configured in settings.json) |
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
- The application will automatically discover the settings file.

If you enable the Admin panel, the settings file must be writable so changes can be persisted.

### Resolution order

Configuration is loaded from the first file found in this order:

- `settings.dev.json`
- `settings.prod.json`
- `settings.json`
- `./settings.json`
- `/app/settings.json`

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
- **Rate limiting**: Production-only (300 requests/minute, exempting health/ready/metrics).
- **CSRF protection**: Required for all state-changing requests.
- **Security headers**: HSTS, X-Frame-Options, X-Content-Type-Options, CSP.
- **Request ID tracking**: Unique IDs for all requests for log correlation.
- **Body size limits**: 1MB maximum request body.
- **Graceful shutdown**: Proper cleanup of HTTP server and background goroutines.

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
- Configurable pagination (25/50/100/all)
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
- `vcv_expiration_threshold_critical_days` — Configured critical threshold.
- `vcv_expiration_threshold_warning_days` — Configured warning threshold.

### Enhanced metrics

- `vcv_certificates_expiry_bucket{vault_id, pki, bucket}` — Certificate distribution by time bucket.
  - Buckets: `0-7d`, `7-30d`, `30-90d`, `90d+`, `expired`, `revoked`.

### Optional metrics (disabled by default)

- `vcv_certificate_expiry_timestamp_seconds{certificate_id, common_name, status, vault_id, pki}`
- `vcv_certificate_days_until_expiry{certificate_id, common_name, status, vault_id, pki}` — Days until expiration (negative if expired).

Enable with:

```json
"metrics": {
  "per_certificate": true,
  "enhanced_metrics": false
}
```

## Build & run

### Production

See README.md on the root path for production deployment instructions.

### Development

A Hashicorp Vault or OpenBao server is required to run the application in development mode. Thus, a container with an init script is provided in `docker-compose.dev.yml`. It will initialize a Vault server with a PKI mount and some certs.

```bash
task dev
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

- `task test-offline`: Run tests without Vault dependency
- `task test-dev`: Run tests against dev Vault instance
