# Testing Patterns

**Analysis Date:** 2026-01-18

## Test Framework

**Runner:**
- Go's built-in `testing` package
- No external test framework dependencies

**Assertion Library:**
- Standard Go testing: `t.Error()`, `t.Errorf()`, `t.Fatal()`, `t.Fatalf()`
- No assertion library (uses native Go conditionals)

**Run Commands:**
```bash
go test ./...                    # Run all tests
go test -v ./bbook               # Run tests in specific package with verbose output
go test -run TestName            # Run specific test
go test -parallel 8              # Run tests in parallel (default behavior)
go test -cover                   # Show coverage
```

**Additional Tools:**
- `golangci-lint` configured to run on test files: `tests: true` in config

## Test File Organization

**Location:**
- Co-located with source code: `engine_test.go` alongside `engine.go`
- Integration/E2E tests in separate directories:
  - `/Users/epic1st/Documents/trading engine/Trading-Engine/backend/test/integration/`
  - `/Users/epic1st/Documents/trading engine/Trading-Engine/backend/test/e2e/`

**Naming:**
- Unit tests: `*_test.go` suffix in same package: `engine_test.go`, `ledger_test.go`, `manager_test.go`
- Integration tests: `*_test.go` in `test/integration/`: `lp_adapter_test.go`
- E2E tests: `*_test.go` in `test/e2e/`: `order_flow_test.go`, `position_management_test.go`
- Manual test scripts: `test_*.go` (not test files): `test_now.go`, `test_all_features.go`

**Structure:**
```
backend/
├── bbook/
│   ├── engine.go
│   ├── engine_test.go         # Unit tests for engine
│   ├── ledger.go
│   └── ledger_test.go         # Unit tests for ledger
├── lpmanager/
│   ├── manager.go
│   └── manager_test.go
└── test/
    ├── integration/
    │   └── lp_adapter_test.go
    └── e2e/
        ├── order_flow_test.go
        └── position_management_test.go
```

## Test Structure

**Suite Organization:**
```go
func TestOrderExecution_MarketOrder(t *testing.T) {
    tests := []struct {
        name              string
        side              string
        volume            float64
        accountBalance    float64
        leverage          float64
        bidPrice          float64
        askPrice          float64
        expectedExecution bool
        expectedFillPrice float64
    }{
        {
            name:              "buy market order executes",
            side:              "BUY",
            volume:            1.0,
            accountBalance:    10000.00,
            leverage:          100,
            bidPrice:          1.0850,
            askPrice:          1.0852,
            expectedExecution: true,
            expectedFillPrice: 1.0852,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        tt := tt // Capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            // Setup
            engine := NewEngine()
            account := engine.CreateAccount("user1", "testuser", "pass", true)
            account.Balance = tt.accountBalance

            // Execute
            position, err := engine.ExecuteMarketOrder(...)

            // Verify
            if tt.expectedExecution {
                if err != nil {
                    t.Fatalf("expected order to execute but got error: %v", err)
                }
                if position.OpenPrice != tt.expectedFillPrice {
                    t.Errorf("fill price: got %.4f, want %.4f", position.OpenPrice, tt.expectedFillPrice)
                }
            }
        })
    }
}
```

**Patterns:**
- Table-driven tests using anonymous structs
- Subtests with `t.Run()` for each test case
- `t.Parallel()` for concurrent test execution
- `tt := tt` to capture loop variable (required for parallel tests)
- Clear test names: `"buy market order executes"`, `"insufficient margin rejects order"`

## Mocking

**Framework:**
- Manual mocks (no mocking library)
- Interface-based mocking for adapters

**Patterns:**
```go
// Mock implementation
type MockAdapter struct {
    id         string
    name       string
    lpType     string
    connected  bool
    quotesChan chan Quote
}

func (m *MockAdapter) ID() string           { return m.id }
func (m *MockAdapter) Name() string         { return m.name }
func (m *MockAdapter) Type() string         { return m.lpType }
func (m *MockAdapter) IsConnected() bool    { return m.connected }
func (m *MockAdapter) Connect() error       { m.connected = true; return nil }
func (m *MockAdapter) Disconnect() error    { m.connected = false; return nil }
```

**What to Mock:**
- External LP adapters: `MockAdapter` in `manager_test.go`
- Price feeds: Price callback function in engine tests
- WebSocket connections (implied, not shown in current tests)

**What NOT to Mock:**
- Internal business logic (Engine, Ledger)
- Data structures (Account, Position, Order)
- Simple utility functions

## Fixtures and Factories

**Test Data:**
```go
func TestMain(m *testing.M) {
    // Initialize logger for tests
    logging.Init(slog.LevelInfo)
    os.Exit(m.Run())
}

// Setup in each test
engine := NewEngine()
account := engine.CreateAccount("user1", "testuser", "pass", true)
account.Balance = 10000.00
engine.ledger.SetBalance(account.ID, 10000.00)
```

**Patterns:**
- `TestMain()` for package-level setup: Initialize logging
- In-test factories: Use real constructors (`NewEngine()`, `NewLedger()`)
- Direct field manipulation for test data: `account.Balance = 10000.00`
- Temporary directories for file-based tests: `t.TempDir()`

**Location:**
- No separate fixture files
- Test data created inline in test functions
- External test data in `/Users/epic1st/Documents/trading engine/Trading-Engine/backend/testdata/`:
  - `test_accounts.json`
  - `test_quotes.json`

## Coverage

**Requirements:**
- No enforced coverage target
- Coverage tracking available but not mandatory

**View Coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Current State:**
- 66 test functions across 13 test files
- 17 test files total (including manual test scripts)
- 141 total Go files in codebase
- Approximately 12% of files have tests (17/141)

## Test Types

**Unit Tests:**
- Scope: Individual functions and methods
- Approach: Isolated testing with minimal dependencies
- Examples:
  - `TestLedgerDeposit` - Tests deposit operation in isolation
  - `TestOrderExecution_VolumeValidation` - Tests volume validation logic
  - `TestClosePosition` - Tests position closing calculations

**Integration Tests:**
- Scope: Multiple components working together
- Approach: Real implementations with mocked external dependencies
- Location: `test/integration/`
- Examples:
  - `lp_adapter_test.go` - Tests LP adapter integration with manager

**E2E Tests:**
- Scope: Full workflow from start to finish
- Approach: End-to-end trading scenarios
- Location: `test/e2e/`
- Examples:
  - `order_flow_test.go` - Complete order execution flow
  - `position_management_test.go` - Full position lifecycle

**Manual Tests:**
- Not part of automated suite
- Located in `backend/fix/`: `test_now.go`, `test_all_features.go`, `test_yofx_443.go`
- Used for FIX protocol testing and debugging

## Common Patterns

**Async Testing:**
```go
func TestLPManager_QuoteAggregation(t *testing.T) {
    // Send quote in goroutine
    go func() {
        oandaAdapter.quotesChan <- testQuote
    }()

    // Receive with timeout
    select {
    case receivedQuote := <-manager.GetQuotesChan():
        if receivedQuote.Symbol != "EURUSD" {
            t.Errorf("got symbol %s, want EURUSD", receivedQuote.Symbol)
        }
    case <-time.After(2 * time.Second):
        t.Fatal("timeout waiting for aggregated quote")
    }
}
```

**Error Testing:**
```go
_, err := ledger.Withdraw(1, 1500.00, "BANK", "REF123", "Test withdrawal", "admin1")

if tt.expectError {
    if err == nil {
        t.Error("expected error but got none")
    }
} else {
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

**Floating Point Comparison:**
```go
tolerance := 0.01
if abs(actualBalance-expectedBalance) > tolerance {
    t.Errorf("balance after commission: got %.2f, want %.2f", actualBalance, expectedBalance)
}

// Helper function
func abs(x float64) float64 {
    if x < 0 {
        return -x
    }
    return x
}
```

**Parallel Testing:**
```go
func TestLedgerDeposit(t *testing.T) {
    tests := []struct {
        name            string
        initialBalance  float64
        depositAmount   float64
        expectedBalance float64
        expectError     bool
    }{
        // ... test cases
    }

    for _, tt := range tests {
        tt := tt // Capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // Run subtests in parallel

            // Test implementation
        })
    }
}
```

**Cleanup:**
```go
tmpDir := t.TempDir() // Automatically cleaned up after test
configPath := filepath.Join(tmpDir, "lp_config.json")

manager := NewManager(configPath)
// Test uses temp config file, no manual cleanup needed
```

## Test Naming Conventions

**Test Functions:**
- `Test<Component>_<Scenario>`: `TestOrderExecution_MarketOrder`
- `Test<Component><Operation>`: `TestLedgerDeposit`, `TestClosePosition`
- Descriptive subtest names: `"buy market order executes"`, `"insufficient margin rejects order"`

**Test Variables:**
- `tt` for table-driven test struct
- `tests` for slice of test cases
- Descriptive field names in test structs: `expectedExecution`, `expectError`, `expectedBalance`

---

*Testing analysis: 2026-01-18*
