---
name: vcv-security-check
description: Security review checklist for vcv before shipping — secrets masking, CSRF, rate limits, sanitized errors, PEM/private-key boundaries, TLS to Vault. Use before finishing any change that touches handlers, middleware, config, admin, or examples.
---

# Security review before shipping

Threat model: read-only cert inventory on a **private network**, fronted by network ACL / reverse proxy. That is a deliberate decision, not an oversight — do not "fix" it unilaterally.

## Checklist

- **Public API auth** — `/api/certs*`, `/api/status`, `/api/config`, `/api/i18n`, `/api/version`, probes, and `/metrics` are intentionally unauthenticated. Adding app-level auth needs an explicit product decision. Admin routes stay behind the session cookie.
- **Secrets** — never log, return, or commit cleartext Vault tokens, admin passwords, or webhook URLs. Admin GET masks (`maskSecrets`); PUT preserves the stored value on blank or `***` (`mergeSecret` / `mergeVaultTokens`).
- **Error sanitization** — public status/error strings are stable and leak no raw Vault internals, mount paths, or upstream URLs.
- **CSRF** — every new state-changing route that reads cookies sits behind `CSRFProtection` (same-origin Origin/Referer).
- **Rate limit + body limit** — always on. Exemptions only for `/api/health`, `/api/ready`, `/metrics`, `/assets/`. Login keeps its tighter limit.
- **Private keys** — PEM endpoints return public X.509 only. Nothing fetches private keys from Vault.
- **TLS to Vault** — prefer CA material in examples and prod guidance; `tls_insecure: true` is lab-only and must stay labeled as such.
- **Examples, logs, screenshots** — no real tokens or hostnames in `settings.example.json`, `settings.enhanced-metrics.example.json`, log samples, or docs images.
- **Request logging** — no request bodies that may carry passwords, no full PEMs.
- **Dependencies** — no new dep pulled in for this change without a look at what it does at runtime.

## Verification

```bash
cd app && go test ./internal/handlers/... ./internal/middleware/... ./internal/config/... -cover
make go-lint-full
make test-offline
git diff --stat && git diff -- '*.example.json'
```

Last pass: skim the diff for hardcoded credentials, a widened CORS origin, a new exempt path, or a removed middleware line.
