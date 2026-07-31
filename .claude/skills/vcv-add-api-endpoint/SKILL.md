---
name: vcv-add-api-endpoint
description: Add or change a JSON API endpoint in the vcv Go backend (app/internal/handlers) and wire it to the Svelte client. Use when adding an HTTP route, changing a response shape, or adding an admin API action.
---

# Add a JSON API endpoint

Backend is JSON-only; there are no server-rendered HTML routes. Everything the UI calls goes through `app/web/frontend/src/lib/api.ts`.

## Steps

1. **Handler** — `app/internal/handlers/<name>.go`.
   - Request/response structs with JSON tags. Return empty slices, never `nil`, for list fields.
   - Take `context.Context` into every Vault call.
   - Match the existing error shape (`{"error": "..."}`) when clients read it. Public error strings stay sanitized — no raw Vault internals.
   - Log with `logger.HTTPError(method, path, status, err)` plus `middleware.GetRequestID(ctx)`.

2. **Route registration** — pick the right register func, do not hand-roll a new router:

   | Kind | Func | File |
   | --- | --- | --- |
   | Cert routes | `RegisterCertRoutes` | `internal/handlers/certs.go` |
   | Admin (session cookie) | `RegisterAdminRoutes` | `internal/handlers/admin.go` |
   | i18n | `RegisterI18nRoutes` | `internal/handlers/i18n.go` |
   | Static/SPA | `RegisterStaticRoutes` | `internal/handlers/static.go` |
   | probes, config, version, metrics | wire in `buildRouter` | `cmd/server/main.go` |

   Never bypass the global chain: RequestID → Logger → Recoverer → SecurityHeaders → CORS → RateLimit → BodyLimit → CSRFProtection. A new **state-changing** route that reads cookies **must** sit behind CSRFProtection.

3. **Auth decision** — cert/status/config/i18n/version/probes/metrics are unauthenticated **by design** (private-network threat model). Do not add auth without an explicit product decision; do not remove it from admin routes.

4. **Frontend client** — add the method to `src/lib/api.ts` via the internal `request<T>()` / `requestVoid()` helpers (they already set `credentials: 'same-origin'` and parse `body.error`). Add types to `src/lib/types.ts`. Components never call `fetch` directly.

5. **i18n** — any new user-visible error or label goes through `t()` and `internal/i18n` in all five languages. See the `vcv-i18n-string` skill.

6. **Tests**
   - Table-driven handler tests, `httptest.NewRequest` + `httptest.NewRecorder`, named `Test<Func>_<Scenario>`.
   - Stub the Vault boundary with `vault.MockClient` (see `vcv-high-coverage-tests`).
   - Admin-only endpoint: cover session **and** CSRF behavior (`internal/handlers` + `internal/middleware`).
   - Assert status, headers, and JSON body.

7. **Docs** — new or changed route: update the API table in `app/README.md`.

## Verification

```bash
cd app && go test ./internal/handlers/... -cover
make go-lint
make test-offline
make web-check
make web-test
```
