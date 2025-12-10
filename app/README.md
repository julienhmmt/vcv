# VaultCertsViewer — Technical overview

This document describes the technical structure of VaultCertsViewer (vcv), a single Go binary that embeds a static HTML/CSS/JS UI to browse and manage certificates from one or more HashiCorp Vault PKI mounts.

## Architecture

- **Backend**: Go + chi router, Vault client (`github.com/hashicorp/vault/api`), zerolog-based logging.
- **Frontend**: Plain `index.html`, `styles.css`, `app.js` served from the embedded filesystem (no Node/bundler).
- **Binary layout**: `app/cmd/server` embeds `/web` assets via Go `embed`; a single executable serves both API and UI.

## Directory layout (app/)

- `cmd/server/main.go` — entrypoint, router, middleware, static file serving, graceful shutdown.
- `cmd/server/web/` — `index.html`, `assets/app.js`, `assets/styles.css`.
- `config/` — environment-backed configuration loading.
- `internal/cache/` — simple in-memory TTL cache (used by Vault client).
- `internal/handlers/` — HTTP handlers (`certs`, `i18n`, `health`, `ready`).
- `internal/logger/` — zerolog initialization and structured helpers (HTTP events, panic).
- `internal/vault/` — Vault client implementations with graceful shutdown support.
- `internal/version/` — build version info (injected via ldflags).
- `middleware/` — request ID, HTTP logging, panic recovery, CORS, security headers.

## API surface

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Embedded UI |
| `/api/cache/invalidate` | POST | Clear Vault cache |
| `/api/certs/{id}/details` | GET | Detailed certificate view |
| `/api/certs/{id}/pem` | GET | PEM content |
| `/api/certs` | GET | List certificates |
| `/api/crl/download` | GET | Download CRL |
| `/api/crl/rotate` | POST | Rotate CRL |
| `/api/health` | GET | Liveness probe |
| `/api/i18n` | GET | UI translations (lang via query param) |
| `/api/ready` | GET | Readiness probe |
| `/api/version` | GET | Application version info |

## Configuration (env vars)

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ENV` | `dev` | Environment: `dev`, `stage`, `prod` |
| `LOG_FILE_PATH` | — | Log file path (if output includes file) |
| `LOG_FORMAT` | `console`/`json` | Log format (env-dependent default) |
| `LOG_LEVEL` | `debug`/`info` | Log level (env-dependent default) |
| `LOG_OUTPUT` | `stdout` | Output: `stdout`, `file`, `both` |
| `PORT` | `52000` | HTTP server port |
| `VAULT_ADDR` | — | Vault server address (required) |
| `VAULT_PKI_MOUNTS` | `pki,pki2` | Comma-separated PKI mount paths |
| `VAULT_READ_TOKEN` | — | Vault read token (required) |
| `VAULT_TLS_INSECURE` | `false` | Skip TLS verification (dev only) |
| `VCV_EXPIRE_CRITICAL` | `7` | Critical expiration threshold (days) |
| `VCV_EXPIRE_WARNING` | `30` | Warning expiration threshold (days) |

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

## Development notes

- No external frontend toolchain; edit `app.js`/`styles.css` directly.
- Request IDs are added to all responses; include them when correlating logs.
- Use `VAULT_TLS_INSECURE=true` only in development environments.
