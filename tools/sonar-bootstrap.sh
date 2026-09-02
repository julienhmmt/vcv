#!/bin/sh

# Bootstrap a local SonarQube container for VCV.
# - Waits for the server to be ready.
# - Sets a strong admin password (randomly generated or from SONAR_ADMIN_PASSWORD).
# - Creates/rotates an analysis token and stores it in .sonar/token.
# - Provisions the project if it does not exist yet.
#
# The script is idempotent: it reuses an existing admin password from
# .sonar/admin-password and rotates the token on each run.

set -u

SONAR_URL="http://localhost:9000"
ADMIN_USER="admin"
DEFAULT_PASS="admin"
TOKEN_NAME="vcv-local"
TOKEN_FILE=".sonar/token"
PASS_FILE=".sonar/admin-password"

# SonarQube projects to provision (one per line, key|name format).
# Newlines separate entries so names with spaces are preserved.
PROJECTS="vcv-server|VaultCertsViewer (Go)
vcv-web|VaultCertsViewer frontend"

mkdir -p .sonar

# Resolve the admin password to use.
if [ -f "$PASS_FILE" ]; then
    target_pass=$(cat "$PASS_FILE")
elif [ -n "${SONAR_ADMIN_PASSWORD:-}" ]; then
    target_pass="$SONAR_ADMIN_PASSWORD"
else
    target_pass="VCV-$(openssl rand -base64 18 | LC_ALL=C tr -dc 'A-Za-z0-9')"
fi

wait_for_sonarqube() {
    echo "Waiting for SonarQube at $SONAR_URL..."
    status=""
    attempts=0
    max_attempts=60
    while [ "$status" != "UP" ] && [ $attempts -lt $max_attempts ]; do
        status=$(curl -s -u "$ADMIN_USER:$DEFAULT_PASS" "$SONAR_URL/api/system/status" 2>/dev/null \
            | python3 -c "import sys, json; print(json.load(sys.stdin).get('status', ''))" 2>/dev/null || true)
        if [ "$status" != "UP" ]; then
            sleep 2
        fi
        attempts=$((attempts + 1))
    done

    if [ "$status" != "UP" ]; then
        echo "SonarQube did not become ready within $((max_attempts * 2)) seconds." >&2
        exit 1
    fi
    echo "SonarQube is up."
}

determine_current_password() {
    # Try the target password first (server may already have been bootstrapped),
    # then fall back to the factory default.
    for try_pass in "$target_pass" "$DEFAULT_PASS"; do
        http_code=$(curl -s -o /dev/null -w "%{http_code}" -u "$ADMIN_USER:$try_pass" \
            "$SONAR_URL/api/user_tokens/search" || echo "000")
        if [ "$http_code" = "200" ]; then
            echo "$try_pass"
            return 0
        fi
    done

    echo "Unable to authenticate to SonarQube." >&2
    echo "Set the current password in SONAR_ADMIN_PASSWORD or remove .sonar/admin-password and start fresh." >&2
    return 1
}

change_default_password() {
    new_pass=$1
    if [ "$new_pass" = "$DEFAULT_PASS" ]; then
        echo "Using default admin password (not recommended for persistent servers)." >&2
        return 0
    fi

    echo "Changing default admin password..."
    http_code=$(curl -s -o /dev/null -w "%{http_code}" -u "$ADMIN_USER:$DEFAULT_PASS" -X POST \
        "$SONAR_URL/api/users/change_password?login=$ADMIN_USER&previousPassword=$DEFAULT_PASS&password=$new_pass" \
        || echo "000")
    if [ "$http_code" != "204" ] && [ "$http_code" != "200" ]; then
        echo "Warning: could not change admin password (HTTP $http_code)." >&2
        return 1
    fi
    return 0
}

create_token() {
    current_pass=$1
    echo "Creating analysis token..."
    curl -s -o /dev/null -w "%{http_code}" -u "$ADMIN_USER:$current_pass" -X POST \
        "$SONAR_URL/api/user_tokens/revoke?name=$TOKEN_NAME" >/dev/null || true

    token_json=$(curl -s -u "$ADMIN_USER:$current_pass" -X POST \
        "$SONAR_URL/api/user_tokens/generate?name=$TOKEN_NAME" \
        || { echo "Failed to create token" >&2; return 1; })
    token=$(echo "$token_json" | python3 -c "import sys, json; print(json.load(sys.stdin).get('token', ''))" 2>/dev/null || true)

    if [ -z "$token" ]; then
        echo "Failed to create token. Response: $token_json" >&2
        return 1
    fi

    printf '%s\n' "$token" > "$TOKEN_FILE"
    echo "Token saved to $TOKEN_FILE"
}

provision_project() {
    current_pass=$1
    echo "Provisioning SonarQube projects..."
    echo "$PROJECTS" | while IFS='|' read -r key name; do
        [ -z "$key" ] && continue
        encoded_name=$(printf '%s' "$name" | python3 -c "import sys, urllib.parse; print(urllib.parse.quote(sys.stdin.read()))")
        http_code=$(curl -s -o /dev/null -w "%{http_code}" -u "$ADMIN_USER:$current_pass" -X POST \
            "$SONAR_URL/api/projects/create?name=${encoded_name}&project=${key}" 2>/dev/null || true)
        if [ "$http_code" = "200" ] || [ "$http_code" = "400" ]; then
            echo "  $key — OK"
        else
            echo "  $key — HTTP $http_code (may already exist)"
        fi
    done
}

main() {
    wait_for_sonarqube

    current_pass=$(determine_current_password)

    if [ "$current_pass" = "$DEFAULT_PASS" ] && [ ! -f "$PASS_FILE" ]; then
        if change_default_password "$target_pass"; then
            current_pass="$target_pass"
            printf '%s\n' "$target_pass" > "$PASS_FILE"
        fi
    elif [ "$current_pass" = "$target_pass" ] && [ ! -f "$PASS_FILE" ]; then
        printf '%s\n' "$target_pass" > "$PASS_FILE"
    fi

    create_token "$current_pass" || exit 1
    provision_project "$current_pass"
    echo "Bootstrap complete. Projects:"
    echo "$PROJECTS" | while IFS='|' read -r key name; do
        [ -z "$key" ] && continue
        echo "  $SONAR_URL/dashboard?id=$key"
    done
}

main "$@"
