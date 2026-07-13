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
- Scopes already used in history: `fix(web):`, `fix(backend):`, `fix(ui):`, `refactor(web):`, `chore(app):`, `docs:`.
- Do not mix unrelated frontend and backend hardening in one PR unless the user asks.
- Do not commit: `coverage.out`, `/vcv` binary, `vcv.log`, real `settings.json` / tokens, `node_modules`, hand-edited `app/web/dist/*` (except `.gitkeep`).

## Invariants (do / don’t)

### Do

- SPA + JSON API only; all HTTP from the UI goes through `src/lib/api.ts`.
- UI strings via `t(key, fallback?, params?)`; source of truth is `internal/i18n`.
- Prefer Make targets (`make web-test`, `make test-offline`, …) over ad-hoc paths.
- Table-driven Go tests; Vitest + colocated `*.test.ts` on the frontend.
- Expiration thresholds from `/api/config` / the `config` store — never hardcode 7/30 in UI logic.
- Encode cert IDs in paths (`encodeURIComponent`); composite IDs are parsed carefully (see `parseCertID` / backend handlers).
- Treat partial vault failure as normal: `GET /api/certs` envelope has `certificates` + `errors[]`.
- **Use graphify memory** (`graphify-frontend/`, `graphify-backend/`, `graphify-full/`) before broad code archaeology — see [Knowledge graphs (graphify)](#knowledge-graphs-graphify).

### Don’t

- Reintroduce HTMX, server-side templates, or `/ui/*` HTML routes.
- Hardcode user-visible English (or any language) in Svelte components.
- Add authentication to public cert/status/metrics/config/i18n/health/ready/version unless product explicitly asks (threat model is private-network ACL).
- Log, return, or commit cleartext Vault tokens. Admin settings **mask** tokens (`***`); preserve stored token on PUT when the field is blank/masked.
- Use npm or yarn — **pnpm only** under `app/web/frontend/`.
- Edit `app/web/dist` by hand (build output; re-embed via `make web-build`).
- Skip middleware (CSRF, RateLimit, BodyLimit) on new state-changing routes that use cookies.
- Fetch private keys from Vault — PEM endpoints return public X.509 only.
- Gate product work on SvelteKit migration without reopening `plans/DECISION-sveltekit.md` with the user.

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
- `make web-dev` alone needs a running Go backend for real API data; `make dev` is heavier (image rebuild).

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
- **Cert routes**: `GET /api/certs` (partial-success envelope), `/api/certs/{id}/details`, `/api/certs/{id}/ca`, `/api/certs/{id}/pem`. Optional `?mounts=` filter.
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

- **Multi-vault**: All instances (including disabled) init at startup so admin can toggle without restart. First enabled vault is primary. Registry remains initialized for disabled vaults.
- **SPA + JSON API only**: no server-rendered HTML for app UI.
- **Embedded FS**: `app/web/dist` via `go:embed` in `app/web/web.go`.
- **Admin**: bcrypt `admin.password` in settings; disabled if missing/invalid. Sessions in-process memory (sticky sessions if scaled horizontally). Mutates writable settings file.
- **Private-network threat model**: unauthenticated cert/status/metrics APIs by design; put network ACL / reverse proxy in front for production. Documented in `app/README.md`.

## Code conventions

### Go

- Prefer table-driven tests; use `_test` package for black-box when practical; name `Test<Func>_<Scenario>`.
- Pass `context.Context` into Vault I/O and long-running work.
- Match existing JSON error shapes in handlers (`error` field when clients expect it).
- Prometheus: register on the **private** registry from `main` — not the global default registry.
- Keep middleware order; rate limit is always on. Login may have tighter limits — follow existing admin patterns.
- Public `/api/status` errors must stay **sanitized** (stable strings, no raw Vault internals).
- Avoid logging secrets, tokens, full PEMs, or request bodies that may contain passwords.
- golangci-lint lives in `app/.golangci.yml` (errcheck, staticcheck, bodyclose, …).

### Frontend (Svelte 5)

- Use runes (`$state`, `$derived`, `$effect`, …). Do not introduce legacy Svelte stores for new state.
- Keep fetch/business logic in `lib/api.ts`, stores, or `lib/utils/*` — not large blocks inside markup.
- Modal/async: ignore stale responses when the open cert id or generation changes (see `CertDetailModal` pattern).
- Toasts: `svelte-sonner`; toaster theme follows the `theme` store.
- New controls: compose bits-ui / existing `ui/*` primitives before inventing styled one-offs.
- Thresholds and status tiers come from the `config` store / `/api/config`.
- `fetch` uses `credentials: 'same-origin'` (see `api.ts`) so admin cookies work.

### i18n checklist

When adding or changing user-visible copy:

1. Add the key to `internal/i18n` for **all** languages: en, fr, de, it, es.
2. Call `t('key', 'English fallback', params?)` in the component or store.
3. No bare user-visible string literals in Svelte (fallbacks in `t()` are OK).
4. Use the same interpolation style already used (`{x}` / `{{x}}`).
5. Drop unused keys — unreferenced messages are dead weight.

## Security rules

| Surface | Auth | Notes |
| --- | --- | --- |
| `/api/certs*`, `/api/status`, `/api/config`, probes, `/api/i18n`, `/api/version` | None | Intentional for private networks |
| `/metrics` | None | Scrape only on private networks |
| Static `/`, `/admin`, `/assets/*` | None | Admin **API** still needs session |
| `/api/admin/*` | Session cookie | bcrypt password; disabled if hash missing/invalid |

Hard rules for agents:

1. **Mask vault tokens** in admin GET settings; on PUT/rename, preserve stored token when client sends blank or masked value.
2. **Never** put cleartext tokens in logs, API responses, examples committed to git, or screenshots docs.
3. **Sanitize** public status/error strings.
4. **PEM** = public certificates only; no private key retrieval.
5. **CSRF**: unsafe methods with cookies require same-origin Origin/Referer (existing middleware).
6. **TLS to Vault**: examples and prod guidance prefer CA material; `tls_insecure: true` is lab-only (runtime warns).
7. Do not “fix” public cert/metrics by adding app-level auth without an explicit product decision and a dedicated plan.

Deep threat-model write-up: `app/README.md` → *Security & deployment assumptions*.

## Verification matrix

| Change type | Minimum verification |
| --- | --- |
| UI string / i18n | Key present in all langs; `make web-check` |
| Frontend logic / components | `make web-test` and `make web-check` |
| Go package (handlers, vault, middleware, config, metrics) | `cd app && go test ./internal/<pkg>/...` then `make test-offline` or `make go-lint` |
| Admin session / CSRF / rate limit | Tests under `internal/handlers` + `internal/middleware` |
| Config schema / settings examples | Ensure examples have no real secrets; related config tests |
| Full local stack | `make dev` (optional; slow — image rebuild) |

“Done” means the relevant checks above are green, not only that files were edited.

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

Edit **examples** (`settings.example.json`, `settings.enhanced-metrics.example.json`) for repo changes — not local files that hold real tokens.

## Knowledge graphs (graphify)

**Use these aggressively.** They are local project memory (gitignored) for architecture, bugs, and cross-file links. Prefer graph query / report / wiki over re-reading large packages from scratch.

### Layout

| Directory | Scope | Prefer for |
| --- | --- | --- |
| `graphify-frontend/` | Svelte SPA (`app/web/frontend`) | UI, stores, `api.ts`, thresholds UX, Track C plans |
| `graphify-backend/` | Go app (`app/cmd`, `app/internal`, embed) | Vault, middleware, admin sessions, metrics, Track D plans |
| `graphify-full/` | FE ∪ BE + cross-stack bridges | API ↔ SPA contracts (list certs, config, admin token mask, i18n, status) |

Each dir typically has:

- `graph.html` — interactive graph  
- `graph.json` — GraphRAG / CLI query target  
- `GRAPH_REPORT.md` — god nodes, surprising edges, suggested questions  
- Optional: `wiki/` (backend), `README.md` (how to rebuild)

Fallback: `graphify-out/` if a one-off pipeline wrote there.

### Mandatory habits (when the folders exist)

1. **Start of a non-trivial task** (bug, feature, audit, “how does X work?”):  
   - Frontend-only → open or query **`graphify-frontend`**.  
   - Backend-only → **`graphify-backend`**.  
   - Spans HTTP (handler ↔ store/component) → **`graphify-full`**, then drill into the layer graph.  
2. **Before large greps** of a package you do not already know: query the graph first (`graphify query` / skill `/graphify query` with `--graph <path>/graph.json`).  
3. **Read `GRAPH_REPORT.md` god nodes + surprising connections** for the layer you touch — often names the hub (`buildRouter`, `createCertsStore`, `certStatus`, …).  
4. **Plans**: if `plans/README.md` exists (gitignored), check active Track C/D rows before inventing duplicate work. Graph findings should feed plans, not replace verification in code.  
5. **Always verify in source** after a graph answer — edges can be INFERRED; line-level truth wins.  
6. **After substantial architecture changes** (new routes, new stores, middleware reorder, multi-vault semantics): rebuild the affected layer graph, then re-merge full if needed (see rebuild). Do not block a small bugfix on a full rebuild.  
7. **Do not commit** `graphify-*/` — already gitignored. Do not invent secrets into graph nodes.

### Which graph to pick

```text
Is the change only under app/web/frontend/ ?
  → graphify-frontend

Is the change only under app/cmd or app/internal (or app/web/web.go embed) ?
  → graphify-backend

Does it cross FE api.ts / stores ↔ Go handlers / vault / middleware ?
  → graphify-full first (bridges), then layer graph for implementation detail

Unsure / onboarding / “map the system” ?
  → GRAPH_REPORT.md in full or both layers; wiki under graphify-backend/wiki/ if present
```

### Query / CLI (repo root)

Default graphify CLI looks at `graphify-out/graph.json`. **Always pass our paths:**

```bash
# Explore (examples)
graphify query "How do cert list errors reach the UI?" --graph graphify-full/graph.json
graphify query "Where is admin session rate limit keyed?" --graph graphify-backend/graph.json
graphify query "expiry thresholds config store" --graph graphify-frontend/graph.json
graphify path "mapVaultsForPut" "mergeVaultTokens" --graph graphify-full/graph.json
graphify explain "buildRouter" --graph graphify-backend/graph.json
```

In agent sessions with the **graphify skill** (`/graphify`): run query/path/explain against the correct folder; if the skill defaults to `graphify-out/`, set the graph path to `graphify-frontend|backend|full/graph.json`.

Open HTML in a browser for orientation:  
`graphify-frontend/graph.html`, `graphify-backend/graph.html`, `graphify-full/graph.html`.

### Rebuild / merge notes

- **Independent layers** stay source of truth for deep work (backend may be directed; frontend undirected).  
- **Full** is an undirected union with `fe_`/`be_` node prefixes + explicit **cross-stack** edges (`source_file: cross-stack`). Plain `graphify merge-graphs A B` fails if one graph is directed and the other is not — convert both to undirected (or use the session merge recipe in `graphify-full/README.md`).  
- Quality levers: scope (no frontend noise in backend graph), deep semantic for backend, edge sanitizer (confidence + `source_file`), optional wiki.  
- Rebuild when user asks `/graphify`, when graphs are missing, or after large refactors. Small PRs may skip rebuild.

### When graphs are missing

If `graphify-frontend/` or `graphify-backend/` is absent: continue with normal code tools, and offer to rebuild (`/graphify` on `app/web/frontend` or `app/` backend scope) so future sessions are faster. Prefer rebuilding **once** over re-deriving architecture every turn.

## Doc map

| Topic | File |
| --- | --- |
| Agent workflow, architecture summary, invariants | **`AGENTS.md`** (this file) |
| Claude pointer | `CLAUDE.md` → use AGENTS.md |
| Deep backend, API table, threat model | `app/README.md` |
| User-facing product / deploy | `README.md` (+ `README.fr.md`, docker-hub readme) |
| Prometheus metrics | `PROMETHEUS_METRICS.md`, `ALERTING.md` |
| Admin operational docs (embedded) | `app/internal/docs/` |
| Implementation plans / improve handoffs | `plans/` (local/gitignored; Track C frontend, Track D backend; execute when present) |
| **Knowledge graphs (query before archaeology)** | `graphify-frontend/`, `graphify-backend/`, `graphify-full/` (local/gitignored) |

When architecture or agent workflow drifts, update **AGENTS.md**. When handler/API/security detail drifts, update **`app/README.md`**. Prefer a **docs branch**, not drive-by edits on an unrelated fix PR.

## Dependencies

- Frontend: **pnpm** only; prefer existing deps before adding new ones.
- Prefer dependency versions published at least ~7 days ago; avoid floating `latest`.
- Do not major-bump Svelte / Vite / bits-ui / Go toolchain inside a product bugfix unless the user asks.
- Go modules: use normal `go get` / `go mod tidy` under `app/`; `make go-update` exists but is broad — don’t run casually in a narrow fix.

## CI / Release

`.github/workflows/`:

- **`go.yml`** — golangci-lint (pinned) + build/test with race + coverage gate
- **`lint.yml`** — repo-wide linting
- **`release.yml`** — on tag: GoReleaser `~> v2` builds frontend (pnpm), embeds dist, cross-compiles `vcv` (`./cmd/server`), publishes archives + Docker (`.goreleaser.yaml`)

Release: push a semver tag (e.g. `1.9`); GoReleaser does the rest.

## Agent habits for this repo

1. Branch off `main` before edits; never commit on `main`. One branch per step.
2. **Query graphify first** when the folders exist (frontend / backend / full) — then read code at the cited paths. Prefer graph + plans over blind multi-package search.
3. Prefer Make targets over ad-hoc `pnpm`/`go` paths when a target exists.
4. Run the verification matrix for the change type before calling the work done.
5. Keep UI strings in `internal/i18n` + frontend `t()` usage in sync (all five languages).
6. Do not reintroduce HTMX or server-side templates.
7. Do not expose secrets in logs, settings examples, or commits; mask admin tokens.
8. When architecture or agent process drifts, update this file (and `app/README.md` for deep detail) on a docs branch.
9. Honor intentional non-goals (public inventory APIs, in-memory admin sessions) unless the user reopens them.
10. After large structural changes, rebuild the affected graphify layer (and full if cross-stack) when practical — do not leave the graph permanently stale for the next agent.
