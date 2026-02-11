#!/usr/bin/env sh
set -eu

# Parameterized Vault dev server initialization script
# Usage: vault-init-template.sh <config-file>
# Config file should be a JSON file with vault configuration

# Check if config file is provided
if [ $# -eq 0 ]; then
  echo "Error: Configuration file required"
  echo "Usage: $0 <config-file.json>"
  exit 1
fi

CONFIG_FILE="$1"

# Check if config file exists
if [ ! -f "${CONFIG_FILE}" ]; then
  echo "Error: Configuration file ${CONFIG_FILE} not found"
  exit 1
fi

# Read configuration from JSON file
# Using basic text processing since we may not have jq available
VAULT_ID=$(grep -o '"vault_id"[[:space:]]*:[[:space:]]*"[^"]*"' "${CONFIG_FILE}" | sed 's/.*"vault_id"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
LOG_PREFIX=$(grep -o '"log_prefix"[[:space:]]*:[[:space:]]*"[^"]*"' "${CONFIG_FILE}" | sed 's/.*"log_prefix"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')

# Default values if not found in config
VAULT_ID=${VAULT_ID:-"vcv"}
LOG_PREFIX=${LOG_PREFIX:-"[vcv]"}

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
printf "${LOG_PREFIX} Waiting for Vault dev to be ready"
while ! vault status >/dev/null 2>&1; do
  printf "."
  sleep 0.5
done
printf " done\n"

# Function to extract array from JSON
extract_array() {
  local key="$1"
  grep -o "\"${key}\"[[:space:]]*:[[:space:]]*\[[^]]*\]" "${CONFIG_FILE}" | sed 's/.*"\([^"]*\)"[[:space:]]*:[[:space:]]*\[\([^]]*\)\].*/\2/' | tr -d '"' | tr ',' ' '
}

# Function to extract object from JSON
extract_object() {
  local key="$1"
  local field="$2"
  grep -A 20 "\"${key}\"[[:space:]]*:[[:space:]]*{" "${CONFIG_FILE}" | grep -o "\"${field}\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" | sed 's/.*"\([^"]*\)".*/\1/' | head -1
}

# Enable PKI engines
echo "Enabling PKI engines..."
PKI_ENGINES=$(extract_array "pki_engines")
for engine in ${PKI_ENGINES}; do
  echo "  Enabling ${engine}"
  vault secrets enable -path="${engine}" pki 2>/dev/null || true
done

# Tune max TTL for all engines
echo "Tuning PKI engines..."
for engine in ${PKI_ENGINES}; do
  vault secrets tune -max-lease-ttl=8760h "${engine}" 2>/dev/null || true
done

# Generate root CAs
echo "Generating root CAs..."
for engine in ${PKI_ENGINES}; do
  common_name=$(extract_object "${engine}" "common_name")
  common_name=${common_name:-"${engine}.local"}
  echo "  Generating CA for ${engine} with CN=${common_name}"
  vault write -force "${engine}/root/generate/internal \
    common_name=\"${common_name}\" \
    ttl=\"8760h\"" >/dev/null 2>&1 || true
done

# Configure URLs
echo "Configuring CRL URLs..."
for engine in ${PKI_ENGINES}; do
  vault write "${engine}/config/urls \
    issuing_certificates=\"${VAULT_ADDR_INTERNAL}/v1/${engine}/ca\" \
    crl_distribution_points=\"${VAULT_ADDR_INTERNAL}/v1/${engine}/crl\"" >/dev/null 2>&1 || true
done

# Create roles
echo "Creating roles..."
for engine in ${PKI_ENGINES}; do
  allowed_domains=$(extract_object "${engine}" "allowed_domains")
  allowed_domains=${allowed_domains:-"local"}
  echo "  Creating role for ${engine} with domains=${allowed_domains}"
  vault write "${engine}/roles/vcv \
    allowed_domains=\"${allowed_domains}\" \
    allow_bare_domains=true \
    allow_subdomains=true \
    max_ttl=\"8760h\" \
    ttl=\"8760h\" \
    not_before_duration=\"30s\"" >/dev/null 2>&1 || true
done

# Issue certificates based on configuration
echo "Issuing certificates..."

# Function to issue certificates with pattern
issue_certificates() {
  local engine="$1"
  local pattern="$2"
  local count="$3"
  local ttl="$4"
  local prefix="$5"
  
  if [ -n "${pattern}" ] && [ "${count}" -gt 0 ]; then
    echo "  Issuing ${count} certificates for ${engine} with pattern=${pattern}"
    for i in $(seq 1 "${count}"); do
      if [ -n "${ttl}" ]; then
        vault write "${engine}/issue/vcv" common_name="${pattern}-${i}.${prefix}" ttl="${ttl}" >/dev/null 2>&1 || true
      else
        vault write "${engine}/issue/vcv" common_name="${pattern}-${i}.${prefix}" >/dev/null 2>&1 || true
      fi
    done
  fi
}

# Process certificate configurations
for engine in ${PKI_ENGINES}; do
  prefix=$(extract_object "${engine}" "prefix")
  prefix=${prefix:-"${engine}.local"}
  
  # Standard certificates
  pattern=$(extract_object "${engine}" "standard_pattern")
  count=$(extract_object "${engine}" "standard_count")
  count=${count:-0}
  issue_certificates "${engine}" "${pattern}" "${count}" "" "${prefix}"
  
  # Expiring soon certificates
  pattern=$(extract_object "${engine}" "expiring_pattern")
  count=$(extract_object "${engine}" "expiring_count")
  ttl=$(extract_object "${engine}" "expiring_ttl")
  count=${count:-0}
  issue_certificates "${engine}" "${pattern}" "${count}" "${ttl}" "${prefix}"
  
  # Expired certificates
  expired_count=$(extract_object "${engine}" "expired_count")
  expired_count=${expired_count:-0}
  if [ "${expired_count}" -gt 0 ]; then
    printf "${LOG_PREFIX} Creating expired certificates (waiting 3s)...\n"
    for i in $(seq 1 "${expired_count}"); do
      vault write "${engine}/issue/vcv" common_name="expired-${i}.${prefix}" ttl="2s" >/dev/null 2>&1 || true
    done
    sleep 3
  fi
done

# Revoke certificates if configured
echo "Processing revocations..."
REVOKE_ENGINES=$(extract_array "revoke_engines")
for engine in ${REVOKE_ENGINES}; do
  revoke_count=$(extract_object "${engine}" "revoke_count")
  revoke_count=${revoke_count:-1}
  prefix=$(extract_object "${engine}" "prefix")
  prefix=${prefix:-"${engine}.local"}
  
  for i in $(seq 1 "${revoke_count}"); do
    REVOKE_OUTPUT=$(vault write -format=json "${engine}/issue/vcv" common_name="revoked-${i}.${prefix}" ttl="720h" 2>/dev/null) || true
    REVOKE_SERIAL=$(echo "${REVOKE_OUTPUT}" | grep -o '"serial_number"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/.*"serial_number"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/' 2>/dev/null) || true
    if [ -n "${REVOKE_SERIAL}" ]; then
      printf "${LOG_PREFIX} Revoking certificate %s from %s\n" "${REVOKE_SERIAL}" "${engine}"
      vault write "${engine}/revoke" serial_number="${REVOKE_SERIAL}" >/dev/null 2>&1 || true
    fi
  done
done

# Force CRL rotation
echo "Rotating CRLs..."
for engine in ${PKI_ENGINES}; do
  vault read "${engine}/crl/rotate" >/dev/null 2>&1 || true
done

printf "${LOG_PREFIX} Vault dev PKI initialized:\n"
printf "  - Engines: ${PKI_ENGINES}\n"
printf "  - Roles: vcv (for all engines)\n"

# Keep the Vault process in foreground
wait "${VAULT_PID}"
