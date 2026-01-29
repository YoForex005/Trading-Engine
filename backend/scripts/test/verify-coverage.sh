#!/bin/bash

# Coverage Verification Script
# Ensures code coverage meets minimum threshold

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

# Coverage threshold (can be overridden via environment)
COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD:-80}

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

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Parse coverage data
parse_coverage() {
    local coverage_file="$COVERAGE_DIR/coverage.out"

    if [[ ! -f "$coverage_file" ]]; then
        print_error "Coverage file not found: $coverage_file"
        return 1
    fi

    # Generate detailed function coverage
    go tool cover -func="$coverage_file" > "$COVERAGE_DIR/coverage-detail.txt"

    # Extract total coverage
    local total_coverage=$(go tool cover -func="$coverage_file" | grep total | awk '{print $3}' | sed 's/%//')

    echo "$total_coverage"
}

# Check coverage threshold
check_threshold() {
    local coverage=$1

    print_header "Coverage Verification"

    print_info "Coverage threshold: ${COVERAGE_THRESHOLD}%"
    print_info "Actual coverage: ${coverage}%"

    # Compare using awk for float comparison
    if awk -v cov="$coverage" -v thresh="$COVERAGE_THRESHOLD" 'BEGIN {exit !(cov >= thresh)}'; then
        print_success "Coverage threshold met! ✨"
        return 0
    else
        print_error "Coverage threshold not met"
        print_warning "Required: ${COVERAGE_THRESHOLD}%, Got: ${coverage}%"
        print_warning "Gap: $(awk -v thresh="$COVERAGE_THRESHOLD" -v cov="$coverage" 'BEGIN {printf "%.2f", thresh - cov}')%"
        return 1
    fi
}

# Identify uncovered code
identify_gaps() {
    local coverage_file="$COVERAGE_DIR/coverage.out"

    print_header "Coverage Gaps"

    # Find files with low coverage
    go tool cover -func="$coverage_file" | \
        awk '$3 != "100.0%" && $3 != "total:" {print $1 " : " $3}' | \
        sort -t: -k2 -n | \
        head -20 > "$COVERAGE_DIR/low-coverage-files.txt"

    if [[ -s "$COVERAGE_DIR/low-coverage-files.txt" ]]; then
        print_info "Files with lowest coverage:"
        cat "$COVERAGE_DIR/low-coverage-files.txt" | while read -r line; do
            echo "  $line"
        done
    fi
}

# Generate coverage badge
generate_badge() {
    local coverage=$1
    local color

    if awk -v cov="$coverage" 'BEGIN {exit !(cov >= 80)}'; then
        color="brightgreen"
    elif awk -v cov="$coverage" 'BEGIN {exit !(cov >= 60)}'; then
        color="yellow"
    else
        color="red"
    fi

    # Create simple SVG badge
    cat > "$COVERAGE_DIR/badge.svg" <<EOF
<svg xmlns="http://www.w3.org/2000/svg" width="120" height="20">
    <rect width="70" height="20" fill="#555"/>
    <rect x="70" width="50" height="20" fill="#${color}"/>
    <text x="35" y="15" fill="#fff" font-family="Arial" font-size="11" text-anchor="middle">coverage</text>
    <text x="95" y="15" fill="#fff" font-family="Arial" font-size="11" text-anchor="middle">${coverage}%</text>
</svg>
EOF

    print_info "Coverage badge: $COVERAGE_DIR/badge.svg"
}

# Generate JSON report
generate_json_report() {
    local coverage=$1
    local threshold_met=$2

    cat > "$COVERAGE_DIR/coverage-report.json" <<EOF
{
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "coverage": ${coverage},
  "threshold": ${COVERAGE_THRESHOLD},
  "threshold_met": ${threshold_met},
  "files": {
    "coverage_out": "$COVERAGE_DIR/coverage.out",
    "coverage_html": "$COVERAGE_DIR/coverage.html",
    "coverage_txt": "$COVERAGE_DIR/coverage.txt",
    "low_coverage_files": "$COVERAGE_DIR/low-coverage-files.txt"
  }
}
EOF

    print_info "JSON report: $COVERAGE_DIR/coverage-report.json"
}

main() {
    cd "$PROJECT_ROOT"

    # Parse coverage
    coverage=$(parse_coverage)

    if [[ -z "$coverage" ]]; then
        print_error "Failed to parse coverage data"
        return 1
    fi

    # Check threshold
    local threshold_met=false
    if check_threshold "$coverage"; then
        threshold_met=true
    fi

    # Identify gaps
    identify_gaps

    # Generate badge
    generate_badge "$coverage"

    # Generate JSON report
    generate_json_report "$coverage" "$threshold_met"

    # Return based on threshold
    if [[ "$threshold_met" == "true" ]]; then
        return 0
    else
        return 1
    fi
}

main "$@"
