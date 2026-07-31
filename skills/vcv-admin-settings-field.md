---
description: Add or change an admin settings field in vcv
---

# Add or change an admin settings field

Use this skill when the admin settings API or UI gains a new field.

## Steps

1. **Config schema** — add the field to the relevant struct in `app/internal/config/config.go` with JSON tags. Keep validation in the config layer; do not scatter checks across handlers.

2. **Merge semantics** — update `mergeAdminSettings` in `app/internal/handlers/admin_api.go`.
   - Use `mergeSecret` / `mergeVaultTokens` for any secret-like value (tokens, webhook URLs).
   - Ensure `maskSecrets` blanks secrets in GET responses and that PUT preserves the stored value when the client sends blank or a mask sentinel.

3. **Admin UI** — update `src/lib/components/admin/AdminPanel.svelte`, `VaultEditor.svelte`, and `AdminDocsModal.svelte` as needed. All copy via `t()` for en, fr, de, it, es.

4. **Examples** — update `settings.example.json` (and `settings.enhanced-metrics.example.json` if relevant). Never put real tokens in examples.

5. **Docs** — if the field is operator-facing, update `app/README.md` and the admin docs under `app/internal/docs/`.

## Verification

```bash
cd app && go test ./internal/config/... ./internal/handlers/... -cover
make web-check
make web-test
make test-offline
```
