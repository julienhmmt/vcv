#!/bin/sh

# Generate Go coverage reports for SonarQube.
# Test failures are logged but do not stop coverage generation, so SonarScanner
# can still run and surface test issues in the SonarQube UI.

set -u

REPO_ROOT="$(pwd)"
GO_TEST_FLAGS="${GO_TEST_FLAGS:-}"
SERVER_DIR="${SERVER_DIR:-app}"

mkdir -p .sonar

echo "Generating server coverage..."
cd "${SERVER_DIR}" || exit 1
go test ${GO_TEST_FLAGS} -count=1 -coverprofile=coverage.out ./...
server_rc=$?
sed -e 's|^vcv/|app/|' coverage.out > "$REPO_ROOT/.sonar/server-coverage.out"

if [ $server_rc -ne 0 ]; then
    echo "Warning: one or more server tests failed. Coverage report was still generated." >&2
fi

echo "Coverage reports ready in .sonar/"
