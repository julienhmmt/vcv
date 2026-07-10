# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

VaultCertsViewer (vcv) is a lightweight web UI that lists and inspects certificates stored in HashiCorp Vault or OpenBao PKI mounts. It's a single Go binary that embeds a compiled Svelte 5 frontend.

Beyond listing, it classifies each certificate by inferred type (machine / user / both / unknown), surfaces the signing authority (intermediate/root CA viewer), and offers a command palette, CSV/JSON export, shareable URL filter state, mobile cards, opt-in auto-refresh, and expiry/connectivity notifications.

## Commands

All project commands are run via Make. Run `make help` to see all available targets.

```bash
# Development: build binary + docker image and start dev stack
make dev

# Lint (go fmt + go vet)
make go-lint

# Run unit tests offline (no Vault required), with coverage
make test-offline

# Run tests against the dev docker-compose stack
make test-dev

# Run tests directly
cd app && go test ./...

# Run a single package test
cd app && go test ./internal/handlers/... -run TestFunctionName

# Frontend (Svelte): install deps, dev server, build to app/web/dist, type-check
make web-install   # pnpm install
make web-dev       # pnpm dev (Vite)
make web-build     # pnpm build → app/web/dist (required before go build / docker)
make web-check     # svelte-check + tsc

# Build multi-arch docker images and push to Docker Hub
VCV_TAG=1.8 make docker-build
```

The frontend lives in `app/web/frontend/` (Vite + pnpm). `make web-build` compiles it into `app/web/dist`, which is embedded via `go:embed` — run it before `go build` or the binary serves a stale/empty UI.

The dev stack starts 5 Vault instances (ports 8200–8204) and 1 OpenBao instance (port 1337), plus the app at `http://localhost:52000`.

## Architecture

### Backend (app/)

- **Entry point**: `cmd/server/main.go` — loads config, creates Vault clients, sets up the chi router with middleware, and starts the HTTP server on port 52000.
- **Router**: Uses `go-chi/chi`. Middleware order: RequestID → Logger → Recoverer → SecurityHeaders → CORS → RateLimit (prod only) → BodyLimit → CSRFProtection.
- **Static serving**: `internal/handlers/static.go` `RegisterStaticRoutes` serves the embedded Vite build — `/` → `dist/index.html`, `/admin` → `dist/admin.html`, hashed assets under `/assets/`. The UI is a client-rendered SPA; the backend exposes only the JSON API.
- **Vault clients**: `internal/vault/` contains `Client` interface, `NewClientFromConfig`, `NewMultiClient` (aggregates multiple vaults), and `NewRegistry` (runtime enable/disable of vault instances). A `DisabledClient` is used as fallback.
- **Configuration**: `internal/config/` loads from `settings.dev.json` → `settings.prod.json` → `settings.json` → `./settings.json` → `/app/settings.json`. Falls back to legacy env vars (`VAULT_ADDR`, `VAULT_READ_TOKEN`, etc.) if no file is found.
- **Handlers**: `internal/handlers/` with separate registration functions per route group: `RegisterStaticRoutes`, `RegisterCertRoutes`, `RegisterAdminRoutes`, `RegisterI18nRoutes`. (The old HTMX `/ui/*` routes and `html/template` rendering have been removed.)
- **Routes**: Cert routes: `GET /api/certs` (list), `/api/certs/{id}/details`, `/api/certs/{id}/ca` (signing authority chain), `/api/certs/{id}/pem`. Plus `/api/health`, `/api/ready`, `/api/status`, `/api/version`, `/api/config` registered in `main.go`.
- **Certificate model**: `internal/certs/certificate.go` — `InferCertType` classifies each cert from its ExtKeyUsage as `machine` (server auth), `user` (client auth), `both`, or `unknown`. `DetailedCertificate` carries `caType` (`intermediate`/`root`) for the authority viewer.
- **Partial-success envelope**: `GET /api/certs` returns a `certsEnvelope` (`certificates` + per-vault `errors []vault.VaultError`) so one failed vault surfaces a warning in the UI instead of failing the whole response. Clients implementing `vault.CertificatesEnvelopeLister` are preferred.
- **i18n**: `internal/i18n/i18n.go` is the source of truth — a `Messages` struct (~175 keys) with maps for en/fr/de/it/es, served at `/api/i18n?lang=`. The frontend never hardcodes UI strings; it fetches these and looks them up by key via the `i18n` store (`t(key, fallback?, params?)`, `{{x}}`/`{x}` interpolation). Keep keys and frontend usage in sync — unreferenced keys are dead weight.
- **Metrics**: `internal/metrics/` — custom Prometheus collector registered against a private `prometheus.Registry` (not the default global one).
- **Caching**: `internal/cache/` — in-memory TTL cache used by Vault client.
- **Logging**: `internal/logger/` — zerolog-based. Initialized via `logger.Init`; log output/format configured via env vars `LOG_OUTPUT`, `LOG_FORMAT`, `LOG_FILE_PATH`.

### Frontend (app/web/frontend/)

- **Stack**: Svelte 5 (runes) + TypeScript + Vite, styled with Tailwind v4 and bits-ui (shadcn-svelte) primitives under `src/lib/components/ui/`. Package manager is pnpm.
- **Entry points**: two mount targets — `src/main.ts` → `App.svelte` (`/`) and `src/admin.ts` → `Admin.svelte` (`/admin`); HTML shells in `src/index.html` / `src/admin.html`.
- **State**: rune-based stores in `src/lib/stores/*.svelte.ts` (`certs`, `status`, `theme`, `i18n`, `admin`). `lib/api.ts` wraps the JSON API.
- **Components**: domain components in `src/lib/components/` — `CertDetailModal` / `CAModal` (detail + signing-authority viewer), `CertCard` (mobile ≤768px), `CertTypeSelect`, `CommandPalette` (Cmd/Ctrl-K), `Donut`, `ActiveFilters`, `StatusOverview`, `MountSelectorDialog`, `VaultStatusPill`, `ErrorBanner`, `TableSkeleton`. shadcn-svelte primitives under `ui/`, admin-only under `admin/`.
- **Tests**: Vitest unit tests + jsdom component-render layer colocated as `*.test.ts` (e.g. `Donut.test.ts`). Run via `pnpm test` / `make web-check`.
- **i18n**: `lib/stores/i18n.svelte.ts` exposes `t(key, fallback?, params?)` with `{{x}}`/`{x}` interpolation, shared down the tree via `setI18nContext` (root) / `getI18n` (children). Every component pulls strings from the Go-served message bundle — do not hardcode UI text.
- **Build output**: `make web-build` emits to `app/web/dist`, embedded by `app/web/web.go`.

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

## CI / Release

GitHub Actions in `.github/workflows/`:

- **`go.yml`** — `golangci-lint` (pinned) + build & test with race detector and coverage threshold gate (`go test -race -coverprofile`).
- **`lint.yml`** — repo-wide linting.
- **`release.yml`** — on tag push, runs GoReleaser (`~> v2`) which builds the frontend (pnpm), embeds `dist`, then cross-compiles the `vcv` binary (`./cmd/server`) and publishes archives + Docker images per `.goreleaser.yaml`.

Cut a release by pushing a semver tag (e.g. `1.8`); GoReleaser handles the rest.
