# VaultCertsViewer — Admin Guide

The admin panel lets you inspect and change the running configuration of
VaultCertsViewer (VCV) without restarting the binary. All changes are written
back to your `settings.json` and applied to the live Vault registry.

## Enabling the admin panel

The panel is only available when `admin.password` in `settings.json` holds a
valid **bcrypt** hash. If the field is missing or is not a bcrypt hash
(`$2a$`, `$2b$`, or `$2y$` prefix), the panel and its API are disabled.

Generate a hash and drop it into `settings.json`:

```json
{
  "admin": {
    "password": "$2b$12$....your-bcrypt-hash...."
  }
}
```

Sign in with username **`admin`** and the password matching that hash.

> Sessions last 12 hours in dev and 4 hours in prod. Sign-in is rate limited
> to 5 attempts per 3 minutes per client IP.

## Expiration thresholds

Two values, in **days**, drive the colored status of every certificate:

- **Critical** — certs expiring within this many days show as critical.
- **Warning** — certs expiring within this many days show as warning.

Defaults are 7 (critical) and 30 (warning). Set Critical below Warning.

## Metrics

- **Per-certificate metrics** — emits one Prometheus series per certificate.
  High cardinality; off by default. Enable only with a bounded cert count.
- **Enhanced metrics** — additional aggregate gauges and counters. On by
  default.

## CORS allowed origins

A comma-separated list of origins permitted to call the JSON API from a
browser, e.g. `https://app.example.com, https://other.example.com`. Leave
empty to keep the API same-origin only.

## Vaults

Each vault row maps to one HashiCorp Vault or OpenBao instance:

- **Address** — the API address, e.g. `https://vault.example.com:8200`.
- **Token** — a read-capable token. Leave the field blank when editing to
  keep the existing token; it is never sent back to the browser.
- **PKI mounts** — one or more PKI mount paths to enumerate (e.g. `pki`,
  `pki_int`).
- **Enabled** — toggle a vault on or off without removing it. Disabled vaults
  stay configured so you can re-enable them instantly.
- **TLS options** — skip-verify and custom CA settings for self-signed setups.

**Add vault** appends a new empty row; fill in an ID, address, and token
before saving. **Remove** deletes the row on save. The first enabled vault is
treated as the *primary*.

Connectivity is checked live: each enabled vault shows whether VCV can reach
it with the configured token.

## Invalidate cache

VCV caches certificate listings in memory with a TTL. **Invalidate cache**
clears it immediately so the next request re-reads from Vault — useful right
after issuing or revoking a certificate.

## Saving

**Save** validates the whole form, writes `settings.json` atomically, then
reloads the Vault registry in place. Invalid input (bad address, missing
token, duplicate vault ID, out-of-range threshold) is rejected with an error
and nothing is written.
