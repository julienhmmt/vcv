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

printf "[vcv-4] Waiting for Vault dev to be ready"
while ! vault status >/dev/null 2>&1; do
  printf "."
  sleep 0.5
done
printf " done\n"

vault secrets enable -path=pki_vault4 pki 2>/dev/null || true
vault secrets enable -path=pki_lab pki 2>/dev/null || true
vault secrets enable -path=pki_qa pki 2>/dev/null || true
vault secrets enable -path=pki_perf pki 2>/dev/null || true

vault secrets tune -max-lease-ttl=8760h pki_vault4 2>/dev/null || true
vault secrets tune -max-lease-ttl=8760h pki_lab 2>/dev/null || true
vault secrets tune -max-lease-ttl=8760h pki_qa 2>/dev/null || true
vault secrets tune -max-lease-ttl=8760h pki_perf 2>/dev/null || true

vault write -force pki_vault4/root/generate/internal common_name="vcv-vault4.local" ttl="8760h" >/dev/null 2>&1 || true
vault write -force pki_lab/root/generate/internal common_name="vcv-lab.local" ttl="8760h" >/dev/null 2>&1 || true
vault write -force pki_qa/root/generate/internal common_name="vcv-qa.local" ttl="8760h" >/dev/null 2>&1 || true
vault write -force pki_perf/root/generate/internal common_name="vcv-perf.local" ttl="8760h" >/dev/null 2>&1 || true

vault write pki_vault4/config/urls issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_vault4/ca" crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_vault4/crl" >/dev/null 2>&1 || true
vault write pki_lab/config/urls issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_lab/ca" crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_lab/crl" >/dev/null 2>&1 || true
vault write pki_qa/config/urls issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_qa/ca" crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_qa/crl" >/dev/null 2>&1 || true
vault write pki_perf/config/urls issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_perf/ca" crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_perf/crl" >/dev/null 2>&1 || true

vault write pki_vault4/roles/vcv allowed_domains="vault4.local" allow_bare_domains=true allow_subdomains=true max_ttl="8760h" ttl="8760h" not_before_duration="30s" >/dev/null 2>&1 || true
vault write pki_lab/roles/vcv allowed_domains="lab.local" allow_bare_domains=true allow_subdomains=true max_ttl="8760h" ttl="8760h" not_before_duration="30s" >/dev/null 2>&1 || true
vault write pki_qa/roles/vcv allowed_domains="qa.local" allow_bare_domains=true allow_subdomains=true max_ttl="8760h" ttl="8760h" not_before_duration="30s" >/dev/null 2>&1 || true
vault write pki_perf/roles/vcv allowed_domains="perf.local" allow_bare_domains=true allow_subdomains=true max_ttl="8760h" ttl="8760h" not_before_duration="30s" >/dev/null 2>&1 || true

for i in 1 2 3 4 5 6 7 8 9 10; do
  vault write pki_vault4/issue/vcv common_name="service-${i}.vault4.local" alt_names="api.service-${i}.vault4.local,www.service-${i}.vault4.local" >/dev/null 2>&1 || true
done
for i in 1 2 3 4 5; do
  vault write pki_vault4/issue/vcv common_name="short-${i}.vault4.local" ttl="120s" >/dev/null 2>&1 || true
done
for i in 1 2 3 4; do
  vault write pki_vault4/issue/vcv common_name="soon-${i}.vault4.local" ttl="48h" >/dev/null 2>&1 || true
done
printf "[vcv-4] Creating expired certificates (waiting 3s)...\n"
vault write pki_vault4/issue/vcv common_name="expired-1.vault4.local" ttl="2s" >/dev/null 2>&1 || true
vault write pki_vault4/issue/vcv common_name="expired-2.vault4.local" ttl="2s" >/dev/null 2>&1 || true
sleep 3

for i in 1 2 3 4 5 6; do
  vault write pki_lab/issue/vcv common_name="lab-${i}.lab.local" ttl="720h" >/dev/null 2>&1 || true
  vault write pki_qa/issue/vcv common_name="qa-${i}.qa.local" ttl="240h" >/dev/null 2>&1 || true
  vault write pki_perf/issue/vcv common_name="perf-${i}.perf.local" ttl="8760h" >/dev/null 2>&1 || true
done

REVOKE_OUTPUT=$(vault write -format=json pki_vault4/issue/vcv common_name="revoked.vault4.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv-4] Revoking certificate %s from pki_vault4\n" "${REVOKE_SERIAL}"
  vault write pki_vault4/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi
vault read pki_vault4/crl/rotate >/dev/null 2>&1 || true
vault read pki_lab/crl/rotate >/dev/null 2>&1 || true
vault read pki_qa/crl/rotate >/dev/null 2>&1 || true
vault read pki_perf/crl/rotate >/dev/null 2>&1 || true

wait "${VAULT_PID}"
