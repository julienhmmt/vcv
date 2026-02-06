# Configuration reference

## 📋 Overview

VaultCertsViewer (VCV) is configured primarily through a `settings.json` file. The admin panel allows you to manage this file directly from the web interface. Environment variables are supported as a legacy fallback when no `settings.json` is found.

VCV uses a server-side rendered architecture powered by [HTMX](https://htmx.org/). All filtering, sorting, and pagination are handled server-side for optimal performance.

> **⚠️ Important:** After saving changes, a server restart may be required for all settings to take effect.

## 🔐 Admin panel access

### VCV_ADMIN_PASSWORD

Environment variable required to enable the admin panel. Must be a **bcrypt hash**.

```bash
# Generate a bcrypt hash (example with htpasswd)
htpasswd -nbBC 10 admin YourSecurePassword | cut -d: -f2

# Or with Python
python3 -c "import bcrypt; print(bcrypt.hashpw(b'YourPassword', bcrypt.gensalt()).decode())"

# Set the environment variable
export VCV_ADMIN_PASSWORD='$2a$10$...'
```

You can also use the 'bcrypt' service of <https://tools.hommet.net/bcrypt> to generate a bcrypt hash (no data is stored).

**Default username:** `admin` (not editable, default value)
**Session duration:** 12 hours (not editable, default value)
**Login rate limiting:** 10 attempts per 5 minutes (not editable, default value)

## 📁 Application settings

### Environment (app.env)

Defines the application environment. Affects security features and logging behavior.

- `dev` - Development mode (verbose logging, no rate limiting)
- `prod` - Production mode (secure cookies, rate limiting enabled)

**Default:** `prod`

### Port (app.port)

HTTP server listening port.

**Default:** `52000`

### Settings file path

The `SETTINGS_PATH` environment variable specifies the path to the `settings.json` file. If not set, VCV searches for files in this order:

1. `settings.<env>.json` (e.g., `settings.dev.json`)
2. `settings.json`
3. `./settings.json`
4. `/etc/vcv/settings.json`

### Logging (app.logging)

Configure application logging behavior:

- **level**: `debug`, `info`, `warn`, `error`
- **format**: `json` or `text`
- **output**: `stdout`, `file`, or `both`
- **file_path**: Log file path when output is `file` or `both`

**Defaults:**

- level: `info`
- format: `json`
- output: `stdout`
- file_path: `/var/log/app/vcv.log`

## 📜 Certificate settings

### Expiration thresholds (certificates.expiration_thresholds)

Configure when certificates are flagged as expiring soon:

- **critical**: Days before expiration to show critical alert
- **warning**: Days before expiration to show warning

These thresholds control:

- Notification banner at the top of the page
- Color coding in the certificate table (red for critical, yellow for warning)
- Timeline visualization on the dashboard
- Prometheus metrics (`vcv_certificates_expiring_critical`, `vcv_certificates_expiring_warning`)

**Defaults:**

- critical: `7`
- warning: `30`

## 🌐 CORS settings (cors)

### Allowed origins (cors.allowed_origins)

Array of allowed CORS origins. Use `["*"]` to allow all origins (not recommended in production).

**Example:**

```json
"allowed_origins": ["https://example.com", "https://app.example.com"]
```

### Allow credentials (cors.allow_credentials)

Boolean to allow credentials in CORS requests.

**Default:** `false`

**Note:** CORS is primarily useful if you embed VCV in another web application or access it from a different domain.

## 🔐 Vault configuration

### Multiple Vault instances

VaultCertsViewer supports monitoring multiple Vault instances simultaneously. Each vault instance requires:

- **ID**: Unique identifier for this Vault instance (required)
- **Display name**: Human-readable name shown in the UI (optional)
- **Address**: Vault server URL (e.g., `https://vault.example.com:8200`)
- **Token**: Read-only Vault token with PKI access (required)
- **PKI mounts**: Array of PKI mount paths (e.g., `["pki", "pki2", "pki-prod"]`)
- **Enabled**: Whether this Vault instance is active

### TLS configuration

For Vaults using custom CA certificates or self-signed certificates:

- **TLS CA cert (Base64)**: Base64-encoded PEM CA bundle (preferred method)
- **TLS CA cert path**: File path to PEM CA bundle
- **TLS CA path**: Directory containing CA certificates
- **TLS server name**: SNI server name override
- **TLS insecure**: Skip TLS verification (⚠️ development only, not recommended)

```bash
# Encode a certificate to base64
cat path-to-cert.pem | base64 | tr -d '\n'
```

**Precedence:** If `tls_ca_cert_base64` is set, it takes priority over file paths.

### Vault token permissions

The Vault token must have read-only access to PKI mounts. Example policy:

```hcl
# Explicit mounts (recommended for production)
vault policy write vcv - <<'EOF'
path "pki/certs"    { capabilities = ["list"] }
path "pki/certs/*"  { capabilities = ["read","list"] }
path "pki2/certs"   { capabilities = ["list"] }
path "pki2/certs/*" { capabilities = ["read","list"] }
path "sys/health"   { capabilities = ["read"] }
EOF

# Wildcard pattern (for dynamic environments)
vault policy write vcv - <<'EOF'
path "pki*/certs"    { capabilities = ["list"] }
path "pki*/certs/*"  { capabilities = ["read","list"] }
path "sys/health"    { capabilities = ["read"] }
EOF

# Create token
vault write auth/token/roles/vcv allowed_policies="vcv" orphan=true period="24h"
vault token create -role="vcv" -policy="vcv" -period="24h" -renewable=true
```

You must replace 'pki' and 'pki2' with the PKI mount paths of your Vault. Add as many PKI mount paths as you have in your Vault.

## ⚡ Performance optimizations

### Caching

VaultCertsViewer implements caching to improve performance:

- **Certificate cache TTL:** 15 minutes (reduces Vault API calls)
- **Health check cache:** 30 seconds (for header status indicator)
- **Parallel fetching:** Multiple Vaults are queried simultaneously
- **Cache invalidation:** Use the refresh button (↻) in the header or `POST /api/cache/invalidate` to clear the certificate cache

With multiple Vaults, parallel fetching provides **3-10× faster** loading times compared to sequential queries.

## 📊 Monitoring & metrics

### Prometheus metrics

Available at `/metrics` endpoint:

- `vcv_certificates_total` - Total number of certificates
- `vcv_certificates_valid` - Number of valid certificates
- `vcv_certificates_expired` - Number of expired certificates
- `vcv_certificates_revoked` - Number of revoked certificates
- `vcv_certificates_expiring_critical` - Certificates expiring within critical threshold
- `vcv_certificates_expiring_warning` - Certificates expiring within warning threshold
- `vcv_vault_connected` - Vault connection status (1=connected, 0=disconnected)
- `vcv_cache_size` - Number of cached entries
- `vcv_last_fetch_timestamp` - Unix timestamp of last certificate fetch

All metrics include labels: `vault_id`, `vault_name`, `pki_mount`

### Health & API endpoints

- `/api/health` - Basic health check (always returns 200 OK)
- `/api/ready` - Readiness probe (checks application state)
- `/api/status` - Detailed status including all Vault connections
- `/api/version` - Application version information
- `/api/config` - Application configuration (expiration thresholds, vault list)
- `/api/i18n` - Translations for the current language
- `/api/certs` - Certificate list (JSON)
- `/api/certs/{id}/details` - Certificate details (JSON)
- `/api/certs/{id}/pem` - Certificate PEM content (JSON)
- `/api/certs/{id}/pem/download` - Download certificate PEM file
- `POST /api/cache/invalidate` - Invalidate the certificate cache

### Rate limiting

In `prod` mode, API rate limiting is enabled at **300 requests per minute** per client. The following paths are exempt:

- `/api/health`, `/api/ready`, `/metrics`
- `/assets/*` (static files)

## 🔒 Security best practices

- Always use `prod` environment in production
- Protect `settings.json` file (contains sensitive tokens)
- Use read-only Vault tokens with minimal permissions
- Rate limiting is automatic in `prod` mode (300 req/min)
- CSRF protection is enabled on all state-changing requests
- Security headers (X-Content-Type-Options, X-Frame-Options, etc.) are set automatically
- Run container with `--read-only` and `--cap-drop=ALL`

## 📝 Example settings.json

```json
{
  "app": {
    "env": "prod",
    "port": 52000,
    "logging": {
      "level": "info",
      "format": "json",
      "output": "stdout",
      "file_path": "/var/log/app/vcv.log"
    }
  },
  "certificates": {
    "expiration_thresholds": {
      "critical": 7,
      "warning": 30
    }
  },
  "cors": {
    "allowed_origins": ["https://example.com"],
    "allow_credentials": false
  },
  "vaults": [
    {
      "id": "vault-prod",
      "display_name": "Production Vault",
      "address": "https://vault.example.com:8200",
      "token": "hvs.xxx",
      "pki_mounts": ["pki", "pki-intermediate"],
      "enabled": true,
      "tls_insecure": false,
      "tls_ca_cert_base64": "LS0tLS1CRUdJTi...",
      "tls_server_name": "vault.example.com"
    },
    {
      "id": "vault-dev",
      "display_name": "Development Vault",
      "address": "https://vault-dev.example.com:8200",
      "token": "hvs.yyy",
      "pki_mounts": ["pki_dev"],
      "enabled": true,
      "tls_insecure": true
    }
  ]
}
```

> **💡 Tip:** Use the admin panel to edit these settings visually. Changes are saved to `settings.json` file.
