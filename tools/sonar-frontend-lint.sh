#!/bin/sh

# Run ESLint on app/web/frontend (including .svelte files) and convert the
# results to SonarQube generic issue format for import as external issues.
#
# SonarQube Community Build cannot parse .svelte files, but eslint-plugin-svelte
# can. This script bridges the gap: ESLint checks all .svelte/.ts/.js files,
# and the results are imported into SonarQube via sonar.externalIssuesReportPaths.
#
# Output files:
#   .sonar/web-eslint.json
#   .sonar/web-lcov.info

set -u

mkdir -p .sonar

REPO_ROOT=$(pwd)
FRONTEND_DIR=app/web/frontend

# Convert ESLint JSON output to SonarQube generic issue format.
# Args: $1 = output file, $2 = input ESLint JSON file, $3 = prefix to strip
convert_eslint_to_sonar() {
    out_file="$1"
    in_file="$2"
    prefix="$3"
    python3 "$REPO_ROOT/tools/eslint-to-sonar.py" "$in_file" "$out_file" "$prefix"
}

echo "Running ESLint on $FRONTEND_DIR/..."
cd "$FRONTEND_DIR" || exit 1
pnpm exec eslint . -f json 2>/dev/null > /tmp/eslint-web.json
cd "$REPO_ROOT" || exit 1
# Strip only the repo root so paths match SonarQube's indexed sources
# (app/web/frontend/src/...); stripping the frontend dir would leave
# src/... paths that SonarQube can't attach to indexed files.
convert_eslint_to_sonar ".sonar/web-eslint.json" /tmp/eslint-web.json "$REPO_ROOT"

rm -f /tmp/eslint-web.json

echo "Running vitest coverage on $FRONTEND_DIR/..."
cd "$FRONTEND_DIR" || exit 1
pnpm test:coverage >/dev/null 2>&1 || echo "Warning: vitest coverage reported failures (report still generated)." >&2
cd "$REPO_ROOT" || exit 1
if [ -f "$FRONTEND_DIR/coverage/lcov.info" ]; then
    # Vitest emits paths relative to the frontend dir (e.g. "src/lib/utils/x.ts");
    # SonarQube's base dir is the repo root and sources are under
    # app/web/frontend/src, so prefix every SF: line accordingly.
    # Drop files SonarQube does not index: .svelte/.svelte.ts/.svelte.js
    # (excluded via sonar.javascript.exclusions) — otherwise SonarQube logs
    # unresolved path warnings and the coverage sensor stalls.
    sed "s|^SF:src/|SF:$FRONTEND_DIR/src/|" "$FRONTEND_DIR/coverage/lcov.info" \
        | grep -v -E '^SF:.*\.(svelte|svelte\.ts|svelte\.js)$' \
        > .sonar/web-lcov.info
else
    echo "Warning: $FRONTEND_DIR/coverage/lcov.info not found; web coverage will be 0%." >&2
    : > .sonar/web-lcov.info
fi

echo "ESLint + coverage reports ready in .sonar/"
