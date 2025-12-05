#!/usr/bin/env sh
set -eu

# Start Vault dev server and run PKI initialization commands.
# This script is meant to be used as the container command in docker-compose.dev.yml.

VAULT_ADDR_INTERNAL="http://127.0.0.1:8200"
VAULT_DEV_LISTEN="0.0.0.0:8200"
VAULT_ROOT_TOKEN="root"

export VAULT_ADDR="${VAULT_ADDR_INTERNAL}"
export VAULT_TOKEN="${VAULT_ROOT_TOKEN}"

# Start Vault dev in background
vault server \
  -dev \
  -dev-root-token-id="${VAULT_ROOT_TOKEN}" \
  -dev-listen-address="${VAULT_DEV_LISTEN}" &
VAULT_PID=$!

# Wait for Vault to be reachable
printf "[vcv] Waiting for Vault dev to be ready"
while ! vault status >/dev/null 2>&1; do
  printf "."
  sleep 0.5
done
printf " done\n"

# Enable and configure PKI
# In dev mode the storage is in-memory, so these commands will run on each start.

# Enable PKI at path pki/ (idempotent: ignore error if already enabled)
vault secrets enable -path=pki pki 2>/dev/null || true

# Tune max TTL
vault secrets tune -max-lease-ttl=8760h pki 2>/dev/null || true

# Generate an internal root CA (force to avoid interactive prompts)
vault write -force pki/root/generate/internal \
  common_name="vcv.local" \
  ttl="8760h" >/dev/null 2>&1 || true

# Configure CRL URLs so Vault auto-generates CRL
vault write pki/config/urls \
  issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki/ca" \
  crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki/crl" >/dev/null 2>&1 || true

# Create a role for issuing test certificates (allow very short TTL for expired certs)
vault write pki/roles/vcv \
  allowed_domains="internal" \
  allow_bare_domains=true \
  allow_subdomains=true \
  max_ttl="720h" \
  ttl="720h" \
  not_before_duration="30s" >/dev/null 2>&1 || true

# Issue valid certificates
vault write pki/issue/vcv \
  common_name="example.internal" \
  alt_names="www.example.internal" >/dev/null 2>&1 || true

vault write pki/issue/vcv \
  common_name="api.internal" \
  alt_names="api.internal" >/dev/null 2>&1 || true

# Issue a certificate expiring soon (24h)
vault write pki/issue/vcv \
  common_name="expiring-soon.internal" \
  ttl="24h" >/dev/null 2>&1 || true

# Issue a certificate expiring in 7 days
vault write pki/issue/vcv \
  common_name="expiring-week.internal" \
  ttl="168h" >/dev/null 2>&1 || true

# Issue an expired certificate (TTL 2s, then wait)
printf "[vcv] Creating expired certificate (waiting 3s)...\n"
vault write pki/issue/vcv \
  common_name="expired.internal" \
  ttl="2s" >/dev/null 2>&1 || true
sleep 3

# Issue a certificate to revoke
REVOKE_OUTPUT=$(vault write -format=json pki/issue/vcv \
  common_name="revoked.internal" \
  ttl="720h" 2>/dev/null) || true

REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true

if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv] Revoking certificate %s\n" "${REVOKE_SERIAL}"
  vault write pki/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi

# Force CRL rotation to include revoked cert
vault read pki/crl/rotate >/dev/null 2>&1 || true

printf "[vcv] Vault dev PKI initialized:\n"
printf "  - Mount: pki/\n"
printf "  - Role: vcv\n"
printf "  - Valid certs: example.internal, api.internal\n"
printf "  - Expiring soon: expiring-soon.internal (24h), expiring-week.internal (7d)\n"
printf "  - Expired cert: expired.internal\n"
printf "  - Revoked cert: revoked.internal\n"

# Keep the Vault process in foreground
wait "${VAULT_PID}"
