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
| `/admin` | GET | Admin page (enabled only if `VCV_ADMIN_PASSWORD` is set) |
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

### Multi-Vault model

The configuration supports defining multiple Vault instances (`vaults[]`).

- The first enabled Vault instance becomes the primary Vault.

### Legacy env fallback

For backward compatibility, configuration can still be provided via environment variables if no settings file is found.

Key legacy variables include `VAULT_ADDR`, `VAULT_READ_TOKEN`, `VAULT_PKI_MOUNTS`, `VCV_EXPIRE_CRITICAL`, `VCV_EXPIRE_WARNING`.

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
