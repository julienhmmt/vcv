# VaultCertsViewer 🔐

VaultCertsViewer (vcv) is a lightweight web UI that lists and inspects certificates stored in one or more Hashicorp Vault or OpenBao PKI mounts, especially their expiration dates and SANs.

OpenBao compatible: VCV works seamlessly with both Hashicorp Vault and OpenBao, as they share the same PKI API. Tested with OpenBao 2.4 and Vault 1.21 (as of 01/2026).

## ✨ What it does

- Discovers all certificates in Vault/OpenBao PKI mounts and shows them in a searchable, filterable table
- Multi-PKI engine support with modal selector to choose which mounts to display
- Shows common names (CN), SANs, and certificate details
- Displays status distribution (valid/expired/revoked) and upcoming expirations
- Highlights certificates expiring soon (7/30 days) with configurable thresholds
- Real-time Vault connection status with toast notifications
- Multi-language support (en, fr, es, de, it) and dark/light theme

## 🚀 Quick start

### Hashicorp Vault

```bash
vault policy write vcv - <<'EOF'
path "pki/certs"    { capabilities = ["list"] }
path "pki/certs/*"  { capabilities = ["read","list"] }
path "sys/health"   { capabilities = ["read"] }
EOF

vault write auth/token/roles/vcv allowed_policies="vcv" orphan=true period="24h"
vault token create -role="vcv" -policy="vcv" -period="24h" -renewable=true
```

### OpenBao

```bash
bao policy write vcv - <<'EOF'
path "pki/certs"    { capabilities = ["list"] }
path "pki/certs/*"  { capabilities = ["read","list"] }
path "sys/health"   { capabilities = ["read"] }
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
    image: jhmmt/vcv:1.5
    container_name: vcv
    restart: unless-stopped
    ports:
      - "52000:52000"
    environment:
      - SETTINGS_PATH=/app/settings.json
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
  -e "SETTINGS_PATH=/app/settings.json" \
  --cap-drop=ALL --read-only --security-opt no-new-privileges:true \
  -p 52000:52000 jhmmt/vcv:1.5
```

## 🔐 TLS configuration

VCV supports custom CA bundles and TLS settings per Vault instance through `settings.json`:

- `tls_ca_cert_base64`: base64-encoded PEM CA bundle
- `tls_ca_cert`: file path to PEM CA bundle  
- `tls_server_name`: TLS SNI server name override
- `tls_insecure`: disable TLS verification (dev only)

## 🎛️ Admin panel (optional)

Enable admin panel for UI-based configuration editing:

```yaml
environment:
  - VCV_ADMIN_PASSWORD=$2y$10$<BCRYPT_HASH_HERE>
volumes:
  - ./settings.json:/app/settings.json:rw  # Must be read-write
```

## 🌍 Features

- **Multi-Vault support**: Connect to multiple Vault/OpenBao instances
- **Multi-PKI engine**: Monitor multiple PKI mounts simultaneously  
- **Real-time status**: Live connection monitoring with notifications
- **Dark mode**: Complete dark theme support
- **Internationalization**: 5 languages with automatic detection
- **Configurable alerts**: Custom expiration thresholds
- **Admin panel**: Web-based configuration management
- **Security hardened**: Read-only container with minimal privileges

## 📚 Documentation

Full documentation available on [GitHub](https://github.com/julienhmmt/vcv)

## 🏷️ Tags

- `jhmmt/vcv:1.5` - Latest stable release
- `jhmmt/vcv:latest` - Always latest version
- `jhmmt/vcv:dev` - Development build
