#!/bin/sh
# Wait until the e2e PKI seeds are fully visible through the app.
# A mid-seed listing would otherwise stick in the 15-minute Vault cache, so
# each poll busts the cache via the admin API first (ending warm for tests).
# Usage: tools/e2e-wait.sh [app_url]
set -eu

APP_URL="${1:-http://localhost:52001}"
ADMIN_PASSWORD="${VCV_E2E_ADMIN_PASSWORD:-e2e-admin-password}"
TIMEOUT_SECS="${VCV_E2E_WAIT_SECS:-600}"

# Deterministic 8760h seeds that used to flake (plus one OpenBao marker).
# The web.* CN of each vault's LAST configured mount is included so "seeded"
# only passes once trailing mounts finished too — init seeds mounts in order.
MARKERS="web.vault-main-pki.local machine-web.vault-main-pki.local user-alice.vault-main-pki.local expiring-24h.vault-main-pki.local expired-1.vault-main-pki.local revoked-1.vault-main-pki.local web.vault-main-pki_production.local web.openbao-dev-1-pki_openbao.local web.openbao-dev-1-pki_openbao_dev.local"

COOKIE_JAR="$(mktemp)"
trap 'rm -f "$COOKIE_JAR"' EXIT INT TERM

deadline=$(($(date +%s) + TIMEOUT_SECS))

printf "waiting for the app at %s" "$APP_URL"
while ! curl -fsS "$APP_URL/api/ready" >/dev/null 2>&1; do
  if [ "$(date +%s)" -ge "$deadline" ]; then
    printf "\napp did not become ready in time\n"
    exit 1
  fi
  printf "."
  sleep 5
done
printf " ready\n"

curl -fsS -c "$COOKIE_JAR" -X POST "$APP_URL/api/admin/login" \
  -H "Content-Type: application/json" \
  -H "Origin: $APP_URL" \
  -d "{\"username\":\"admin\",\"password\":\"$ADMIN_PASSWORD\"}" >/dev/null

printf "waiting for PKI seeds"
while :; do
  # Bust the Vault-list cache so polls always read fresh mounts.
  # (Origin header: cookie-bearing POSTs require a same-origin value.)
  curl -fsS -b "$COOKIE_JAR" -X POST "$APP_URL/api/cache/invalidate" -H "Origin: $APP_URL" >/dev/null 2>&1 || true
  body="$(curl -fsS "$APP_URL/api/certs?page_size=all" 2>/dev/null)" || body=""
  missing=""
  for marker in $MARKERS; do
    case "$body" in
      *"$marker"*) ;;
      *) missing="$missing $marker" ;;
    esac
  done
  if [ -z "$missing" ]; then
    printf " seeded\n"
    exit 0
  fi
  if [ "$(date +%s)" -ge "$deadline" ]; then
    printf "\nseeds missing in time:%s\n" "$missing"
    exit 1
  fi
  printf "."
  sleep 10
done
