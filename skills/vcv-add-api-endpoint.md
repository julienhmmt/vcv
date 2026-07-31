---
description: Add a JSON API endpoint to vcv end-to-end
---

# Add a JSON API endpoint

Use this skill when adding or changing an HTTP API in `app/`.

## Steps

1. **Handler** — add the handler in `app/internal/handlers/<name>.go`.
   - Define request/response structs with JSON tags.
   - Match existing error shapes (`error` field when clients expect it) and sanitize public errors.
   - Log with `logger.HTTPError` and the request ID from `middleware.GetRequestID`.

2. **Route registration**
   - Cert-related: `RegisterCertRoutes`.
   - Admin: `RegisterAdminRoutes` (session cookie required).
   - i18n: `RegisterI18nRoutes`.
   - Probes/config/version: wire in `cmd/server/main.go`.
   - Do not bypass global middleware (RequestID, Logger, Recoverer, SecurityHeaders, CORS, RateLimit, BodyLimit, CSRF).

3. **Frontend client** — add a method to `app/web/frontend/src/lib/api.ts` using `request` or `requestVoid` (already sets `credentials: 'same-origin'`). Add types to `lib/types.ts`.

4. **Tests**
   - Table-driven handler tests with `httptest`.
   - Use `vault.MockClient` for Vault I/O; return empty slices, not `nil`.
   - If the endpoint is admin-only, cover session/CSRF behavior under `internal/handlers` + `internal/middleware`.

## Verification

```bash
cd app && go test ./internal/handlers/... -cover
make test-offline
make web-check
make web-test
```
