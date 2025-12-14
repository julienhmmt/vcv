#!/usr/bin/env sh
set -eu

# Start Vault dev server and run PKI initialization commands for vault-dev-2.
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
printf "[vcv-2] Waiting for Vault dev to be ready"
while ! vault status >/dev/null 2>&1; do
  printf "."
  sleep 0.5
done
printf " done\n"

# Enable and configure PKI
# In dev mode the storage is in-memory, so these commands will run on each start.

# Enable PKI at path pki_vault2/ (idempotent: ignore error if already enabled)
vault secrets enable -path=pki_vault2 pki 2>/dev/null || true

# Enable additional PKI engines for vault-dev-2
vault secrets enable -path=pki_corporate pki 2>/dev/null || true
vault secrets enable -path=pki_external pki 2>/dev/null || true
vault secrets enable -path=pki_partners pki 2>/dev/null || true

# Tune max TTL for all engines
vault secrets tune -max-lease-ttl=8760h pki_vault2 2>/dev/null || true
vault secrets tune -max-lease-ttl=8760h pki_corporate 2>/dev/null || true
vault secrets tune -max-lease-ttl=8760h pki_external 2>/dev/null || true
vault secrets tune -max-lease-ttl=8760h pki_partners 2>/dev/null || true

# Generate root CAs for all engines (force to avoid interactive prompts)
vault write -force pki_vault2/root/generate/internal \
  common_name="vcv-vault2.local" \
  ttl="8760h" >/dev/null 2>&1 || true

vault write -force pki_corporate/root/generate/internal \
  common_name="vcv-corporate.local" \
  ttl="8760h" >/dev/null 2>&1 || true

vault write -force pki_external/root/generate/internal \
  common_name="vcv-external.local" \
  ttl="8760h" >/dev/null 2>&1 || true

vault write -force pki_partners/root/generate/internal \
  common_name="vcv-partners.local" \
  ttl="8760h" >/dev/null 2>&1 || true

# Configure CRL URLs for all engines
vault write pki_vault2/config/urls \
  issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_vault2/ca" \
  crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_vault2/crl" >/dev/null 2>&1 || true

vault write pki_corporate/config/urls \
  issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_corporate/ca" \
  crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_corporate/crl" >/dev/null 2>&1 || true

vault write pki_external/config/urls \
  issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_external/ca" \
  crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_external/crl" >/dev/null 2>&1 || true

vault write pki_partners/config/urls \
  issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_partners/ca" \
  crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_partners/crl" >/dev/null 2>&1 || true

# Create roles for issuing test certificates in all engines
vault write pki_vault2/roles/vcv \
  allowed_domains="vault2.local" \
  allow_bare_domains=true \
  allow_subdomains=true \
  max_ttl="8760h" \
  ttl="8760h" \
  not_before_duration="30s" >/dev/null 2>&1 || true

vault write pki_corporate/roles/vcv \
  allowed_domains="corp.local,corporate.local,enterprise.local" \
  allow_bare_domains=true \
  allow_subdomains=true \
  max_ttl="8760h" \
  ttl="8760h" \
  not_before_duration="30s" >/dev/null 2>&1 || true

vault write pki_external/roles/vcv \
  allowed_domains="external.local,public.local,api-gateway.local" \
  allow_bare_domains=true \
  allow_subdomains=true \
  max_ttl="8760h" \
  ttl="8760h" \
  not_before_duration="30s" >/dev/null 2>&1 || true

vault write pki_partners/roles/vcv \
  allowed_domains="partners.local,thirdparty.local,integration.local" \
  allow_bare_domains=true \
  allow_subdomains=true \
  max_ttl="8760h" \
  ttl="8760h" \
  not_before_duration="30s" >/dev/null 2>&1 || true

# Issue certificates for PKI engine (pki_vault2)
echo "Creating certificates for pki_vault2 engine..."

# Valid certificates
vault write pki_vault2/issue/vcv common_name="main.vault2.local" alt_names="www.main.vault2.local,api.main.vault2.local" >/dev/null 2>&1 || true
vault write pki_vault2/issue/vcv common_name="services.vault2.local" alt_names="svc1.services.vault2.local,svc2.services.vault2.local" >/dev/null 2>&1 || true
vault write pki_vault2/issue/vcv common_name="admin.vault2.local" alt_names="console.admin.vault2.local,manage.admin.vault2.local" >/dev/null 2>&1 || true
vault write pki_vault2/issue/vcv common_name="data.vault2.local" alt_names="primary.data.vault2.local,backup.data.vault2.local" >/dev/null 2>&1 || true

# Short-lived certificates (minutes)
vault write pki_vault2/issue/vcv common_name="short-1.vault2.local" ttl="120s" >/dev/null 2>&1 || true
vault write pki_vault2/issue/vcv common_name="short-2.vault2.local" ttl="180s" >/dev/null 2>&1 || true
vault write pki_vault2/issue/vcv common_name="short-3.vault2.local" ttl="300s" >/dev/null 2>&1 || true

# Long-lived certificates (years)
vault write pki_vault2/issue/vcv common_name="long-1.vault2.local" ttl="8760h" >/dev/null 2>&1 || true
vault write pki_vault2/issue/vcv common_name="long-2.vault2.local" ttl="8760h" >/dev/null 2>&1 || true

# Certificates expiring soon (24h-72h)
vault write pki_vault2/issue/vcv common_name="critical-expiring-1.vault2.local" ttl="48h" >/dev/null 2>&1 || true
vault write pki_vault2/issue/vcv common_name="critical-expiring-2.vault2.local" ttl="72h" >/dev/null 2>&1 || true
vault write pki_vault2/issue/vcv common_name="critical-expiring-3.vault2.local" ttl="24h" >/dev/null 2>&1 || true

# Certificates expiring in 7-30 days
vault write pki_vault2/issue/vcv common_name="warning-expiring-1.vault2.local" ttl="168h" >/dev/null 2>&1 || true
vault write pki_vault2/issue/vcv common_name="warning-expiring-2.vault2.local" ttl="240h" >/dev/null 2>&1 || true

# Expired certificates (TTL 2s, then wait)
printf "[vcv-2] Creating expired certificates (waiting 3s)...\n"
vault write pki_vault2/issue/vcv common_name="expired-1.vault2.local" ttl="2s" >/dev/null 2>&1 || true
vault write pki_vault2/issue/vcv common_name="expired-2.vault2.local" ttl="2s" >/dev/null 2>&1 || true
sleep 3

# Issue certificates for PKI CORPORATE engine
echo "Creating certificates for pki_corporate engine..."

vault write pki_corporate/issue/vcv common_name="intranet.corp.local" alt_names="portal.intranet.corp.local,hr.intranet.corp.local" >/dev/null 2>&1 || true
vault write pki_corporate/issue/vcv common_name="email.corp.local" alt_names="smtp.email.corp.local,imap.email.corp.local" >/dev/null 2>&1 || true
vault write pki_corporate/issue/vcv common_name="finance.corp.local" alt_names="erp.finance.corp.local,accounting.finance.corp.local" >/dev/null 2>&1 || true
vault write pki_corporate/issue/vcv common_name="hr.corp.local" alt_names="recruitment.hr.corp.local,payroll.hr.corp.local" >/dev/null 2>&1 || true

# Corporate certificates expiring soon
vault write pki_corporate/issue/vcv common_name="corp-expiring-1.local" ttl="36h" >/dev/null 2>&1 || true
vault write pki_corporate/issue/vcv common_name="corp-expiring-2.local" ttl="60h" >/dev/null 2>&1 || true

# Issue certificates for PKI EXTERNAL engine
echo "Creating certificates for pki_external engine..."

vault write pki_external/issue/vcv common_name="customer-api.external.local" alt_names="v1.customer-api.external.local,v2.customer-api.external.local" >/dev/null 2>&1 || true
vault write pki_external/issue/vcv common_name="public-portal.external.local" alt_names="www.public-portal.external.local,mobile.public-portal.external.local" >/dev/null 2>&1 || true
vault write pki_external/issue/vcv common_name="partner-gateway.external.local" alt_names="rest.partner-gateway.external.local,soap.partner-gateway.external.local" >/dev/null 2>&1 || true

# External certificates expiring soon
vault write pki_external/issue/vcv common_name="external-expiring-1.local" ttl="48h" >/dev/null 2>&1 || true
vault write pki_external/issue/vcv common_name="external-expiring-2.local" ttl="72h" >/dev/null 2>&1 || true

# External long-lived certificate
vault write pki_external/issue/vcv common_name="external-long-1.local" ttl="8760h" >/dev/null 2>&1 || true

# Issue certificates for PKI PARTNERS engine
echo "Creating certificates for pki_partners engine..."

vault write pki_partners/issue/vcv common_name="sso.partners.local" alt_names="auth.sso.partners.local,oauth.sso.partners.local" >/dev/null 2>&1 || true
vault write pki_partners/issue/vcv common_name="integration.partners.local" alt_names="api.integration.partners.local,webhook.integration.partners.local" >/dev/null 2>&1 || true
vault write pki_partners/issue/vcv common_name="thirdparty.partners.local" alt_names="vendor.thirdparty.partners.local,supplier.thirdparty.partners.local" >/dev/null 2>&1 || true

# Partners certificates expiring soon
vault write pki_partners/issue/vcv common_name="partners-expiring-1.local" ttl="24h" >/dev/null 2>&1 || true
vault write pki_partners/issue/vcv common_name="partners-expiring-2.local" ttl="96h" >/dev/null 2>&1 || true

# Create certificates to revoke in each engine
echo "Creating certificates to revoke..."

# Revoke from pki_vault2
REVOKE_OUTPUT=$(vault write -format=json pki_vault2/issue/vcv common_name="revoked.vault2.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv-2] Revoking certificate %s from pki_vault2\n" "${REVOKE_SERIAL}"
  vault write pki_vault2/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi

# Revoke from pki_corporate
REVOKE_OUTPUT=$(vault write -format=json pki_corporate/issue/vcv common_name="revoked.corporate.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv-2] Revoking certificate %s from pki_corporate\n" "${REVOKE_SERIAL}"
  vault write pki_corporate/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi

# Revoke from pki_external
REVOKE_OUTPUT=$(vault write -format=json pki_external/issue/vcv common_name="revoked.external.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv-2] Revoking certificate %s from pki_external\n" "${REVOKE_SERIAL}"
  vault write pki_external/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi

# Revoke from pki_partners
REVOKE_OUTPUT=$(vault write -format=json pki_partners/issue/vcv common_name="revoked.partners.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv-2] Revoking certificate %s from pki_partners\n" "${REVOKE_SERIAL}"
  vault write pki_partners/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi

# Force CRL rotation to include revoked certs
vault read pki_vault2/crl/rotate >/dev/null 2>&1 || true
vault read pki_corporate/crl/rotate >/dev/null 2>&1 || true
vault read pki_external/crl/rotate >/dev/null 2>&1 || true
vault read pki_partners/crl/rotate >/dev/null 2>&1 || true

printf "[vcv-2] Vault dev PKI initialized:\n"
printf "  - Mounts: pki_vault2/, pki_corporate/, pki_external/, pki_partners/\n"
printf "  - Roles: vcv (for all engines)\n"
printf "  - Certificates issued:\n"
printf "      pki_vault2:     vault2.* (valid/expiring/expired), revoked.vault2.local\n"
printf "      pki_corporate:  corp.* (valid/expiring), revoked.corporate.local\n"
printf "      pki_external:   external.* (valid/expiring), revoked.external.local\n"
printf "      pki_partners:   partners.* (valid/expiring), revoked.partners.local\n"

# Keep the Vault process in foreground
wait "${VAULT_PID}"
