# Automated Test Suite Plan

## Version: 1.0.0 | Date: 2026-01-30

## 1. Test Architecture

```
tests/
├── unit/          # Go unit tests
├── integration/   # Bash integration tests  
├── e2e/           # Playwright E2E tests
└── fixtures/      # Test data
```

## 2. Unit Tests (Go)

### FIX Gateway Tests
```go
// backend/fix/gateway_test.go
func TestFIXGateway_NoSimulationData(t *testing.T) {
    gateway := fix.NewFIXGateway()
    gateway.Connect("YOFX2")
    
    // Collect 100 ticks
    for i := 0; i < 100; i++ {
        tick := <-gateway.GetMarketData()
        assert.Equal(t, "YOFX", tick.LP) // Must be YOFX
        assert.NotContains(t, tick.LP, "SIM")
    }
}
```

### Tick Store Tests
```go
func TestTickStore_RejectsSimulation(t *testing.T) {
    store := tickstore.NewOptimizedTickStore("TEST")
    
    // Attempt to store SIM tick (should fail in production)
    err := store.StoreTick("EURUSD", 1.10, 1.11, 0.01, "SIM", time.Now())
    assert.Error(t, err)
}
```

## 3. Integration Tests (Bash)

### FIX Pipeline Test
```bash
#!/bin/bash
# tests/integration/test_fix_pipeline.sh

# 1. Verify FIX connection
STATUS=$(curl -s http://localhost:7999/admin/fix/status | jq -r '.sessions.YOFX2')
[ "$STATUS" = "LOGGED_IN" ] || exit 1

# 2. Subscribe and receive ticks
curl -X POST http://localhost:7999/api/symbols/subscribe \
  -d '{"symbol":"EURUSD"}'
sleep 5

# 3. Verify no SIM data
DB="data/ticks/ticks_BROKER-001_$(date +%Y%m%d).db"
SIM_COUNT=$(sqlite3 "$DB" "SELECT COUNT(*) FROM ticks WHERE lp LIKE '%SIM%';")
[ "$SIM_COUNT" -eq 0 ] || exit 1
```

## 4. E2E Tests (Playwright)

```typescript
// tests/e2e/playwright/lp-display.spec.ts
test('all LP tags show YOFX', async ({ page }) => {
  await page.goto('http://localhost:5173');
  
  const lpCells = page.locator('[data-testid="lp-cell"]');
  const count = await lpCells.count();
  
  for (let i = 0; i < count; i++) {
    const text = await lpCells.nth(i).textContent();
    expect(text).toBe('YOFX');
  }
});
```

## 5. Run Tests

```bash
# Unit tests
go test ./... -v -cover

# Integration tests
bash tests/integration/test_fix_pipeline.sh

# E2E tests
npx playwright test
```

## 6. Coverage Targets

| Component | Target |
|-----------|--------|
| FIX Gateway | 80% |
| Tick Store | 85% |
| WebSocket Hub | 75% |
| **Overall** | **75%** |

## 7. CI/CD Integration

```yaml
# .github/workflows/test.yml
name: Tests
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Tests
        run: go test ./... -v
```
