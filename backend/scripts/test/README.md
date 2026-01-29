# RTX Trading Engine - Automated Test Suite

Comprehensive automated testing framework with **ZERO human intervention** required.

## 🚀 Quick Start

Run all tests with a single command:

```bash
cd backend
./scripts/test/run-all-tests.sh
```

That's it! The script will:
- ✅ Check prerequisites
- ✅ Start test services (Redis, mock servers)
- ✅ Run all test suites
- ✅ Verify coverage thresholds
- ✅ Generate HTML reports
- ✅ Clean up automatically

## 📋 Available Test Scripts

### 1. Master Test Runner
```bash
./scripts/test/run-all-tests.sh [OPTIONS]
```

**Options:**
- `--parallel` - Run test suites in parallel
- `--fail-fast` - Stop on first failure
- `--skip-services` - Skip starting Docker test services
- `--coverage-threshold N` - Set coverage threshold (default: 80)
- `--help` - Show help message

**Examples:**
```bash
# Run all tests with default settings
./scripts/test/run-all-tests.sh

# Run tests in parallel mode
./scripts/test/run-all-tests.sh --parallel

# Stop on first failure with custom coverage threshold
./scripts/test/run-all-tests.sh --fail-fast --coverage-threshold 85

# Skip Docker services (for CI environments)
./scripts/test/run-all-tests.sh --skip-services
```

### 2. Backend Unit Tests
```bash
./scripts/test/run-backend-tests.sh
```

Runs Go unit tests with coverage reporting:
- Race detection enabled
- Coverage reports (HTML, text, JSON)
- Atomic coverage mode
- 5-minute timeout

### 3. Integration Tests
```bash
./scripts/test/run-integration-tests.sh
```

Runs API integration tests:
- Redis integration
- HTTP API endpoints
- WebSocket connections
- Authentication flows
- Order placement workflows

### 4. End-to-End Tests
```bash
./scripts/test/run-e2e-tests.sh
```

Runs complete workflow tests:
- Full server startup
- Health checks
- Login flows
- Market order placement
- Position management
- WebSocket streaming

### 5. Load & Performance Tests
```bash
./scripts/test/run-load-tests.sh
```

Runs performance benchmarks:
- Tick ingestion throughput
- OHLC generation performance
- Order placement latency
- Memory profiling
- CPU profiling

### 6. Coverage Verification
```bash
./scripts/test/verify-coverage.sh
```

Verifies coverage meets threshold:
- Parses coverage data
- Identifies coverage gaps
- Generates coverage badge
- Creates JSON report
- Exits with error if threshold not met

### 7. Test Report Generator
```bash
./scripts/test/generate-test-report.sh
```

Generates beautiful HTML test report:
- Coverage metrics
- Test suite results
- Performance benchmarks
- Links to detailed reports

## 📊 Test Reports

After running tests, reports are generated in `test-reports/`:

```
test-reports/
├── index.html              # Main test report (open in browser)
├── coverage/
│   ├── coverage.html       # Interactive coverage report
│   ├── coverage.out        # Go coverage data
│   ├── coverage.txt        # Coverage by function
│   ├── badge.svg          # Coverage badge
│   └── coverage-report.json
├── benchmarks/
│   ├── report.html        # Performance report
│   ├── benchmark.log      # Benchmark results
│   ├── cpu.prof          # CPU profile
│   └── mem.prof          # Memory profile
├── backend-tests.log      # Unit test logs
├── integration-tests.log  # Integration test logs
└── test-run.log          # Master test log
```

**View main report:**
```bash
open test-reports/index.html
```

## 🎯 Performance Thresholds

Default performance requirements:

| Metric | Threshold |
|--------|-----------|
| Code Coverage | ≥ 80% |
| Tick Throughput | ≥ 50,000 ticks/sec |
| Order Latency | ≤ 10ms |
| Memory Limit | ≤ 500MB |

## 🐳 Docker Test Services

Test services are managed via `docker-compose.test.yml`:

```bash
# Start services manually
docker-compose -f docker-compose.test.yml up -d

# Stop services
docker-compose -f docker-compose.test.yml down

# View logs
docker-compose -f docker-compose.test.yml logs -f
```

Services included:
- **Redis** - In-memory data store (port 6379)
- **Mock FIX Server** - FIX protocol testing

## 🔄 CI/CD Integration

### GitHub Actions

Tests run automatically on:
- ✅ Push to `main` or `develop`
- ✅ Pull requests
- ✅ Daily schedule (2 AM UTC)

**Workflow file:** `.github/workflows/tests.yml`

**Artifacts uploaded:**
- Test results
- Coverage reports
- Benchmark results
- Security scan results

### Local CI Simulation

Simulate CI environment locally:

```bash
CI=true REDIS_URL=redis://localhost:6379 ./scripts/test/run-all-tests.sh --skip-services
```

## 🧪 Writing Tests

### Unit Test Example

```go
// backend/mypackage/service_test.go
package mypackage

import "testing"

func TestMyFunction(t *testing.T) {
    result := MyFunction(42)

    if result != "expected" {
        t.Errorf("Expected 'expected', got '%s'", result)
    }
}

// Benchmark example
func BenchmarkMyFunction(b *testing.B) {
    for i := 0; i < b.N; i++ {
        MyFunction(42)
    }
}
```

### Integration Test Example

```go
// backend/tests/integration/api_test.go
// +build integration

package integration

import "testing"

func TestAPIEndpoint(t *testing.T) {
    ts := SetupTestServer(t)
    defer ts.Cleanup()

    // Test logic here
}
```

## 🔍 Troubleshooting

### Tests Failing Locally

1. **Check prerequisites:**
   ```bash
   go version  # Should be 1.24.0+
   docker --version
   redis-cli --version
   ```

2. **Clear test cache:**
   ```bash
   go clean -testcache
   rm -rf test-reports/
   ```

3. **Restart test services:**
   ```bash
   docker-compose -f docker-compose.test.yml down
   docker-compose -f docker-compose.test.yml up -d
   ```

4. **Check logs:**
   ```bash
   cat test-reports/test-run.log
   ```

### Coverage Too Low

1. **View coverage report:**
   ```bash
   open test-reports/coverage/coverage.html
   ```

2. **Identify gaps:**
   ```bash
   cat test-reports/coverage/low-coverage-files.txt
   ```

3. **Write tests for uncovered code**

### Performance Tests Failing

1. **View benchmark results:**
   ```bash
   open test-reports/benchmarks/report.html
   ```

2. **Profile CPU usage:**
   ```bash
   go tool pprof test-reports/benchmarks/cpu.prof
   ```

3. **Profile memory:**
   ```bash
   go tool pprof test-reports/benchmarks/mem.prof
   ```

## 🎨 Colorful Output

Tests provide colorful terminal output:

- 🟢 **Green** - Success
- 🔴 **Red** - Failure
- 🟡 **Yellow** - Warning
- 🔵 **Blue** - Info

## 📈 Best Practices

1. **Run tests before commits:**
   ```bash
   ./scripts/test/run-all-tests.sh --fail-fast
   ```

2. **Check coverage regularly:**
   ```bash
   ./scripts/test/verify-coverage.sh
   ```

3. **Benchmark performance changes:**
   ```bash
   ./scripts/test/run-load-tests.sh
   ```

4. **Review test reports:**
   ```bash
   open test-reports/index.html
   ```

## 🤝 Contributing

When adding new features:

1. ✅ Write unit tests
2. ✅ Write integration tests if needed
3. ✅ Ensure coverage ≥ 80%
4. ✅ Run full test suite
5. ✅ Check performance impact

## 📝 Environment Variables

Customize test behavior:

```bash
# Set coverage threshold
export COVERAGE_THRESHOLD=85

# Enable parallel testing
export PARALLEL=1

# Fail fast mode
export FAIL_FAST=1

# Redis URL
export REDIS_URL=redis://localhost:6379

# CI mode
export CI=true
```

## 🚨 Exit Codes

All scripts use consistent exit codes:

- `0` - All tests passed ✅
- `1` - Some tests failed ❌

Perfect for CI/CD and automation!

## 📞 Support

- **Issues:** Open a GitHub issue
- **Documentation:** See main README.md
- **Logs:** Check `test-reports/test-run.log`

---

**Built with ❤️ for zero-intervention testing**
