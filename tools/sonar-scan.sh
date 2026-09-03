#!/bin/sh

# Run the SonarScanner CLI against the local SonarQube container.
# Expects .sonar/token to exist (created by `make sonar-bootstrap`).
#
# Usage:
#   tools/sonar-scan.sh              # scan all projects in sonar-projects/
#   tools/sonar-scan.sh server       # scan one project by short name
#   tools/sonar-scan.sh web

set -u

TOKEN_FILE=".sonar/token"
SONAR_URL="http://localhost:9001"
PROPERTIES_DIR="sonar-projects"

# Map short names to properties files.
props_for() {
    case "$1" in
        server) echo "$PROPERTIES_DIR/server.properties" ;;
        web)    echo "$PROPERTIES_DIR/web.properties" ;;
        *)      echo "" ;;
    esac
}

# Resolve which properties files to scan.
if [ $# -ge 1 ]; then
    props_file=$(props_for "$1")
    if [ -z "$props_file" ] || [ ! -f "$props_file" ]; then
        echo "Unknown project '$1'. Use: server, web" >&2
        exit 2
    fi
    props_files="$props_file"
else
    props_files=""
    for f in "$PROPERTIES_DIR"/*.properties; do
        [ -f "$f" ] || continue
        props_files="$props_files $f"
    done
fi

if [ ! -f "$TOKEN_FILE" ]; then
    echo "SonarQube token not found. Run 'make sonar-bootstrap' first." >&2
    exit 1
fi

SONAR_TOKEN=$(tr -d '[:space:]' < "$TOKEN_FILE")
export SONAR_TOKEN

echo "Waiting for SonarQube at $SONAR_URL..."
status=""
attempts=0
max_attempts=60
while [ "$status" != "UP" ] && [ $attempts -lt $max_attempts ]; do
    status=$(curl -s "$SONAR_URL/api/system/status" 2>/dev/null \
        | python3 -c "import sys, json; print(json.load(sys.stdin).get('status', ''))" 2>/dev/null || true)
    if [ "$status" != "UP" ]; then
        sleep 2
    fi
    attempts=$((attempts + 1))
done

if [ "$status" != "UP" ]; then
    echo "SonarQube is not ready. Run 'make sonar-up' first." >&2
    exit 1
fi

overall_rc=0
scanned_keys=""

for props_file in $props_files; do
    [ -f "$props_file" ] || continue
    project_key=$(grep '^sonar.projectKey=' "$props_file" | head -1 | cut -d= -f2 | tr -d '[:space:]')
    scanned_keys="$scanned_keys $project_key"
    echo ""
    echo "=== Scanning $project_key ($props_file) ==="
    docker compose -f docker-compose.sonarqube.yml run --rm --no-deps \
        -e SONAR_TOKEN \
        sonar-scanner \
        -Dproject.settings="$props_file"
    rc=$?
    if [ $rc -ne 0 ]; then
        echo "Scan for $project_key failed (exit $rc)." >&2
        overall_rc=$rc
    else
        echo "Scan for $project_key complete: $SONAR_URL/dashboard?id=$project_key"
    fi
    # Brief pause so the server can process the report before the next scan
    # hammers it, reducing OOM risk on memory-constrained hosts.
    sleep 3
done

echo ""
# Print a summary table by querying the SonarQube API.
echo "=== Scan summary ==="
admin_pass=""
if [ -f ".sonar/admin-password" ]; then
    admin_pass=$(cat ".sonar/admin-password")
fi

printf "%-20s  %-8s  %-8s  %-8s  %-8s\n" "PROJECT" "GATE" "BUGS" "SMELLS" "COVERAGE"
printf "%-20s  %-8s  %-8s  %-8s  %-8s\n" "-------" "----" "----" "------" "--------"

for key in $scanned_keys; do
    if [ -n "$admin_pass" ]; then
        metrics=$(curl -s -u "admin:$admin_pass" \
            "$SONAR_URL/api/measures/component?component=$key&metricKeys=bugs,code_smells,coverage,security_hotspots" 2>/dev/null)
        gate=$(curl -s -u "admin:$admin_pass" \
            "$SONAR_URL/api/qualitygates/project_status?projectKey=$key" 2>/dev/null \
            | python3 -c "import sys, json; print(json.load(sys.stdin).get('projectStatus', {}).get('status', '?'))" 2>/dev/null)
        bugs=$(echo "$metrics" | python3 -c "
import sys, json
try:
    m = {x['metric']: x['value'] for x in json.load(sys.stdin).get('component', {}).get('measures', [])}
    print(m.get('bugs', '0'))
except Exception:
    print('0')
" 2>/dev/null)
        smells=$(echo "$metrics" | python3 -c "
import sys, json
try:
    m = {x['metric']: x['value'] for x in json.load(sys.stdin).get('component', {}).get('measures', [])}
    print(m.get('code_smells', '0'))
except Exception:
    print('0')
" 2>/dev/null)
        coverage=$(echo "$metrics" | python3 -c "
import sys, json
try:
    m = {x['metric']: x['value'] for x in json.load(sys.stdin).get('component', {}).get('measures', [])}
    print(m.get('coverage', '-'))
except Exception:
    print('-')
" 2>/dev/null)
    else
        gate="?"; bugs="?"; smells="?"; coverage="-"
    fi
    printf "%-20s  %-8s  %-8s  %-8s  %-8s\n" "$key" "$gate" "$bugs" "$smells" "$coverage"
    echo "  $SONAR_URL/dashboard?id=$key"
done

echo ""
if [ $overall_rc -ne 0 ]; then
    echo "One or more scans failed (exit $overall_rc)." >&2
    exit $overall_rc
fi

echo "All scans complete."
