#!/bin/bash

# Backend Unit Tests Runner
# Runs Go unit tests with coverage reporting

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
COVERAGE_DIR="$REPORT_DIR/coverage"

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

main() {
    cd "$PROJECT_ROOT"

    print_header "Backend Unit Tests"

    # Create directories
    mkdir -p "$COVERAGE_DIR"

    # Get all packages with tests
    print_info "Discovering test packages..."
    packages=$(go list ./... 2>/dev/null | grep -v /vendor/ | grep -v /tests/integration || true)

    if [[ -z "$packages" ]]; then
        print_info "No unit test packages found, creating sample tests..."
        # We'll handle this case gracefully
        packages="./..."
    fi

    # Run tests with coverage
    print_info "Running unit tests with coverage..."

    if go test -v \
        -race \
        -coverprofile="$COVERAGE_DIR/coverage.out" \
        -covermode=atomic \
        -timeout=5m \
        $packages \
        2>&1 | tee "$REPORT_DIR/backend-tests.log"; then

        print_success "Unit tests passed"

        # Generate coverage reports
        if [[ -f "$COVERAGE_DIR/coverage.out" ]]; then
            print_info "Generating coverage reports..."

            # HTML report
            go tool cover -html="$COVERAGE_DIR/coverage.out" -o "$COVERAGE_DIR/coverage.html"
            print_success "HTML coverage report: $COVERAGE_DIR/coverage.html"

            # Text summary
            go tool cover -func="$COVERAGE_DIR/coverage.out" > "$COVERAGE_DIR/coverage.txt"

            # Extract coverage percentage
            coverage=$(go tool cover -func="$COVERAGE_DIR/coverage.out" | grep total | awk '{print $3}' | sed 's/%//')
            echo "$coverage" > "$COVERAGE_DIR/coverage-percent.txt"

            print_success "Total coverage: ${coverage}%"
        fi

        return 0
    else
        print_error "Unit tests failed"
        return 1
    fi
}

main "$@"
