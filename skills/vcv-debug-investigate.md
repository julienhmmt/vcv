---
description: Systematic debugging workflow for vcv
---

# Systematic debugging

Use this skill when investigating a bug, regression, or unexpected behavior.

## Steps

1. **Orient with graphify** — if `graphify-frontend/`, `graphify-backend/`, or `graphify-full/` exist, query them first to find god nodes and surprising edges, then read the cited code. Verify in source afterward.

2. **Reproduce** — build a minimal repro (request, component state, or config). Capture structured logs with the request ID (`HTTPEvent` / `HTTPError` / `PanicEvent`).

3. **Root cause** — identify the upstream cause; prefer a minimal upstream fix over a downstream workaround. Do not silence errors without understanding them.

4. **Fix and regression test** — add or update a table-driven test that fails before the fix and passes after. Keep the change scoped to the cause.

5. **Scope discipline** — one branch per fix; do not pile unrelated changes together.

## Verification

Run the relevant checks from the AGENTS.md verification matrix for the change type, then the broader suites:

```bash
make test-offline
make web-check
make web-test
```
