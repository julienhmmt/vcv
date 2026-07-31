---
name: vcv-frontend-ui-change
description: Add or change a Svelte 5 component, store, or util in app/web/frontend. Use for any UI work in vcv — new component, filter/table/modal change, styling, or frontend state.
---

# Add or change a Svelte UI component

Stack: Svelte 5 runes + TypeScript + Vite, Tailwind v4, bits-ui primitives under `src/lib/components/ui/`, `@lucide/svelte`, `svelte-sonner`. **pnpm only.**

## Steps

1. **Structure** — runes (`$state`, `$derived`, `$effect`); no legacy Svelte stores for new state. Fetch and business logic live in `src/lib/api.ts`, `src/lib/stores/*.svelte.ts`, or `src/lib/utils/*` — not in markup.

2. **Reuse before inventing** — compose existing `src/lib/components/ui/*` primitives and `vcv-*` CSS classes. Check `src/lib/utils/` first: `cert-filter`, `cert-status`, `cert-label`, `cert-icons`, `url-state`, `export`, `expiry-notify`, `config-thresholds`, `clipboard`. Respect `.stylelintrc.json`.

3. **i18n** — every user-visible string via `t('key', 'English fallback', params?)`. No bare literals in markup. New key ⇒ add it to `internal/i18n` for **en, fr, de, it, es** (see `vcv-i18n-string` — a missing translation renders blank, it does not fall back).

4. **Async safety** — modal/async loads must ignore stale responses when the open cert id or a generation counter changes (`CertDetailModal` pattern). Otherwise a slow response overwrites a newer one.

5. **Config, not constants** — expiration thresholds and page sizes come from the `config` store / `/api/config`. Never hardcode 7/30.

6. **Cert IDs** — `encodeURIComponent` in paths; composite IDs go through `parseCertID`.

7. **Partial failure is normal** — `GET /api/certs` returns `certificates` + `errors[]`. One dead vault warns; it does not empty the table.

8. **Accessibility** — labels on controls, keyboard reachable, focus trapped in dialogs (bits-ui handles this if you use the primitives).

9. **Test** — colocated `*.test.ts` (Vitest + jsdom) for non-trivial components and every new util.

## Verification

```bash
make web-check
make web-test
```
