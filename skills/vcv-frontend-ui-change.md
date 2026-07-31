---
description: Add or change a Svelte UI component in vcv
---

# Add or change a Svelte UI component

Use this skill for frontend work in `app/web/frontend/`.

## Steps

1. **Structure** — use Svelte 5 runes (`$state`, `$derived`, `$effect`). Keep fetch/business logic in `src/lib/api.ts`, stores under `src/lib/stores/*.svelte.ts`, or `src/lib/utils/*`; avoid large logic blocks in markup.

2. **Components** — compose bits-ui / existing `src/lib/components/ui/*` primitives before inventing styled one-offs. Reuse existing `vcv-*` CSS classes and Tailwind utilities; respect `.stylelintrc.json`.

3. **i18n** — every user-visible string goes through `t(key, fallback?, params?)`.
   - Add the key to `internal/i18n` for **all** languages: en, fr, de, it, es.
   - No bare user-visible string literals in Svelte.
   - Drop unused keys.

4. **Async safety** — for modal/async loads, ignore stale responses when the open cert id or generation changes (see `CertDetailModal` pattern).

5. **State** — for shareable filter/pagination state, use the URL-state utilities; page sizes come from `config` store / `/api/config` thresholds.

## Verification

```bash
make web-check
make web-test
```

Add a colocated `*.test.ts` for new components or utils when behavior is non-trivial.
