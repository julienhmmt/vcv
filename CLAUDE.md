# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

VaultCertsViewer (vcv) is a lightweight web UI that lists and inspects certificates stored in HashiCorp Vault or OpenBao PKI mounts. It's a single Go binary that embeds a static HTML/CSS/JS frontend.

## Commands

All tasks are run via [Task](https://taskfile.dev). Run `task --list` to see all available tasks.

```bash
# Development: build binary + docker image and start dev stack
task dev

# Lint (go fmt + go vet)
task lint

# Run unit tests offline (no Vault required), with coverage
task test-offline

# Run tests against the dev docker-compose stack
task test-dev

# Run tests directly
cd app && go test ./...

# Run a single package test
cd app && go test ./internal/handlers/... -run TestFunctionName

# Build multi-arch docker images and push to Docker Hub
VCV_TAG=1.7 task docker-build
```

The dev stack starts 5 Vault instances (ports 8200–8204) and 1 OpenBao instance (port 1337), plus the app at `http://localhost:52000`.

## Architecture

### Backend (app/)

- **Entry point**: `cmd/server/main.go` — loads config, creates Vault clients, sets up the chi router with middleware, and starts the HTTP server on port 52000.
- **Router**: Uses `go-chi/chi`. Middleware order: RequestID → Logger → Recoverer → SecurityHeaders → CORS → RateLimit (prod only) → BodyLimit → CSRFProtection.
- **Vault clients**: `internal/vault/` contains `Client` interface, `NewClientFromConfig`, `NewMultiClient` (aggregates multiple vaults), and `NewRegistry` (runtime enable/disable of vault instances). A `DisabledClient` is used as fallback.
- **Configuration**: `internal/config/` loads from `settings.dev.json` → `settings.prod.json` → `settings.json` → `./settings.json` → `/app/settings.json`. Falls back to legacy env vars (`VAULT_ADDR`, `VAULT_READ_TOKEN`, etc.) if no file is found.
- **Handlers**: `internal/handlers/` with separate registration functions per route group: `RegisterCertRoutes`, `RegisterUIRoutes`, `RegisterAdminRoutes`, `RegisterI18nRoutes`.
- **Metrics**: `internal/metrics/` — custom Prometheus collector registered against a private `prometheus.Registry` (not the default global one).
- **Caching**: `internal/cache/` — in-memory TTL cache used by Vault client.
- **Logging**: `internal/logger/` — zerolog-based. Initialized via `logger.Init`; log output/format configured via env vars `LOG_OUTPUT`, `LOG_FORMAT`, `LOG_FILE_PATH`.

### Frontend (app/web/)

- `index.html` and `admin.html` are the root pages, served directly.
- `assets/app-htmx.js` and `assets/styles.css` are the frontend assets.
- `templates/` contains HTMX HTML fragments rendered server-side by Go's `html/template`.
- No Node.js or build step — plain HTML/CSS/JS with HTMX.

### Key architectural patterns

- **Multi-vault model**: All vault instances (including disabled ones) are initialized at startup so they can be toggled at runtime via the admin panel without a restart. The first enabled vault is the "primary".
- **HTMX fragments**: UI updates use server-rendered HTML fragments at `/ui/*` endpoints, not a JSON API. The JSON API at `/api/*` is available for external consumers.
- **Embedded filesystem**: The `web/` directory is embedded via `go:embed` in `web/web.go` and served from the binary.
- **Admin panel**: Protected by bcrypt password in `settings.json`. Disabled if `admin.password` is missing or not a valid bcrypt hash. Admin can mutate `settings.json` at runtime.
- **Version injection**: Build version is injected via `-ldflags` into `internal/version/`.

## Configuration

Copy `settings.example.json` to `settings.json` (or `settings.dev.json` for dev). Key fields:

- `app.env`: `"dev"` or `"prod"` (rate limiting only active in prod)
- `app.port`: default `52000`
- `vaults[]`: list of vault instances with `address`, `token`, `pki_mounts`, TLS options
- `certificates.expiration_thresholds.critical` / `warning`: days before expiry (default 7/30)
- `admin.password`: bcrypt hash to enable admin panel
- `metrics.per_certificate`: boolean, disabled by default (high cardinality)
