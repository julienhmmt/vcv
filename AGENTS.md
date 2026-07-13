# AGENTS.md

Guidance for agents working in this repository. Keep this file accurate as the app changes.

## Git workflow (mandatory)

- **Never commit to the default branch** (`main`). Always create a feature branch first.
- **One branch per logical step or commit.** Do not pile unrelated changes onto an existing branch or commit multiple independent fixes in one commit on `main`.
- Before any commit:
  1. Confirm current branch is **not** `main` (`git branch --show-current`).
  2. If on `main` (or another protected/shared base), create a branch:

     ```bash
     git checkout main
     git pull --ff-only
     git checkout -b <type>/<short-description>
     ```

  3. Use conventional branch names, e.g. `fix/...`, `feat/...`, `refactor/...`, `docs/...`, `chore/...`.
- Prefer small, reviewable PRs: one concern per branch (docs-only, one bugfix, one feature slice).
- Do **not** push unless the user asks. Do **not** force-push or rewrite history on shared branches.
- Commit message: conventional commits when possible (`fix:`, `feat:`, `docs:`, …); focus on why.

## Project Overview

VaultCertsViewer (vcv) is a lightweight web UI that lists and inspects certificates stored in HashiCorp Vault or OpenBao PKI mounts. It ships as a **single Go binary** that embeds a compiled Svelte 5 SPA.

Beyond listing, it:

- Classifies each certificate by inferred type (`machine` / `user` / `both` / `unknown`)
- Surfaces signing authority (intermediate/root CA) in the certificate detail modal
- Offers command palette (Cmd/Ctrl-K), CSV/JSON export, shareable URL filter state
- Mobile card list (≤768px), opt-in auto-refresh, expiry toasts (sonner), connectivity warnings
- Multi-vault / multi-PKI mount selection, dashboard status overview, pagination

Read-only inventory UI; cert and metrics APIs are intentionally unauthenticated for private networks (see `app/README.md` security section).

## Commands

All project commands run via Make. Run `make help` for targets.

```bash
# Development: build frontend + Go binary + docker image, start dev stack
make dev

# Lint (go fmt + go vet)
make go-lint

# Go unit tests offline (no Vault), with coverage
make test-offline

# Go tests against the dev docker-compose stack
make test-dev

# Direct Go tests
cd app && go test ./...
cd app && go test ./internal/handlers/... -run TestFunctionName

# Frontend (Svelte): install, Vite dev server, build, type-check, tests
make web-install         # pnpm install
make web-dev             # pnpm dev (proxies /api to :52000)
make web-build           # pnpm build → app/web/dist (required before go build / docker)
make web-check           # svelte-check + tsc
make web-test            # vitest run
make web-test-coverage   # vitest + coverage

# Multi-arch docker images push
VCV_TAG=1.9 make docker-build
```

- Frontend source: `app/web/frontend/` (Vite + pnpm).
- `make web-build` writes `app/web/dist`, embedded via `go:embed` in `app/web/web.go`. Always rebuild before `go build` or the binary serves a stale/empty UI.
- Dev stack: 5 Vault instances (ports 8200–8204), 1 OpenBao (port 1337), app at `http://localhost:52000`.

## Architecture

### Backend (`app/`)

| Path | Role |
| --- | --- |
| `cmd/server/main.go` | Entry: load config, Vault clients, chi router, graceful shutdown |
| `internal/cache/` | In-memory TTL cache (Vault client) |
| `internal/config/` | Settings-file loading, multi-vault, thresholds |
| `internal/certs/` | Certificate model + `InferCertType` |
| `internal/errors/` | Shared error types |
| `internal/handlers/` | HTTP handlers (certs, admin, i18n, health, ready, config, static) |
| `internal/httputil/` | Client IP helpers (rate limit) |
| `internal/i18n/` | Message bundles (en/fr/de/it/es) — source of truth for UI strings |
| `internal/logger/` | zerolog; `LOG_OUTPUT` / `LOG_FORMAT` / `LOG_FILE_PATH` |
| `internal/metrics/` | Prometheus collectors (private registry; enhanced optional) |
| `internal/middleware/` | RequestID, Logger, Recoverer, SecurityHeaders, CORS, RateLimit, BodyLimit, CSRF |
| `internal/vault/` | Client interface, real/multi/registry/disabled, status checks |
| `internal/version/` | Version via `-ldflags` |
| `web/` | Embedded Vite `dist/` |

- **Router middleware order** (always registered): RequestID → Logger → Recoverer → SecurityHeaders → CORS → **RateLimit (always on)** → BodyLimit → CSRFProtection. Rate limit exempts `/api/health`, `/api/ready`, `/metrics`, and `/assets/`.
- **Static**: `RegisterStaticRoutes` serves `/` → `dist/index.html`, `/admin` → `dist/admin.html`, hashed `/assets/*`. Client-rendered SPA; backend is JSON API only. Old HTMX `/ui/*` and `html/template` rendering are gone.
- **Vault clients**: `Client` interface, `NewClientFromConfig`, `NewMultiClient`, `NewRegistry` (runtime enable/disable). `DisabledClient` as fallback. Per-vault status via `CheckInstances`.
- **Config resolution** (first file found): `settings.dev.json` → `settings.prod.json` → `settings.json` → `./settings.json` → `/app/config/settings.json`. **No Vault env-var-only config path** — a settings file is required. Logger env vars are still used and are set from the file by `config.Load`.
- **Handler registration**: `RegisterStaticRoutes`, `RegisterCertRoutes`, `RegisterAdminRoutes`, `RegisterI18nRoutes`; health/ready/status/version/config/metrics wired in `main.go`.
- **Cert routes**: `GET /api/certs` (partial-success envelope), `/api/certs/{id}/details`, `/api/certs/{id}/ca`, `/api/certs/{id}/pem`.
- **Admin API** (session cookie): `/api/admin/session`, `login`, `logout`, `docs`, `settings` GET/PUT, `POST /api/admin/vault`, `DELETE /api/admin/vault/{id}`; optional `POST /api/cache/invalidate`.
- **Certificate model**: `InferCertType` from ExtKeyUsage → `machine` / `user` / `both` / `unknown`. CA viewer uses `caType` intermediate/root on details.
- **Partial-success envelope**: `GET /api/certs` returns `certificates` + per-vault `errors []vault.VaultError` so one failed vault warns in UI instead of failing the whole list.
- **i18n**: `internal/i18n` Messages + maps; `GET /api/i18n?lang=`. Frontend uses `t(key, fallback?, params?)` only — do not hardcode UI strings.

### Frontend (`app/web/frontend/`)

- **Stack**: Svelte 5 (runes) + TypeScript + Vite, Tailwind v4, bits-ui (shadcn-svelte) under `src/lib/components/ui/`, `@lucide/svelte`, `svelte-sonner`. Package manager: **pnpm**.
- **Entry points**: `src/main.ts` → `App.svelte` (`/`); `src/admin.ts` → `Admin.svelte` (`/admin`). HTML shells: `src/index.html`, `src/admin.html`.
- **Stores** (`src/lib/stores/*.svelte.ts`): `certs`, `config` (expiration thresholds from `/api/config`), `status`, `theme`, `i18n`, `admin`. API wrapper: `lib/api.ts`. Types: `lib/types.ts`.
- **Domain components** (`src/lib/components/`):
  - `CertTable` / `CertMobileList` / `CertCard` / `CertStatusBadge` / `PaginationBar`
  - `CertDetailModal` (details + CA/signing authority + PEM copy)
  - `CertTypeSelect`, `CommandPalette`, `ActiveFilters`, `StatusOverview`
  - `MountSelectorDialog`, `VaultStatusPill`, `ErrorBanner`, `TableSkeleton`
  - Admin: `admin/AdminLogin`, `AdminPanel`, `VaultEditor`, `AdminDocsModal`
  - UI primitives under `ui/` (button, dialog, select, command, sonner Toaster, …)
- **Utils** (`src/lib/utils/`): `cert-filter`, `cert-status`, `cert-label`, `url-state`, `export`, `expiry-notify`, `config-thresholds`, clipboard, icons.
- **Tests**: Vitest + jsdom, colocated `*.test.ts`. Run `make web-test` / `make web-check`.
- **i18n**: `setI18nContext` / `getI18n`; every visible string via `t(...)`.

### Key patterns

- **Multi-vault**: All instances (including disabled) init at startup so admin can toggle without restart. First enabled vault is primary.
- **SPA + JSON API only**: no server-rendered HTML for app UI.
- **Embedded FS**: `app/web/dist` via `go:embed` in `app/web/web.go`.
- **Admin**: bcrypt `admin.password` in settings; disabled if missing/invalid. Sessions in-process memory (sticky sessions if scaled horizontally). Mutates writable settings file.
- **Private-network threat model**: unauthenticated cert/status/metrics APIs by design; put network ACL / reverse proxy in front for production. Documented in `app/README.md`.

## Configuration

Copy `settings.example.json` → `settings.json` (or `settings.dev.json` for local). Key fields:

| Field | Notes |
| --- | --- |
| `app.env` | `"dev"` / `"prod"` (affects defaults, secure cookies, logging defaults — not rate-limit enablement) |
| `app.port` | Default `52000` |
| `app.logging` | `level`, `format`, `output`, `file_path` |
| `vaults[]` | `id`, `address`, `token`, `pki_mounts`, `display_name`, TLS fields, `enabled` |
| `certificates.expiration_thresholds` | `critical` / `warning` days (default 7 / 30); UI loads via `/api/config` |
| `admin.password` | bcrypt hash to enable admin |
| `metrics.per_certificate` | Default false (high cardinality) |
| `metrics.enhanced_metrics` | Optional richer collectors |
| `cors.allowed_origins` / `allow_credentials` | Browser cross-origin |

TLS: prefer `tls_ca_cert_base64`; `tls_insecure: true` only for labs (runtime warns).

## CI / Release

`.github/workflows/`:

- **`go.yml`** — golangci-lint (pinned) + build/test with race + coverage gate
- **`lint.yml`** — repo-wide linting
- **`release.yml`** — on tag: GoReleaser `~> v2` builds frontend (pnpm), embeds dist, cross-compiles `vcv` (`./cmd/server`), publishes archives + Docker (`.goreleaser.yaml`)

Release: push a semver tag (e.g. `1.8`); GoReleaser does the rest.

## Agent habits for this repo

1. Branch off `main` before edits; never commit on `main`.
2. Prefer Make targets over ad-hoc `pnpm`/`go` paths when a target exists.
3. After frontend changes: `make web-check` and/or `make web-test`. After Go changes: `make go-lint` and `make test-offline` (or package-scoped tests).
4. Keep UI strings in `internal/i18n` + frontend `t()` usage in sync.
5. Do not reintroduce HTMX or server-side templates.
6. Do not expose secrets in logs, settings examples, or commits.
7. When architecture drifts, update **this file** (and `app/README.md` for deep technical detail) in a **docs branch**, not as a drive-by on an unrelated fix.
