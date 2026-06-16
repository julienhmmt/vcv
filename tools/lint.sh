#!/bin/sh

# Lint and format CSS, Markdown, JSON, YAML files for VCV project
# Uses project's local node_modules (bun)
# Runs stylelint for CSS and prettier for formatting

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$ROOT_DIR"

if [ ! -f "package.json" ]; then
    echo "  ❌ Error: package.json not found in $ROOT_DIR"
    exit 1
fi

echo "==> Linting and formatting VCV project"
echo ""

# Check if node_modules exists
if [ ! -d "node_modules" ]; then
    echo "  ⚠️  node_modules not found. Running 'bun install'..."
    bun install
    echo ""
fi

# Run stylelint for CSS
echo "  [1/2] Stylelint (CSS)..."
if bunx stylelint "app/web/frontend/src/**/*.css" --fix; then
    echo "  ✅ Stylelint: no errors"
else
    echo "  ❌ Stylelint: errors found"
    exit 1
fi
echo ""

# Run prettier for formatting
echo "  [2/2] Prettier (CSS, MD, YAML, JSON)..."
if bunx prettier --write "**/*.{css,md,json,yml,yaml}" --log-level warn; then
    echo "  ✅ Prettier: files formatted"
else
    echo "  ❌ Prettier: formatting failed"
    exit 1
fi
echo ""

echo "==> Done!"
