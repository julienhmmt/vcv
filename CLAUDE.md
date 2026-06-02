# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

VaultCertsViewer (vcv) is a lightweight web UI that lists and inspects certificates stored in HashiCorp Vault or OpenBao PKI mounts. It's a single Go binary that embeds a compiled Svelte 5 frontend.

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

# Frontend (Svelte): install deps, dev server, build to app/web/dist, type-check
task web-install   # pnpm install
task web-dev       # pnpm dev (Vite)
task web-build     # pnpm build → app/web/dist (required before go build / docker)
task web-check     # svelte-check + tsc

# Build multi-arch docker images and push to Docker Hub
VCV_TAG=1.7 task docker-build
```

The frontend lives in `app/web/frontend/` (Vite + pnpm). `task web-build` compiles it into `app/web/dist`, which is embedded via `go:embed` — run it before `go build` or the binary serves a stale/empty UI.

The dev stack starts 5 Vault instances (ports 8200–8204) and 1 OpenBao instance (port 1337), plus the app at `http://localhost:52000`.

## Architecture

### Backend (app/)

- **Entry point**: `cmd/server/main.go` — loads config, creates Vault clients, sets up the chi router with middleware, and starts the HTTP server on port 52000.
- **Router**: Uses `go-chi/chi`. Middleware order: RequestID → Logger → Recoverer → SecurityHeaders → CORS → RateLimit (prod only) → BodyLimit → CSRFProtection.
- **Static serving**: `internal/handlers/static.go` `RegisterStaticRoutes` serves the embedded Vite build — `/` → `dist/index.html`, `/admin` → `dist/admin.html`, hashed assets under `/assets/`. The UI is a client-rendered SPA; the backend exposes only the JSON API.
- **Vault clients**: `internal/vault/` contains `Client` interface, `NewClientFromConfig`, `NewMultiClient` (aggregates multiple vaults), and `NewRegistry` (runtime enable/disable of vault instances). A `DisabledClient` is used as fallback.
- **Configuration**: `internal/config/` loads from `settings.dev.json` → `settings.prod.json` → `settings.json` → `./settings.json` → `/app/settings.json`. Falls back to legacy env vars (`VAULT_ADDR`, `VAULT_READ_TOKEN`, etc.) if no file is found.
- **Handlers**: `internal/handlers/` with separate registration functions per route group: `RegisterStaticRoutes`, `RegisterCertRoutes`, `RegisterAdminRoutes`, `RegisterI18nRoutes`. (The old HTMX `/ui/*` routes and `html/template` rendering have been removed.)
- **i18n**: `internal/i18n/i18n.go` is the source of truth — a `Messages` struct (~210 keys) with full maps for en/fr/de/it/es, served at `/api/i18n?lang=`. The frontend never hardcodes UI strings; it fetches these and looks them up by key.
- **Metrics**: `internal/metrics/` — custom Prometheus collector registered against a private `prometheus.Registry` (not the default global one).
- **Caching**: `internal/cache/` — in-memory TTL cache used by Vault client.
- **Logging**: `internal/logger/` — zerolog-based. Initialized via `logger.Init`; log output/format configured via env vars `LOG_OUTPUT`, `LOG_FORMAT`, `LOG_FILE_PATH`.

### Frontend (app/web/frontend/)

- **Stack**: Svelte 5 (runes) + TypeScript + Vite, styled with Tailwind v4 and bits-ui (shadcn-svelte) primitives under `src/lib/components/ui/`. Package manager is pnpm.
- **Entry points**: two mount targets — `src/main.ts` → `App.svelte` (`/`) and `src/admin.ts` → `Admin.svelte` (`/admin`); HTML shells in `src/index.html` / `src/admin.html`.
- **State**: rune-based stores in `src/lib/stores/*.svelte.ts` (`certs`, `status`, `theme`, `i18n`, `admin`). `lib/api.ts` wraps the JSON API.
- **i18n**: `lib/stores/i18n.svelte.ts` exposes `t(key, fallback?, params?)` with `{{x}}`/`{x}` interpolation, shared down the tree via `setI18nContext` (root) / `getI18n` (children). Every component pulls strings from the Go-served message bundle — do not hardcode UI text.
- **Build output**: `task web-build` emits to `app/web/dist`, embedded by `app/web/web.go`.

### Key architectural patterns

- **Multi-vault model**: All vault instances (including disabled ones) are initialized at startup so they can be toggled at runtime via the admin panel without a restart. The first enabled vault is the "primary".
- **SPA + JSON API**: The Svelte app talks exclusively to the JSON API under `/api/*` (certs, status, version, i18n, admin). There is no server-side HTML rendering.
- **Embedded filesystem**: The compiled `app/web/dist` is embedded via `go:embed` in `app/web/web.go` and served from the binary.
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
