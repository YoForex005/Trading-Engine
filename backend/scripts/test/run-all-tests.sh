#!/bin/bash

# Master Test Runner - Runs all test suites with zero human intervention
# Exit codes: 0 = success, 1 = failure

set -e  # Exit on error
set -o pipefail  # Catch errors in pipelines

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Test configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
COVERAGE_THRESHOLD=80
PARALLEL_JOBS=4

# Output files
REPORT_DIR="$PROJECT_ROOT/test-reports"
COVERAGE_DIR="$REPORT_DIR/coverage"
LOG_FILE="$REPORT_DIR/test-run.log"

# Test results tracking
TESTS_PASSED=0
TESTS_FAILED=0
START_TIME=$(date +%s)

# Print banner
print_banner() {
    echo -e "${CYAN}${BOLD}"
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║                  RTX Trading Engine Test Suite                ║"
    echo "║                    Automated Test Runner                       ║"
    echo "╚═══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Print section header
print_header() {
    echo -e "\n${BLUE}${BOLD}▶ $1${NC}"
    echo -e "${BLUE}$(printf '─%.0s' {1..70})${NC}"
}

# Print success message
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Print error message
print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Print warning message
print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Print info message
print_info() {
    echo -e "${CYAN}ℹ${NC} $1"
}

# Cleanup function
cleanup() {
    local exit_code=$?

    print_header "Cleanup"

    # Stop test services
    if [[ -f "$SCRIPT_DIR/.test-services-running" ]]; then
        print_info "Stopping test services..."
        docker-compose -f "$PROJECT_ROOT/docker-compose.test.yml" down 2>/dev/null || true
        rm -f "$SCRIPT_DIR/.test-services-running"
    fi

    # Kill any remaining test processes
    pkill -f "go test" 2>/dev/null || true

    if [[ $exit_code -eq 0 ]]; then
        print_success "Cleanup completed"
    else
        print_warning "Cleanup completed with errors (exit code: $exit_code)"
    fi
}

# Setup trap for cleanup
trap cleanup EXIT INT TERM

# Check prerequisites
check_prerequisites() {
    print_header "Checking Prerequisites"

    local missing=()

    # Check Go
    if ! command -v go &> /dev/null; then
        missing+=("go (https://golang.org/dl/)")
    else
        print_success "Go $(go version | awk '{print $3}')"
    fi

    # Check Docker
    if ! command -v docker &> /dev/null; then
        missing+=("docker (https://docs.docker.com/get-docker/)")
    else
        print_success "Docker $(docker --version | awk '{print $3}' | sed 's/,//')"
    fi

    # Check Redis (optional)
    if ! command -v redis-cli &> /dev/null; then
        print_warning "redis-cli not found (optional for local testing)"
    else
        print_success "Redis CLI $(redis-cli --version | awk '{print $2}')"
    fi

    if [[ ${#missing[@]} -gt 0 ]]; then
        print_error "Missing required tools:"
        for tool in "${missing[@]}"; do
            echo "  - $tool"
        done
        exit 1
    fi

    print_success "All prerequisites satisfied"
}

# Setup test environment
setup_environment() {
    print_header "Setting Up Test Environment"

    # Create report directories
    mkdir -p "$REPORT_DIR"
    mkdir -p "$COVERAGE_DIR"

    # Clear old reports
    rm -f "$REPORT_DIR"/*.xml "$REPORT_DIR"/*.html "$REPORT_DIR"/*.json
    rm -f "$COVERAGE_DIR"/*

    # Start log file
    echo "Test run started at $(date)" > "$LOG_FILE"

    print_success "Report directory: $REPORT_DIR"
    print_success "Coverage directory: $COVERAGE_DIR"
}

# Start test services (Redis, mock servers, etc.)
start_test_services() {
    print_header "Starting Test Services"

    # Check if docker-compose.test.yml exists
    if [[ ! -f "$PROJECT_ROOT/docker-compose.test.yml" ]]; then
        print_warning "docker-compose.test.yml not found, skipping service startup"
        return 0
    fi

    # Start services
    print_info "Starting test containers..."
    cd "$PROJECT_ROOT"
    if docker-compose -f docker-compose.test.yml up -d &>> "$LOG_FILE"; then
        touch "$SCRIPT_DIR/.test-services-running"
        print_success "Test services started"

        # Wait for services to be ready
        print_info "Waiting for services to be ready..."
        sleep 5

        # Check Redis
        if docker-compose -f docker-compose.test.yml ps | grep -q redis; then
            print_success "Redis is ready"
        fi
    else
        print_warning "Failed to start test services, continuing without them"
    fi
}

# Run a test suite with retry logic
run_test_suite() {
    local name=$1
    local script=$2
    local max_retries=2
    local retry=0

    print_header "Running: $name"

    while [[ $retry -le $max_retries ]]; do
        if [[ $retry -gt 0 ]]; then
            print_warning "Retry attempt $retry/$max_retries"
            sleep 2
        fi

        if bash "$script" >> "$LOG_FILE" 2>&1; then
            print_success "$name passed"
            ((TESTS_PASSED++))
            return 0
        fi

        ((retry++))
    done

    print_error "$name failed after $max_retries retries"
    ((TESTS_FAILED++))
    return 1
}

# Run all test suites
run_tests() {
    print_header "Running Test Suites"

    local test_suites=(
        "Backend Unit Tests:$SCRIPT_DIR/run-backend-tests.sh"
        "Integration Tests:$SCRIPT_DIR/run-integration-tests.sh"
        "E2E Tests:$SCRIPT_DIR/run-e2e-tests.sh"
    )

    # Run tests in parallel if requested
    if [[ "${PARALLEL:-0}" == "1" ]]; then
        print_info "Running tests in parallel (jobs=$PARALLEL_JOBS)"

        for suite in "${test_suites[@]}"; do
            IFS=':' read -r name script <<< "$suite"
            (run_test_suite "$name" "$script") &
        done

        wait
    else
        # Sequential execution
        for suite in "${test_suites[@]}"; do
            IFS=':' read -r name script <<< "$suite"
            if ! run_test_suite "$name" "$script"; then
                if [[ "${FAIL_FAST:-0}" == "1" ]]; then
                    print_error "Fail-fast enabled, stopping test run"
                    exit 1
                fi
            fi
        done
    fi
}

# Verify coverage threshold
verify_coverage() {
    print_header "Verifying Coverage"

    if bash "$SCRIPT_DIR/verify-coverage.sh" >> "$LOG_FILE" 2>&1; then
        print_success "Coverage threshold met (≥${COVERAGE_THRESHOLD}%)"
        return 0
    else
        print_error "Coverage threshold not met (<${COVERAGE_THRESHOLD}%)"
        return 1
    fi
}

# Generate test report
generate_report() {
    print_header "Generating Test Report"

    if bash "$SCRIPT_DIR/generate-test-report.sh" >> "$LOG_FILE" 2>&1; then
        print_success "Test report generated"
        print_info "Report location: $REPORT_DIR/index.html"
        return 0
    else
        print_warning "Failed to generate test report"
        return 1
    fi
}

# Print summary
print_summary() {
    local end_time=$(date +%s)
    local duration=$((end_time - START_TIME))
    local minutes=$((duration / 60))
    local seconds=$((duration % 60))

    echo
    echo -e "${CYAN}${BOLD}╔═══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}${BOLD}║                       Test Summary                             ║${NC}"
    echo -e "${CYAN}${BOLD}╚═══════════════════════════════════════════════════════════════╝${NC}"
    echo

    echo -e "  ${BOLD}Test Suites:${NC}"
    echo -e "    ${GREEN}✓ Passed:${NC}  $TESTS_PASSED"
    echo -e "    ${RED}✗ Failed:${NC}  $TESTS_FAILED"
    echo

    echo -e "  ${BOLD}Duration:${NC} ${minutes}m ${seconds}s"
    echo -e "  ${BOLD}Reports:${NC}  $REPORT_DIR"
    echo

    if [[ $TESTS_FAILED -eq 0 ]]; then
        echo -e "${GREEN}${BOLD}  ╔═══════════════════════════════════════╗${NC}"
        echo -e "${GREEN}${BOLD}  ║     ALL TESTS PASSED SUCCESSFULLY     ║${NC}"
        echo -e "${GREEN}${BOLD}  ╚═══════════════════════════════════════╝${NC}"
        echo
        return 0
    else
        echo -e "${RED}${BOLD}  ╔═══════════════════════════════════════╗${NC}"
        echo -e "${RED}${BOLD}  ║       SOME TESTS FAILED               ║${NC}"
        echo -e "${RED}${BOLD}  ╚═══════════════════════════════════════╝${NC}"
        echo
        echo -e "  ${YELLOW}Check logs:${NC} $LOG_FILE"
        echo
        return 1
    fi
}

# Main execution
main() {
    cd "$PROJECT_ROOT"

    print_banner

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --parallel)
                PARALLEL=1
                shift
                ;;
            --fail-fast)
                FAIL_FAST=1
                shift
                ;;
            --skip-services)
                SKIP_SERVICES=1
                shift
                ;;
            --coverage-threshold)
                COVERAGE_THRESHOLD=$2
                shift 2
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo
                echo "Options:"
                echo "  --parallel              Run tests in parallel"
                echo "  --fail-fast             Stop on first failure"
                echo "  --skip-services         Skip starting test services"
                echo "  --coverage-threshold N  Set coverage threshold (default: 80)"
                echo "  --help                  Show this help message"
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done

    # Run test pipeline
    check_prerequisites
    setup_environment

    if [[ "${SKIP_SERVICES:-0}" != "1" ]]; then
        start_test_services
    fi

    run_tests
    verify_coverage || true  # Don't fail on coverage
    generate_report || true  # Don't fail on report generation

    # Print summary and exit
    if print_summary; then
        exit 0
    else
        exit 1
    fi
}

# Run main function
main "$@"
