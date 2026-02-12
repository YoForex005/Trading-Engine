#!/bin/bash
# verify_build.sh
# Comprehensive build and code quality verification

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  RTX5 Backend Build Verification Suite    ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""

# Function to run check
run_check() {
    local name="$1"
    local command="$2"
    local critical="$3" # true/false

    ((TOTAL_CHECKS++))
    echo -e "${BLUE}[$TOTAL_CHECKS]${NC} Checking: $name..."

    if eval "$command" &> /tmp/check_output.txt; then
        echo -e "    ${GREEN}✓ PASS${NC}"
        ((PASSED_CHECKS++))
        return 0
    else
        if [ "$critical" = "true" ]; then
            echo -e "    ${RED}✗ FAIL (CRITICAL)${NC}"
            cat /tmp/check_output.txt
        else
            echo -e "    ${YELLOW}⚠ WARN${NC}"
        fi
        ((FAILED_CHECKS++))
        return 1
    fi
}

# Check 1: Go installation
run_check "Go installation" "command -v go" true

if [ $? -eq 0 ]; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "    ${GREEN}Go version: $GO_VERSION${NC}"
fi

# Check 2: Go module dependencies
run_check "Go module integrity" "go mod verify" true

# Check 3: Build main server
run_check "Build cmd/server" "go build -o /tmp/rtx5-server ./cmd/server/" true

# Check 4: Run go vet
run_check "go vet analysis" "go vet ./..." true

# Check 5: Run go fmt check
run_check "go fmt formatting" "test -z \$(gofmt -l .)" false

# Check 6: Check main.go size
MAIN_SIZE=$(wc -l < cmd/server/main.go)
run_check "main.go size (<500 lines)" "[ $MAIN_SIZE -lt 500 ]" false
echo -e "    Current size: ${YELLOW}$MAIN_SIZE${NC} lines"

# Check 7: Check for fmt.Println in production code
PRINTLN_COUNT=$(grep -r "fmt.Println" --include="*.go" . | \
    grep -v "./cmd/migrate" | \
    grep -v "./cmd/test_e2e" | \
    grep -v "./logging/examples" | \
    wc -l || echo 0)
run_check "No debug fmt.Println" "[ $PRINTLN_COUNT -eq 0 ]" false
echo -e "    Found: ${YELLOW}$PRINTLN_COUNT${NC} occurrences"

# Check 8: Check for panic calls
PANIC_COUNT=$(grep -r "panic(" --include="*.go" . | \
    grep -v "_test.go" | \
    grep -v "logging/examples" | \
    wc -l || echo 0)
run_check "No unsafe panic()" "[ $PANIC_COUNT -eq 0 ]" false
echo -e "    Found: ${YELLOW}$PANIC_COUNT${NC} panic calls"

# Check 9: Run unit tests
run_check "Unit tests" "go test ./... -short -timeout 30s" false

# Check 10: Check for TODO comments
TODO_COUNT=$(grep -r "TODO" --include="*.go" . | wc -l || echo 0)
echo -e "${BLUE}[INFO]${NC} Found ${YELLOW}$TODO_COUNT${NC} TODO comments"

# Check 11: Check for FIXME comments
FIXME_COUNT=$(grep -r "FIXME" --include="*.go" . | wc -l || echo 0)
echo -e "${BLUE}[INFO]${NC} Found ${YELLOW}$FIXME_COUNT${NC} FIXME comments"

# Check 12: Check for unused imports (via go build)
echo -e "${BLUE}[$((TOTAL_CHECKS+1))]${NC} Checking for unused imports..."
if go build ./... &> /tmp/unused_imports.txt; then
    echo -e "    ${GREEN}✓ No unused imports${NC}"
else
    if grep -q "imported and not used" /tmp/unused_imports.txt; then
        echo -e "    ${YELLOW}⚠ Found unused imports:${NC}"
        grep "imported and not used" /tmp/unused_imports.txt
    fi
fi

# Check 13: Check for race conditions
echo -e "${BLUE}[$((TOTAL_CHECKS+2))]${NC} Checking for race conditions (light check)..."
if go build -race ./cmd/server/ &> /dev/null; then
    echo -e "    ${GREEN}✓ Race detector compiled successfully${NC}"
else
    echo -e "    ${YELLOW}⚠ Race detector compilation failed${NC}"
fi

# Summary
echo ""
echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║           Verification Summary             ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""
echo -e "Total checks:  ${BLUE}$TOTAL_CHECKS${NC}"
echo -e "Passed:        ${GREEN}$PASSED_CHECKS${NC}"
echo -e "Failed:        ${RED}$FAILED_CHECKS${NC}"
echo ""

# Exit code
if [ $FAILED_CHECKS -eq 0 ]; then
    echo -e "${GREEN}✓ All checks passed!${NC}"
    echo ""
    echo "Build is ready for deployment."
    exit 0
else
    echo -e "${YELLOW}⚠ Some checks failed.${NC}"
    echo ""
    echo "Review the failures above and run the cleanup scripts:"
    echo "  - ./scripts/cleanup_debug_statements.sh"
    echo "  - ./scripts/fix_panic_calls.sh"
    echo "  - ./scripts/split_main_go.sh"
    exit 1
fi
