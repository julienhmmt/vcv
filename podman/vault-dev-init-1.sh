#!/usr/bin/env sh
set -eu

# Start Vault dev server and run PKI initialization commands.
# This script is meant to be used as the container command in docker-compose.dev.yml.

VAULT_ADDR_INTERNAL="http://127.0.0.1:8201"
VAULT_DEV_LISTEN="0.0.0.0:8201"
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

# Enable additional PKI engines
vault secrets enable -path=pki_dev pki 2>/dev/null || true
vault secrets enable -path=pki_stage pki 2>/dev/null || true
vault secrets enable -path=pki_production pki 2>/dev/null || true

# Tune max TTL for all engines
vault secrets tune -max-lease-ttl=8760h pki 2>/dev/null || true
vault secrets tune -max-lease-ttl=8760h pki_dev 2>/dev/null || true
vault secrets tune -max-lease-ttl=8760h pki_stage 2>/dev/null || true
vault secrets tune -max-lease-ttl=8760h pki_production 2>/dev/null || true

# Generate root CAs for all engines (force to avoid interactive prompts)
vault write -force pki/root/generate/internal \
  common_name="vcv.local" \
  ttl="8760h" >/dev/null 2>&1 || true

vault write -force pki_dev/root/generate/internal \
  common_name="vcv-dev.local" \
  ttl="8760h" >/dev/null 2>&1 || true

vault write -force pki_stage/root/generate/internal \
  common_name="vcv-stage.local" \
  ttl="8760h" >/dev/null 2>&1 || true

vault write -force pki_production/root/generate/internal \
  common_name="vcv-production.local" \
  ttl="8760h" >/dev/null 2>&1 || true

# Configure CRL URLs for all engines
vault write pki/config/urls \
  issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki/ca" \
  crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki/crl" >/dev/null 2>&1 || true

vault write pki_dev/config/urls \
  issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_dev/ca" \
  crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_dev/crl" >/dev/null 2>&1 || true

vault write pki_stage/config/urls \
  issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_stage/ca" \
  crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_stage/crl" >/dev/null 2>&1 || true

vault write pki_production/config/urls \
  issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_production/ca" \
  crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_production/crl" >/dev/null 2>&1 || true

# Create roles for issuing test certificates in all engines
vault write pki/roles/vcv \
  allowed_domains="production.local" \
  allow_bare_domains=true \
  allow_subdomains=true \
  max_ttl="8760h" \
  ttl="8760h" \
  not_before_duration="30s" >/dev/null 2>&1 || true

vault write pki_dev/roles/vcv \
  allowed_domains="dev.local,development.local" \
  allow_bare_domains=true \
  allow_subdomains=true \
  max_ttl="8760h" \
  ttl="8760h" \
  not_before_duration="30s" >/dev/null 2>&1 || true

vault write pki_stage/roles/vcv \
  allowed_domains="staging.local" \
  allow_bare_domains=true \
  allow_subdomains=true \
  max_ttl="8760h" \
  ttl="8760h" \
  not_before_duration="30s" >/dev/null 2>&1 || true

vault write pki_production/roles/vcv \
  allowed_domains="production.local" \
  allow_bare_domains=true \
  allow_subdomains=true \
  max_ttl="8760h" \
  ttl="8760h" \
  not_before_duration="30s" >/dev/null 2>&1 || true

# Issue certificates for PKI engine (pki)
echo "Creating certificates for pki engine..."

# Valid certificates
vault write pki/issue/vcv common_name="web.production.local" alt_names="www.web.production.local,api.web.production.local" >/dev/null 2>&1 || true
vault write pki/issue/vcv common_name="api.production.local" alt_names="api.production.local,admin.api.production.local" >/dev/null 2>&1 || true
vault write pki/issue/vcv common_name="admin.production.local" alt_names="admin.production.local,console.admin.production.local" >/dev/null 2>&1 || true
vault write pki/issue/vcv common_name="database.production.local" alt_names="db.database.production.local,replica.database.production.local" >/dev/null 2>&1 || true
vault write pki/issue/vcv common_name="cache.production.local" alt_names="redis.cache.production.local,memcache.cache.production.local" >/dev/null 2>&1 || true

# Certificates expiring soon (24h-72h)
vault write pki/issue/vcv common_name="expiring-soon-1.production.local" ttl="48h" >/dev/null 2>&1 || true
vault write pki/issue/vcv common_name="expiring-soon-2.production.local" ttl="72h" >/dev/null 2>&1 || true
vault write pki/issue/vcv common_name="expiring-soon-3.production.local" ttl="24h" >/dev/null 2>&1 || true

# Certificates expiring in 7-30 days
vault write pki/issue/vcv common_name="expiring-week-1.production.local" ttl="168h" >/dev/null 2>&1 || true
vault write pki/issue/vcv common_name="expiring-week-2.production.local" ttl="240h" >/dev/null 2>&1 || true
vault write pki/issue/vcv common_name="expiring-month-1.production.local" ttl="720h" >/dev/null 2>&1 || true
vault write pki/issue/vcv common_name="expiring-month-2.production.local" ttl="500h" >/dev/null 2>&1 || true

# Expired certificates (TTL 2s, then wait)
printf "[vcv] Creating expired certificates (waiting 3s)...\n"
vault write pki/issue/vcv common_name="expired-1.production.local" ttl="2s" >/dev/null 2>&1 || true
vault write pki/issue/vcv common_name="expired-2.production.local" ttl="2s" >/dev/null 2>&1 || true
sleep 3

# Diverse TTL certificates for production mount (pki)
vault write pki/issue/vcv common_name="no-expiry.production.local" ttl="8760h" >/dev/null 2>&1 || true
vault write pki/issue/vcv common_name="expire-2m.production.local" ttl="120s" >/dev/null 2>&1 || true
vault write pki/issue/vcv common_name="expire-1d.production.local" ttl="24h" >/dev/null 2>&1 || true
vault write pki/issue/vcv common_name="expire-4d.production.local" ttl="96h" >/dev/null 2>&1 || true
vault write pki/issue/vcv common_name="expire-20d.production.local" ttl="480h" >/dev/null 2>&1 || true
vault write pki/issue/vcv common_name="expire-31d.production.local" ttl="744h" >/dev/null 2>&1 || true

# Issue certificates for PKI DEV engine
echo "Creating certificates for pki_dev engine..."

# Valid certificates for dev
vault write pki_dev/issue/vcv common_name="web.development.local" alt_names="www.web.development.local,api.web.development.local" >/dev/null 2>&1 || true
vault write pki_dev/issue/vcv common_name="api.development.local" alt_names="api.development.local,admin.api.development.local" >/dev/null 2>&1 || true
vault write pki_dev/issue/vcv common_name="microservice-1.dev.local" alt_names="ms1.microservice-1.dev.local,ms1-backup.microservice-1.dev.local" >/dev/null 2>&1 || true
vault write pki_dev/issue/vcv common_name="microservice-2.dev.local" alt_names="ms2.microservice-2.dev.local,ms2-backup.microservice-2.dev.local" >/dev/null 2>&1 || true
vault write pki_dev/issue/vcv common_name="testing.development.local" alt_names="test.testing.development.local,staging.testing.development.local" >/dev/null 2>&1 || true

# Dev certificates expiring soon
vault write pki_dev/issue/vcv common_name="dev-expiring-soon-1.local" ttl="36h" >/dev/null 2>&1 || true
vault write pki_dev/issue/vcv common_name="dev-expiring-soon-2.local" ttl="60h" >/dev/null 2>&1 || true

# Dev certificates expiring in weeks
vault write pki_dev/issue/vcv common_name="dev-expiring-week-1.local" ttl="200h" >/dev/null 2>&1 || true
vault write pki_dev/issue/vcv common_name="dev-expiring-week-2.local" ttl="300h" >/dev/null 2>&1 || true

# Diverse TTL certificates for dev mount
vault write pki_dev/issue/vcv common_name="no-expiry.development.local" ttl="8760h" >/dev/null 2>&1 || true
vault write pki_dev/issue/vcv common_name="expire-2m.development.local" ttl="120s" >/dev/null 2>&1 || true
vault write pki_dev/issue/vcv common_name="expire-1d.development.local" ttl="24h" >/dev/null 2>&1 || true
vault write pki_dev/issue/vcv common_name="expire-4d.development.local" ttl="96h" >/dev/null 2>&1 || true
vault write pki_dev/issue/vcv common_name="expire-20d.development.local" ttl="480h" >/dev/null 2>&1 || true
vault write pki_dev/issue/vcv common_name="expire-31d.development.local" ttl="744h" >/dev/null 2>&1 || true

# Issue certificates for PKI STAGE engine
echo "Creating certificates for pki_stage engine..."

# Valid certificates for staging
vault write pki_stage/issue/vcv common_name="web.staging.local" alt_names="www.web.staging.local,api.web.staging.local" >/dev/null 2>&1 || true
vault write pki_stage/issue/vcv common_name="api.staging.local" alt_names="api.staging.local,admin.api.staging.local" >/dev/null 2>&1 || true
vault write pki_stage/issue/vcv common_name="loadbalancer.staging.local" alt_names="lb1.loadbalancer.staging.local,lb2.loadbalancer.staging.local" >/dev/null 2>&1 || true
vault write pki_stage/issue/vcv common_name="monitoring.staging.local" alt_names="prom.monitoring.staging.local,grafana.monitoring.staging.local" >/dev/null 2>&1 || true
vault write pki_stage/issue/vcv common_name="ci-cd.staging.local" alt_names="jenkins.ci-cd.staging.local,gitlab.ci-cd.staging.local" >/dev/null 2>&1 || true

# Staging certificates expiring soon
vault write pki_stage/issue/vcv common_name="stage-expiring-soon-1.local" ttl="24h" >/dev/null 2>&1 || true
vault write pki_stage/issue/vcv common_name="stage-expiring-soon-2.local" ttl="48h" >/dev/null 2>&1 || true

# Staging certificates expiring in weeks
vault write pki_stage/issue/vcv common_name="stage-expiring-week-1.local" ttl="180h" >/dev/null 2>&1 || true
vault write pki_stage/issue/vcv common_name="stage-expiring-week-2.local" ttl="250h" >/dev/null 2>&1 || true

# Diverse TTL certificates for stage mount
vault write pki_stage/issue/vcv common_name="no-expiry.staging.local" ttl="8760h" >/dev/null 2>&1 || true
vault write pki_stage/issue/vcv common_name="expire-2m.staging.local" ttl="120s" >/dev/null 2>&1 || true
vault write pki_stage/issue/vcv common_name="expire-1d.staging.local" ttl="24h" >/dev/null 2>&1 || true
vault write pki_stage/issue/vcv common_name="expire-4d.staging.local" ttl="96h" >/dev/null 2>&1 || true
vault write pki_stage/issue/vcv common_name="expire-20d.staging.local" ttl="480h" >/dev/null 2>&1 || true
vault write pki_stage/issue/vcv common_name="expire-31d.staging.local" ttl="744h" >/dev/null 2>&1 || true

# Issue certificates for PKI PRODUCTION engine (sample set)
echo "Creating certificates for pki_production engine..."

vault write pki_production/issue/vcv common_name="core.production.local" alt_names="core.api.production.local" >/dev/null 2>&1 || true
vault write pki_production/issue/vcv common_name="edge.production.local" alt_names="edge.lb.production.local" >/dev/null 2>&1 || true
vault write pki_production/issue/vcv common_name="analytics.production.local" alt_names="clickhouse.analytics.production.local" >/dev/null 2>&1 || true
vault write pki_production/issue/vcv common_name="observability.production.local" alt_names="logs.observability.production.local,metrics.observability.production.local" >/dev/null 2>&1 || true
vault write pki_production/issue/vcv common_name="backup.production.local" alt_names="backup1.production.local,backup2.production.local" >/dev/null 2>&1 || true

# Production certificates expiring soon
vault write pki_production/issue/vcv common_name="prod-expiring-soon-1.local" ttl="36h" >/dev/null 2>&1 || true
vault write pki_production/issue/vcv common_name="prod-expiring-soon-2.local" ttl="60h" >/dev/null 2>&1 || true

# Diverse TTL certificates for production mount (pki_production)
vault write pki_production/issue/vcv common_name="no-expiry.production.alt" ttl="8760h" >/dev/null 2>&1 || true
vault write pki_production/issue/vcv common_name="expire-2m.production.alt" ttl="120s" >/dev/null 2>&1 || true
vault write pki_production/issue/vcv common_name="expire-1d.production.alt" ttl="24h" >/dev/null 2>&1 || true
vault write pki_production/issue/vcv common_name="expire-4d.production.alt" ttl="96h" >/dev/null 2>&1 || true
vault write pki_production/issue/vcv common_name="expire-20d.production.alt" ttl="480h" >/dev/null 2>&1 || true
vault write pki_production/issue/vcv common_name="expire-31d.production.alt" ttl="744h" >/dev/null 2>&1 || true

# Production certificate to revoke
REVOKE_OUTPUT=$(vault write -format=json pki_production/issue/vcv common_name="revoked.prod.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv] Revoking certificate %s from pki_production\n" "${REVOKE_SERIAL}"
  vault write pki_production/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi

# Create certificates to revoke in each engine
echo "Creating certificates to revoke..."

# Revoke from pki
REVOKE_OUTPUT=$(vault write -format=json pki/issue/vcv common_name="revoked.pki.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv] Revoking certificate %s from pki\n" "${REVOKE_SERIAL}"
  vault write pki/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi
REVOKE_OUTPUT=$(vault write -format=json pki/issue/vcv common_name="revoked2.pki.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv] Revoking certificate %s from pki\n" "${REVOKE_SERIAL}"
  vault write pki/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi

# Revoke from pki_dev
REVOKE_OUTPUT=$(vault write -format=json pki_dev/issue/vcv common_name="revoked.dev.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv] Revoking certificate %s from pki_dev\n" "${REVOKE_SERIAL}"
  vault write pki_dev/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi
REVOKE_OUTPUT=$(vault write -format=json pki_dev/issue/vcv common_name="revoked2.dev.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv] Revoking certificate %s from pki_dev\n" "${REVOKE_SERIAL}"
  vault write pki_dev/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi

# Revoke from pki_stage
REVOKE_OUTPUT=$(vault write -format=json pki_stage/issue/vcv common_name="revoked.stage.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv] Revoking certificate %s from pki_stage\n" "${REVOKE_SERIAL}"
  vault write pki_stage/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi
REVOKE_OUTPUT=$(vault write -format=json pki_stage/issue/vcv common_name="revoked2.stage.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv] Revoking certificate %s from pki_stage\n" "${REVOKE_SERIAL}"
  vault write pki_stage/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi

# Revoke from pki_production
REVOKE_OUTPUT=$(vault write -format=json pki_production/issue/vcv common_name="revoked.prod.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv] Revoking certificate %s from pki_production\n" "${REVOKE_SERIAL}"
  vault write pki_production/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi
REVOKE_OUTPUT=$(vault write -format=json pki_production/issue/vcv common_name="revoked2.prod.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv] Revoking certificate %s from pki_production\n" "${REVOKE_SERIAL}"
  vault write pki_production/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi

# Force CRL rotation to include revoked certs
vault read pki/crl/rotate >/dev/null 2>&1 || true
vault read pki_dev/crl/rotate >/dev/null 2>&1 || true
vault read pki_stage/crl/rotate >/dev/null 2>&1 || true
vault read pki_production/crl/rotate >/dev/null 2>&1 || true

printf "[vcv] Vault dev PKI initialized:\n"
printf "  - Mounts: pki/, pki_dev/, pki_stage/, pki_production/\n"
printf "  - Roles: vcv (for all engines)\n"
printf "  - Certificates issued:\n"
printf "      pki:             production.* (valid/expiring/expired), revoked.pki.local\n"
printf "      pki_dev:         development.* (valid/expiring), revoked.dev.local\n"
printf "      pki_stage:       staging.* (valid/expiring), revoked.stage.local\n"
printf "      pki_production:  production.* (valid/expiring), revoked.prod.local\n"

# Keep the Vault process in foreground
wait "${VAULT_PID}"
