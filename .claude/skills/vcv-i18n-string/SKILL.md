---
name: vcv-i18n-string
description: Add, rename, or remove a user-visible string in vcv across all five languages (en, fr, de, it, es). Use whenever UI copy changes — Go compiles fine with a missing translation, so this needs a checklist.
---

# Add or change a UI string

Source of truth is `app/internal/i18n/i18n.go`. The frontend only reads `/api/i18n?lang=` and calls `t(key, fallback?, params?)`.

## The trap

`Messages` is a **struct**, not a map. Forgetting a language leaves that field as `""` — it compiles, it serializes as `"key": ""`, and `t()` returns `""` because `messages[key] ?? fallback` only falls back on `null`/`undefined`. Result: a **blank label in production for that language only**. Nothing catches this but this checklist.

## Steps

1. **Struct field** — add to `type Messages struct` in `internal/i18n/i18n.go`. Go field PascalCase, JSON tag camelCase. The **JSON tag is the key the frontend uses**.

2. **Five values** — add an entry to all five literals in the same file:
   `englishMessages`, `frenchMessages`, `spanishMessages`, `germanMessages`, `italianMessages`.
   Translate for real; do not paste English into the other four.

3. **Interpolation** — `t()` supports `{name}` and `{{name}}`. Match the style already used by neighboring keys. Unknown placeholders are left visible on purpose.

4. **Use it** — `t('exportSuccess', 'Export complete')` in the component or store. Never a bare literal. The fallback argument is for first paint only, not a substitute for step 2.

5. **Plurals** — this codebase uses separate keys (`daysRemaining` / `daysRemainingSingular`, `expiredDays` / `expiredDaysSingular`). Follow that, do not invent a plural engine.

6. **Remove dead keys** — renaming or deleting copy means dropping the struct field **and** all five entries. Unreferenced messages are dead weight.

7. **Adding a language** (rare) — new `Language` const, new `Messages` var, `MessagesForLanguage` + `GetLanguage` + `FromAcceptLanguage` cases, and `LANGUAGES` in `src/lib/stores/i18n.svelte.ts`. Then update every doc that says "five languages".

## Verification

```bash
cd app && go test ./internal/i18n/... -cover
make web-check
make web-test
```

Manual spot check: `curl -s 'localhost:52000/api/i18n?lang=de' | jq '.messages.<yourKey>'` — must not be `""`. Repeat for fr, it, es.
