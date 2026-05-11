#!/usr/bin/env sh
set -eu

# Generic Vault dev init script.
# Usage: vault-dev-init.sh <instance_id> <mount1> [mount2 ...]
# Starts vault dev server then enables PKI mounts, generates root CAs,
# creates roles, issues test certs (valid/expiring/expired/revoked/type coverage).

INSTANCE_ID="${1:-vcv}"
shift || true
MOUNTS="${*:-pki pki_dev pki_stage pki_production}"

VAULT_ADDR_INTERNAL="http://127.0.0.1:8200"
VAULT_DEV_LISTEN="0.0.0.0:8200"
VAULT_ROOT_TOKEN="root"

export VAULT_ADDR="${VAULT_ADDR_INTERNAL}"
export VAULT_TOKEN="${VAULT_ROOT_TOKEN}"

vault server -dev \
  -dev-root-token-id="${VAULT_ROOT_TOKEN}" \
  -dev-listen-address="${VAULT_DEV_LISTEN}" &
VAULT_PID=$!

printf "[%s] waiting for vault" "${INSTANCE_ID}"
while ! vault status >/dev/null 2>&1; do
  printf "."
  sleep 0.5
done
printf " ready\n"

issue_cert() {
  mount="$1"; cn="$2"; ttl="$3"; alt="${4:-}"
  issue_cert_with_role "${mount}" "vcv" "${cn}" "${ttl}" "${alt}"
}

issue_cert_with_role() {
  mount="$1"; role="$2"; cn="$3"; ttl="$4"; alt="${5:-}"
  if [ -n "${alt}" ]; then
    vault write "${mount}/issue/${role}" common_name="${cn}" alt_names="${alt}" ttl="${ttl}" >/dev/null 2>&1 || true
  else
    vault write "${mount}/issue/${role}" common_name="${cn}" ttl="${ttl}" >/dev/null 2>&1 || true
  fi
}

issue_and_revoke() {
  mount="$1"; cn="$2"
  out=$(vault write -format=json "${mount}/issue/vcv" common_name="${cn}" ttl="720h" 2>/dev/null) || return 0
  serial=$(echo "${out}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
  [ -n "${serial}" ] && vault write "${mount}/revoke" serial_number="${serial}" >/dev/null 2>&1 || true
}

for mount in ${MOUNTS}; do
  printf "[%s] configuring mount %s\n" "${INSTANCE_ID}" "${mount}"

  vault secrets enable -path="${mount}" pki 2>/dev/null || true
  vault secrets tune -max-lease-ttl=8760h "${mount}" 2>/dev/null || true

  vault write -force "${mount}/root/generate/internal" \
    common_name="${INSTANCE_ID}-${mount}.local" \
    ttl="8760h" >/dev/null 2>&1 || true

  vault write "${mount}/config/urls" \
    issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/${mount}/ca" \
    crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/${mount}/crl" >/dev/null 2>&1 || true

  vault write "${mount}/roles/vcv" \
    allow_any_name=true \
    allow_bare_domains=true \
    allow_subdomains=true \
    enforce_hostnames=false \
    max_ttl="8760h" \
    ttl="8760h" \
    not_before_duration="30s" >/dev/null 2>&1 || true

  vault write "${mount}/roles/vcv-machine" \
    allow_any_name=true \
    allow_bare_domains=true \
    allow_subdomains=true \
    enforce_hostnames=false \
    server_flag=true \
    client_flag=false \
    code_signing_flag=false \
    email_protection_flag=false \
    max_ttl="8760h" \
    ttl="8760h" \
    not_before_duration="30s" >/dev/null 2>&1 || true

  vault write "${mount}/roles/vcv-user" \
    allow_any_name=true \
    allow_bare_domains=true \
    allow_subdomains=true \
    enforce_hostnames=false \
    server_flag=false \
    client_flag=true \
    code_signing_flag=false \
    email_protection_flag=false \
    max_ttl="8760h" \
    ttl="8760h" \
    not_before_duration="30s" >/dev/null 2>&1 || true

  vault write "${mount}/roles/vcv-both" \
    allow_any_name=true \
    allow_bare_domains=true \
    allow_subdomains=true \
    enforce_hostnames=false \
    server_flag=true \
    client_flag=true \
    code_signing_flag=false \
    email_protection_flag=false \
    max_ttl="8760h" \
    ttl="8760h" \
    not_before_duration="30s" >/dev/null 2>&1 || true

  vault write "${mount}/roles/vcv-unknown" \
    allow_any_name=true \
    allow_bare_domains=true \
    allow_subdomains=true \
    enforce_hostnames=false \
    server_flag=false \
    client_flag=false \
    code_signing_flag=true \
    email_protection_flag=false \
    max_ttl="8760h" \
    ttl="8760h" \
    not_before_duration="30s" >/dev/null 2>&1 || true

  base="${INSTANCE_ID}-${mount}"
  issue_cert_with_role "${mount}" "vcv-machine" "machine-web.${base}.local" "8760h" "www.machine-web.${base}.local,api.machine-web.${base}.local"
  issue_cert_with_role "${mount}" "vcv-machine" "machine-api.${base}.local" "720h" "admin.machine-api.${base}.local"
  issue_cert_with_role "${mount}" "vcv-user" "user-alice.${base}.local" "8760h" ""
  issue_cert_with_role "${mount}" "vcv-user" "user-bob.${base}.local" "720h" ""
  issue_cert_with_role "${mount}" "vcv-both" "mtls-service.${base}.local" "8760h" "client.mtls-service.${base}.local"
  issue_cert_with_role "${mount}" "vcv-unknown" "codesign.${base}.local" "8760h" ""

  issue_cert "${mount}" "web.${base}.local"   "8760h" "www.web.${base}.local,api.web.${base}.local"
  issue_cert "${mount}" "api.${base}.local"   "8760h" "admin.api.${base}.local"
  issue_cert "${mount}" "admin.${base}.local" "8760h" ""
  issue_cert "${mount}" "db.${base}.local"    "8760h" "replica.db.${base}.local"

  issue_cert "${mount}" "expiring-24h.${base}.local"  "24h"   ""
  issue_cert "${mount}" "expiring-48h.${base}.local"  "48h"   ""
  issue_cert "${mount}" "expiring-week.${base}.local" "168h"  ""
  issue_cert "${mount}" "expiring-month.${base}.local" "720h" ""
  issue_cert "${mount}" "longterm.${base}.local"      "8760h" ""

  issue_cert "${mount}" "expired-1.${base}.local" "2s" ""
  issue_cert "${mount}" "expired-2.${base}.local" "2s" ""

  issue_and_revoke "${mount}" "revoked-1.${base}.local"
  issue_and_revoke "${mount}" "revoked-2.${base}.local"

  vault read "${mount}/crl/rotate" >/dev/null 2>&1 || true
done

sleep 3

printf "[%s] PKI init done. Mounts: %s\n" "${INSTANCE_ID}" "${MOUNTS}"

wait "${VAULT_PID}"
