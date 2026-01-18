#!/bin/bash
# Performance Test Suite Runner
# Runs all load tests and generates comprehensive reports

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
RESULTS_DIR="$BACKEND_DIR/tests/performance/results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULTS_SUBDIR="$RESULTS_DIR/$TIMESTAMP"

# Server configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
WS_URL="${WS_URL:-ws://localhost:8080/ws}"

# Test configuration
RUN_LOAD_TEST="${RUN_LOAD_TEST:-true}"
RUN_STRESS_TEST="${RUN_STRESS_TEST:-true}"
RUN_SPIKE_TEST="${RUN_SPIKE_TEST:-true}"
RUN_SOAK_TEST="${RUN_SOAK_TEST:-false}"  # Disabled by default (24h)
RUN_GO_BENCHMARKS="${RUN_GO_BENCHMARKS:-true}"

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║          Trading Engine Performance Test Suite            ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Create results directory
mkdir -p "$RESULTS_SUBDIR"
echo -e "${GREEN}✓${NC} Results directory: $RESULTS_SUBDIR"
echo ""

# Check dependencies
echo -e "${YELLOW}Checking dependencies...${NC}"

if ! command -v k6 &> /dev/null; then
    echo -e "${RED}✗ k6 not found${NC}"
    echo "Installing k6..."

    if [[ "$OSTYPE" == "darwin"* ]]; then
        brew install k6
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
        echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
        sudo apt-get update
        sudo apt-get install k6
    fi
fi

echo -e "${GREEN}✓${NC} k6 installed"

# Check if server is running
echo -e "${YELLOW}Checking server...${NC}"
if curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health" | grep -q "200"; then
    echo -e "${GREEN}✓${NC} Server is running at $BASE_URL"
else
    echo -e "${RED}✗ Server not responding at $BASE_URL${NC}"
    echo "Please start the server before running tests"
    exit 1
fi

echo ""

# Function to run a test
run_test() {
    local test_name=$1
    local test_file=$2
    local duration=$3

    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}Running: $test_name${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    local output_file="$RESULTS_SUBDIR/$test_name"

    # Run k6 test with environment variables
    if k6 run \
        --out json="$output_file.json" \
        --summary-export="$output_file-summary.json" \
        --env BASE_URL="$BASE_URL" \
        --env WS_URL="$WS_URL" \
        "$BACKEND_DIR/tests/performance/$test_file" \
        | tee "$output_file.log"; then

        echo -e "${GREEN}✓ $test_name completed successfully${NC}"
        return 0
    else
        echo -e "${RED}✗ $test_name failed${NC}"
        return 1
    fi
}

# Track test results
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Run Load Test
if [[ "$RUN_LOAD_TEST" == "true" ]]; then
    if run_test "load-test" "load-test.js" "25m"; then
        ((TESTS_PASSED++))
    else
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))
    echo ""
fi

# Run Stress Test
if [[ "$RUN_STRESS_TEST" == "true" ]]; then
    if run_test "stress-test" "stress-test.js" "19m"; then
        ((TESTS_PASSED++))
    else
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))
    echo ""
fi

# Run Spike Test
if [[ "$RUN_SPIKE_TEST" == "true" ]]; then
    if run_test "spike-test" "spike-test.js" "20m"; then
        ((TESTS_PASSED++))
    else
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))
    echo ""
fi

# Run Soak Test (24 hours - only if explicitly enabled)
if [[ "$RUN_SOAK_TEST" == "true" ]]; then
    echo -e "${YELLOW}⚠ Starting 24-hour soak test${NC}"
    echo -e "${YELLOW}  This will run for 24 hours. Press Ctrl+C to cancel.${NC}"
    sleep 5

    if run_test "soak-test" "soak-test.js" "24h"; then
        ((TESTS_PASSED++))
    else
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))
    echo ""
fi

# Run Go Benchmarks
if [[ "$RUN_GO_BENCHMARKS" == "true" ]]; then
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}Running: Go Benchmarks${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    cd "$BACKEND_DIR/tests/performance"

    if go test -bench=. -benchmem -benchtime=10s -cpu=1,2,4,8 \
        -timeout=30m \
        > "$RESULTS_SUBDIR/go-benchmarks.txt" 2>&1; then

        echo -e "${GREEN}✓ Go benchmarks completed successfully${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ Go benchmarks failed${NC}"
        ((TESTS_FAILED++))
    fi
    ((TESTS_RUN++))

    cd "$BACKEND_DIR"
    echo ""
fi

# Generate summary report
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                    Test Summary                            ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "Tests Run:    ${BLUE}$TESTS_RUN${NC}"
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [[ $TESTS_FAILED -eq 0 ]]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
else
    echo -e "${RED}✗ Some tests failed${NC}"
fi

echo ""
echo -e "${YELLOW}Results saved to:${NC} $RESULTS_SUBDIR"
echo ""

# Run analysis script
if [[ -f "$SCRIPT_DIR/analyze-results.sh" ]]; then
    echo -e "${YELLOW}Running results analysis...${NC}"
    "$SCRIPT_DIR/analyze-results.sh" "$RESULTS_SUBDIR"
fi

# Exit with appropriate code
if [[ $TESTS_FAILED -eq 0 ]]; then
    exit 0
else
    exit 1
fi
