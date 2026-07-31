---
description: Add high-coverage Go tests to vcv using testify/mock
---

# Add high-coverage Go tests to vcv

Use this skill when asked to add or improve Go test coverage in `app/`.

## Goal

- Target **>= 90% package coverage** for new or refactored packages.
- Keep tests offline-runnable (`make test-offline`).
- Use `github.com/stretchr/testify/assert` and `testify/mock` for external boundaries.

## Steps

1. Mock the Vault boundary. `internal/vault` exposes a `MockClient` based on testify/mock. Stub the methods the code under test calls, returning **empty slices** rather than `nil` to avoid type-assertion panics.

   ```go
   primary := &vault.MockClient{}
   primary.On("CheckConnection", mock.Anything).Return(nil)

   multi := &vault.MockClient{}
   multi.On("ListCertificates", mock.Anything).Return([]certs.Certificate{}, nil)
   ```

2. Build the router or handler.
   - Handler-level tests: `httptest.NewRecorder` + `httptest.NewRequest`.
   - Router-level tests: `buildRouter` from `cmd/server` with an in-memory `fstest.MapFS` web filesystem and a fresh `prometheus.NewRegistry()`.

   ```go
   webFS := fstest.MapFS{
       "dist/index.html":    &fstest.MapFile{Data: []byte("<!doctype html><html><body>ok</body></html>")},
       "dist/admin.html":    &fstest.MapFile{Data: []byte("<!doctype html><html><body>admin</body></html>")},
       "dist/assets/app.js": &fstest.MapFile{Data: []byte("console.log('ok')")},
   }
   registry := prometheus.NewRegistry()
   router, err := buildRouter(cfg, primary, statusClients, multi, registry, webFS, "", nil)
   ```

3. Write table-driven subtests named `Test<Func>_<Scenario>`; compare status, headers, and JSON body. Keep the `_test` package suffix for black-box tests.

4. For `internal/vault` integration-style tests, spin up an `httptest.Server` and point a real client at it instead of hitting a real Vault.

## Verification

```bash
cd app && go test ./internal/<pkg>/... -cover
make test-offline
```
