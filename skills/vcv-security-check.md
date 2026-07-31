---
description: Security review checklist for vcv changes
---

# Security review before shipping

Run this checklist before calling a change done.

## Checklist

- **Public API auth** — `/api/certs*`, `/api/status`, `/api/config`, probes, `/api/i18n`, `/api/version`, and `/metrics` are intentionally unauthenticated for private networks. Do not add app-level auth without an explicit product decision.
- **Secrets** — never log, return, or commit cleartext Vault tokens or webhook URLs. Admin GET masks secrets; PUT preserves stored values when the field is blank or masked (`***`).
- **Error sanitization** — public status/error strings are stable and contain no raw Vault internals.
- **CSRF** — unsafe methods with cookies require same-origin Origin/Referer via existing middleware. New state-changing routes that use cookies must sit behind CSRFProtection.
- **Rate limit and body limit** — always on. Exempt paths only for health/ready/metrics/assets.
- **PEM endpoints** — return public X.509 only; never fetch private keys from Vault.
- **TLS to Vault** — examples and prod guidance prefer CA material; `tls_insecure: true` is lab-only.
- **Examples and logs** — no secrets in `settings.example.json`, logs, or screenshots docs.

## Verification

```bash
cd app && go test ./internal/handlers/... ./internal/middleware/... -cover
make go-lint
make test-offline
```
