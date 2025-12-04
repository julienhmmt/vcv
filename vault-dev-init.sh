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

# Create a role for issuing test certificates
vault write pki/roles/vcv \
  allowed_domains="internal" \
  allow_bare_domains=true \
  allow_subdomains=true \
  max_ttl="720h" >/dev/null 2>&1 || true

# Issue a few test certificates
vault write pki/issue/vcv \
  common_name="example.internal" \
  alt_names="www.example.internal" >/dev/null 2>&1 || true

vault write pki/issue/vcv \
  common_name="api.internal" \
  alt_names="api.internal" >/dev/null 2>&1 || true

vault write pki/issue/vcv \
  common_name="old.internal" \
  ttl="24h" >/dev/null 2>&1 || true

printf "[vcv] Vault dev PKI initialized (pki/, role vcv, example certs)\n"

# Keep the Vault process in foreground
wait "${VAULT_PID}"
