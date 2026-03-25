# VaultCertsViewer 🔐

VaultCertsViewer (vcv) is a lightweight web UI that lists and inspects certificates stored in one or more HashiCorp Vault or OpenBao PKI mounts, especially their expiration dates and SANs.

OpenBao compatible: VCV works seamlessly with both HashiCorp Vault and OpenBao, as they share the same PKI API. Tested with OpenBao 2.4+ and Vault 1.20+ (as of 02/2026).

## ✨ What it does

- Discovers all certificates in Vault/OpenBao PKI mounts and shows them in a searchable, filterable table with pagination (25/50/100/all)
- Multi-Vault support: connect to multiple Vault/OpenBao instances simultaneously
- Multi-PKI engine support with modal selector to choose which mounts to display
- Shows common names (CN), SANs, and certificate details with creation/expiration dates
- Displays status distribution (valid/expired/revoked) and upcoming expirations
- Highlights certificates expiring soon with configurable thresholds (default: 7/30 days)
- Real-time Vault connection status with header indicator
- Multi-language support (en, fr, es, de, it) and dark/light theme
- Admin panel for web-based configuration management (optional, bcrypt-protected)
- Prometheus metrics endpoint with enhanced metrics and per-certificate options
- Security hardened: rate limiting (300 req/min), CSRF protection, security headers

## 🚀 Quick start

### Hashicorp Vault

Option 1: Explicit mounts (recommended for production):

```bash
vault policy write vcv - <<'EOF'
path "pki/certs"    { capabilities = ["list"] }
path "pki/certs/*"  { capabilities = ["read","list"] }
path "pki2/certs"   { capabilities = ["list"] }
path "pki2/certs/*" { capabilities = ["read","list"] }
path "sys/health"   { capabilities = ["read"] }
EOF

vault write auth/token/roles/vcv allowed_policies="vcv" orphan=true period="24h"
vault token create -role="vcv" -policy="vcv" -period="24h" -renewable=true
```

Option 2: Wildcard pattern (for dynamic environments):

```bash
vault policy write vcv - <<'EOF'
path "pki*/certs"    { capabilities = ["list"] }
path "pki*/certs/*"  { capabilities = ["read","list"] }
path "sys/health"    { capabilities = ["read"] }
EOF

vault write auth/token/roles/vcv allowed_policies="vcv" orphan=true period="24h"
vault token create -role="vcv" -policy="vcv" -period="24h" -renewable=true
```

### OpenBao

Option 1: Explicit mounts (recommended for production):

```bash
bao policy write vcv - <<'EOF'
path "pki/certs"    { capabilities = ["list"] }
path "pki/certs/*"  { capabilities = ["read","list"] }
path "pki2/certs"   { capabilities = ["list"] }
path "pki2/certs/*" { capabilities = ["read","list"] }
path "sys/health"   { capabilities = ["read"] }
EOF

bao write auth/token/roles/vcv allowed_policies="vcv" orphan=true period="24h"
bao token create -role="vcv" -policy="vcv" -period="24h" -renewable=true
```

Option 2: Wildcard pattern (for dynamic environments):

```bash
bao policy write vcv - <<'EOF'
path "pki*/certs"    { capabilities = ["list"] }
path "pki*/certs/*"  { capabilities = ["read","list"] }
path "sys/health"    { capabilities = ["read"] }
EOF

bao write auth/token/roles/vcv allowed_policies="vcv" orphan=true period="24h"
bao token create -role="vcv" -policy="vcv" -period="24h" -renewable=true
```

## 📋 Configuration

Create a `settings.json` file:

```json
{
  "app": {
    "env": "prod",
    "port": 52000,
    "logging": {
      "level": "info",
      "format": "json",
      "output": "stdout"
    }
  },
  "certificates": {
    "expiration_thresholds": {
      "critical": 7,
      "warning": 30
    }
  },
  "metrics": {
    "per_certificate": false,
    "enhanced_metrics": true
  },
  "cors": {
    "enabled": false,
    "allowed_origins": []
  },
  "vaults": [
    {
      "id": "main",
      "enabled": true,
      "display_name": "Main Vault",
      "address": "https://vault.example.com:8200",
      "token": "s.REPLACE_ME",
      "pki_mounts": ["pki", "pki2"],
      "tls_insecure": false
    }
  ]
}
```

## 🐳 Docker compose

```yaml
services:
  vcv:
    image: jhmmt/vcv:1.6.1.1.1.1
    container_name: vcv
    restart: unless-stopped
    ports:
      - "52000:52000"
    volumes:
      - ./settings.json:/app/settings.json:rw
```

```bash
docker compose up -d
```

## 🐳 Docker run

```bash
docker run -d \
  --name vcv \
  -v "$(pwd)/settings.json:/app/settings.json:rw" \
  --cap-drop=ALL --read-only --security-opt no-new-privileges:true \
  -p 52000:52000 jhmmt/vcv:1.6.1
```

## 🔐 TLS configuration

VCV supports custom CA bundles and TLS settings per Vault instance through `settings.json`:

- `tls_ca_cert_base64`: base64-encoded PEM CA bundle (preferred)
- `tls_ca_cert`: file path to PEM CA bundle
- `tls_ca_path`: directory containing CA certs
- `tls_server_name`: TLS SNI server name override
- `tls_insecure`: disable TLS verification (dev only)

Precedence: `tls_ca_cert_base64` overrides `tls_ca_cert`/`tls_ca_path`.

## 📊 Prometheus metrics

VCV exposes metrics at `/metrics` endpoint with comprehensive certificate monitoring:

**Key metrics:**

- `vcv_certificates_total{vault_id, pki, status}` - Certificate counts by status
- `vcv_certificates_expiring_soon_count{vault_id, pki, level}` - Expiring certificates
- `vcv_vault_connected{vault_id}` - Connection status
- `vcv_certificates_last_fetch_timestamp_seconds` - Last successful scrape

**Configuration:**

```json
"metrics": {
  "per_certificate": false,    // High cardinality metrics
  "enhanced_metrics": true     // Detailed bucket metrics
}
```

## 🎛️ Admin panel (optional)

To enable the administration panel, you have to edit the `settings.json` file and write the admin bcrypt hash password in the `admin.password` field.

```json
...
  "admin": {
    "password": "$2y$10$.changeme"
  },
...
```

Also, do not forget to make the `settings.json` file read-write, to be able to edit and save parameters.

```yaml
volumes:
  - ./settings.json:/app/settings.json:rw
```

## 🌍 Features

- **Multi-Vault support**: Connect to multiple Vault/OpenBao instances simultaneously
- **Multi-PKI engine**: Monitor multiple PKI mounts with modal selector and certificate counts
- **Real-time search**: Instant filtering with 300ms debouncing
- **Status filtering**: Quick filters for valid/expired/revoked certificates
- **Pagination**: Configurable page sizes (25/50/100/all) with navigation
- **Sort options**: By vault, PKI mount, common name, creation/expiration dates
- **Real-time status**: Live Vault connection monitoring with header indicator
- **Internationalization**: 5 languages with URL parameter support (?lang=xx)
- **Configurable thresholds**: Custom expiration thresholds (critical/warning)
- **Admin panel**: Web-based configuration management (bcrypt-protected)
- **Prometheus metrics**: Built-in metrics endpoint with enhanced options
- **Read-only container**: Supports running with minimal privileges

## 📚 Documentation

Full documentation available on [GitHub](https://github.com/julienhmmt/vcv)

## 🏷️ Tags

- `jhmmt/vcv:1.6.1` - Latest stable release
- `jhmmt/vcv:latest` - Always latest version
- `jhmmt/vcv:beta` - Development build
