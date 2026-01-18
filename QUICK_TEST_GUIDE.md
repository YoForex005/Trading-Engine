# Quick Test Guide 🚀

## TL;DR - Run All Tests

```bash
./run_all_tests.sh
```

That's it! This will run all 86 tests (53 integration + 33 E2E).

---

## Individual Test Suites

### Integration Tests (Go)

```bash
cd backend
./tests/integration/run_tests.sh all
```

**Specific tests:**
```bash
./tests/integration/run_tests.sh api          # API tests (17 tests)
./tests/integration/run_tests.sh websocket    # WebSocket tests (10 tests)
./tests/integration/run_tests.sh order        # Order flow tests (11 tests)
./tests/integration/run_tests.sh admin        # Admin tests (15 tests)
```

**Other options:**
```bash
./tests/integration/run_tests.sh coverage     # Coverage report
./tests/integration/run_tests.sh benchmark    # Performance benchmarks
./tests/integration/run_tests.sh race         # Race detection
```

### E2E Tests (JavaScript/Playwright)

```bash
cd tests/e2e
./run_tests.sh all
```

**Specific tests:**
```bash
./run_tests.sh trading      # Trading workflows (10 tests)
./run_tests.sh admin        # Admin workflows (23 tests)
```

**Debug modes:**
```bash
./run_tests.sh headed       # See browser
./run_tests.sh debug        # Step through
./run_tests.sh ui           # Playwright UI
./run_tests.sh report       # View results
```

---

## First Time Setup

### E2E Tests Only (Integration tests need no setup)

```bash
cd tests/e2e
npm install
npx playwright install
```

---

## What Gets Tested?

### Integration Tests (53 tests)
✅ All API endpoints
✅ WebSocket real-time streaming
✅ Complete order flows
✅ Admin operations
✅ Error handling
✅ Performance

### E2E Tests (33 tests)
✅ Complete user workflows
✅ Trading scenarios
✅ Admin configuration
✅ Multi-step processes

---

## Quick Commands Reference

```bash
# RUN ALL TESTS
./run_all_tests.sh

# INTEGRATION ONLY
./run_all_tests.sh --integration-only

# E2E ONLY
./run_all_tests.sh --e2e-only

# SKIP E2E (integration only, faster)
./run_all_tests.sh --skip-e2e
```

---

## Expected Output

```
╔═══════════════════════════════════════════════════════════╗
║     RTX Trading Engine - Complete Test Suite             ║
╚═══════════════════════════════════════════════════════════╝

✓ Backend server is running

╔═══════════════════════════════════════════════════════════╗
║         Integration Tests (Go)                            ║
╚═══════════════════════════════════════════════════════════╝

Running API Tests...
✓ API Tests passed

Running WebSocket Tests...
✓ WebSocket Tests passed

Running Order Flow Tests...
✓ Order Flow Tests passed

Running Admin Flow Tests...
✓ Admin Flow Tests passed

✓ Integration tests passed

╔═══════════════════════════════════════════════════════════╗
║         E2E Tests (Playwright)                            ║
╚═══════════════════════════════════════════════════════════╝

✓ E2E tests passed

╔═══════════════════════════════════════════════════════════╗
║         Test Summary                                      ║
╚═══════════════════════════════════════════════════════════╝

✓ Integration Tests: PASSED
✓ E2E Tests: PASSED

╔═══════════════════════════════════════════════════════════╗
║     All Tests Completed Successfully! 🎉                  ║
╚═══════════════════════════════════════════════════════════╝
```

---

## Troubleshooting

**Backend not running?**
```bash
cd backend
go run cmd/server/main.go
```

**Port 7999 in use?**
```bash
lsof -ti:7999 | xargs kill -9
```

**E2E dependencies missing?**
```bash
cd tests/e2e
npm install
npx playwright install
```

---

## Performance Expectations

- Integration tests: ~2-5 seconds total
- E2E tests: ~30-60 seconds total
- Total runtime: ~1-2 minutes

---

## More Details

- **Full Guide**: See `TESTING.md`
- **Test Summary**: See `TEST_SUMMARY.md`
- **Integration Tests**: See `backend/tests/integration/`
- **E2E Tests**: See `tests/e2e/`

---

**86 Tests | 100% Endpoint Coverage | Production Ready ✅**
