#!/bin/bash
# Performance Test Results Analyzer
# Analyzes test results and generates comprehensive reports with baseline comparison

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
RESULTS_DIR="${1:-$BACKEND_DIR/tests/performance/results/latest}"
BASELINE_FILE="$BACKEND_DIR/tests/performance/baseline.json"

echo -e "${CYAN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║        Performance Test Results Analysis                  ║${NC}"
echo -e "${CYAN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""

if [[ ! -d "$RESULTS_DIR" ]]; then
    echo -e "${RED}✗ Results directory not found: $RESULTS_DIR${NC}"
    exit 1
fi

echo -e "${BLUE}Results Directory:${NC} $RESULTS_DIR"
echo ""

# Check for required tools
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}⚠ jq not installed, installing...${NC}"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        brew install jq
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        sudo apt-get install -y jq
    fi
fi

# Load baseline if exists
BASELINE_EXISTS=false
if [[ -f "$BASELINE_FILE" ]]; then
    BASELINE_EXISTS=true
    echo -e "${GREEN}✓ Baseline file found${NC}"
else
    echo -e "${YELLOW}⚠ No baseline file found - this will be the new baseline${NC}"
fi
echo ""

# Analysis function
analyze_test() {
    local test_name=$1
    local summary_file="$RESULTS_DIR/${test_name}-summary.json"

    if [[ ! -f "$summary_file" ]]; then
        echo -e "${YELLOW}⚠ Summary file not found: $summary_file${NC}"
        return 1
    fi

    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}${test_name^^} Analysis${NC}"
    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    # Extract key metrics
    local p95=$(jq -r '.metrics.http_req_duration.values."p(95)"' "$summary_file" 2>/dev/null || echo "N/A")
    local p99=$(jq -r '.metrics.http_req_duration.values."p(99)"' "$summary_file" 2>/dev/null || echo "N/A")
    local error_rate=$(jq -r '.metrics.http_req_failed.values.rate' "$summary_file" 2>/dev/null || echo "0")
    local total_requests=$(jq -r '.metrics.http_reqs.values.count' "$summary_file" 2>/dev/null || echo "0")
    local rps=$(jq -r '.metrics.http_reqs.values.rate' "$summary_file" 2>/dev/null || echo "0")

    # Display metrics
    echo -e "Performance Metrics:"
    echo -e "  ${BLUE}Response Time p95:${NC} ${p95} ms"
    echo -e "  ${BLUE}Response Time p99:${NC} ${p99} ms"
    echo -e "  ${BLUE}Error Rate:${NC} $(echo "$error_rate * 100" | bc -l | xargs printf "%.4f")%"
    echo -e "  ${BLUE}Total Requests:${NC} $total_requests"
    echo -e "  ${BLUE}Requests/Second:${NC} $(printf "%.2f" "$rps")"
    echo ""

    # Compare with baseline
    if [[ "$BASELINE_EXISTS" == "true" ]]; then
        local baseline_p95=$(jq -r ".${test_name}.p95" "$BASELINE_FILE" 2>/dev/null || echo "0")
        local baseline_p99=$(jq -r ".${test_name}.p99" "$BASELINE_FILE" 2>/dev/null || echo "0")
        local baseline_error=$(jq -r ".${test_name}.error_rate" "$BASELINE_FILE" 2>/dev/null || echo "0")

        if [[ "$baseline_p95" != "null" && "$baseline_p95" != "0" ]]; then
            local p95_change=$(echo "scale=2; (($p95 - $baseline_p95) / $baseline_p95) * 100" | bc -l)
            local p99_change=$(echo "scale=2; (($p99 - $baseline_p99) / $baseline_p99) * 100" | bc -l)
            local error_change=$(echo "scale=2; (($error_rate - $baseline_error) / $baseline_error) * 100" | bc -l)

            echo -e "Baseline Comparison:"

            # p95 comparison
            if (( $(echo "$p95_change > 10" | bc -l) )); then
                echo -e "  ${RED}✗ p95 degraded by ${p95_change}%${NC} (was ${baseline_p95}ms)"
            elif (( $(echo "$p95_change < -10" | bc -l) )); then
                echo -e "  ${GREEN}✓ p95 improved by ${p95_change#-}%${NC} (was ${baseline_p95}ms)"
            else
                echo -e "  ${YELLOW}≈ p95 stable (${p95_change}% change)${NC}"
            fi

            # p99 comparison
            if (( $(echo "$p99_change > 10" | bc -l) )); then
                echo -e "  ${RED}✗ p99 degraded by ${p99_change}%${NC} (was ${baseline_p99}ms)"
            elif (( $(echo "$p99_change < -10" | bc -l) )); then
                echo -e "  ${GREEN}✓ p99 improved by ${p99_change#-}%${NC} (was ${baseline_p99}ms)"
            else
                echo -e "  ${YELLOW}≈ p99 stable (${p99_change}% change)${NC}"
            fi

            # Overall verdict
            echo ""
            if (( $(echo "$p95_change > 10 || $p99_change > 10" | bc -l) )); then
                echo -e "${RED}⚠ PERFORMANCE REGRESSION DETECTED (>10% degradation)${NC}"
                echo -e "${RED}  This build should not be deployed${NC}"
                return 1
            elif (( $(echo "$error_rate > $baseline_error * 1.5" | bc -l) )); then
                echo -e "${RED}⚠ ERROR RATE INCREASED${NC}"
                echo -e "${YELLOW}  Review for stability issues${NC}"
                return 1
            else
                echo -e "${GREEN}✓ Performance within acceptable range${NC}"
            fi
        fi
    fi

    echo ""
    return 0
}

# Analyze test-specific metrics
analyze_custom_metrics() {
    local test_name=$1
    local summary_file="$RESULTS_DIR/${test_name}-summary.json"

    if [[ ! -f "$summary_file" ]]; then
        return 0
    fi

    case "$test_name" in
        "load-test")
            local order_success=$(jq -r '.metrics.order_placement_success.values.rate' "$summary_file" 2>/dev/null || echo "0")
            local order_p95=$(jq -r '.metrics.order_execution_time.values."p(95)"' "$summary_file" 2>/dev/null || echo "N/A")
            local ws_latency=$(jq -r '.metrics.websocket_message_latency.values."p(95)"' "$summary_file" 2>/dev/null || echo "N/A")

            echo -e "Trading-Specific Metrics:"
            echo -e "  ${BLUE}Order Success Rate:${NC} $(echo "$order_success * 100" | bc -l | xargs printf "%.2f")%"
            echo -e "  ${BLUE}Order Execution p95:${NC} ${order_p95} ms ${order_p95 < 50 && echo "${GREEN}✓${NC}" || echo "${YELLOW}⚠${NC}"}"
            echo -e "  ${BLUE}WebSocket Latency p95:${NC} ${ws_latency} ms ${ws_latency < 10 && echo "${GREEN}✓${NC}" || echo "${YELLOW}⚠${NC}"}"
            echo ""
            ;;

        "stress-test")
            local breaking_point=$(jq -r '.breaking_point_vus' "$RESULTS_DIR/stress-test-results.json" 2>/dev/null || echo "N/A")
            local graceful=$(jq -r '.system_behavior.graceful_degradation' "$RESULTS_DIR/stress-test-results.json" 2>/dev/null || echo "false")

            echo -e "Stress Test Analysis:"
            echo -e "  ${BLUE}Breaking Point:${NC} ~${breaking_point} concurrent users"
            echo -e "  ${BLUE}Degradation:${NC} ${graceful == "true" && echo "${GREEN}Graceful ✓${NC}" || echo "${RED}Catastrophic ✗${NC}"}"
            echo ""
            ;;

        "spike-test")
            local recovery=$(jq -r '.metrics.recovery_success_rate' "$RESULTS_DIR/spike-test-results.json" 2>/dev/null || echo "0")
            local spike_errors=$(jq -r '.metrics.spike_errors' "$RESULTS_DIR/spike-test-results.json" 2>/dev/null || echo "0")

            echo -e "Spike Test Analysis:"
            echo -e "  ${BLUE}Recovery Success:${NC} $(echo "$recovery * 100" | bc -l | xargs printf "%.2f")%"
            echo -e "  ${BLUE}Errors During Spikes:${NC} $spike_errors"
            echo ""
            ;;

        "soak-test")
            local memory_leak=$(jq -r '.analysis.memory_leak_detected' "$RESULTS_DIR/soak-test-results.json" 2>/dev/null || echo "false")
            local perf_deg=$(jq -r '.analysis.performance_degradation_detected' "$RESULTS_DIR/soak-test-results.json" 2>/dev/null || echo "false")
            local conn_leaks=$(jq -r '.metrics.connection_leaks' "$RESULTS_DIR/soak-test-results.json" 2>/dev/null || echo "0")

            echo -e "Soak Test Analysis:"
            echo -e "  ${BLUE}Memory Leak:${NC} ${memory_leak == "false" && echo "${GREEN}None ✓${NC}" || echo "${RED}Detected ✗${NC}"}"
            echo -e "  ${BLUE}Performance Degradation:${NC} ${perf_deg == "false" && echo "${GREEN}Stable ✓${NC}" || echo "${RED}Detected ✗${NC}"}"
            echo -e "  ${BLUE}Connection Leaks:${NC} $conn_leaks ${conn_leaks < 10 && echo "${GREEN}✓${NC}" || echo "${RED}✗${NC}"}"
            echo ""
            ;;
    esac
}

# Analyze Go benchmarks
analyze_go_benchmarks() {
    local bench_file="$RESULTS_DIR/go-benchmarks.txt"

    if [[ ! -f "$bench_file" ]]; then
        echo -e "${YELLOW}⚠ Go benchmark results not found${NC}"
        return 0
    fi

    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}GO BENCHMARKS Analysis${NC}"
    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    # Parse key benchmarks
    echo -e "Critical Path Benchmarks:"

    # Order Execution
    local order_exec=$(grep "BenchmarkOrderExecution-" "$bench_file" | head -1 | awk '{print $3}')
    echo -e "  ${BLUE}Order Execution:${NC} $order_exec ns/op ${order_exec < 50000000 && echo "${GREEN}✓${NC}" || echo "${YELLOW}⚠${NC}"}"

    # WebSocket Broadcast
    local ws_broadcast=$(grep "BenchmarkWebSocketBroadcast-" "$bench_file" | head -1 | awk '{print $3}')
    echo -e "  ${BLUE}WebSocket Broadcast:${NC} $ws_broadcast ns/op ${ws_broadcast < 10000000 && echo "${GREEN}✓${NC}" || echo "${YELLOW}⚠${NC}"}"

    # Memory allocations
    echo ""
    echo -e "Memory Allocations:"
    grep "allocs/op" "$bench_file" | head -5 | while read -r line; do
        local benchmark=$(echo "$line" | awk '{print $1}')
        local allocs=$(echo "$line" | awk '{print $5}')
        echo -e "  ${BLUE}$benchmark:${NC} $allocs"
    done

    echo ""
}

# Main analysis
REGRESSION_DETECTED=false

# Analyze each test
for test in "load-test" "stress-test" "spike-test" "soak-test"; do
    if analyze_test "$test"; then
        analyze_custom_metrics "$test"
    else
        REGRESSION_DETECTED=true
    fi
done

# Analyze Go benchmarks
analyze_go_benchmarks

# Generate consolidated report
REPORT_FILE="$RESULTS_DIR/analysis-report.md"

cat > "$REPORT_FILE" << EOF
# Performance Test Analysis Report

**Generated:** $(date)

## Summary

$(if [[ "$REGRESSION_DETECTED" == "true" ]]; then
    echo "⚠️ **PERFORMANCE REGRESSION DETECTED**"
    echo ""
    echo "This build has performance degradation >10% compared to baseline."
    echo "Recommend not deploying until issues are resolved."
else
    echo "✅ **All tests passed**"
    echo ""
    echo "Performance is within acceptable range."
fi)

## Test Results

### Load Test
- **Status:** $(test -f "$RESULTS_DIR/load-test-summary.json" && echo "✅ Passed" || echo "⚠️ Not Run")
- **Key Metrics:** See detailed results below

### Stress Test
- **Status:** $(test -f "$RESULTS_DIR/stress-test-results.json" && echo "✅ Passed" || echo "⚠️ Not Run")
- **Breaking Point:** $(test -f "$RESULTS_DIR/stress-test-results.json" && jq -r '.breaking_point_vus' "$RESULTS_DIR/stress-test-results.json" || echo "N/A") users

### Spike Test
- **Status:** $(test -f "$RESULTS_DIR/spike-test-results.json" && echo "✅ Passed" || echo "⚠️ Not Run")
- **Recovery Rate:** $(test -f "$RESULTS_DIR/spike-test-results.json" && jq -r '.metrics.recovery_success_rate' "$RESULTS_DIR/spike-test-results.json" | awk '{printf "%.2f%%", $1*100}' || echo "N/A")

### Soak Test
- **Status:** $(test -f "$RESULTS_DIR/soak-test-results.json" && echo "✅ Passed" || echo "⚠️ Not Run")
- **Duration:** $(test -f "$RESULTS_DIR/soak-test-results.json" && jq -r '.duration_hours' "$RESULTS_DIR/soak-test-results.json" | awk '{printf "%.2f hours", $1}' || echo "N/A")

## Recommendations

$(if [[ "$REGRESSION_DETECTED" == "true" ]]; then
    echo "1. Investigate performance bottlenecks"
    echo "2. Profile slow endpoints"
    echo "3. Review recent code changes"
    echo "4. Optimize database queries"
else
    echo "1. Continue monitoring production metrics"
    echo "2. Run soak test before major releases"
    echo "3. Update baseline if consistent improvement"
fi)

## Detailed Results

Full test results available in: \`$RESULTS_DIR\`
EOF

echo -e "${GREEN}✓ Analysis report generated:${NC} $REPORT_FILE"
echo ""

# Update baseline if no regression
if [[ "$REGRESSION_DETECTED" == "false" && -f "$RESULTS_DIR/load-test-summary.json" ]]; then
    echo -e "${YELLOW}Update baseline with these results? [y/N]${NC}"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        cat > "$BASELINE_FILE" << EOF
{
  "load-test": {
    "p95": $(jq -r '.metrics.http_req_duration.values."p(95)"' "$RESULTS_DIR/load-test-summary.json"),
    "p99": $(jq -r '.metrics.http_req_duration.values."p(99)"' "$RESULTS_DIR/load-test-summary.json"),
    "error_rate": $(jq -r '.metrics.http_req_failed.values.rate' "$RESULTS_DIR/load-test-summary.json")
  },
  "timestamp": "$(date -Iseconds)"
}
EOF
        echo -e "${GREEN}✓ Baseline updated${NC}"
    fi
fi

echo ""
echo -e "${CYAN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║                 Analysis Complete                         ║${NC}"
echo -e "${CYAN}╚═══════════════════════════════════════════════════════════╝${NC}"

# Exit with error if regression detected
if [[ "$REGRESSION_DETECTED" == "true" ]]; then
    exit 1
else
    exit 0
fi
