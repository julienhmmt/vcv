#!/usr/bin/env sh
set -eu

# Start OpenBao dev server and run PKI initialization commands.
# This script is meant to be used as the container command in docker-compose.dev.yml.
# OpenBao is a community-driven fork of HashiCorp Vault

OPENBAO_ADDR_INTERNAL="http://127.0.0.1:1337"
OPENBAO_DEV_LISTEN="0.0.0.0:1337"
OPENBAO_ROOT_TOKEN="root"

export OPENBAO_ADDR="${OPENBAO_ADDR_INTERNAL}"
export OPENBAO_TOKEN="${OPENBAO_ROOT_TOKEN}"

# Start OpenBao dev in background
bao server \
  -dev \
  -dev-root-token-id="${OPENBAO_ROOT_TOKEN}" \
  -dev-listen-address="${OPENBAO_DEV_LISTEN}" &
OPENBAO_PID=$!

# Wait for OpenBao to be reachable
printf "[vcv-openbao] Waiting for OpenBao dev to be ready"
while true; do
  if wget --no-check-certificate --timeout=1 --tries=1 -q -O /dev/null "${OPENBAO_ADDR_INTERNAL}/v1/sys/health" 2>/dev/null; then
    break
  fi
  printf "."
  sleep 0.5
done
printf " done\n"
# Additional wait for OpenBao to be fully ready
sleep 3

# Enable and configure PKI using wget (BusyBox version)
# In dev mode the storage is in-memory, so these commands will run on each start.
# OpenBao PKI engines for testing VCV certificate management

# Enable PKI at path pki_openbao/
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"type":"pki"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/sys/mounts/pki_openbao" 2>/dev/null || true

# Enable additional PKI engines for OpenBao testing
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"type":"pki"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/sys/mounts/pki_openbao_dev" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"type":"pki"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/sys/mounts/pki_openbao_stage" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"type":"pki"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/sys/mounts/pki_openbao_prod" 2>/dev/null || true

# Tune max TTL for all OpenBao engines
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"max_lease_ttl":"8760h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/sys/mounts/pki_openbao/tune" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"max_lease_ttl":"8760h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/sys/mounts/pki_openbao_dev/tune" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"max_lease_ttl":"8760h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/sys/mounts/pki_openbao_stage/tune" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"max_lease_ttl":"8760h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/sys/mounts/pki_openbao_prod/tune" 2>/dev/null || true

# Generate root CAs for all OpenBao engines
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"openbao-vcv.local","ttl":"24h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao/root/generate/internal" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"openbao-vcv-dev.local","ttl":"24h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_dev/root/generate/internal" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"openbao-vcv-stage.local","ttl":"24h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_stage/root/generate/internal" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"openbao-vcv-prod.local","ttl":"24h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_prod/root/generate/internal" 2>/dev/null || true

# Configure CRL URLs for all OpenBao engines
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"issuing_certificates":"http://127.0.0.1:1337/v1/pki_openbao/ca","crl_distribution_points":"http://127.0.0.1:1337/v1/pki_openbao/crl"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao/config/urls" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"issuing_certificates":"http://127.0.0.1:1337/v1/pki_openbao_dev/ca","crl_distribution_points":"http://127.0.0.1:1337/v1/pki_openbao_dev/crl"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_dev/config/urls" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"issuing_certificates":"http://127.0.0.1:1337/v1/pki_openbao_stage/ca","crl_distribution_points":"http://127.0.0.1:1337/v1/pki_openbao_stage/crl"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_stage/config/urls" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"issuing_certificates":"http://127.0.0.1:1337/v1/pki_openbao_prod/ca","crl_distribution_points":"http://127.0.0.1:1337/v1/pki_openbao_prod/crl"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_prod/config/urls" 2>/dev/null || true

# Create roles for issuing test certificates in all OpenBao engines
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"allowed_domains":"openbao-production.local","allow_bare_domains":true,"allow_subdomains":true,"max_ttl":"24h","ttl":"24h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao/roles/openbao-vcv" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"allowed_domains":"openbao-dev.local,openbao-development.local","allow_bare_domains":true,"allow_subdomains":true,"max_ttl":"24h","ttl":"24h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_dev/roles/openbao-vcv" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"allowed_domains":"openbao-staging.local","allow_bare_domains":true,"allow_subdomains":true,"max_ttl":"24h","ttl":"24h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_stage/roles/openbao-vcv" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"allowed_domains":"openbao-production.local","allow_bare_domains":true,"allow_subdomains":true,"max_ttl":"24h","ttl":"24h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_prod/roles/openbao-vcv" 2>/dev/null || true

# Issue certificates for OpenBao PKI engines
echo "Creating certificates for OpenBao pki_openbao engine..."

# Valid certificates with different domains for OpenBao testing
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"web.openbao-production.local","alt_names":"www.web.openbao-production.local,api.web.openbao-production.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao/issue/openbao-vcv" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"api.openbao-production.local","alt_names":"api.openbao-production.local,admin.api.openbao-production.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao/issue/openbao-vcv" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"admin.openbao-production.local","alt_names":"admin.openbao-production.local,console.admin.openbao-production.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao/issue/openbao-vcv" 2>/dev/null || true

# Certificates expiring soon (24h-72h) for VCV testing
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"expiring-soon.openbao-production.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao/issue/openbao-vcv" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"critical-expiry.openbao-production.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao/issue/openbao-vcv" 2>/dev/null || true

# Certificates expiring in 7-30 days for VCV warning testing
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"warning-expiry.openbao-production.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao/issue/openbao-vcv" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"monitor-expiry.openbao-production.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao/issue/openbao-vcv" 2>/dev/null || true

# Expired certificates (TTL 2s, then wait)
printf "[vcv-openbao] Creating expired certificates (waiting 3s)...\n"
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"expired.openbao-production.local","ttl":"2s"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao/issue/openbao-vcv" 2>/dev/null || true
sleep 3

# Long-term certificate for testing
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"longterm.openbao-production.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao/issue/openbao-vcv" 2>/dev/null || true

# Issue certificates for OpenBao DEV engine
echo "Creating certificates for OpenBao pki_openbao_dev engine..."

# Valid certificates for dev environment
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"web.openbao-dev.local","alt_names":"www.web.openbao-dev.local,api.web.openbao-dev.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_dev/issue/openbao-vcv" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"microservice.openbao-dev.local","alt_names":"ms1.microservice.openbao-dev.local,ms2.microservice.openbao-dev.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_dev/issue/openbao-vcv" 2>/dev/null || true

# Dev certificate expiring soon for testing
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"dev-expiring.openbao-dev.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_dev/issue/openbao-vcv" 2>/dev/null || true

# Dev certificate with longer expiry
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"longterm.openbao-dev.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_dev/issue/openbao-vcv" 2>/dev/null || true

# Issue certificates for OpenBao STAGE engine  
echo "Creating certificates for OpenBao pki_openbao_stage engine..."

# Valid certificates for staging environment
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"web.openbao-staging.local","alt_names":"www.web.openbao-staging.local,api.web.openbao-staging.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_stage/issue/openbao-vcv" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"loadbalancer.openbao-staging.local","alt_names":"lb1.loadbalancer.openbao-staging.local,lb2.loadbalancer.openbao-staging.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_stage/issue/openbao-vcv" 2>/dev/null || true

# Staging certificate expiring soon for testing
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"stage-expiring.openbao-staging.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_stage/issue/openbao-vcv" 2>/dev/null || true

# Staging certificate with longer expiry
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"longterm.openbao-staging.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_stage/issue/openbao-vcv" 2>/dev/null || true

# Issue certificates for OpenBao PRODUCTION engine
echo "Creating certificates for OpenBao pki_openbao_prod engine..."

# Valid certificates for production environment
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"core.openbao-prod.local","alt_names":"core.api.openbao-prod.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_prod/issue/openbao-vcv" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"edge.openbao-prod.local","alt_names":"edge.lb.openbao-prod.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_prod/issue/openbao-vcv" 2>/dev/null || true

# Production certificate expiring soon for testing
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"prod-expiring.openbao-prod.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_prod/issue/openbao-vcv" 2>/dev/null || true

# Production certificate with longer expiry
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"longterm.openbao-prod.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_prod/issue/openbao-vcv" 2>/dev/null || true

# Create certificates to revoke in OpenBao engines for testing
echo "Creating certificates to revoke..."

# Revoke from pki_openbao
REVOKE_OUTPUT=$(wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"revoked.openbao.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao/issue/openbao-vcv" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv-openbao] Revoking certificate %s from pki_openbao\n" "${REVOKE_SERIAL}"
  wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"serial_number":"'"${REVOKE_SERIAL}"'"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao/revoke" 2>/dev/null || true
fi

# Revoke from pki_openbao_prod
REVOKE_OUTPUT=$(wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"common_name":"revoked-prod.openbao.local","ttl":"12h"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_prod/issue/openbao-vcv" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv-openbao] Revoking certificate %s from pki_openbao_prod\n" "${REVOKE_SERIAL}"
  wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{"serial_number":"'"${REVOKE_SERIAL}"'"}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_prod/revoke" 2>/dev/null || true
fi

# Force CRL rotation to include revoked certs
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao/crl/rotate" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_dev/crl/rotate" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_stage/crl/rotate" 2>/dev/null || true
wget --no-check-certificate --timeout=5 --tries=1 -q -O - --post-data='{}' --header='X-Vault-Token: root' --header='Content-Type: application/json' "http://127.0.0.1:1337/v1/pki_openbao_prod/crl/rotate" 2>/dev/null || true

printf "[vcv-openbao] OpenBao dev PKI initialized:\n"
printf "  - OpenBao Address: %s\n" "${OPENBAO_ADDR_INTERNAL}"
printf "  - Mounts: pki_openbao/, pki_openbao_dev/, pki_openbao_stage/, pki_openbao_prod/\n"
printf "  - Roles: openbao-vcv (for all engines)\n"
printf "  - Certificates issued:\n"
printf "      pki_openbao:         openbao-production.* (valid/expiring/expired), revoked.openbao.local\n"
printf "      pki_openbao_dev:     openbao-dev.* (valid/expiring), \n"
printf "      pki_openbao_stage:   openbao-staging.* (valid/expiring), \n"
printf "      pki_openbao_prod:    openbao-prod.* (valid/expiring), revoked-prod.openbao.local\n"
printf "  - Test scenarios: Critical expiry (24h), Warning expiry (7-30d), Expired, Revoked\n"

# Keep OpenBao running
wait ${OPENBAO_PID}
