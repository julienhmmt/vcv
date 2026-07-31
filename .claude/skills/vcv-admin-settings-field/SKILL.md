---
name: vcv-admin-settings-field
description: Add or change a field in vcv admin settings (config struct, merge/mask semantics, admin UI, examples). Use when the admin settings API or panel gains a field — especially secret-like values such as tokens or webhook URLs.
---

# Add or change an admin settings field

Admin settings are read from and written back to a settings file. Secrets are masked on GET and preserved on PUT, so a new field has to opt into that machinery explicitly.

## Steps

1. **Config schema** — add the field to the relevant struct in `app/internal/config/config.go` with a JSON tag. Validation stays in the config layer; do not scatter checks across handlers.

2. **Merge semantics** — `mergeAdminSettings` in `app/internal/handlers/admin_api.go` (`admin_api.go:237`).
   - Secret-like value (token, webhook URL, password)? Route it through `mergeSecret` (`:254`) or `mergeVaultTokens` (`:295`) so a blank or masked (`***`) incoming value keeps the stored one.
   - Add it to `maskSecrets` (`:266`) so GET never returns cleartext.
   - Plain value: merge it explicitly. A field missing from `mergeAdminSettings` is silently dropped on every PUT.

3. **Admin UI** — `src/lib/components/admin/AdminPanel.svelte`, `VaultEditor.svelte`, `AdminDocsModal.svelte` as applicable. All copy via `t()` in five languages (`vcv-i18n-string`).

4. **Examples** — update `settings.example.json` (and `settings.enhanced-metrics.example.json` if relevant). Placeholder values only, never a real token.

5. **Docs** — operator-facing field: update `app/README.md` and the embedded admin docs in `app/internal/docs/ADMIN.md`.

## Test what actually breaks

- GET returns the field masked, not cleartext.
- PUT with blank ⇒ stored value survives.
- PUT with `***` ⇒ stored value survives.
- PUT with a new value ⇒ new value persists.
- Round-trip: PUT then GET does not corrupt unrelated fields.

## Verification

```bash
cd app && go test ./internal/config/... ./internal/handlers/... -cover
make go-lint
make test-offline
make web-check
make web-test
```
