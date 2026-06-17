# VaultCertsViewer — Frontend / UI / UX Audit & Enhancement Plan

_Date: 2026-06-17 · Scope: `app/web/frontend/` (Svelte 5 + TS + Vite + Tailwind v4 + bits-ui)_

## Verdict

The frontend is **already mature and well-built**: rune-based stores, full i18n (5 langs, Go-served keys), dark mode via `data-theme`, accessibility basics (skip link, `aria-sort`, `aria-pressed`, `/` search shortcut), skeleton loading, toasts, a polished "passport" cert detail modal, donut overview, and an admin panel. This is not a rescue job — it is a polish-and-harden job.

**One real bug found (mobile cert list is invisible).** Fix that first. The rest is genuine enhancement.

---

## P0 — Bugs (ship immediately)

### 1. Mobile certificate list renders nothing (≤768px)
- **Evidence:** [vcv.css:611-630](app/web/frontend/src/styles/vcv.css#L611-L630) hides `.vcv-table-wrapper` with `display:none` at ≤768px and styles `.vcv-certs-mobile-cards` / `.vcv-cert-card`. **No component renders that markup** — grep of all `.svelte` files returns zero matches. [App.svelte:431-560](app/web/frontend/src/App.svelte#L431-L560) only renders the table.
- **Impact:** On phones/narrow viewports the entire cert table disappears with no fallback. Core feature unusable on mobile.
- **Fix:** Add a mobile card list in `App.svelte` (extract `CertCard.svelte`) that maps the same `paged` rows into the already-styled `.vcv-cert-card` markup. Reuse `expiryLabel`, `certStatus`, `statusMeta`. Sort/pagination controls stay shared.

---

## P1 — High-value enhancements

### 2. Shareable / bookmarkable filter state (URL sync)
- **Today:** `search`, `statusFilters`, `certTypeFilter`, `mountFilter`, `sortKey/Dir`, `pageSize` are local `$state` only — no `URLSearchParams` usage anywhere. Reloading or sharing a link loses the view.
- **Plan:** Add a small `lib/utils/url-state.ts` that serializes filter/sort/page state to the query string and rehydrates on mount via an `$effect`. Enables "send teammate the link to all critical certs in mount X".

### 3. Command palette (Cmd/Ctrl-K)
- **Today:** The cmdk wrapper is already vendored under [ui/command/](app/web/frontend/src/lib/components/ui/command/) but **unused**.
- **Plan:** Wire a `CommandPalette.svelte` for quick jump-to-cert (by CN/serial/SAN), quick status filters, theme toggle, language switch, open admin. Big UX win for power users with many certs; low cost since the primitive exists.

### 4. Export filtered inventory (CSV / JSON)
- **Today:** Only single-cert PEM download exists ([CertDetailModal.svelte:81-93](app/web/frontend/src/lib/components/CertDetailModal.svelte#L81-L93)). No way to export the list.
- **Plan:** Add "Export" button near pagination → CSV/JSON of the **currently filtered+sorted** set (CN, SANs, mount, vault, expiry, status, serial). Cert inventory/audit is the product's core job; export closes the loop. Client-side blob, no backend change.

### 5. `prefers-reduced-motion` support
- **Today:** 51 `transition`/`animation`/`@keyframes` in `vcv.css`, **zero** `prefers-reduced-motion` guards. Accessibility + vestibular gap (spinner, donut, row hovers, modal transitions).
- **Plan:** Add one global `@media (prefers-reduced-motion: reduce)` block that neutralizes non-essential motion. ~10 lines.

---

## P2 — Maintainability & quality

### 6. Split the 3,944-line `vcv.css` monolith
- **Evidence:** [vcv.css](app/web/frontend/src/styles/vcv.css) is one 3.9k-line file — violates the project's own "800 lines max / many small files" rule.
- **Plan:** Split by domain into `styles/` partials imported from one entry: `tokens.css` (the `:root` + `[data-theme="dark"]` variables), `base.css`, `header.css`, `table.css`, `cards.css`, `overview.css`, `detail-modal.css`, `admin.css`, `footer.css`, `responsive.css`. Pure mechanical move; no visual change. Tailwind v4 `@import` already supported (see recent commit `3f3c805`).

### 7. Remove dead scaffolding
- `lib/Counter.svelte`, `assets/vite.svg`, `assets/svelte.svg` — Vite-template leftovers, zero references in source.
- **Plus:** the orphaned `.vcv-certs-mobile-cards` CSS becomes live once #1 lands (don't delete — use it).

### 8. Table performance for large cert sets
- **Today:** Full filter→sort→paginate over `certs.certificates` recomputes on every keystroke; fine for hundreds, degrades for thousands. No virtualization.
- **Plan (defer until needed):** Debounce search input; if real deployments exceed ~2k certs, add row virtualization (e.g. `@tanstack/svelte-virtual`) to the desktop table. Measure before adding the dep.

---

## P3 — Nice-to-have polish

- **Richer empty/error states.** Table empty row is plain text ([App.svelte:505-512](app/web/frontend/src/App.svelte#L505-L512)). Add an illustrated empty state distinguishing "no certs at all" vs "none match filters" (with a one-tap Clear-filters).
- **Configurable cert auto-refresh.** Status polls every 10s ([App.svelte:119](app/web/frontend/src/App.svelte#L119)) but certs only refresh manually. Offer an opt-in interval (off by default) with a "last updated" timestamp.
- **Donut interactivity.** Pure CSS conic-gradient ([Donut.svelte](app/web/frontend/src/lib/components/Donut.svelte)) — add segment hover tooltips + click-to-filter to match the stat buttons.
- **HTML `<meta>` polish.** [index.html](app/web/frontend/index.html) lacks `description` and `theme-color` (light/dark) meta tags. `theme-color` improves mobile browser chrome.
- **Semantics of clickable rows.** `<tr role="button" tabindex="0">` ([App.svelte:517-524](app/web/frontend/src/App.svelte#L517-L524)) works but is unusual for SR users; consider a focusable cell button or keep and document.

---

## Suggested execution order

| Phase | Items | Risk | Notes |
|-------|-------|------|-------|
| 1 | #1 mobile cards | low | **Bug** — do first, markup + reuse existing CSS |
| 2 | #5 reduced-motion, #7 dead files, #9/#12 meta+empty states | low | quick wins, no logic change |
| 3 | #2 URL state, #4 export | med | new utils, well-isolated |
| 4 | #3 command palette | med | primitive exists, mostly wiring |
| 5 | #6 CSS split | low/tedious | mechanical, do when not mid-feature |
| 6 | #8 virtualization | med | only if cert counts demand it |

## Out of scope / already good

i18n coverage, dark mode, toasts, skeletons, cert detail modal design, admin panel, focus-ring tokens, keyboard `/` shortcut — leave as-is.

## Test plan (per item)

- **#1:** Chrome DevTools responsive 375/768px — cert list visible, tappable, opens detail modal; Lighthouse mobile pass.
- **#2:** Apply filters → copy URL → reload/new tab → identical view.
- **#3:** Cmd/Ctrl-K opens, fuzzy-finds by CN/serial/SAN, Enter opens cert.
- **#4:** Export with filters active → file rows == on-screen filtered rows.
- **#5:** OS "reduce motion" on → no spinner spin / donut animation; layout intact.
- **#6:** Visual diff before/after split — pixel-identical in light + dark.
