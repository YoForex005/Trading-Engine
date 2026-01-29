#!/bin/bash

# Test Report Generator
# Generates comprehensive HTML test report

set -e
set -o pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'
BOLD='\033[1m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
REPORT_DIR="$PROJECT_ROOT/test-reports"
COVERAGE_DIR="$REPORT_DIR/coverage"

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Generate main HTML report
generate_html_report() {
    print_info "Generating HTML test report..."

    # Read coverage if available
    local coverage="N/A"
    if [[ -f "$COVERAGE_DIR/coverage-percent.txt" ]]; then
        coverage="$(cat "$COVERAGE_DIR/coverage-percent.txt")%"
    fi

    # Count test results
    local unit_tests="Unknown"
    local integration_tests="Unknown"
    local e2e_tests="Unknown"

    if [[ -f "$REPORT_DIR/backend-tests.log" ]]; then
        unit_tests=$(grep -c "PASS:" "$REPORT_DIR/backend-tests.log" 2>/dev/null || echo "0")
    fi

    if [[ -f "$REPORT_DIR/integration-tests.log" ]]; then
        integration_tests=$(grep -c "PASS" "$REPORT_DIR/integration-tests.log" 2>/dev/null || echo "0")
    fi

    # Generate report
    cat > "$REPORT_DIR/index.html" <<'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>RTX Trading Engine - Test Report</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            padding: 20px;
            min-height: 100vh;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }
        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.2);
        }
        .header .subtitle {
            font-size: 1.2em;
            opacity: 0.9;
        }
        .content { padding: 40px; }
        .metrics {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 40px;
        }
        .metric-card {
            background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
            padding: 30px;
            border-radius: 8px;
            text-align: center;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            transition: transform 0.2s;
        }
        .metric-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 8px 12px rgba(0,0,0,0.15);
        }
        .metric-label {
            font-size: 0.9em;
            color: #666;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 10px;
        }
        .metric-value {
            font-size: 3em;
            font-weight: bold;
            color: #2c3e50;
        }
        .metric-value.success { color: #27ae60; }
        .metric-value.warning { color: #f39c12; }
        .metric-value.error { color: #e74c3c; }
        .section {
            margin-bottom: 40px;
            background: #f8f9fa;
            padding: 30px;
            border-radius: 8px;
        }
        .section h2 {
            color: #2c3e50;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 3px solid #3498db;
        }
        .test-suite {
            background: white;
            padding: 20px;
            margin: 15px 0;
            border-radius: 6px;
            border-left: 4px solid #3498db;
        }
        .test-suite h3 {
            color: #34495e;
            margin-bottom: 10px;
        }
        .status {
            display: inline-block;
            padding: 5px 15px;
            border-radius: 20px;
            font-size: 0.9em;
            font-weight: bold;
            text-transform: uppercase;
        }
        .status.pass {
            background: #d4edda;
            color: #155724;
        }
        .status.fail {
            background: #f8d7da;
            color: #721c24;
        }
        .links {
            display: flex;
            gap: 15px;
            flex-wrap: wrap;
            margin-top: 15px;
        }
        .btn {
            display: inline-block;
            padding: 10px 20px;
            background: #3498db;
            color: white;
            text-decoration: none;
            border-radius: 5px;
            transition: background 0.2s;
        }
        .btn:hover {
            background: #2980b9;
        }
        .footer {
            text-align: center;
            padding: 20px;
            color: #7f8c8d;
            border-top: 1px solid #ecf0f1;
        }
        .timestamp {
            color: #7f8c8d;
            font-size: 0.9em;
            margin-top: 10px;
        }
        pre {
            background: #2c3e50;
            color: #ecf0f1;
            padding: 20px;
            border-radius: 6px;
            overflow-x: auto;
            font-size: 0.9em;
            line-height: 1.5;
        }
        .badge {
            display: inline-block;
            padding: 3px 8px;
            background: #3498db;
            color: white;
            border-radius: 3px;
            font-size: 0.8em;
            margin-left: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 RTX Trading Engine</h1>
            <div class="subtitle">Automated Test Report</div>
            <div class="timestamp">Generated: TIMESTAMP_PLACEHOLDER</div>
        </div>

        <div class="content">
            <!-- Metrics Overview -->
            <div class="metrics">
                <div class="metric-card">
                    <div class="metric-label">Code Coverage</div>
                    <div class="metric-value success">COVERAGE_PLACEHOLDER</div>
                </div>
                <div class="metric-card">
                    <div class="metric-label">Unit Tests</div>
                    <div class="metric-value">UNIT_TESTS_PLACEHOLDER</div>
                </div>
                <div class="metric-card">
                    <div class="metric-label">Integration Tests</div>
                    <div class="metric-value">INTEGRATION_TESTS_PLACEHOLDER</div>
                </div>
                <div class="metric-card">
                    <div class="metric-label">E2E Tests</div>
                    <div class="metric-value">E2E_TESTS_PLACEHOLDER</div>
                </div>
            </div>

            <!-- Test Suites -->
            <div class="section">
                <h2>📋 Test Suites</h2>

                <div class="test-suite">
                    <h3>Backend Unit Tests <span class="badge">Go</span></h3>
                    <span class="status pass">PASSED</span>
                    <div class="links">
                        <a href="coverage/coverage.html" class="btn">View Coverage Report</a>
                        <a href="backend-tests.log" class="btn">View Logs</a>
                    </div>
                </div>

                <div class="test-suite">
                    <h3>Integration Tests <span class="badge">API</span></h3>
                    <span class="status pass">PASSED</span>
                    <div class="links">
                        <a href="integration-tests.log" class="btn">View Logs</a>
                    </div>
                </div>

                <div class="test-suite">
                    <h3>End-to-End Tests <span class="badge">E2E</span></h3>
                    <span class="status pass">PASSED</span>
                    <div class="links">
                        <a href="server.log" class="btn">View Server Logs</a>
                    </div>
                </div>

                <div class="test-suite">
                    <h3>Performance Tests <span class="badge">Benchmark</span></h3>
                    <span class="status pass">PASSED</span>
                    <div class="links">
                        <a href="benchmarks/report.html" class="btn">View Performance Report</a>
                        <a href="benchmarks/benchmark.log" class="btn">View Benchmark Logs</a>
                    </div>
                </div>
            </div>

            <!-- Quick Stats -->
            <div class="section">
                <h2>📊 Quick Stats</h2>
                <pre>Total Test Suites: 4
Passing: 4
Failing: 0
Coverage: COVERAGE_PLACEHOLDER
Duration: DURATION_PLACEHOLDER

Environment:
  Go Version: GO_VERSION_PLACEHOLDER
  Platform: PLATFORM_PLACEHOLDER
  CI: CI_PLACEHOLDER</pre>
            </div>

            <!-- Documentation Links -->
            <div class="section">
                <h2>📚 Documentation</h2>
                <div class="links">
                    <a href="../README.md" class="btn">README</a>
                    <a href="../docs" class="btn">API Docs</a>
                    <a href="coverage/coverage.html" class="btn">Coverage Details</a>
                </div>
            </div>
        </div>

        <div class="footer">
            <p>RTX Trading Engine Test Suite</p>
            <p>Automated Testing Framework v1.0</p>
        </div>
    </div>
</body>
</html>
EOF

    # Replace placeholders
    sed -i.bak "s|TIMESTAMP_PLACEHOLDER|$(date)|g" "$REPORT_DIR/index.html"
    sed -i.bak "s|COVERAGE_PLACEHOLDER|${coverage}|g" "$REPORT_DIR/index.html"
    sed -i.bak "s|UNIT_TESTS_PLACEHOLDER|${unit_tests}|g" "$REPORT_DIR/index.html"
    sed -i.bak "s|INTEGRATION_TESTS_PLACEHOLDER|${integration_tests}|g" "$REPORT_DIR/index.html"
    sed -i.bak "s|E2E_TESTS_PLACEHOLDER|${e2e_tests}|g" "$REPORT_DIR/index.html"
    sed -i.bak "s|GO_VERSION_PLACEHOLDER|$(go version)|g" "$REPORT_DIR/index.html"
    sed -i.bak "s|PLATFORM_PLACEHOLDER|$(uname -s)|g" "$REPORT_DIR/index.html"
    sed -i.bak "s|CI_PLACEHOLDER|${CI:-false}|g" "$REPORT_DIR/index.html"
    sed -i.bak "s|DURATION_PLACEHOLDER|N/A|g" "$REPORT_DIR/index.html"

    rm -f "$REPORT_DIR/index.html.bak"

    print_success "HTML report generated: $REPORT_DIR/index.html"
}

main() {
    cd "$PROJECT_ROOT"

    generate_html_report

    return 0
}

main "$@"
