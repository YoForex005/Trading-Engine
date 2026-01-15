# Logging Guidelines

This document describes logging best practices for the Trading Engine backend.

## Overview

The Trading Engine uses Go's standard library `log/slog` for structured logging. All logs are output in JSON format to stdout for aggregation and analysis by monitoring systems (Prometheus, Grafana, CloudWatch, etc.).

## Quick Start

```go
import "github.com/epic1st/rtx/backend/internal/logging"

// In main.go - initialize logger once at startup
logging.Init(slog.LevelInfo)

// Use the global logger
logging.Default.Info("position opened",
    "position_id", position.ID,
    "account_id", account.ID,
    "symbol", position.Symbol,
    "volume", position.Volume,
)
```

## Log Levels

### Info
Use for normal operational events that indicate system health and business operations.

**When to use:**
- Position opened/closed
- Order placed/filled
- Account created
- Service started
- Configuration loaded

**Example:**
```go
logging.Default.Info("position opened",
    "position_id", pos.ID,
    "account_id", pos.AccountID,
    "symbol", pos.Symbol,
    "volume", pos.Volume,
    "open_price", pos.OpenPrice,
)
```

### Warn
Use for unexpected conditions that don't prevent operation but require attention.

**When to use:**
- Order rejected (insufficient margin)
- LP connection lost (with retry)
- Invalid request parameters
- Configuration issues with fallback

**Example:**
```go
logging.Default.Warn("order rejected",
    "account_id", order.AccountID,
    "symbol", order.Symbol,
    "reason", "insufficient_margin",
    "required_margin", margin,
)
```

### Error
Use for failures that require immediate attention but don't crash the service.

**When to use:**
- Database query failures
- LP API errors (after retries exhausted)
- Failed margin calculations
- Critical business logic errors

**Example:**
```go
logging.Default.Error("failed to update margin state",
    "account_id", accountID,
    "error", err,
)
```

**CRITICAL:** Every `Error` log **MUST** include an `"error"` field with the actual error object.

### Debug
Use for detailed troubleshooting information. Disabled in production by default.

**When to use:**
- Tick pipeline status
- Detailed request/response data
- Algorithm intermediate values
- Performance metrics

**Example:**
```go
logging.Default.Debug("tick pipeline status",
    "ticks_received", counter,
    "clients_connected", clientCount,
    "latest_symbol", tick.Symbol,
)
```

**Enable debug logging:**
```bash
# Set DEBUG=true in .env or environment
DEBUG=true ./server
```

## Required Context Fields

Always include these fields when logging events for specific entities:

### Trading Operations
- **account_id** - Every operation on an account
- **position_id** - Position open/close/modify operations
- **order_id** - Order placement/cancellation/fill operations
- **symbol** - Market data or trading operations on a symbol
- **trade_id** - Trade execution records

### Admin Operations
- **admin_id** - Admin actions (deposit, withdrawal, adjustment)
- **method** - Payment/transfer method
- **amount** - Financial amounts

### WebSocket/Network
- **remote_addr** - Client IP address
- **origin** - CORS origin header
- **client_id** - WebSocket client identifier

### Liquidity Providers
- **lp_id** - LP identifier (OANDA, Binance, etc.)
- **provider** - LP provider name
- **retry_count** - Retry attempts

## Anti-Patterns

### ❌ DON'T: String concatenation or formatting
```go
// BAD
logging.Default.Info(fmt.Sprintf("Position %d opened", id))
```

### ✅ DO: Use structured fields
```go
// GOOD
logging.Default.Info("position opened", "position_id", id)
```

---

### ❌ DON'T: Missing error field
```go
// BAD
logging.Default.Error("database failed")
```

### ✅ DO: Always include error
```go
// GOOD
logging.Default.Error("database query failed", "query", sql, "error", err)
```

---

### ❌ DON'T: Log sensitive data
```go
// BAD - exposes password
logging.Default.Info("user login", "password", password)
```

### ✅ DO: Exclude sensitive fields
```go
// GOOD
logging.Default.Info("user login",
    "username", username,
    "ip", r.RemoteAddr,
)
```

**Never log:**
- Passwords (plain or hashed)
- API keys
- JWT tokens
- Credit card numbers
- Personal identification numbers

---

### ❌ DON'T: Excessive logging
```go
// BAD - logs on every function call
func calculateMargin() {
    logging.Default.Debug("calculateMargin called")
    // ...
}
```

### ✅ DO: Log meaningful events
```go
// GOOD - logs result only
func calculateMargin(accountID int64) error {
    margin, err := doCalculation()
    if err != nil {
        logging.Default.Error("margin calculation failed",
            "account_id", accountID,
            "error", err,
        )
        return err
    }
    return nil
}
```

---

### ❌ DON'T: Unstructured logging
```go
// BAD - not queryable
log.Println("[B-Book] Position opened")
fmt.Printf("Account: %v\n", account)
```

### ✅ DO: Use slog exclusively
```go
// GOOD - structured and queryable
logging.Default.Info("position opened",
    "position_id", pos.ID,
    "account_id", pos.AccountID,
)
```

## Request-Scoped Logging

For HTTP handlers, create request-scoped loggers with trace information:

```go
func (h *APIHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
    // Create request-scoped logger
    reqLogger := logging.WithContext(
        "request_id", r.Header.Get("X-Request-ID"),
        "ip", r.RemoteAddr,
    )

    reqLogger.Info("place order request received")

    // Use reqLogger throughout the request
    reqLogger.Info("order placed",
        "order_id", order.ID,
        "symbol", order.Symbol,
    )
}
```

## Log Output Format

Logs are output as JSON lines to stdout:

```json
{"time":"2026-01-16T10:30:00Z","level":"INFO","msg":"position opened","position_id":12345,"account_id":1,"symbol":"BTCUSD","volume":0.1,"open_price":95000.50}
```

**Benefits:**
- Easily parsed by log aggregation tools
- Queryable in monitoring dashboards
- Consistent structure across services
- Machine-readable for automated alerts

## Querying Logs

### Find all errors for an account
```bash
jq 'select(.account_id == 1 and .level == "ERROR")' logs.json
```

### Find position operations
```bash
jq 'select(.msg | contains("position"))' logs.json
```

### Count errors by type
```bash
jq -r 'select(.level == "ERROR") | .msg' logs.json | sort | uniq -c
```

## Configuration

### Log Level via Environment
```bash
# Production - INFO level
./server

# Development - DEBUG level
DEBUG=true ./server
```

### Programmatic Configuration
```go
// In main.go
logLevel := slog.LevelInfo
if os.Getenv("DEBUG") == "true" {
    logLevel = slog.LevelDebug
}

logging.Init(logLevel)
```

## Testing

### Silence logs in tests
```go
func TestSomething(t *testing.T) {
    // Only show errors in tests
    logging.Init(slog.LevelError)

    // ... test code ...
}
```

### Capture and verify logs
```go
func TestLoggingBehavior(t *testing.T) {
    var buf bytes.Buffer
    handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
        Level: slog.LevelDebug,
    })
    logging.Default = slog.New(handler)

    // Trigger code that logs
    engine.ExecuteOrder(...)

    // Verify log output
    logs := buf.String()
    assert.Contains(t, logs, "position opened")
    assert.Contains(t, logs, "account_id")
}
```

## Migration from Unstructured Logging

### Before (unstructured)
```go
log.Printf("[B-Book] Position #%d opened for account %d", posID, accountID)
fmt.Println("Closing position", position.ID)
```

### After (structured)
```go
logging.Default.Info("position opened",
    "position_id", posID,
    "account_id", accountID,
)

logging.Default.Info("closing position", "position_id", position.ID)
```

## Monitoring Integration

### Prometheus Alerts
```yaml
# Alert on error rate
- alert: HighErrorRate
  expr: rate(log_messages_total{level="ERROR"}[5m]) > 10
  annotations:
    summary: "High error rate detected"
```

### Grafana Dashboard
```sql
-- Show errors by account
SELECT account_id, COUNT(*) as error_count
FROM logs
WHERE level = 'ERROR'
AND time > now() - 1h
GROUP BY account_id
ORDER BY error_count DESC
```

## References

- [Go slog package documentation](https://pkg.go.dev/log/slog)
- [Better Stack Go logging guide](https://betterstack.com/community/guides/logging/logging-in-go/)
- Phase 16 Research: `.planning/phases/16-code-organization-best-practices/16-RESEARCH.md`

---

**Last Updated:** 2026-01-16
**Phase:** 16 - Code Organization & Best Practices
