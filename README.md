# VaultCertsViewer

VaultCertsViewer (vcv) is a lightweight web UI that lists and inspects certificates stored in a HashiCorp Vault PKI mount, especially their expiration dates and SANs.

Currently, VaultCertsViewer (vcv) can only view one mount at a time. If you have (for example) 4 mounts, you'll need 4 instances of vcv.

## What it does

- Discovers all certificates in a Vault PKI and shows them in a searchable, filterable table.
- Shows common names (CN) and SANs.
- Displays status distribution (valid / expired / revoked) and upcoming expirations.
- Highlights certificates expiring soon (7/30 days) and shows details (CN, SAN, fingerprints, issuer, validity).
- Lets you pick UI language (en, fr, es, de, it) and theme (light/dark).

## Why it exists

The native Vault UI is heavy and not convenient for quickly checking certificate expirations and details. VaultCertsViewer gives platform / security / ops teams a fast, **read-only** view of the Vault PKI inventory with only the essential information.

## Who should use it

- Teams operating Vault PKI who need visibility on their certificates.
- Operators who want a ready-to-use browser view alongside Vault CLI or Web UI.

## How to deploy and use

In HashiCorp Vault, create a read-only role and token for the API to reach the target PKI engine (adjust `pki` if you use another mount):

```bash
vault policy write vcv - <<'EOF'
path "pki/certs"   { capabilities = ["list"] }
path "pki/cert/*"  { capabilities = ["read"] }
path "sys/health"  { capabilities = ["read"] }
EOF
vault write auth/token/roles/vcv allowed_policies="vcv" orphan=true period="24h"
vault token create -role="vcv" -policy="vcv" -period="24h" -renewable=true
```

This dedicated token limits permissions to certificate listing/reading, can be renewed, and is used as `VAULT_READ_TOKEN` by the app.

### docker-compose

Grab `docker-compose.yml`, put it in a directory, then run:

```bash
docker compose up -d
```

No storage needed, unless you want to log to a file.

### docker run

Start the container with this command:

```bash
docker run -d \
  -e "APP_ENV=prod" \
  -e "LOG_FORMAT=json" \
  -e "LOG_OUTPUT=stdout" \
  -e "VAULT_ADDR=http://changeme:8200" \
  -e "VAULT_READ_TOKEN=changeme" \
  -e "VAULT_PKI_MOUNT=changeme" \
  -e "VAULT_TLS_INSECURE=true" \
  -e "LOG_LEVEL=info" \
  --cap-drop=ALL --read-only --security-opt no-new-privileges:true \
  -p 52000:52000 jhmmt/vcv:1.0
```

## Translations

The UI is localized in English, French, Spanish, German, and Italian. Language is selectable in the header or via `?lang=xx`.

## More details

- Technical documentation: [app/README.md](app/README.md)
- French overview: [README.fr.md](README.fr.md)
