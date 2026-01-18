# 🚀 Quick Start - Automated Testing

## One Command to Rule Them All

```bash
cd backend
./scripts/test/run-all-tests.sh
```

## Common Commands

### Run Everything
```bash
./scripts/test/run-all-tests.sh
```

### Run Tests Fast (Parallel)
```bash
./scripts/test/run-all-tests.sh --parallel
```

### Stop on First Failure
```bash
./scripts/test/run-all-tests.sh --fail-fast
```

### Individual Test Suites
```bash
# Unit tests only
./scripts/test/run-backend-tests.sh

# Integration tests only
./scripts/test/run-integration-tests.sh

# E2E tests only
./scripts/test/run-e2e-tests.sh

# Performance benchmarks only
./scripts/test/run-load-tests.sh
```

### Check Coverage
```bash
./scripts/test/verify-coverage.sh
```

### View Reports
```bash
# Open main report in browser
open test-reports/index.html

# Open coverage report
open test-reports/coverage/coverage.html

# Open performance report
open test-reports/benchmarks/report.html
```

## CI/CD

Tests run automatically on GitHub Actions:
- ✅ Every push to main/develop
- ✅ Every pull request
- ✅ Daily at 2 AM UTC

## What Gets Tested?

### ✅ Unit Tests
- All Go packages
- Business logic
- Data structures
- Utilities

### ✅ Integration Tests
- API endpoints
- Database operations
- Redis integration
- WebSocket connections

### ✅ E2E Tests
- Complete workflows
- User authentication
- Order placement
- Position management

### ✅ Performance Tests
- Tick throughput (50K+ ticks/sec)
- Order latency (<10ms)
- Memory usage
- CPU profiling

## Exit Codes

- `0` = ✅ All tests passed
- `1` = ❌ Some tests failed

## Prerequisites

Required:
- ✅ Go 1.24.0+
- ✅ Docker

Optional:
- Redis CLI
- wscat (for WebSocket tests)

## Troubleshooting

### Tests failing?
```bash
# Clear cache and retry
go clean -testcache
rm -rf test-reports/
./scripts/test/run-all-tests.sh
```

### Coverage too low?
```bash
# See what's missing
open test-reports/coverage/coverage.html
cat test-reports/coverage/low-coverage-files.txt
```

### Performance issues?
```bash
# Check benchmarks
open test-reports/benchmarks/report.html

# Profile CPU
go tool pprof test-reports/benchmarks/cpu.prof
```

## Need Help?

1. Check logs: `cat test-reports/test-run.log`
2. See detailed README: `cat scripts/test/README.md`
3. Open an issue on GitHub

---

**Zero configuration. Zero intervention. Just run it!** ✨
