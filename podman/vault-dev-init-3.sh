#!/usr/bin/env sh
set -eu

# Start Vault dev server and run PKI initialization commands for vault-dev-3.
# This script is meant to be used as the container command in docker-compose.dev.yml.

VAULT_ADDR_INTERNAL="http://127.0.0.1:8203"
VAULT_DEV_LISTEN="0.0.0.0:8203"
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
printf "[vcv-3] Waiting for Vault dev to be ready"
while ! vault status >/dev/null 2>&1; do
  printf "."
  sleep 0.5
done
printf " done\n"

# Enable and configure PKI
# In dev mode the storage is in-memory, so these commands will run on each start.

# Enable PKI at path pki_vault3/ (idempotent: ignore error if already enabled)
vault secrets enable -path=pki_vault3 pki 2>/dev/null || true

# Enable additional PKI engines for vault-dev-3
vault secrets enable -path=pki_cloud pki 2>/dev/null || true
vault secrets enable -path=pki_edge pki 2>/dev/null || true
vault secrets enable -path=pki_iot pki 2>/dev/null || true
vault secrets enable -path=pki_blockchain pki 2>/dev/null || true

# Tune max TTL for all engines
vault secrets tune -max-lease-ttl=8760h pki_vault3 2>/dev/null || true
vault secrets tune -max-lease-ttl=8760h pki_cloud 2>/dev/null || true
vault secrets tune -max-lease-ttl=8760h pki_edge 2>/dev/null || true
vault secrets tune -max-lease-ttl=8760h pki_iot 2>/dev/null || true
vault secrets tune -max-lease-ttl=8760h pki_blockchain 2>/dev/null || true

# Generate root CAs for all engines (force to avoid interactive prompts)
vault write -force pki_vault3/root/generate/internal \
  common_name="vcv-vault3.local" \
  ttl="8760h" >/dev/null 2>&1 || true

vault write -force pki_cloud/root/generate/internal \
  common_name="vcv-cloud.local" \
  ttl="8760h" >/dev/null 2>&1 || true

vault write -force pki_edge/root/generate/internal \
  common_name="vcv-edge.local" \
  ttl="8760h" >/dev/null 2>&1 || true

vault write -force pki_iot/root/generate/internal \
  common_name="vcv-iot.local" \
  ttl="8760h" >/dev/null 2>&1 || true

vault write -force pki_blockchain/root/generate/internal \
  common_name="vcv-blockchain.local" \
  ttl="8760h" >/dev/null 2>&1 || true

# Configure CRL URLs for all engines
vault write pki_vault3/config/urls \
  issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_vault3/ca" \
  crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_vault3/crl" >/dev/null 2>&1 || true

vault write pki_cloud/config/urls \
  issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_cloud/ca" \
  crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_cloud/crl" >/dev/null 2>&1 || true

vault write pki_edge/config/urls \
  issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_edge/ca" \
  crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_edge/crl" >/dev/null 2>&1 || true

vault write pki_iot/config/urls \
  issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_iot/ca" \
  crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_iot/crl" >/dev/null 2>&1 || true

vault write pki_blockchain/config/urls \
  issuing_certificates="${VAULT_ADDR_INTERNAL}/v1/pki_blockchain/ca" \
  crl_distribution_points="${VAULT_ADDR_INTERNAL}/v1/pki_blockchain/crl" >/dev/null 2>&1 || true

# Create roles for issuing test certificates in all engines
vault write pki_vault3/roles/vcv \
  allowed_domains="vault3.local" \
  allow_bare_domains=true \
  allow_subdomains=true \
  max_ttl="8760h" \
  ttl="8760h" \
  not_before_duration="30s" >/dev/null 2>&1 || true

vault write pki_cloud/roles/vcv \
  allowed_domains="cloud.local,k8s.local,container.local,microservice.local" \
  allow_bare_domains=true \
  allow_subdomains=true \
  max_ttl="8760h" \
  ttl="8760h" \
  not_before_duration="30s" >/dev/null 2>&1 || true

vault write pki_edge/roles/vcv \
  allowed_domains="edge.local,cdn.local,gateway.local,proxy.local" \
  allow_bare_domains=true \
  allow_subdomains=true \
  max_ttl="8760h" \
  ttl="8760h" \
  not_before_duration="30s" >/dev/null 2>&1 || true

vault write pki_iot/roles/vcv \
  allowed_domains="iot.local,device.local,sensor.local,embedded.local" \
  allow_bare_domains=true \
  allow_subdomains=true \
  max_ttl="8760h" \
  ttl="8760h" \
  not_before_duration="30s" >/dev/null 2>&1 || true

vault write pki_blockchain/roles/vcv \
  allowed_domains="blockchain.local,web3.local,dapp.local,nft.local" \
  allow_bare_domains=true \
  allow_subdomains=true \
  max_ttl="8760h" \
  ttl="8760h" \
  not_before_duration="30s" >/dev/null 2>&1 || true

# Issue certificates for PKI engine (pki_vault3)
echo "Creating certificates for pki_vault3 engine..."

# Valid certificates
vault write pki_vault3/issue/vcv common_name="core.vault3.local" alt_names="www.core.vault3.local,api.core.vault3.local" >/dev/null 2>&1 || true
vault write pki_vault3/issue/vcv common_name="cluster.vault3.local" alt_names="node1.cluster.vault3.local,node2.cluster.vault3.local" >/dev/null 2>&1 || true
vault write pki_vault3/issue/vcv common_name="monitor.vault3.local" alt_names="prom.monitor.vault3.local,grafana.monitor.vault3.local" >/dev/null 2>&1 || true
vault write pki_vault3/issue/vcv common_name="backup.vault3.local" alt_names="primary.backup.vault3.local,secondary.backup.vault3.local" >/dev/null 2>&1 || true

# Short-lived certificates (minutes)
vault write pki_vault3/issue/vcv common_name="short-1.vault3.local" ttl="120s" >/dev/null 2>&1 || true
vault write pki_vault3/issue/vcv common_name="short-2.vault3.local" ttl="180s" >/dev/null 2>&1 || true
vault write pki_vault3/issue/vcv common_name="short-3.vault3.local" ttl="300s" >/dev/null 2>&1 || true

# Long-lived certificates (years)
vault write pki_vault3/issue/vcv common_name="long-1.vault3.local" ttl="8760h" >/dev/null 2>&1 || true
vault write pki_vault3/issue/vcv common_name="long-2.vault3.local" ttl="8760h" >/dev/null 2>&1 || true

# Certificates expiring soon (24h-72h)
vault write pki_vault3/issue/vcv common_name="urgent-expiring-1.vault3.local" ttl="48h" >/dev/null 2>&1 || true
vault write pki_vault3/issue/vcv common_name="urgent-expiring-2.vault3.local" ttl="72h" >/dev/null 2>&1 || true
vault write pki_vault3/issue/vcv common_name="urgent-expiring-3.vault3.local" ttl="24h" >/dev/null 2>&1 || true

# Certificates expiring in 7-30 days
vault write pki_vault3/issue/vcv common_name="scheduled-expiring-1.vault3.local" ttl="168h" >/dev/null 2>&1 || true
vault write pki_vault3/issue/vcv common_name="scheduled-expiring-2.vault3.local" ttl="240h" >/dev/null 2>&1 || true

# Expired certificates (TTL 2s, then wait)
printf "[vcv-3] Creating expired certificates (waiting 3s)...\n"
vault write pki_vault3/issue/vcv common_name="expired-1.vault3.local" ttl="2s" >/dev/null 2>&1 || true
vault write pki_vault3/issue/vcv common_name="expired-2.vault3.local" ttl="2s" >/dev/null 2>&1 || true
sleep 3

# Issue certificates for PKI CLOUD engine
echo "Creating certificates for pki_cloud engine..."

vault write pki_cloud/issue/vcv common_name="kubernetes.cloud.local" alt_names="api.kubernetes.cloud.local,etcd.kubernetes.cloud.local" >/dev/null 2>&1 || true
vault write pki_cloud/issue/vcv common_name="docker.cloud.local" alt_names="registry.docker.cloud.local,swarm.docker.cloud.local" >/dev/null 2>&1 || true
vault write pki_cloud/issue/vcv common_name="microservice.cloud.local" alt_names="svc1.microservice.cloud.local,svc2.microservice.cloud.local" >/dev/null 2>&1 || true
vault write pki_cloud/issue/vcv common_name="container.cloud.local" alt_names="app1.container.cloud.local,app2.container.cloud.local" >/dev/null 2>&1 || true

# Cloud certificates expiring soon
vault write pki_cloud/issue/vcv common_name="cloud-expiring-1.local" ttl="36h" >/dev/null 2>&1 || true
vault write pki_cloud/issue/vcv common_name="cloud-expiring-2.local" ttl="60h" >/dev/null 2>&1 || true

# Cloud long-lived certificate
vault write pki_cloud/issue/vcv common_name="cloud-long-1.local" ttl="8760h" >/dev/null 2>&1 || true

# Issue certificates for PKI EDGE engine
echo "Creating certificates for pki_edge engine..."

vault write pki_edge/issue/vcv common_name="cdn.edge.local" alt_names="pop1.cdn.edge.local,pop2.cdn.edge.local" >/dev/null 2>&1 || true
vault write pki_edge/issue/vcv common_name="gateway.edge.local" alt_names="north.gateway.edge.local,south.gateway.edge.local" >/dev/null 2>&1 || true
vault write pki_edge/issue/vcv common_name="proxy.edge.local" alt_names="reverse.proxy.edge.local,forward.proxy.edge.local" >/dev/null 2>&1 || true
vault write pki_edge/issue/vcv common_name="loadbalancer.edge.local" alt_names="lb1.loadbalancer.edge.local,lb2.loadbalancer.edge.local" >/dev/null 2>&1 || true

# Edge certificates expiring soon
vault write pki_edge/issue/vcv common_name="edge-expiring-1.local" ttl="48h" >/dev/null 2>&1 || true
vault write pki_edge/issue/vcv common_name="edge-expiring-2.local" ttl="96h" >/dev/null 2>&1 || true

# Issue certificates for PKI IOT engine
echo "Creating certificates for pki_iot engine..."

vault write pki_iot/issue/vcv common_name="sensor.iot.local" alt_names="temp.sensor.iot.local,humidity.sensor.iot.local" >/dev/null 2>&1 || true
vault write pki_iot/issue/vcv common_name="device.iot.local" alt_names="camera.device.iot.local,lock.device.iot.local" >/dev/null 2>&1 || true
vault write pki_iot/issue/vcv common_name="embedded.iot.local" alt_names="controller.embedded.iot.local,actuator.embedded.iot.local" >/dev/null 2>&1 || true
vault write pki_iot/issue/vcv common_name="gateway.iot.local" alt_names="mqtt.gateway.iot.local,coap.gateway.iot.local" >/dev/null 2>&1 || true

# IoT certificates expiring soon
vault write pki_iot/issue/vcv common_name="iot-expiring-1.local" ttl="24h" >/dev/null 2>&1 || true
vault write pki_iot/issue/vcv common_name="iot-expiring-2.local" ttl="72h" >/dev/null 2>&1 || true

# Issue certificates for PKI BLOCKCHAIN engine
echo "Creating certificates for pki_blockchain engine..."

vault write pki_blockchain/issue/vcv common_name="web3.blockchain.local" alt_names="eth.web3.blockchain.local,polygon.web3.blockchain.local" >/dev/null 2>&1 || true
vault write pki_blockchain/issue/vcv common_name="dapp.blockchain.local" alt_names="defi.dapp.blockchain.local,nft.dapp.blockchain.local" >/dev/null 2>&1 || true
vault write pki_blockchain/issue/vcv common_name="node.blockchain.local" alt_names="validator.node.blockchain.local,miner.node.blockchain.local" >/dev/null 2>&1 || true
vault write pki_blockchain/issue/vcv common_name="wallet.blockchain.local" alt_names="hot.wallet.blockchain.local,cold.wallet.blockchain.local" >/dev/null 2>&1 || true

# Blockchain certificates expiring soon
vault write pki_blockchain/issue/vcv common_name="blockchain-expiring-1.local" ttl="36h" >/dev/null 2>&1 || true
vault write pki_blockchain/issue/vcv common_name="blockchain-expiring-2.local" ttl="84h" >/dev/null 2>&1 || true

# Create certificates to revoke in each engine
echo "Creating certificates to revoke..."

# Revoke from pki_vault3
REVOKE_OUTPUT=$(vault write -format=json pki_vault3/issue/vcv common_name="revoked.vault3.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv-3] Revoking certificate %s from pki_vault3\n" "${REVOKE_SERIAL}"
  vault write pki_vault3/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi

# Revoke from pki_cloud
REVOKE_OUTPUT=$(vault write -format=json pki_cloud/issue/vcv common_name="revoked.cloud.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv-3] Revoking certificate %s from pki_cloud\n" "${REVOKE_SERIAL}"
  vault write pki_cloud/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi

# Revoke from pki_edge
REVOKE_OUTPUT=$(vault write -format=json pki_edge/issue/vcv common_name="revoked.edge.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv-3] Revoking certificate %s from pki_edge\n" "${REVOKE_SERIAL}"
  vault write pki_edge/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi

# Revoke from pki_iot
REVOKE_OUTPUT=$(vault write -format=json pki_iot/issue/vcv common_name="revoked.iot.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv-3] Revoking certificate %s from pki_iot\n" "${REVOKE_SERIAL}"
  vault write pki_iot/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi

# Revoke from pki_blockchain
REVOKE_OUTPUT=$(vault write -format=json pki_blockchain/issue/vcv common_name="revoked.blockchain.local" ttl="720h" 2>/dev/null) || true
REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
if [ -n "${REVOKE_SERIAL}" ]; then
  printf "[vcv-3] Revoking certificate %s from pki_blockchain\n" "${REVOKE_SERIAL}"
  vault write pki_blockchain/revoke serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
fi

# Force CRL rotation to include revoked certs
vault read pki_vault3/crl/rotate >/dev/null 2>&1 || true
vault read pki_cloud/crl/rotate >/dev/null 2>&1 || true
vault read pki_edge/crl/rotate >/dev/null 2>&1 || true
vault read pki_iot/crl/rotate >/dev/null 2>&1 || true
vault read pki_blockchain/crl/rotate >/dev/null 2>&1 || true

printf "[vcv-3] Vault dev PKI initialized:\n"
printf "  - Mounts: pki_vault3/, pki_cloud/, pki_edge/, pki_iot/, pki_blockchain/\n"
printf "  - Roles: vcv (for all engines)\n"
printf "  - Certificates issued:\n"
printf "      pki_vault3:     vault3.* (valid/expiring/expired), revoked.vault3.local\n"
printf "      pki_cloud:      cloud.* (valid/expiring), revoked.cloud.local\n"
printf "      pki_edge:       edge.* (valid/expiring), revoked.edge.local\n"
printf "      pki_iot:        iot.* (valid/expiring), revoked.iot.local\n"
printf "      pki_blockchain: blockchain.* (valid/expiring), revoked.blockchain.local\n"

# Keep the Vault process in foreground
wait "${VAULT_PID}"
