# Frontend Improvement Plan

Transform the VaultCertsViewer frontend from "Good" (15/20) to "Excellent" (18-20/20) through systematic improvements across accessibility, performance, theming, and code architecture.

## Audit Summary

### Current Score: 15/20 (Good)

| Dimension | Score | Status |
| --------- | ----- | ------ |
| Accessibility | 3/4 | Critical modal focus trap issue |
| Performance | 3/4 | Large CSS bundle |
| Theming | 3/4 | Hardcoded values persist |
| Responsive Design | 3/4 | Touch targets too small |
| Anti-Patterns | 3/4 | Clean design, mixed patterns |

### Target Score: 18-20/20 (Excellent)

---

## Phase 1: Foundation (Colors & Tokens) — P0

### 1.1 Unify Color System

**Goal**: Eliminate hardcoded hex values in `vcv.css`, align with OKLCH theme

**Changes**:

- Map all `vcv.css` hex colors to CSS custom properties
- Create semantic color tokens: `--vcv-status-valid`, `--vcv-status-warning`, etc.
- Ensure dark mode uses proper OKLCH values
- Keep legacy CSS variables as fallbacks during transition

**Files**:

- `app/web/frontend/src/styles/vcv.css` — refactor color definitions
- `app/web/frontend/src/app.css` — add vcv-* tokens to @theme

**Verification**: No hex values in vcv.css except in comments showing OKLCH equivalents

---

## Phase 2: Accessibility Critical — P0

### 2.1 Migrate Modal System to Bits-UI

**Goal**: Remove broken custom focus trap, use proven primitives

**Changes**:

- Delete `Modal.svelte` (or deprecate with warning)
- Update `CertDetailModal.svelte` to use `Dialog.Root` from bits-ui
- Update `CAModal.svelte` similarly
- Verify `MountSelectorDialog.svelte` uses bits-ui correctly
- Ensure `Dialog.Content` has `showCloseButton={true}` where needed

**Files**:

- `app/web/frontend/src/lib/components/Modal.svelte` — delete
- `app/web/frontend/src/lib/components/CertDetailModal.svelte` — migrate
- `app/web/frontend/src/lib/components/CAModal.svelte` — migrate

**Verification**: Tab navigation cycles through modal elements, Escape closes, focus returns to trigger

### 2.2 Add Non-Color Status Indicators

**Goal**: WCAG 1.4.1 compliance — status not conveyed by color alone

**Changes**:

- Add Lucide icons to status badges: `CheckCircle`, `AlertTriangle`, `AlertOctagon`, `XCircle`, `Ban`
- Update `CertStatusBadge.svelte` to show icon + text
- Update `StatusOverview.svelte` segments to include icons alongside dots
- Update filter chips with icons

**Files**:

- `app/web/frontend/src/lib/components/CertStatusBadge.svelte` — add icon prop
- `app/web/frontend/src/lib/components/StatusOverview.svelte` — add icons to segments
- `app/web/frontend/src/lib/utils/cert-status.ts` — add `statusIcon()` function
- `app/web/frontend/src/app.css` — icon color styles

**Verification**: Status badges display icon + text; colorblind simulation shows distinguishable states

---

## Phase 3: Responsive & Touch — P1

### 3.1 Increase Touch Targets

**Goal**: WCAG 2.5.5 compliance — minimum 44px touch targets

**Changes**:

- Filter chips: increase padding to `min-height: 44px`
- Mount selector items: `padding: 12px 16px` minimum
- Table rows: verify clickable area
- Pagination buttons: ensure 44px height
- Language switcher: verify touch target

**Files**:

- `app/web/frontend/src/app.css` — update `.vcv-filter-chip`, `.vcv-msd-mount`, `.vcv-msd-vault`
- `app/web/frontend/src/styles/vcv.css` — verify existing patterns

**Verification**: Chrome DevTools mobile simulation shows all interactive elements >=44px

---

## Phase 4: Architecture & Performance — P1

### 4.1 Component-Scoped CSS

**Goal**: Reduce CSS bundle, improve maintainability

**Changes**:

- Extract component-specific styles from `vcv.css` to co-located `<style>` blocks
- Priority components: `CertDetailModal`, `MountSelectorDialog`, `StatusOverview`
- Keep shared utilities (badges, buttons, tables) in vcv.css
- Use Svelte's scoped styles with `:global()` for deep selectors

**Files**:

- `app/web/frontend/src/styles/vcv.css` — reduce to shared utilities only
- `app/web/frontend/src/lib/components/CertDetailModal.svelte` — add scoped styles
- `app/web/frontend/src/lib/components/MountSelectorDialog.svelte` — add scoped styles
- `app/web/frontend/src/lib/components/StatusOverview.svelte` — add scoped styles

**Verification**: `vcv.css` under 500 lines; no visual regressions

---

## Phase 5: Semantic Markup — P2

### 5.1 Proper Table Structure

**Goal**: WCAG 1.3.1 compliance — programmatic table relationships

**Changes**:

- Wrap headers in `<thead>`
- Wrap body in `<tbody>`
- Add `scope="col"` to header cells
- Consider `aria-sort` for sortable columns

**Files**:

- `app/web/frontend/src/App.svelte` — table structure around lines 330-420

**Verification**: Screen reader announces "column header, sortable" correctly

---

## Phase 6: Polish & Cleanup — P2

### 6.1 i18n String Cleanup

**Goal**: Consistent internationalization patterns

**Changes**:

- Review all `i18n.t(key, fallback)` calls
- Move descriptive fallbacks to i18n bundle
- Keep minimal fallbacks (empty string or key itself) in code

**Files**:

- `app/web/frontend/src/App.svelte` — audit all i18n calls
- `app/web/frontend/src/lib/components/*.svelte` — audit i18n calls

### 6.2 Final Verification

**Goal**: Confirm all improvements landed

**Verification Checklist**:

- [ ] Audit score 18+ / 20
- [ ] No hex colors in vcv.css
- [ ] All modals use bits-ui with proper focus trap
- [ ] Status indicators have icons
- [ ] All touch targets >=44px
- [ ] Table has proper semantics
- [ ] Zero console warnings
- [ ] TypeScript strict mode passes
- [ ] Visual regression: side-by-side before/after screenshots

---

## Execution Order

| Phase | Task | Effort | Dependencies |
| ----- | ---- | ------ | ------------ |
| 1 | Color system unification | Medium | None |
| 2.1 | Modal migration | Small | Phase 1 (color tokens) |
| 2.2 | Status icons | Small | Phase 1 |
| 3 | Touch targets | Small | None |
| 4 | CSS splitting | Large | Phase 1, 2 |
| 5 | Table semantics | Small | None |
| 6 | i18n cleanup | Small | None |
| 6.2 | Final verification | Small | All above |

---

## Risk Mitigation

1. **Visual Regression**: Take screenshots before each phase; compare pixel-by-pixel
2. **Dark Mode Breakage**: Test theme switching after color changes
3. **Bundle Size**: Monitor CSS bundle size; target <50KB gzipped
4. **Accessibility**: Run axe-core after each accessibility fix

---

## Files to Modify

```
app/web/frontend/src/
├── styles/vcv.css                    # Phase 1, 4: Color tokens, extract component styles
├── app.css                           # Phase 1: Add vcv-* tokens
├── lib/components/
│   ├── Modal.svelte                  # Phase 2.1: Delete
│   ├── CertDetailModal.svelte        # Phase 2.1, 4: Migrate to bits-ui, scoped styles
│   ├── CAModal.svelte                # Phase 2.1: Migrate to bits-ui
│   ├── MountSelectorDialog.svelte    # Phase 3, 4: Touch targets, scoped styles
│   ├── StatusOverview.svelte         # Phase 2.2, 4: Status icons, scoped styles
│   ├── CertStatusBadge.svelte        # Phase 2.2: Add icons
│   └── ui/dialog/dialog-content.svelte # Phase 2.1: Verify close button behavior
├── lib/utils/cert-status.ts          # Phase 2.2: Add statusIcon() function
└── App.svelte                        # Phase 5, 6: Table semantics, i18n cleanup
```

---

## Success Criteria

The implementation is complete when:

1. **Audit Score**: Frontend audit score reaches 18+/20
2. **Color System**: Zero hardcoded hex values in component CSS; all colors use semantic tokens
3. **Accessibility**: All P0/P1 issues resolved; axe-core passes with zero violations
4. **Performance**: CSS bundle <50KB gzipped; no render-blocking resources
5. **Responsive**: All touch targets >=44px on mobile viewports
6. **Maintainability**: Component styles co-located; shared utilities clearly documented

---

*Generated from impeccable audit on 2026-06-16*
