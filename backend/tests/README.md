

# RTX Trading Engine - Automated Test Suite

Comprehensive automated testing for all API endpoints, WebSocket connections, integration workflows, and load testing WITHOUT human intervention.

## 📋 Table of Contents

- [Overview](#overview)
- [Test Coverage](#test-coverage)
- [Quick Start](#quick-start)
- [Test Suites](#test-suites)
- [Running Tests](#running-tests)
- [CI/CD Integration](#cicd-integration)
- [Test Reports](#test-reports)
- [Performance Benchmarks](#performance-benchmarks)
- [Troubleshooting](#troubleshooting)

## 🎯 Overview

This test suite provides **100% automated testing** for the RTX Trading Engine backend, covering:

- **40+ REST API endpoints** (authentication, orders, positions, market data, admin)
- **WebSocket real-time connections** (subscriptions, streaming, reconnection)
- **Integration workflows** (complete trading lifecycles)
- **Load testing** (100, 1000, 10000 concurrent users)
- **Performance benchmarks** (throughput, latency, resource usage)

**Zero human intervention required** - all tests run automatically via command line or CI/CD pipelines.

## ✅ Test Coverage

### 1. REST API Tests (`api_test.go`)

| Category | Endpoints | Tests |
|----------|-----------|-------|
| **Authentication** | `/login` | 4 tests |
| **Health & Config** | `/health`, `/api/config` | 2 tests |
| **Orders** | `/order/*`, `/orders/*` | 12 tests |
| **Positions** | `/position/*`, `/positions/*` | 6 tests |
| **Market Data** | `/ticks`, `/ohlc` | 4 tests |
| **Risk Management** | `/risk/*` | 4 tests |
| **Admin** | `/admin/*` | 3 tests |
| **Concurrent** | All endpoints | 2 tests |
| **Benchmarks** | Critical paths | 3 benchmarks |

**Total: 40+ endpoint tests**

### 2. WebSocket Tests (`websocket_test.go`)

- Connection tests (single, multiple, reconnect)
- Subscription management (subscribe, unsubscribe, multi-symbol)
- Real-time data streaming (high-frequency ticks)
- Error handling (invalid messages, unknown types)
- Performance tests (concurrent clients, throughput)
- Benchmarks (broadcast, subscribe/unsubscribe)

**Total: 15+ WebSocket tests**

### 3. Integration Tests (`integration_flow_test.go`)

Complete trading workflows:

- **Complete Trading Workflow**: Login → Market Data → Orders → Positions → Risk Management
- **Position Lifecycle**: Open → Monitor → Modify → Trailing Stop → Close
- **Multi-Symbol Trading**: Concurrent trading across 5+ symbols
- **Order Management**: Pending orders, cancellations, bulk operations
- **Risk Management**: Lot calculation, margin preview, exposure monitoring
- **Real-Time Data Flow**: Continuous price updates, tick queries
- **Admin Operations**: Config management, execution mode toggle

**Total: 8 integration workflows**

### 4. Load Tests (`load_test.go`)

Stress testing at scale:

- **100 concurrent users** (realistic load)
- **1,000 concurrent users** (high load)
- **10,000 concurrent users** (stress test)
- **High-frequency tick streaming** (1000+ ticks/sec)
- **Mixed operations** (realistic workload simulation)
- **Database load** (query performance)
- **Memory usage** (resource consumption)
- **30-second stress test** (sustained load)

**Performance metrics tracked:**
- Total requests / Success rate / Failed requests
- Min / Max / Avg latency
- P50 / P95 / P99 percentiles
- Requests per second (RPS)
- Memory and CPU usage

## 🚀 Quick Start

### Prerequisites

```bash
# Ensure Go is installed
go version  # Should be 1.19+

# Navigate to backend directory
cd backend
```

### Run All Tests (Automated)

```bash
# Run comprehensive test suite
./scripts/test/run-api-tests.sh

# Quick mode (skip load tests)
./scripts/test/run-api-tests.sh --quick

# Generate HTML + coverage reports
./scripts/test/run-api-tests.sh --coverage --html

# CI/CD mode (fail fast)
./scripts/test/run-api-tests.sh --ci --json
```

### Run Specific Test Suites

```bash
# API unit tests only
go test -v -run "Test(Auth|Order|Position)" ./tests/

# WebSocket tests only
go test -v -run "TestWS_" ./tests/

# Integration workflows only
go test -v -run "TestWorkflow_" ./tests/

# Load tests only (skip in short mode)
go test -v -timeout 30m -run "TestLoad_" ./tests/

# Benchmarks only
go test -bench=. -benchmem ./tests/
```

## 📊 Test Suites

### 1. API Unit Tests

```bash
# Run all API tests
go test -v ./tests/ -run "Test(Auth|Order|Position|Market|Risk|Admin)"

# Run with coverage
go test -v -coverprofile=coverage.out ./tests/

# View coverage
go tool cover -html=coverage.out
```

**Example output:**
```
=== RUN   TestAuth_Login_Success
--- PASS: TestAuth_Login_Success (0.05s)
=== RUN   TestOrder_PlaceMarketOrder_Buy
--- PASS: TestOrder_PlaceMarketOrder_Buy (0.12s)
=== RUN   TestMarket_GetTicks
--- PASS: TestMarket_GetTicks (0.08s)
...
PASS
coverage: 78.3% of statements
```

### 2. WebSocket Tests

```bash
# Run WebSocket suite
go test -v ./tests/ -run "TestWS_"

# Test specific scenarios
go test -v ./tests/ -run "TestWS_Subscribe_MultipleSymbols"
go test -v ./tests/ -run "TestWS_HighFrequencyTicks"
```

**Example output:**
```
=== RUN   TestWS_Connect
--- PASS: TestWS_Connect (0.02s)
=== RUN   TestWS_Subscribe_SingleSymbol
--- PASS: TestWS_Subscribe_SingleSymbol (0.15s)
=== RUN   TestWS_RealTimeTickStream
    websocket_test.go:245: Successfully received 20 high-frequency ticks
--- PASS: TestWS_RealTimeTickStream (1.05s)
```

### 3. Integration Workflows

```bash
# Run all workflows
go test -v ./tests/ -run "TestWorkflow_"

# Run specific workflow
go test -v ./tests/ -run "TestWorkflow_CompleteTrading"
```

**Example output:**
```
=== RUN   TestWorkflow_CompleteTrading
    integration_flow_test.go:25: Step 1: Login
    integration_flow_test.go:30: Step 2: Get account summary
    integration_flow_test.go:33: Step 3: Inject market data
    integration_flow_test.go:39: Step 4: Place market orders
    integration_flow_test.go:64: Placed BUY EURUSD 0.10 lots
    integration_flow_test.go:64: Placed SELL GBPUSD 0.20 lots
    ...
    integration_flow_test.go:152: Complete trading workflow test passed
--- PASS: TestWorkflow_CompleteTrading (2.34s)
```

### 4. Load Tests

```bash
# Run load tests (may take 5-30 minutes)
go test -v -timeout 30m ./tests/ -run "TestLoad_"

# Run specific load test
go test -v -timeout 10m ./tests/ -run "TestLoad_PlaceOrders_100Concurrent"

# Skip load tests (use -short flag)
go test -v -short ./tests/
```

**Example output:**
```
=== RUN   TestLoad_PlaceOrders_100Concurrent
    load_test.go:180:
    === Load Test Results ===
    Total Requests:    1000
    Successful:        998 (99.80%)
    Failed:            2 (0.20%)
    Duration:          5.234s
    Requests/sec:      191.05

    Latency Statistics:
      Min:             2ms
      Max:             156ms
      Avg:             45ms
      P50:             42ms
      P95:             89ms
      P99:             123ms
--- PASS: TestLoad_PlaceOrders_100Concurrent (5.23s)
```

## 🔧 Running Tests

### Command Line Options

```bash
# Automated test script
./scripts/test/run-api-tests.sh [OPTIONS]

Options:
  --quick          Skip load tests (fast)
  --load-only      Run only load tests
  --verbose        Detailed output
  --coverage       Generate coverage report
  --html           Generate HTML reports
  --json           Generate JSON output
  --ci             CI/CD mode (fail fast)
  --bench          Run benchmarks
  --help           Show help
```

### Examples

```bash
# Quick smoke test (1-2 minutes)
./scripts/test/run-api-tests.sh --quick

# Full test suite with reports (10-30 minutes)
./scripts/test/run-api-tests.sh --coverage --html --json

# Load tests only
./scripts/test/run-api-tests.sh --load-only

# CI/CD pipeline
./scripts/test/run-api-tests.sh --ci --json --coverage
```

### Manual Go Test Commands

```bash
# All tests with verbose output
go test -v ./tests/

# Specific test pattern
go test -v -run "TestOrder_" ./tests/

# With coverage
go test -v -coverprofile=coverage.out ./tests/
go tool cover -html=coverage.out -o coverage.html

# JSON output
go test -json ./tests/ > test_results.json

# Parallel execution (faster)
go test -v -parallel 8 ./tests/

# Timeout for long-running tests
go test -v -timeout 30m ./tests/

# Skip long tests
go test -v -short ./tests/
```

## 🔄 CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/api-tests.yml
name: API Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run tests
        run: |
          cd backend
          ./scripts/test/run-api-tests.sh --ci --coverage --json

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./backend/tests/reports/coverage.out

      - name: Upload test results
        uses: actions/upload-artifact@v3
        with:
          name: test-results
          path: ./backend/tests/reports/
```

### GitLab CI

```yaml
# .gitlab-ci.yml
test:
  image: golang:1.21
  script:
    - cd backend
    - ./scripts/test/run-api-tests.sh --ci --coverage --json
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: backend/tests/reports/coverage.xml
    paths:
      - backend/tests/reports/
```

### Jenkins

```groovy
// Jenkinsfile
pipeline {
    agent any
    stages {
        stage('Test') {
            steps {
                sh '''
                    cd backend
                    ./scripts/test/run-api-tests.sh --ci --coverage --json
                '''
            }
        }
        stage('Reports') {
            steps {
                publishHTML([
                    reportDir: 'backend/tests/reports',
                    reportFiles: 'test_summary.html',
                    reportName: 'Test Report'
                ])
            }
        }
    }
}
```

## 📈 Test Reports

### Generated Reports

After running tests with `--html --json --coverage`, you'll find:

```
backend/tests/reports/
├── api_tests.log              # API test output
├── websocket_tests.log        # WebSocket test output
├── integration_tests.log      # Integration test output
├── load_tests.log             # Load test output
├── benchmark.log              # Benchmark results
├── coverage.out               # Coverage data
├── coverage.txt               # Coverage summary
├── coverage.html              # Interactive coverage report
├── test_results.json          # JSON test results
└── test_summary.html          # HTML summary dashboard
```

### Viewing Reports

```bash
# Open HTML summary
open backend/tests/reports/test_summary.html

# Open coverage report
open backend/tests/reports/coverage.html

# View coverage in terminal
go tool cover -func=backend/tests/reports/coverage.out

# Parse JSON results
cat backend/tests/reports/test_results.json | jq '.[] | select(.Action == "pass")'
```

### Coverage Metrics

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./tests/
go tool cover -func=coverage.out

# Example output:
# github.com/epic1st/rtx/backend/api/server.go:42:    HandleLogin          100.0%
# github.com/epic1st/rtx/backend/api/server.go:94:    HandlePlaceOrder     95.2%
# github.com/epic1st/rtx/backend/api/server.go:131:   HandleGetTicks       88.7%
# total:                                              (statements)         78.3%
```

## ⚡ Performance Benchmarks

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./tests/

# Specific benchmark
go test -bench=BenchmarkPlaceMarketOrder -benchmem ./tests/

# With CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./tests/
go tool pprof cpu.prof

# With memory profiling
go test -bench=. -memprofile=mem.prof ./tests/
go tool pprof mem.prof
```

### Benchmark Output

```
BenchmarkPlaceMarketOrder-8           5000    245632 ns/op    12456 B/op    145 allocs/op
BenchmarkGetTicks-8                  10000    156234 ns/op     8924 B/op     98 allocs/op
BenchmarkCalculateLot-8              20000     78456 ns/op     4512 B/op     56 allocs/op
BenchmarkWS_TickBroadcast-8          50000     34567 ns/op     2345 B/op     23 allocs/op
BenchmarkLoad_PlaceOrder-8           3000    389234 ns/op    15678 B/op    178 allocs/op
```

### Performance Targets

| Operation | Target | P95 | P99 |
|-----------|--------|-----|-----|
| Place Order | < 50ms | < 100ms | < 200ms |
| Get Ticks | < 30ms | < 60ms | < 100ms |
| Calculate Risk | < 20ms | < 40ms | < 80ms |
| WebSocket Broadcast | < 10ms | < 20ms | < 50ms |
| Concurrent (100 users) | > 100 RPS | < 150ms | < 250ms |
| Concurrent (1000 users) | > 500 RPS | < 300ms | < 500ms |

## 🐛 Troubleshooting

### Common Issues

#### Tests Fail to Start

```bash
# Check Go installation
go version

# Ensure dependencies are installed
cd backend
go mod download
go mod tidy

# Clean build cache
go clean -testcache
```

#### Port Already in Use

```bash
# Find process using port 7999
lsof -i :7999

# Kill the process
kill -9 <PID>

# Or change test port in test setup
```

#### WebSocket Connection Timeout

```bash
# Increase timeout in test
# Edit websocket_test.go, increase timeout values

# Check firewall settings
sudo ufw status
```

#### Load Tests OOM (Out of Memory)

```bash
# Reduce concurrent users
# Edit load_test.go, decrease numConcurrent

# Increase memory limit
GOMEMLIMIT=4GiB go test -v -run TestLoad_ ./tests/

# Run load tests individually
go test -v -run TestLoad_PlaceOrders_100Concurrent ./tests/
```

### Debug Mode

```bash
# Run with verbose logging
go test -v -run TestOrder_PlaceMarketOrder ./tests/

# Print additional debug info
DEBUG=true go test -v ./tests/

# Run single test
go test -v -run "^TestAuth_Login_Success$" ./tests/
```

### Test Isolation Issues

```bash
# Clean test cache
go clean -testcache

# Run tests sequentially (no parallel)
go test -v -parallel 1 ./tests/

# Run with fresh environment
go test -v -count=1 ./tests/
```

## 📝 Writing New Tests

### Test Structure

```go
func TestYourFeature(t *testing.T) {
    // Setup
    tc := SetupTest(t)
    tc.InjectPrice("EURUSD", 1.10000, 1.10020)

    // Execute
    reqBody := map[string]interface{}{
        "symbol": "EURUSD",
        "side":   "BUY",
        "volume": 0.1,
    }
    body, _ := json.Marshal(reqBody)

    req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    tc.Server.HandlePlaceOrder(w, req)

    // Assert
    if w.Code != http.StatusOK {
        t.Errorf("Expected 200, got %d", w.Code)
    }

    var resp map[string]interface{}
    json.NewDecoder(w.Body).Decode(&resp)

    if !resp["success"].(bool) {
        t.Error("Expected success to be true")
    }
}
```

### Best Practices

1. **Use descriptive test names**: `TestOrder_PlaceMarketOrder_InvalidVolume`
2. **Clean up resources**: Always defer cleanup
3. **Test edge cases**: Null values, invalid inputs, boundary conditions
4. **Use table-driven tests**: For multiple similar test cases
5. **Mock external dependencies**: Don't rely on real LP connections
6. **Measure performance**: Add benchmarks for critical paths
7. **Document expected behavior**: Add comments for complex tests

## 📚 Additional Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [httptest Package](https://pkg.go.dev/net/http/httptest)
- [WebSocket Testing](https://pkg.go.dev/github.com/gorilla/websocket)
- [Benchmark Guide](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)

## 🎯 Summary

This automated test suite provides:

- ✅ **100% automation** - No manual intervention required
- ✅ **Comprehensive coverage** - 40+ endpoints, WebSocket, workflows
- ✅ **Load testing** - Up to 10,000 concurrent users
- ✅ **CI/CD ready** - JSON/HTML reports, coverage, benchmarks
- ✅ **Performance metrics** - Latency, throughput, resource usage
- ✅ **Easy to run** - Single command execution

**Run all tests:**
```bash
./scripts/test/run-api-tests.sh --coverage --html
```

**For CI/CD:**
```bash
./scripts/test/run-api-tests.sh --ci --json --coverage
```
