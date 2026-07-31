---
name: vcv-debug-investigate
description: Systematic root-cause debugging in vcv — reproduce, trace with request IDs, fix upstream, add a regression test. Use when investigating a bug, regression, flaky test, or unexpected UI/API behavior.
---

# Systematic debugging

**Iron law: no fix without a root cause.** A symptom patch in the caller leaves every sibling caller broken.

## Steps

1. **Orient** — if `graphify-frontend/`, `graphify-backend/`, or `graphify-full/` exist, query them before broad code archaeology:

   ```bash
   graphify query "How do cert list errors reach the UI?" --graph graphify-full/graph.json
   graphify explain "buildRouter" --graph graphify-backend/graph.json
   ```

   Then read the cited files. Always confirm in source — the graph can be stale.

2. **Reproduce** — smallest possible repro: one request, one component state, one settings file. Capture structured logs and correlate on the request ID (`HTTPEvent` / `HTTPError` / `PanicEvent`). No repro means no verified fix.

3. **Localize** — which layer actually owns the behavior?
   | Symptom | Look at |
   | --- | --- |
   | Wrong/missing certs, partial list | `internal/vault` (multi/registry), `certs.go` envelope |
   | 403 on a POST | `middleware` CSRF, session cookie |
   | 429 | RateLimit, exempt paths |
   | Setting reverts after save | `mergeAdminSettings` / `mergeSecret` |
   | Blank label in one language | `internal/i18n` missing entry (see `vcv-i18n-string`) |
   | Stale modal data | missing generation/id guard in the Svelte component |
   | Stale or empty UI after build | `app/web/dist` not rebuilt (`make web-build`) |

4. **Root cause** — grep every caller of the function before editing. Fix upstream where all callers route through; do not silence an error you do not understand.

5. **Regression test** — a table-driven Go test or a colocated `*.test.ts` that **fails before the fix and passes after**. Write it first, watch it fail.

6. **Scope discipline** — one branch, one cause. Unrelated cleanups go on their own branch.

## Verification

Run the checks for the layers you touched, then:

```bash
make go-lint
make test-offline
make web-check
make web-test
```
