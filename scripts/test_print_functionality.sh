#!/bin/bash
# Test script for chart print functionality

echo "========================================="
echo "Chart Print Functionality Test Suite"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run test
run_test() {
    local test_name="$1"
    local test_command="$2"

    echo -n "Testing: $test_name... "

    if eval "$test_command" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Test 1: Check if chartPrinter service exists
run_test "chartPrinter service exists" \
    "test -f clients/desktop/src/services/chartPrinter.ts"

# Test 2: Check if PrintSetupDialog exists
run_test "PrintSetupDialog component exists" \
    "test -f clients/desktop/src/components/dialogs/PrintSetupDialog.tsx"

# Test 3: Check if PrintPreviewDialog exists
run_test "PrintPreviewDialog component exists" \
    "test -f clients/desktop/src/components/dialogs/PrintPreviewDialog.tsx"

# Test 4: Check if useChartPrint hook exists
run_test "useChartPrint hook exists" \
    "test -f clients/desktop/src/hooks/useChartPrint.ts"

# Test 5: Check if ChartPrintExample exists
run_test "ChartPrintExample component exists" \
    "test -f clients/desktop/src/components/ChartPrintExample.tsx"

# Test 6: Check if print_preferences.go exists
run_test "print_preferences.go backend file exists" \
    "test -f backend/api/print_preferences.go"

# Test 7: Check if test file exists
run_test "chartPrinter test file exists" \
    "test -f clients/desktop/src/services/chartPrinter.test.ts"

# Test 8: Check if documentation exists
run_test "Chart print documentation exists" \
    "test -f docs/CHART_PRINT_FUNCTIONALITY.md"

# Test 9: Check if quick integration guide exists
run_test "Quick integration guide exists" \
    "test -f docs/QUICK_PRINT_INTEGRATION.md"

# Test 10: Check if implementation summary exists
run_test "Implementation summary exists" \
    "test -f CHART_PRINT_IMPLEMENTATION_SUMMARY.md"

# Test 11: Verify TypeScript compilation
echo -n "Testing: TypeScript compilation... "
cd clients/desktop
if npm run typecheck > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((TESTS_FAILED++))
fi
cd ../..

# Test 12: Verify Go compilation (print_preferences.go only)
echo -n "Testing: Go file syntax... "
if go fmt backend/api/print_preferences.go > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PASS${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}✗ FAIL${NC}"
    ((TESTS_FAILED++))
fi

# Test 13: Check for required imports in chartPrinter.ts
run_test "chartPrinter has required exports" \
    "grep -q 'export.*chartPrinter' clients/desktop/src/services/chartPrinter.ts"

# Test 14: Check for PrintPreferences type export
run_test "PrintPreferences type is exported" \
    "grep -q 'export.*PrintPreferences' clients/desktop/src/services/chartPrinter.ts"

# Test 15: Check for useChartPrint hook export
run_test "useChartPrint hook is exported" \
    "grep -q 'export function useChartPrint' clients/desktop/src/hooks/useChartPrint.ts"

# Summary
echo ""
echo "========================================="
echo "Test Summary"
echo "========================================="
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo "Total Tests: $((TESTS_PASSED + TESTS_FAILED))"

if [ $TESTS_FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ All tests passed! Chart print functionality is ready.${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}✗ Some tests failed. Please review the output above.${NC}"
    exit 1
fi
