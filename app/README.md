# VaultCertsViewer — Technical overview

This document describes the technical structure of VaultCertsViewer (vcv), a single Go binary that embeds a compiled Svelte SPA to browse certificates from one or more HashiCorp Vault or OpenBao PKI mounts.

## Architecture

- **Backend**: Go + chi router, Vault/OpenBao client (`github.com/hashicorp/vault/api`), zerolog-based logging.
- **Frontend**: Svelte 5 + TypeScript + Vite (built into `app/web/dist`, embedded via `go:embed`).
- **Binary layout**: `app/cmd/server` embeds the Vite build; a single executable serves both the JSON API and static SPA.
- **Security**: Rate limiting (always on), CSRF protection for cookie-bearing state changes, security headers, request ID tracking, body size limits.
- **OpenBao compatibility**: Uses the Vault client library, which works with both HashiCorp Vault and OpenBao.

## Directory layout (app/)

- `cmd/server/main.go` — entrypoint, router, middleware, static file serving, graceful shutdown.
- `internal/cache/` — simple in-memory TTL cache (used by Vault client).
- `internal/config/` — settings-file configuration loading with expiration threshold support.
- `internal/errors/` — custom error types and helpers.
- `internal/handlers/` — HTTP handlers (`certs`, `i18n`, `health`, `ready`, `admin`, `config`).
- `internal/httputil/` — HTTP utility functions for client IP extraction (rate limiting).
- `internal/i18n/` — Internationalization message bundles.
- `internal/logger/` — zerolog initialization and structured helpers.
- `internal/metrics/` — Prometheus collectors.
- `internal/middleware/` — request ID, HTTP logging, panic recovery, CORS, security headers, rate limiting, CSRF, body limit.
- `internal/vault/` — Vault client implementations with multi-vault registry.
- `internal/version/` — build version info (injected via ldflags).
- `web/` — embedded frontend (`dist/` from Vite build).

## API surface

| Endpoint                  | Method  | Description                                              |
| ------------------------- | ------- | -------------------------------------------------------- |
| `/`                       | GET     | SPA shell (`index.html`)                                 |
| `/admin`                  | GET     | Admin SPA shell (`admin.html`)                           |
| `/assets/*`               | GET     | Hashed static assets                                     |
| `/api/certs`              | GET     | List certificates (partial-success envelope)             |
| `/api/certs/{id}/details` | GET     | Detailed certificate view                                |
| `/api/certs/{id}/pem`     | GET     | PEM content (JSON)                                       |
| `/api/certs/{id}/ca`      | GET     | Signing authority (intermediate/root)                    |
| `/api/config`             | GET     | Public application configuration (thresholds, mounts)    |
| `/api/health`             | GET     | Liveness probe                                           |
| `/api/i18n`               | GET     | UI translations (`?lang=`)                               |
| `/api/ready`              | GET     | Readiness probe                                          |
| `/api/status`             | GET     | Vault connection status (per vault; sanitized errors)    |
| `/api/version`            | GET     | Application version info                                 |
| `/metrics`                | GET     | Prometheus metrics                                       |
| `/api/admin/session`      | GET     | Admin session status                                     |
| `/api/admin/login`        | POST    | Admin login (JSON)                                       |
| `/api/admin/logout`       | POST    | Admin logout (JSON)                                      |
| `/api/admin/settings`     | GET/PUT | Admin settings (JSON, requires auth)                     |
| `/api/admin/docs`         | GET     | Admin documentation HTML (requires auth)                 |

## Configuration (settings.json)

Configuration **requires** a settings JSON file. There is no Vault env-var-only config path.

Recommended deployment pattern:

- Mount a `settings.json` file into the container under `/app/config/settings.json` (or provide one of the candidate paths below).
- If you enable the Admin panel, the settings file must be writable so changes can be persisted.

### Resolution order

Configuration is loaded from the first file found in this order (see `settingsCandidates()` in `internal/config`):

- `settings.dev.json`
- `settings.prod.json`
- `settings.json`
- `./settings.json`
- `/app/config/settings.json`

### Schema overview

- `app.env`, `app.port`
- `app.logging.level`, `app.logging.format`, `app.logging.output`, `app.logging.file_path`
- `cors.allowed_origins`, `cors.allow_credentials`
- `certificates.expiration_thresholds.critical`, `certificates.expiration_thresholds.warning`
- `metrics.per_certificate` (default false; high cardinality when true), `metrics.enhanced_metrics`
- `vaults[]`: list of Vault instances
  - `address`, `token`
  - `pki_mounts` (source of truth; recommended)
  - `pki_mount` (deprecated singular alias; accepted on read when `pki_mounts` is empty)
  - `tls_insecure` (default false; prefer CA material — see security notes)
  - `tls_ca_cert_base64` (preferred; base64-encoded PEM CA bundle)
  - `tls_ca_cert` (file path to a PEM CA bundle)
  - `tls_ca_path` (directory containing CA certs)
  - `tls_server_name` (SNI override)

Precedence rules for TLS material:

- If `tls_ca_cert_base64` is set, it is used and `tls_ca_cert` / `tls_ca_path` are ignored.
- Otherwise, `tls_ca_cert` / `tls_ca_path` are used (if set).

### Multi-Vault model

The configuration supports defining multiple Vault instances (`vaults[]`).

- The first enabled Vault instance becomes the primary Vault.
- Disabled vaults remain initialized so they can be toggled at runtime via the admin panel.

### Production TLS expectations

- Set `tls_insecure` to `false` in production and provide CA material (`tls_ca_cert_base64` preferred).
- Use a least-privilege Vault token (list/read PKI certs only — not issue/sign/revoke).
- `tls_insecure: true` is only for throwaway lab use; the runtime emits a warning when it is enabled.

### Logging initialization note

`internal/logger.Init` reads output settings from environment variables (`LOG_OUTPUT`, `LOG_FORMAT`, `LOG_FILE_PATH`). When loading from `settings.json`, `config.Load` sets those env vars so logging matches file configuration. Logger env vars are unrelated to Vault connectivity.

## Security & deployment assumptions

vcv is designed for **private networks**. Do not expose the listen port to the public internet without additional controls.

### Trust boundary

| Surface | Auth | Notes |
| --- | --- | --- |
| `/api/certs*`, `/api/status`, `/api/config`, `/api/health`, `/api/ready`, `/api/version`, `/api/i18n` | Unauthenticated | Intentional for internal inventory UI and probes |
| `/metrics` | Unauthenticated | Scrape only from private Prometheus / mesh |
| Static SPA `/`, `/admin`, `/assets/*` | Unauthenticated | Admin *API* still requires session |
| `/api/admin/*` | Session cookie (`vcv_admin_session`) | bcrypt password in settings; disabled if password missing/invalid |

### What PEMs are

`GET /api/certs/{id}/pem` returns **public** X.509 certificates as stored in Vault PKI. Private keys are never retrieved. Certificate inventories are still sensitive in many organizations.

### Recommended controls

1. Network policy / VPN / private mesh only — do not bind publicly without an ACL edge.
2. Reverse proxy (Traefik, nginx, etc.) with IP allowlist or SSO in front of the entire app.
3. Scrape `/metrics` only from Prometheus on a private network; consider proxy auth for metrics.
4. Vault token: least-privilege policy, for example:

    ```hcl
    path "pki/certs"   { capabilities = ["list"] }
    path "pki/cert/*"  { capabilities = ["read"] }
    path "sys/health"  { capabilities = ["read"] }
    ```

5. Admin panel: set `admin.password` to a bcrypt hash to enable; omit the field (or use an invalid hash) to disable. Sessions are in-process memory (sticky sessions or external store needed for horizontal scale).
6. TLS to Vault: `tls_insecure: false` plus CA material in production.

### App-layer controls already present

- Global rate limiting (always on; health/ready/metrics and `/assets/` exempt).
- CSRF: unsafe methods with cookies require same-origin Origin/Referer (cookieless clients still allowed for automation).
- Security headers, body size limits, request IDs.
- Public `/api/status` error strings are sanitized (no raw Vault internals).

### Reverse proxy / `trust_proxy`

Rate limiting keys and CSRF target-origin construction honor `X-Forwarded-For`, `X-Forwarded-Host`, and `X-Forwarded-Proto` **only** when `app.trust_proxy` is `true` (default **`false`**).

Set `app.trust_proxy: true` **only** when a reverse proxy (nginx, Traefik, etc.) sits in front of vcv and **overwrites** client-supplied `X-Forwarded-*` headers with trusted values. If clients can reach vcv directly while this flag is true, they can spoof per-IP rate-limit buckets and skew CSRF origin checks.

Lab `docker-compose` without a stripping proxy should leave `trust_proxy` false.

## Logging

- Initialized in `cmd/server/main.go` via `internal/logger.Init`.
- Middlewares emit structured HTTP events with `request_id`, status, duration.
- Handlers use `HTTPEvent`/`HTTPError` helpers; panic recovery logs stack traces.
- Version info logged at startup.

## Internationalization

- Languages: en, fr, es, de, it.
- `/api/i18n` returns messages; the SPA selects language and interpolates keys via the i18n store.
- Backend message structs in `internal/i18n` are the source of truth for UI strings.
