---
name: vcv-high-coverage-tests
description: Write or improve Go tests in vcv/app using table-driven subtests, vault.MockClient, httptest, and an in-memory web FS. Use when asked to add coverage, test a handler or router, or fix a failing/flaky Go test.
---

# Add high-coverage Go tests

## Goal

- **≥ 90% package coverage** for new or refactored packages.
- Offline-runnable (`make test-offline`) — no real Vault, no network.
- `testify/assert` + `testify/mock` at external boundaries only.

## Steps

1. **Mock the Vault boundary.** `internal/vault` ships `MockClient` (`internal/vault/mock_client.go`) implementing `Client`.

   ```go
   primary := &vault.MockClient{}
   primary.On("CheckConnection", mock.Anything).Return(nil)

   multi := &vault.MockClient{}
   multi.On("ListCertificates", mock.Anything).Return([]certs.Certificate{}, nil)
   ```

   `ListCertificates` type-asserts safely, but `GetCertificateDetails`, `GetCertificatePEM`, and `GetIntermediateCA` do a bare `args.Get(0).(T)` — returning `nil` there **panics**. Always return a zero-value struct or an empty slice, never `nil`.

2. **Pick the level.**
   - Handler: `httptest.NewRequest` + `httptest.NewRecorder`.
   - Router (middleware chain, static routes, wiring): `buildRouter` from `cmd/server` with an in-memory FS and a fresh registry.

   ```go
   webFS := fstest.MapFS{
       "dist/index.html":    &fstest.MapFile{Data: []byte("<!doctype html><html><body>ok</body></html>")},
       "dist/admin.html":    &fstest.MapFile{Data: []byte("<!doctype html><html><body>admin</body></html>")},
       "dist/assets/app.js": &fstest.MapFile{Data: []byte("console.log('ok')")},
   }
   registry := prometheus.NewRegistry() // never the global default registry
   router, err := buildRouter(cfg, primary, statusClients, multi, registry, webFS, "", nil)
   ```

3. **Shape.** Table-driven subtests named `Test<Func>_<Scenario>`. Assert status, headers, and JSON body. Prefer the `_test` package suffix for black-box tests.

4. **Vault package itself.** For `internal/vault`, stand up an `httptest.Server` and point a real client at it rather than mocking — that is where the HTTP contract is worth exercising.

5. **Cover the failure paths.** The interesting coverage here is partial-vault failure (`certificates` + `errors[]` envelope), CSRF rejection, rate-limit exemptions, and sanitized error strings — not more happy-path getters.

## Verification

```bash
cd app && go test ./internal/<pkg>/... -cover
cd app && go test -race ./...
make go-lint
make test-offline
```
