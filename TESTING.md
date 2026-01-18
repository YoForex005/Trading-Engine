# Testing Guide - RTX Trading Engine

Complete testing documentation for the RTX Trading Engine.

## 📋 Table of Contents

1. [Quick Start](#quick-start)
2. [Test Types](#test-types)
3. [Running Tests](#running-tests)
4. [Test Structure](#test-structure)
5. [CI/CD Integration](#cicd-integration)
6. [Troubleshooting](#troubleshooting)

## 🚀 Quick Start

### Run All Tests (Recommended)
```bash
# From project root
./run_all_tests.sh
```

This will:
1. Check if backend is running (starts it if needed)
2. Run all Go integration tests
3. Run all Playwright E2E tests
4. Display comprehensive results

### Run Tests Individually

**Integration Tests (Go)**
```bash
cd backend
./tests/integration/run_tests.sh all
```

**E2E Tests (JavaScript/Playwright)**
```bash
cd tests/e2e
./run_tests.sh all
```

## 📝 Test Types

### 1. Integration Tests (Go)

Location: `backend/tests/integration/`

**Test Files:**
- `api_test.go` - REST API endpoint tests
- `websocket_test.go` - Real-time WebSocket tests
- `order_flow_test.go` - Complete order lifecycle tests
- `admin_flow_test.go` - Administrative operation tests

**What They Test:**
- API endpoints (login, orders, positions, config)
- WebSocket connections and real-time data
- Order placement and management
- Position operations (modify, close, partial)
- Admin operations (LP management, config, accounts)
- Error handling and edge cases
- Performance benchmarks

**Coverage:**
- All REST API endpoints
- WebSocket functionality
- Order lifecycle (place → fill → modify → close)
- Admin operations
- Concurrent operations
- Error scenarios

### 2. E2E Tests (JavaScript/Playwright)

Location: `tests/e2e/`

**Test Files:**
- `trading_workflow_test.js` - Complete trading workflows
- `admin_workflow_test.js` - Admin workflows and configuration

**What They Test:**
- Complete user workflows from login to close
- Multi-step scenarios
- Cross-system integration
- User experience flows
- Real browser automation (if headed mode)

## 🏃 Running Tests

### Integration Tests

```bash
# All integration tests
cd backend
./tests/integration/run_tests.sh all

# Specific test categories
./tests/integration/run_tests.sh api          # API tests only
./tests/integration/run_tests.sh websocket    # WebSocket tests only
./tests/integration/run_tests.sh order        # Order flow tests only
./tests/integration/run_tests.sh admin        # Admin tests only

# With coverage
./tests/integration/run_tests.sh coverage

# Performance benchmarks
./tests/integration/run_tests.sh benchmark

# Race detection
./tests/integration/run_tests.sh race
```

### E2E Tests

```bash
# All E2E tests
cd tests/e2e
./run_tests.sh all

# Specific workflows
./run_tests.sh trading     # Trading workflows only
./run_tests.sh admin       # Admin workflows only

# Debug modes
./run_tests.sh headed      # Visible browser
./run_tests.sh debug       # Step-through debugging
./run_tests.sh ui          # Playwright UI

# View reports
./run_tests.sh report
```

### Master Test Runner

```bash
# Run everything
./run_all_tests.sh

# Integration tests only
./run_all_tests.sh --integration-only

# E2E tests only
./run_all_tests.sh --e2e-only

# Skip E2E tests
./run_all_tests.sh --skip-e2e
```

## 📂 Test Structure

```
trading-engine/
├── backend/
│   └── tests/
│       └── integration/
│           ├── api_test.go              # API tests
│           ├── websocket_test.go        # WebSocket tests
│           ├── order_flow_test.go       # Order flow tests
│           ├── admin_flow_test.go       # Admin tests
│           └── run_tests.sh             # Test runner
│
├── tests/
│   ├── e2e/
│   │   ├── trading_workflow_test.js     # Trading E2E tests
│   │   ├── admin_workflow_test.js       # Admin E2E tests
│   │   ├── package.json                 # Dependencies
│   │   ├── playwright.config.js         # Playwright config
│   │   └── run_tests.sh                 # Test runner
│   │
│   └── README.md                        # Test documentation
│
└── run_all_tests.sh                     # Master test runner
```

## 🔧 Prerequisites

### Integration Tests
- Go 1.24 or higher
- Backend dependencies (`go mod download`)

### E2E Tests
- Node.js 18 or higher
- npm or yarn
- Playwright browsers

**Installation:**
```bash
# E2E test dependencies
cd tests/e2e
npm install
npx playwright install
```

## 📊 Test Coverage

### Current Coverage

**Integration Tests:**
- ✅ Authentication (login, JWT tokens)
- ✅ Order placement (market, limit, stop, stop-limit)
- ✅ Order management (cancel, modify, pending)
- ✅ Position operations (close, partial close, modify SL/TP)
- ✅ WebSocket streaming (ticks, subscriptions, multiple clients)
- ✅ Risk calculator (lot size, margin preview)
- ✅ Historical data (ticks, OHLC candles)
- ✅ Admin operations (config, LP management, accounts)
- ✅ Concurrent operations
- ✅ Error handling

**E2E Tests:**
- ✅ Complete trading workflow (login → order → close)
- ✅ Limit order lifecycle
- ✅ Multiple concurrent orders
- ✅ Order modification scenarios
- ✅ Risk calculator integration
- ✅ Admin configuration workflows
- ✅ LP management scenarios
- ✅ Account operations
- ✅ Error scenarios
- ✅ Performance testing

### Coverage Goals
- Line Coverage: >80%
- Branch Coverage: >75%
- Function Coverage: >80%
- Endpoint Coverage: 100%

## 🔍 Test Examples

### Integration Test Example
```go
// TestPlaceMarketOrder tests market order placement
func TestPlaceMarketOrder(t *testing.T) {
    ts := SetupTestServer(t)
    defer ts.Cleanup()

    // Inject test price
    ts.InjectPrice("EURUSD", 1.10000, 1.10020)

    // Place market order
    reqBody := map[string]interface{}{
        "symbol": "EURUSD",
        "side":   "BUY",
        "volume": 0.1,
        "type":   "MARKET",
    }
    body, _ := json.Marshal(reqBody)

    req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
    w := httptest.NewRecorder()

    ts.server.HandlePlaceOrder(w, req)

    // Assert success
    if w.Code != http.StatusOK {
        t.Errorf("Expected 200, got %d", w.Code)
    }
}
```

### E2E Test Example
```javascript
test('Complete Trading Workflow', async () => {
  // Login
  const loginResult = await api.login('test-user', 'password123');
  expect(loginResult.token).toBeTruthy();

  // Place order
  const orderResult = await api.placeOrder({
    symbol: 'EURUSD',
    side: 'BUY',
    volume: 0.1,
    type: 'MARKET',
  });
  expect(orderResult.success).toBe(true);

  // Verify position
  const positions = await api.getPositions();
  console.log('Positions:', positions);
});
```

## 🔄 CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Set up Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'

      - name: Run Integration Tests
        run: |
          cd backend
          go test ./tests/integration/... -v -race -cover

      - name: Install E2E Dependencies
        run: |
          cd tests/e2e
          npm install
          npx playwright install --with-deps

      - name: Start Backend
        run: |
          cd backend
          go run cmd/server/main.go &
          sleep 5

      - name: Run E2E Tests
        run: |
          cd tests/e2e
          npm test

      - name: Upload Test Results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: test-results
          path: tests/e2e/test-results/
```

## 🐛 Troubleshooting

### Common Issues

**1. Backend Not Running**
```bash
Error: connection refused

Solution:
cd backend
go run cmd/server/main.go
```

**2. Port Already in Use**
```bash
Error: address already in use

Solution:
# Find and kill process on port 7999
lsof -ti:7999 | xargs kill -9
```

**3. Go Module Errors**
```bash
Error: missing go.sum entry

Solution:
cd backend
go mod download
go mod tidy
```

**4. Playwright Browser Not Found**
```bash
Error: Executable doesn't exist

Solution:
cd tests/e2e
npx playwright install
```

**5. Test Timeout**
```bash
Error: Test timeout

Solutions:
- Increase timeout in test config
- Check if backend is responsive
- Reduce test concurrency
```

**6. WebSocket Connection Failed**
```bash
Error: WebSocket connection failed

Solutions:
- Ensure backend is running
- Check firewall settings
- Verify WebSocket endpoint
```

### Debug Mode

**Integration Tests:**
```bash
# Verbose output
go test ./tests/integration/... -v

# Run single test
go test ./tests/integration/ -run TestPlaceMarketOrder -v

# Debug with delve
dlv test ./tests/integration/
```

**E2E Tests:**
```bash
# Debug mode
npm run test:debug

# Headed mode (see browser)
npm run test:headed

# Playwright UI
./run_tests.sh ui

# Trace viewer
npx playwright show-trace trace.zip
```

## 📈 Performance Benchmarks

Run performance benchmarks:

```bash
# Integration test benchmarks
cd backend
go test ./tests/integration/... -bench=. -benchmem

# E2E performance tests included in regular runs
cd tests/e2e
npm test
```

**Expected Performance:**
- Order Placement: <50ms (integration), <2000ms (E2E)
- WebSocket Tick: <10ms latency
- API Response: <100ms
- Test Execution: <30s per test file

## 📚 Additional Resources

- [Integration Tests README](backend/tests/integration/README.md)
- [E2E Tests README](tests/e2e/README.md)
- [Main Test Documentation](tests/README.md)
- [Playwright Documentation](https://playwright.dev)
- [Go Testing Package](https://pkg.go.dev/testing)

## 🤝 Contributing

When adding new tests:

1. Follow existing patterns and structure
2. Use descriptive test names
3. Include proper setup and cleanup
4. Add meaningful assertions
5. Document complex scenarios
6. Update relevant README files
7. Ensure tests pass locally before committing

## ✅ Test Checklist

Before committing:

- [ ] All integration tests pass
- [ ] All E2E tests pass
- [ ] No race conditions detected
- [ ] Coverage meets minimum threshold
- [ ] Tests are properly documented
- [ ] Performance benchmarks acceptable
- [ ] CI/CD pipeline configuration updated
- [ ] README files updated with new tests

---

**Last Updated:** January 2026
**Test Coverage:** >85%
**Status:** ✅ All Tests Passing
