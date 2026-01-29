# Coding Conventions

**Analysis Date:** 2026-01-18

## Naming Patterns

**Files:**
- Go source files: lowercase with underscores: `engine.go`, `daily_store.go`, `ohlc_cache.go`
- Test files: `*_test.go` suffix: `engine_test.go`, `ledger_test.go`, `manager_test.go`
- Main entry points: `main.go` in `cmd/server/` directory
- Test utilities: `test_*.go` prefix for manual test scripts: `test_now.go`, `test_all_features.go`, `test_yofx_443.go`

**Functions:**
- Exported functions: PascalCase: `NewEngine()`, `CreateAccount()`, `ExecuteMarketOrder()`
- Unexported functions: camelCase: `calculateMargin()`, `getAccountSummaryUnlocked()`, `saveConfigLocked()`
- Factory constructors: `New*` prefix: `NewEngine()`, `NewLedger()`, `NewManager()`, `NewHub()`
- Handler methods: `Handle*` prefix: `HandleGetAccountSummary()`, `HandlePlaceMarketOrder()`, `HandleClosePosition()`

**Variables:**
- Exported fields: PascalCase in structs: `Account.ID`, `Position.OpenPrice`, `Order.Status`
- Unexported fields: camelCase in structs: `Engine.mu`, `Hub.latestPrices`, `Manager.activeAggregators`
- Constants: SCREAMING_SNAKE_CASE: `OANDA_API_KEY`, `OANDA_ACCOUNT_ID` (though rarely used)
- Package-level vars: camelCase: `upgrader`, `tickCounter`, `executionMode`, `brokerConfig`

**Types:**
- Structs: PascalCase: `Account`, `Position`, `Order`, `Trade`, `LedgerEntry`, `MarketTick`
- Interfaces: PascalCase: `LPAdapter` (implied from usage)
- Type aliases: PascalCase for exported types

## Code Style

**Formatting:**
- Tool: `gofmt` (standard Go formatting)
- Indentation: Tabs (Go standard)
- Line length: No strict limit, but generally kept reasonable (< 120 chars for readability)

**Linting:**
- Tool: `golangci-lint` with custom configuration
- Config: `/Users/epic1st/Documents/trading engine/Trading-Engine/backend/.golangci.yml`
- Enabled linters:
  - `govet` - Critical syntax and type checking
  - `ineffassign` - Detect ineffectual assignments
  - `unused` - Find unused code
  - `misspell` - Spell checking
  - `errcheck` - Error handling validation (with exclusions for JSON encoding, defer Close)
- Disabled linters (for gradual improvement):
  - `staticcheck`, `gocritic`, `gocyclo`, `bodyclose`, `errorlint`, `gosec`, `noctx`

## Import Organization

**Order:**
1. Standard library imports: `encoding/json`, `log`, `sync`, `time`, `errors`, `fmt`
2. External dependencies: `github.com/gorilla/websocket`, `github.com/golang-jwt/jwt/v5`, `golang.org/x/crypto`
3. Internal packages: `github.com/epic1st/rtx/backend/...`

**Example from `backend/cmd/server/main.go`:**
```go
import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/epic1st/rtx/backend/api"
    "github.com/epic1st/rtx/backend/auth"
    "github.com/epic1st/rtx/backend/internal/api/handlers"
    "github.com/epic1st/rtx/backend/internal/core"
    "github.com/epic1st/rtx/backend/lpmanager"
    "github.com/epic1st/rtx/backend/lpmanager/adapters"
    "github.com/epic1st/rtx/backend/tickstore"
    "github.com/epic1st/rtx/backend/ws"
)
```

**Path Aliases:**
- None detected. Uses full import paths.

## Error Handling

**Patterns:**
- Return errors as last return value: `func Execute(...) (*Position, error)`
- Check errors immediately after function calls
- Log errors with context before returning: `log.Printf("[B-Book] Failed to...: %v", err)`
- Use `errors.New()` for simple errors: `return errors.New("account not found")`
- Use `fmt.Errorf()` for formatted errors: `return fmt.Errorf("symbol %s not found", symbol)`
- Named errors for reusable error types: `ErrLPNotFound`, `ErrLPAlreadyExists`

**Error checking:**
```go
position, err := engine.ExecuteMarketOrder(accountID, symbol, side, volume, sl, tp)
if err != nil {
    return nil, fmt.Errorf("failed to execute order: %v", err)
}
```

**Validation pattern:**
```go
if volume < spec.MinVolume || volume > spec.MaxVolume {
    return nil, fmt.Errorf("volume must be between %.2f and %.2f", spec.MinVolume, spec.MaxVolume)
}
```

## Logging

**Framework:** Standard library `log` package

**Patterns:**
- Prefix logs with component in brackets: `[B-Book]`, `[Hub]`, `[LPManager]`, `[Ledger]`, `[FIX]`
- Log levels indicated by convention:
  - `log.Println()` for info-level logs
  - `log.Printf()` for formatted info logs
  - `[WARN]` prefix for warnings: `log.Printf("[WARN] Admin login failed")`
  - `[CRITICAL]` prefix for critical errors: `log.Printf("[CRITICAL] JWT Generation failed")`
  - `[INFO]` prefix for informational messages
- Structured logging partially adopted with `slog` in tests: `logging.Init(slog.LevelInfo)`

**Example logging patterns:**
```go
log.Printf("[B-Book] EXECUTED: %s %s %.2f lots @ %.5f (Position #%d)", side, symbol, volume, fillPrice, positionID)
log.Printf("[Ledger] DEPOSIT: Account #%d +%.2f via %s | Balance: %.2f", accountID, amount, method, newBalance)
log.Printf("[Hub] Client connected. Total clients: %d", clientCount)
```

## Comments

**When to Comment:**
- Exported types and functions (minimal but present)
- Complex business logic (sparse)
- Non-obvious code patterns
- Critical sections: `// CRITICAL: Persist tick for chart history`
- Performance-sensitive code: `// BUFFERED: Prevent blocking engine`, `// NON-BLOCKING SEND`
- TODOs for incomplete features: `// TODO: Implement account-specific WebSocket`

**JSDoc/GoDoc:**
- Minimal GoDoc comments present
- Struct fields use inline comments: `` `json:"id"` `` tags with field descriptions
- Package-level comments mostly absent

**Example:**
```go
// Engine is the B-Book execution engine
type Engine struct {
    mu             sync.RWMutex
    accounts       map[int64]*Account
    positions      map[int64]*Position
    orders         map[int64]*Order
    trades         []Trade
    symbols        map[string]*SymbolSpec
    nextPositionID int64
    nextOrderID    int64
    nextTradeID    int64
    priceCallback  func(symbol string) (bid, ask float64, ok bool)
    ledger         *Ledger
}
```

## Function Design

**Size:**
- Most functions 20-100 lines
- Large functions tolerated for handlers (200+ lines in `main.go`)
- Complex logic broken into helper functions: `calculateMargin()`, `calculatePnL()`, `getAccountSummaryUnlocked()`

**Parameters:**
- Prefer named parameters over structs for simple functions
- Use struct parameters for complex configuration: `LPConfig`, `BrokerConfig`
- Context not consistently used (disabled in linter)

**Return Values:**
- Multiple return values common: `(value, error)`, `(bid, ask, ok)`
- Named return values avoided
- Pointer returns for structs: `(*Position, error)`, `(*Account, bool)`
- Boolean flag for existence checks: `GetAccount(id) (*Account, bool)`

## Module Design

**Exports:**
- Export only what's needed by other packages
- Internal implementations unexported: `saveConfigLocked()`, `getAccountSummaryUnlocked()`
- Public API through exported methods on exported types

**Barrel Files:**
- Not used (Go doesn't have barrel file pattern)
- Each package exports from individual files

**Package Organization:**
- `backend/bbook/` - B-Book trading engine (internal execution)
- `backend/internal/core/` - Core business logic
- `backend/internal/api/handlers/` - HTTP API handlers
- `backend/lpmanager/` - Liquidity provider management
- `backend/lpmanager/adapters/` - LP-specific adapters
- `backend/ws/` - WebSocket hub for real-time updates
- `backend/auth/` - Authentication and JWT
- `backend/tickstore/` - Market data storage
- `backend/fix/` - FIX protocol integration
- `backend/cmd/server/` - Main application entry point

## Concurrency Patterns

**Mutexes:**
- Use `sync.RWMutex` for read-heavy data: `Engine.mu`, `Hub.mu`, `Ledger.mu`
- Lock/defer unlock pattern:
```go
e.mu.Lock()
defer e.mu.Unlock()
```
- Separate read/write locks for performance:
```go
h.mu.RLock()
defer h.mu.RUnlock()
```

**Channels:**
- Buffered channels for non-blocking sends: `make(chan []byte, 2048)`, `make(chan Quote, 1000)`
- Select with default for non-blocking operations:
```go
select {
case h.broadcast <- data:
default:
    // Drop tick if buffer full
}
```

**Goroutines:**
- Launched for background tasks: `go hub.Run()`, `go lpMgr.StartQuoteAggregation()`
- Defer cleanup in goroutines: `defer conn.Close()`

## JSON Handling

**Struct Tags:**
- Use `json` tags on all exported fields: `` `json:"accountId"` ``
- Use camelCase for JSON field names
- Use `omitempty` for optional fields: `` `json:"closePrice,omitempty"` ``

**Encoding/Decoding:**
- `json.NewEncoder(w).Encode(data)` for HTTP responses
- `json.NewDecoder(r.Body).Decode(&req)` for HTTP requests
- Errcheck excludes these operations (logged at higher level)

---

*Convention analysis: 2026-01-18*
