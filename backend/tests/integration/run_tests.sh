#!/bin/bash

# Integration Test Runner Script
# Runs all Go integration tests with various options

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== RTX Trading Engine Integration Tests ===${NC}\n"

# Parse command line arguments
MODE="${1:-all}"

# Test directory
TEST_DIR="./tests/integration"

# Function to run tests
run_tests() {
    local test_file=$1
    local test_name=$2

    echo -e "${YELLOW}Running ${test_name}...${NC}"

    if go test "${TEST_DIR}/${test_file}" -v -race -timeout 30s; then
        echo -e "${GREEN}✓ ${test_name} passed${NC}\n"
    else
        echo -e "${RED}✗ ${test_name} failed${NC}\n"
        exit 1
    fi
}

case $MODE in
    all)
        echo "Running all integration tests..."
        echo ""
        run_tests "api_test.go" "API Tests"
        run_tests "websocket_test.go" "WebSocket Tests"
        run_tests "order_flow_test.go" "Order Flow Tests"
        run_tests "admin_flow_test.go" "Admin Flow Tests"
        ;;

    api)
        run_tests "api_test.go" "API Tests"
        ;;

    websocket|ws)
        run_tests "websocket_test.go" "WebSocket Tests"
        ;;

    order|orders)
        run_tests "order_flow_test.go" "Order Flow Tests"
        ;;

    admin)
        run_tests "admin_flow_test.go" "Admin Flow Tests"
        ;;

    coverage)
        echo "Running tests with coverage..."
        go test ${TEST_DIR}/... -cover -coverprofile=coverage.out
        echo ""
        echo -e "${GREEN}Coverage report:${NC}"
        go tool cover -func=coverage.out
        echo ""
        echo "To view HTML coverage report, run:"
        echo "  go tool cover -html=coverage.out"
        ;;

    bench|benchmark)
        echo "Running benchmarks..."
        go test ${TEST_DIR}/... -bench=. -benchmem -run=^$
        ;;

    race)
        echo "Running race detection tests..."
        go test ${TEST_DIR}/... -race -v
        ;;

    short)
        echo "Running short tests only..."
        go test ${TEST_DIR}/... -short -v
        ;;

    *)
        echo -e "${RED}Unknown test mode: $MODE${NC}"
        echo ""
        echo "Usage: $0 [MODE]"
        echo ""
        echo "Modes:"
        echo "  all         - Run all tests (default)"
        echo "  api         - Run API tests only"
        echo "  websocket   - Run WebSocket tests only"
        echo "  order       - Run order flow tests only"
        echo "  admin       - Run admin flow tests only"
        echo "  coverage    - Run with coverage report"
        echo "  benchmark   - Run performance benchmarks"
        echo "  race        - Run with race detection"
        echo "  short       - Run short tests only"
        echo ""
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}=== All Tests Completed Successfully ===${NC}"
