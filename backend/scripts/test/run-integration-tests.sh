#!/bin/bash

# Integration Tests Runner
# Runs API integration tests with test database

set -e
set -o pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'
BOLD='\033[1m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
REPORT_DIR="$PROJECT_ROOT/test-reports"

print_header() {
    echo -e "${BLUE}${BOLD}$1${NC}"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Setup test environment
setup_test_env() {
    print_info "Setting up integration test environment..."

    # Export test environment variables
    export TEST_MODE=true
    export REDIS_URL="redis://localhost:6379"
    export API_PORT=8081

    # Check if Redis is available
    if command -v redis-cli &> /dev/null; then
        if redis-cli ping &> /dev/null; then
            print_success "Redis is available"
        else
            print_info "Starting Redis..."
            redis-server --daemonize yes --port 6379 2>/dev/null || true
            sleep 2
        fi
    fi
}

# Cleanup test environment
cleanup_test_env() {
    print_info "Cleaning up test environment..."

    # Stop any test servers
    pkill -f "go run.*server" 2>/dev/null || true

    # Flush test Redis data
    if command -v redis-cli &> /dev/null && redis-cli ping &> /dev/null; then
        redis-cli flushdb &> /dev/null || true
    fi
}

main() {
    cd "$PROJECT_ROOT"

    print_header "Integration Tests"

    # Setup
    setup_test_env

    # Ensure cleanup on exit
    trap cleanup_test_env EXIT

    # Run integration tests
    print_info "Running integration tests..."

    if go test -v \
        -tags=integration \
        -timeout=10m \
        ./tests/integration/... \
        2>&1 | tee "$REPORT_DIR/integration-tests.log"; then

        print_success "Integration tests passed"
        return 0
    else
        print_error "Integration tests failed"
        return 1
    fi
}

main "$@"
