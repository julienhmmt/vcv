#!/usr/bin/env sh
set -eu

# Generic OpenBao dev init script.
# Usage: openbao-dev-init.sh <instance_id> <mount1> [mount2 ...]
# Starts bao dev server then enables PKI mounts, generates root CAs,
# creates roles, issues test certs (valid/expiring/expired/revoked).

INSTANCE_ID="${1:-openbao}"
shift || true
MOUNTS="${*:-pki_openbao pki_openbao_dev pki_openbao_stage pki_openbao_prod}"

BAO_ADDR_INTERNAL="http://127.0.0.1:1337"
BAO_DEV_LISTEN="0.0.0.0:1337"
BAO_ROOT_TOKEN="root"

export BAO_ADDR="${BAO_ADDR_INTERNAL}"
export BAO_TOKEN="${BAO_ROOT_TOKEN}"
export VAULT_ADDR="${BAO_ADDR_INTERNAL}"
export VAULT_TOKEN="${BAO_ROOT_TOKEN}"

bao server -dev \
  -dev-root-token-id="${BAO_ROOT_TOKEN}" \
  -dev-listen-address="${BAO_DEV_LISTEN}" &
BAO_PID=$!

printf "[%s] waiting for openbao" "${INSTANCE_ID}"
while ! bao status >/dev/null 2>&1; do
  printf "."
  sleep 0.5
done
printf " ready\n"

issue_cert() {
  mount="$1"; cn="$2"; ttl="$3"; alt="${4:-}"
  if [ -n "${alt}" ]; then
    bao write "${mount}/issue/vcv" common_name="${cn}" alt_names="${alt}" ttl="${ttl}" >/dev/null 2>&1 || true
  else
    bao write "${mount}/issue/vcv" common_name="${cn}" ttl="${ttl}" >/dev/null 2>&1 || true
  fi
}

issue_and_revoke() {
  mount="$1"; cn="$2"
  out=$(bao write -format=json "${mount}/issue/vcv" common_name="${cn}" ttl="720h" 2>/dev/null) || return 0
  serial=$(echo "${out}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
  [ -n "${serial}" ] && bao write "${mount}/revoke" serial_number="${serial}" >/dev/null 2>&1 || true
}

for mount in ${MOUNTS}; do
  printf "[%s] configuring mount %s\n" "${INSTANCE_ID}" "${mount}"

  bao secrets enable -path="${mount}" pki 2>/dev/null || true
  bao secrets tune -max-lease-ttl=8760h "${mount}" 2>/dev/null || true

  bao write -force "${mount}/root/generate/internal" \
    common_name="${INSTANCE_ID}-${mount}.local" \
    ttl="8760h" >/dev/null 2>&1 || true

  bao write "${mount}/config/urls" \
    issuing_certificates="${BAO_ADDR_INTERNAL}/v1/${mount}/ca" \
    crl_distribution_points="${BAO_ADDR_INTERNAL}/v1/${mount}/crl" >/dev/null 2>&1 || true

  bao write "${mount}/roles/vcv" \
    allow_any_name=true \
    allow_bare_domains=true \
    allow_subdomains=true \
    enforce_hostnames=false \
    max_ttl="8760h" \
    ttl="8760h" \
    not_before_duration="30s" >/dev/null 2>&1 || true

  base="${INSTANCE_ID}-${mount}"
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

  bao read "${mount}/crl/rotate" >/dev/null 2>&1 || true
done

sleep 3

printf "[%s] PKI init done. Mounts: %s\n" "${INSTANCE_ID}" "${MOUNTS}"

wait "${BAO_PID}"
