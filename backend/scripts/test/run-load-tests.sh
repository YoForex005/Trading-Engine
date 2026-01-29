#!/bin/bash

# Load Tests Runner
# Runs performance and load testing with benchmarks

set -e
set -o pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m'
BOLD='\033[1m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
REPORT_DIR="$PROJECT_ROOT/test-reports"
BENCHMARK_DIR="$REPORT_DIR/benchmarks"

# Performance thresholds
TICK_THROUGHPUT_MIN=50000  # ticks/sec
ORDER_LATENCY_MAX=10       # milliseconds
MEMORY_LIMIT_MB=500        # MB

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

# Run Go benchmarks
run_benchmarks() {
    print_info "Running Go benchmarks..."

    mkdir -p "$BENCHMARK_DIR"

    # Run all benchmarks
    go test -bench=. -benchmem -benchtime=5s \
        -cpuprofile="$BENCHMARK_DIR/cpu.prof" \
        -memprofile="$BENCHMARK_DIR/mem.prof" \
        ./datapipeline/... \
        2>&1 | tee "$BENCHMARK_DIR/benchmark.log"

    local exit_code=${PIPESTATUS[0]}

    # Parse results
    if [[ -f "$BENCHMARK_DIR/benchmark.log" ]]; then
        print_info "Benchmark results:"

        # Extract key metrics
        grep "Benchmark" "$BENCHMARK_DIR/benchmark.log" | while read -r line; do
            echo "  $line"
        done
    fi

    return $exit_code
}

# Verify performance thresholds
verify_thresholds() {
    print_header "Verifying Performance Thresholds"

    local failed=0

    # Check tick throughput
    if [[ -f "$BENCHMARK_DIR/benchmark.log" ]]; then
        local throughput=$(grep "BenchmarkTickIngestion" "$BENCHMARK_DIR/benchmark.log" | awk '{print $3}' | head -1)

        if [[ -n "$throughput" ]]; then
            # Remove any non-numeric suffix
            throughput=$(echo "$throughput" | sed 's/[^0-9]//g')

            if [[ $throughput -ge $TICK_THROUGHPUT_MIN ]]; then
                print_success "Tick throughput: ${throughput} ops/sec (≥ ${TICK_THROUGHPUT_MIN})"
            else
                print_error "Tick throughput: ${throughput} ops/sec (< ${TICK_THROUGHPUT_MIN})"
                ((failed++))
            fi
        else
            print_warning "Tick throughput metric not found"
        fi
    fi

    # Check memory usage
    if [[ -f "$BENCHMARK_DIR/benchmark.log" ]]; then
        local mem_alloc=$(grep "B/op" "$BENCHMARK_DIR/benchmark.log" | awk '{print $5}' | head -1)

        if [[ -n "$mem_alloc" ]]; then
            print_info "Memory allocation: ${mem_alloc} B/op"
        fi
    fi

    return $failed
}

# Generate performance report
generate_perf_report() {
    print_info "Generating performance report..."

    cat > "$BENCHMARK_DIR/report.html" <<EOF
<!DOCTYPE html>
<html>
<head>
    <title>RTX Trading Engine - Performance Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #2c3e50; border-bottom: 3px solid #3498db; padding-bottom: 10px; }
        h2 { color: #34495e; margin-top: 30px; }
        .metric { background: #ecf0f1; padding: 15px; margin: 10px 0; border-radius: 4px; }
        .metric-name { font-weight: bold; color: #2c3e50; }
        .metric-value { color: #27ae60; font-size: 1.2em; }
        .threshold { color: #7f8c8d; font-size: 0.9em; }
        .pass { color: #27ae60; }
        .fail { color: #e74c3c; }
        pre { background: #2c3e50; color: #ecf0f1; padding: 15px; border-radius: 4px; overflow-x: auto; }
        .timestamp { color: #7f8c8d; font-size: 0.9em; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 RTX Trading Engine - Performance Report</h1>
        <p class="timestamp">Generated: $(date)</p>

        <h2>📊 Benchmark Results</h2>
        <pre>$(cat "$BENCHMARK_DIR/benchmark.log" 2>/dev/null || echo "No benchmark data available")</pre>

        <h2>⚡ Performance Thresholds</h2>
        <div class="metric">
            <div class="metric-name">Tick Throughput</div>
            <div class="threshold">Minimum: ${TICK_THROUGHPUT_MIN} ticks/sec</div>
        </div>

        <div class="metric">
            <div class="metric-name">Order Latency</div>
            <div class="threshold">Maximum: ${ORDER_LATENCY_MAX} ms</div>
        </div>

        <div class="metric">
            <div class="metric-name">Memory Limit</div>
            <div class="threshold">Maximum: ${MEMORY_LIMIT_MB} MB</div>
        </div>

        <h2>📈 CPU Profile</h2>
        <p>CPU profile saved to: <code>$BENCHMARK_DIR/cpu.prof</code></p>
        <p>View with: <code>go tool pprof $BENCHMARK_DIR/cpu.prof</code></p>

        <h2>💾 Memory Profile</h2>
        <p>Memory profile saved to: <code>$BENCHMARK_DIR/mem.prof</code></p>
        <p>View with: <code>go tool pprof $BENCHMARK_DIR/mem.prof</code></p>
    </div>
</body>
</html>
EOF

    print_success "Performance report: $BENCHMARK_DIR/report.html"
}

main() {
    cd "$PROJECT_ROOT"

    print_header "Load & Performance Tests"

    mkdir -p "$BENCHMARK_DIR"

    # Run benchmarks
    if run_benchmarks; then
        print_success "Benchmarks completed"
    else
        print_error "Benchmarks failed"
        return 1
    fi

    # Verify thresholds
    if verify_thresholds; then
        print_success "All performance thresholds met"
    else
        print_warning "Some performance thresholds not met"
    fi

    # Generate report
    generate_perf_report

    print_success "Load tests completed"
    return 0
}

main "$@"
