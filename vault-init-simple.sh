#!/usr/bin/env sh
set -eu

# Simple Vault init script that uses the original vault-dev-init.sh as base
# Usage: vault-init-simple.sh <vault-id>

VAULT_ID="$1"

# Copy the original script and modify it for this vault instance
cp /vault-dev-init.sh "/tmp/vault-init-${VAULT_ID}.sh"

# Simple replacements based on vault ID
case "${VAULT_ID}" in
  "vcv")
    # Keep original settings
    ;;
  "vcv-2")
    sed -i 's/pki_dev/pki_vault2/g' "/tmp/vault-init-${VAULT_ID}.sh"
    sed -i 's/pki_stage/pki_corporate/g' "/tmp/vault-init-${VAULT_ID}.sh"
    sed -i 's/pki_production/pki_external/g' "/tmp/vault-init-${VAULT_ID}.sh"
    sed -i 's|\[vcv\]|[vcv-2]|g' "/tmp/vault-init-${VAULT_ID}.sh"
    ;;
  "vcv-3")
    sed -i 's/pki_dev/pki_vault3/g' "/tmp/vault-init-${VAULT_ID}.sh"
    sed -i 's/pki_stage/pki_cloud/g' "/tmp/vault-init-${VAULT_ID}.sh"
    sed -i 's/pki_production/pki_edge/g' "/tmp/vault-init-${VAULT_ID}.sh"
    sed -i 's|\[vcv\]|[vcv-3]|g' "/tmp/vault-init-${VAULT_ID}.sh"
    ;;
  "vcv-4")
    sed -i 's/pki_dev/pki_vault4/g' "/tmp/vault-init-${VAULT_ID}.sh"
    sed -i 's/pki_stage/pki_lab/g' "/tmp/vault-init-${VAULT_ID}.sh"
    sed -i 's/pki_production/pki_qa/g' "/tmp/vault-init-${VAULT_ID}.sh"
    sed -i 's|\[vcv\]|[vcv-4]|g' "/tmp/vault-init-${VAULT_ID}.sh"
    ;;
  "vcv-5")
    sed -i 's/pki_dev/pki_vault5/g' "/tmp/vault-init-${VAULT_ID}.sh"
    sed -i 's/pki_stage/pki_internal/g' "/tmp/vault-init-${VAULT_ID}.sh"
    sed -i 's/pki_production/pki_dmz/g' "/tmp/vault-init-${VAULT_ID}.sh"
    sed -i 's|\[vcv\]|[vcv-5]|g' "/tmp/vault-init-${VAULT_ID}.sh"
    ;;
esac

# Execute the modified script
sh "/tmp/vault-init-${VAULT_ID}.sh"
