#!/bin/bash
# Quick test to verify scripts work

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🧪 Testing script functionality..."

# Test 1: Check scripts are executable
echo "✓ Checking permissions..."
for script in run-all-tests.sh run-backend-tests.sh run-integration-tests.sh run-e2e-tests.sh run-load-tests.sh verify-coverage.sh generate-test-report.sh; do
    if [[ ! -x "$SCRIPT_DIR/$script" ]]; then
        echo "✗ Script not executable: $script"
        exit 1
    fi
done
echo "  All scripts are executable"

# Test 2: Check help messages work
echo "✓ Checking help messages..."
if "$SCRIPT_DIR/run-all-tests.sh" --help > /dev/null 2>&1; then
    echo "  Help message works"
else
    echo "✗ Help message failed"
    exit 1
fi

# Test 3: Verify directory structure
echo "✓ Checking directory structure..."
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
if [[ -f "$PROJECT_ROOT/go.mod" ]]; then
    echo "  Go module found"
else
    echo "✗ Go module not found"
    exit 1
fi

echo ""
echo "✅ All checks passed!"
echo ""
echo "Ready to run tests with:"
echo "  ./scripts/test/run-all-tests.sh"
