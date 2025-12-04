# vcv

VaultCertsViewer is a Go application to list and manage certs in Vault, the easiest possible for SRE.

## Development

```bash
docker compose -f docker-compose.dev.yml up -d --build
```

### Vault setup

```bash
docker exec -it vault ash
$ export VAULT_ADDR=http://127.0.0.1:8200
$ export VAULT_TOKEN=root
$ vault secrets enable -path=pki pki
$ vault secrets tune -max-lease-ttl=8760h pki
$ vault write pki/root/generate/internal common_name="vcv.local" ttl="8760h"
$ vault write pki/roles/vcv allowed_domains="internal" allow_bare_domains=true allow_subdomains=true max_ttl="720h"
$ vault write pki/issue/vcv common_name="example.internal" alt_names="www.example.internal"
$ vault write pki/issue/vcv common_name="api.internal" alt_names="api.internal"
$ vault write pki/issue/vcv common_name="old.internal" ttl="24h"
# create a policy 'read-only'
$ vault policy write vcv-read - <<'EOF'
path "pki/certs" {
  capabilities = ["list"]
}

path "pki/certs/revoked" {
  capabilities = ["list"]
}

path "pki/cert/*" {
  capabilities = ["read"]
}
EOF
# create a token with the policy 'read-only'
$ apk add --no-cache jq
$ vault token create -policy=vcv-read -display-name=vcv-read -format=json | jq -r '.auth.client_token'
# copy the value and paste it in the .env file on the 'VAULT_READ_TOKEN' line
```

To stop the containers:

```bash
docker compose -f docker-compose.dev.yml down
```

To remove the volumes:

```bash
docker compose -f docker-compose.dev.yml down -v
```

## Production

```bash
docker compose -f docker-compose.yml up -d
```
