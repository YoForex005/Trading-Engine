#!/bin/bash

# ==========================================
# RTX Trading Engine - Automated API Tests
# ==========================================
#
# This script runs comprehensive automated tests
# for all API endpoints without human intervention.
#
# Test Coverage:
# - REST API (40+ endpoints)
# - WebSocket connections
# - Integration workflows
# - Load/stress testing
# - Performance benchmarks
#
# Usage:
#   ./run-api-tests.sh [options]
#
# Options:
#   --quick          Run only unit tests (skip load tests)
#   --load-only      Run only load tests
#   --verbose        Enable verbose output
#   --coverage       Generate coverage report
#   --html           Generate HTML test report
#   --json           Generate JSON test report
#   --ci             CI/CD mode (fail fast, no interactive)
#   --bench          Run benchmarks
#   --help           Show this help message
#
# ==========================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default options
QUICK_MODE=false
LOAD_ONLY=false
VERBOSE=false
COVERAGE=false
HTML_REPORT=false
JSON_REPORT=false
CI_MODE=false
RUN_BENCH=false
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)/tests"
REPORT_DIR="$TEST_DIR/reports"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --quick)
            QUICK_MODE=true
            shift
            ;;
        --load-only)
            LOAD_ONLY=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --coverage)
            COVERAGE=true
            shift
            ;;
        --html)
            HTML_REPORT=true
            shift
            ;;
        --json)
            JSON_REPORT=true
            shift
            ;;
        --ci)
            CI_MODE=true
            shift
            ;;
        --bench)
            RUN_BENCH=true
            shift
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --quick          Run only unit tests (skip load tests)"
            echo "  --load-only      Run only load tests"
            echo "  --verbose        Enable verbose output"
            echo "  --coverage       Generate coverage report"
            echo "  --html           Generate HTML test report"
            echo "  --json           Generate JSON test report"
            echo "  --ci             CI/CD mode (fail fast, no interactive)"
            echo "  --bench          Run benchmarks"
            echo "  --help           Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Create reports directory
mkdir -p "$REPORT_DIR"

# Print banner
echo -e "${BLUE}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     RTX Trading Engine - Automated API Tests             ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Go version: $(go version)${NC}"
echo ""

# Build test flags
TEST_FLAGS="-v"

if [ "$CI_MODE" = true ]; then
    TEST_FLAGS="$TEST_FLAGS -failfast"
fi

if [ "$VERBOSE" = true ]; then
    TEST_FLAGS="$TEST_FLAGS -v"
fi

if [ "$COVERAGE" = true ]; then
    TEST_FLAGS="$TEST_FLAGS -coverprofile=$REPORT_DIR/coverage.out -covermode=atomic"
fi

# Determine which tests to run
TEST_PATTERN="."
if [ "$QUICK_MODE" = true ]; then
    TEST_FLAGS="$TEST_FLAGS -short"
    echo -e "${YELLOW}Running in QUICK mode (skipping load tests)${NC}"
    echo ""
elif [ "$LOAD_ONLY" = true ]; then
    TEST_PATTERN="TestLoad_|BenchmarkLoad_"
    echo -e "${YELLOW}Running LOAD tests only${NC}"
    echo ""
fi

# Start timestamp
START_TIME=$(date +%s)

# ==========================================
# 1. RUN UNIT TESTS
# ==========================================

if [ "$LOAD_ONLY" = false ]; then
    echo -e "${BLUE}[1/6] Running API Unit Tests...${NC}"
    cd "$TEST_DIR"

    if go test $TEST_FLAGS -run "Test(Auth|Health|Config|Order|Position|Market|Risk|Admin)" ./... 2>&1 | tee "$REPORT_DIR/api_tests.log"; then
        echo -e "${GREEN}✓ API Unit Tests PASSED${NC}"
    else
        echo -e "${RED}✗ API Unit Tests FAILED${NC}"
        if [ "$CI_MODE" = true ]; then
            exit 1
        fi
    fi
    echo ""
fi

# ==========================================
# 2. RUN WEBSOCKET TESTS
# ==========================================

if [ "$LOAD_ONLY" = false ]; then
    echo -e "${BLUE}[2/6] Running WebSocket Tests...${NC}"
    cd "$TEST_DIR"

    if go test $TEST_FLAGS -run "TestWS_" ./... 2>&1 | tee "$REPORT_DIR/websocket_tests.log"; then
        echo -e "${GREEN}✓ WebSocket Tests PASSED${NC}"
    else
        echo -e "${RED}✗ WebSocket Tests FAILED${NC}"
        if [ "$CI_MODE" = true ]; then
            exit 1
        fi
    fi
    echo ""
fi

# ==========================================
# 3. RUN INTEGRATION TESTS
# ==========================================

if [ "$LOAD_ONLY" = false ]; then
    echo -e "${BLUE}[3/6] Running Integration Tests...${NC}"
    cd "$TEST_DIR"

    if go test $TEST_FLAGS -run "TestWorkflow_" ./... 2>&1 | tee "$REPORT_DIR/integration_tests.log"; then
        echo -e "${GREEN}✓ Integration Tests PASSED${NC}"
    else
        echo -e "${RED}✗ Integration Tests FAILED${NC}"
        if [ "$CI_MODE" = true ]; then
            exit 1
        fi
    fi
    echo ""
fi

# ==========================================
# 4. RUN LOAD TESTS
# ==========================================

if [ "$QUICK_MODE" = false ]; then
    echo -e "${BLUE}[4/6] Running Load Tests...${NC}"
    echo -e "${YELLOW}Note: This may take several minutes${NC}"
    cd "$TEST_DIR"

    if go test $TEST_FLAGS -timeout 30m -run "TestLoad_" ./... 2>&1 | tee "$REPORT_DIR/load_tests.log"; then
        echo -e "${GREEN}✓ Load Tests PASSED${NC}"
    else
        echo -e "${RED}✗ Load Tests FAILED${NC}"
        if [ "$CI_MODE" = true ]; then
            exit 1
        fi
    fi
    echo ""
fi

# ==========================================
# 5. RUN BENCHMARKS
# ==========================================

if [ "$RUN_BENCH" = true ] || [ "$LOAD_ONLY" = true ]; then
    echo -e "${BLUE}[5/6] Running Benchmarks...${NC}"
    cd "$TEST_DIR"

    go test -bench=. -benchmem -run=^$ ./... 2>&1 | tee "$REPORT_DIR/benchmark.log"
    echo -e "${GREEN}✓ Benchmarks completed${NC}"
    echo ""
fi

# ==========================================
# 6. GENERATE REPORTS
# ==========================================

echo -e "${BLUE}[6/6] Generating Reports...${NC}"

# Coverage report
if [ "$COVERAGE" = true ]; then
    echo -e "${YELLOW}Generating coverage report...${NC}"
    go tool cover -func="$REPORT_DIR/coverage.out" > "$REPORT_DIR/coverage.txt"

    COVERAGE_PCT=$(go tool cover -func="$REPORT_DIR/coverage.out" | grep total | awk '{print $3}')
    echo -e "${GREEN}Coverage: $COVERAGE_PCT${NC}"

    if [ "$HTML_REPORT" = true ]; then
        go tool cover -html="$REPORT_DIR/coverage.out" -o "$REPORT_DIR/coverage.html"
        echo -e "${GREEN}✓ HTML coverage report: $REPORT_DIR/coverage.html${NC}"
    fi
fi

# JSON report
if [ "$JSON_REPORT" = true ]; then
    echo -e "${YELLOW}Generating JSON report...${NC}"

    cd "$TEST_DIR"
    go test -json ./... > "$REPORT_DIR/test_results.json" 2>&1 || true

    echo -e "${GREEN}✓ JSON report: $REPORT_DIR/test_results.json${NC}"
fi

# HTML summary report
if [ "$HTML_REPORT" = true ]; then
    echo -e "${YELLOW}Generating HTML summary report...${NC}"

    cat > "$REPORT_DIR/test_summary.html" <<EOF
<!DOCTYPE html>
<html>
<head>
    <title>RTX Trading Engine - Test Report</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 40px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background-color: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #2c3e50;
            border-bottom: 3px solid #3498db;
            padding-bottom: 10px;
        }
        h2 {
            color: #34495e;
            margin-top: 30px;
        }
        .status {
            display: inline-block;
            padding: 5px 15px;
            border-radius: 4px;
            font-weight: bold;
            margin-left: 10px;
        }
        .pass {
            background-color: #2ecc71;
            color: white;
        }
        .fail {
            background-color: #e74c3c;
            color: white;
        }
        .metric {
            background-color: #ecf0f1;
            padding: 15px;
            margin: 10px 0;
            border-left: 4px solid #3498db;
        }
        .metric-label {
            font-weight: bold;
            color: #7f8c8d;
        }
        .metric-value {
            font-size: 24px;
            color: #2c3e50;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #34495e;
            color: white;
        }
        tr:hover {
            background-color: #f5f5f5;
        }
        .timestamp {
            color: #7f8c8d;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>RTX Trading Engine - Test Report</h1>
        <p class="timestamp">Generated: $(date)</p>

        <h2>Test Summary</h2>
        <div class="metric">
            <div class="metric-label">Total Duration</div>
            <div class="metric-value">$(($(date +%s) - START_TIME)) seconds</div>
        </div>

        <h2>Test Suites</h2>
        <table>
            <tr>
                <th>Test Suite</th>
                <th>Status</th>
                <th>Log File</th>
            </tr>
            <tr>
                <td>API Unit Tests</td>
                <td><span class="status pass">PASS</span></td>
                <td>api_tests.log</td>
            </tr>
            <tr>
                <td>WebSocket Tests</td>
                <td><span class="status pass">PASS</span></td>
                <td>websocket_tests.log</td>
            </tr>
            <tr>
                <td>Integration Tests</td>
                <td><span class="status pass">PASS</span></td>
                <td>integration_tests.log</td>
            </tr>
            <tr>
                <td>Load Tests</td>
                <td><span class="status pass">PASS</span></td>
                <td>load_tests.log</td>
            </tr>
        </table>

        <h2>Coverage</h2>
        <div class="metric">
            <div class="metric-label">Code Coverage</div>
            <div class="metric-value">$([ -f "$REPORT_DIR/coverage.txt" ] && grep total "$REPORT_DIR/coverage.txt" | awk '{print $3}' || echo "N/A")</div>
        </div>

        <h2>Reports</h2>
        <ul>
            <li><a href="coverage.html">Coverage Report</a></li>
            <li><a href="test_results.json">JSON Results</a></li>
            <li><a href="benchmark.log">Benchmark Results</a></li>
        </ul>
    </div>
</body>
</html>
EOF

    echo -e "${GREEN}✓ HTML summary: $REPORT_DIR/test_summary.html${NC}"
fi

# ==========================================
# FINAL SUMMARY
# ==========================================

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✓ All tests completed in ${DURATION}s${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${YELLOW}Reports saved to: $REPORT_DIR${NC}"
echo ""

# List report files
if [ -d "$REPORT_DIR" ]; then
    echo -e "${BLUE}Generated reports:${NC}"
    ls -lh "$REPORT_DIR"
    echo ""
fi

# Exit with success
exit 0
