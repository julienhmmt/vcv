#!/usr/bin/env sh
set -eu

VAULT_ADDR_INTERNAL="http://127.0.0.1:8200"
VAULT_DEV_LISTEN="0.0.0.0:8200"
VAULT_ROOT_TOKEN="root"

export VAULT_ADDR="${VAULT_ADDR_INTERNAL}"
export VAULT_TOKEN="${VAULT_ROOT_TOKEN}"

vault server \
  -dev \
  -dev-root-token-id="${VAULT_ROOT_TOKEN}" \
  -dev-listen-address="${VAULT_DEV_LISTEN}" &
VAULT_PID=$!

printf "[vcv-5] Waiting for Vault dev to be ready"
while ! vault status >/dev/null 2>&1; do
  printf "."
  sleep 0.5
done
printf " done\n"

vault secrets enable -path=pki_vault5 pki 2>/dev/null || true
vault secrets enable -path=pki_internal pki 2>/dev/null || true
vault secrets enable -path=pki_dmz pki 2>/dev/null || true
vault secrets enable -path=pki_shared pki 2>/dev/null || true

vault secrets tune -max-lease-ttl=8760h pki_vault5 2>/dev/null || true
vault secrets tune -max-lease-ttl=8760h pki_internal 2>/dev/null || true
vault secrets tune -max-lease-ttl=8760h pki_dmz 2>/dev/null || true
vault secrets tune -max-lease-ttl=8760h pki_shared 2>/dev/null || true

vault write -force pki_vault5/root/generate/internal common_name="vcv-vault5.local" ttl="8760h" >/dev/null 2>&1 || true
vault write -force pki_internal/root/generate/internal common_name="vcv-internal.local" ttl="8760h" >/dev/null 2>&1 || true
vault write -force pki_dmz/root/generate/internal common_name="vcv-dmz.local" ttl="8760h" >/dev/null 2>&1 || true
vault write -force pki_shared/root/generate/internal common_name="vcv-shared.local" ttl="8760h" >/dev/null 2>&1 || true

vault write pki_vault5/config/urls issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_vault5/ca" crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_vault5/crl" >/dev/null 2>&1 || true
vault write pki_internal/config/urls issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_internal/ca" crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_internal/crl" >/dev/null 2>&1 || true
vault write pki_dmz/config/urls issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_dmz/ca" crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_dmz/crl" >/dev/null 2>&1 || true
vault write pki_shared/config/urls issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_shared/ca" crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_shared/crl" >/dev/null 2>&1 || true

vault write pki_vault5/roles/vcv allowed_domains="vault5.local" allow_bare_domains=true allow_subdomains=true max_ttl="8760h" ttl="8760h" not_before_duration="30s" >/dev/null 2>&1 || true
vault write pki_internal/roles/vcv allowed_domains="internal.local" allow_bare_domains=true allow_subdomains=true max_ttl="8760h" ttl="8760h" not_before_duration="30s" >/dev/null 2>&1 || true
vault write pki_dmz/roles/vcv allowed_domains="dmz.local" allow_bare_domains=true allow_subdomains=true max_ttl="8760h" ttl="8760h" not_before_duration="30s" >/dev/null 2>&1 || true
vault write pki_shared/roles/vcv allowed_domains="shared.local" allow_bare_domains=true allow_subdomains=true max_ttl="8760h" ttl="8760h" not_before_duration="30s" >/dev/null 2>&1 || true

for i in 1 2 3 4 5 6 7 8 9 10 11 12; do
  vault write pki_vault5/issue/vcv common_name="svc-${i}.vault5.local" alt_names="api.svc-${i}.vault5.local,www.svc-${i}.vault5.local" >/dev/null 2>&1 || true
done
for i in 1 2 3 4 5 6; do
  vault write pki_vault5/issue/vcv common_name="short-${i}.vault5.local" ttl="90s" >/dev/null 2>&1 || true
done
for i in 1 2 3 4; do
  vault write pki_vault5/issue/vcv common_name="soon-${i}.vault5.local" ttl="36h" >/dev/null 2>&1 || true
done
for i in 1 2 3 4; do
  vault write pki_vault5/issue/vcv common_name="long-${i}.vault5.local" ttl="8760h" >/dev/null 2>&1 || true
done
printf "[vcv-5] Creating expired certificates (waiting 3s)...\n"
vault write pki_vault5/issue/vcv common_name="expired-1.vault5.local" ttl="2s" >/dev/null 2>&1 || true
vault write pki_vault5/issue/vcv common_name="expired-2.vault5.local" ttl="2s" >/dev/null 2>&1 || true
sleep 3

for i in 1 2 3 4 5; do
  vault write pki_internal/issue/vcv common_name="internal-${i}.internal.local" ttl="240h" >/dev/null 2>&1 || true
  vault write pki_dmz/issue/vcv common_name="dmz-${i}.dmz.local" ttl="72h" >/dev/null 2>&1 || true
  vault write pki_shared/issue/vcv common_name="shared-${i}.shared.local" ttl="720h" >/dev/null 2>&1 || true
done

REVOKE_OUTPUT=$(vault write -format=json pki_vault5/issue/vcv common_name="revoked.vault5.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv-5] Revoking certificate %s from pki_vault5\n" "${REVOKE_SERIAL}"
  vault write pki_vault5/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi
REVOKE_OUTPUT=$(vault write -format=json pki_internal/issue/vcv common_name="revoked.internal.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv-5] Revoking certificate %s from pki_internal\n" "${REVOKE_SERIAL}"
  vault write pki_internal/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi

vault read pki_vault5/crl/rotate >/dev/null 2>&1 || true
vault read pki_internal/crl/rotate >/dev/null 2>&1 || true
vault read pki_dmz/crl/rotate >/dev/null 2>&1 || true
vault read pki_shared/crl/rotate >/dev/null 2>&1 || true

wait "${VAULT_PID}"
