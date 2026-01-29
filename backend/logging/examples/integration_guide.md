# Integration Guide: Adding Structured Logging to Trading Engine

This guide shows how to integrate the new logging system into the existing trading engine codebase.

## Step 1: Update Dependencies

Add the Sentry SDK to go.mod:

```bash
cd /Users/epic1st/Documents/trading\ engine/backend
go get github.com/getsentry/sentry-go
go mod tidy
```

## Step 2: Initialize Logging in main.go

Update `cmd/server/main.go`:

```go
package main

import (
    // ... existing imports
    "github.com/epic1st/rtx/backend/logging"
)

func main() {
    // Initialize logging FIRST
    setupLogging()

    logging.Info("Starting RTX Trading Engine",
        logging.String("version", "3.0"),
        logging.String("mode", brokerConfig.ExecutionMode),
        logging.String("lp", brokerConfig.PriceFeedLP),
    )

    // ... rest of initialization
}

func setupLogging() {
    // Create rotating file writer
    rotatingWriter, err := logging.NewRotatingFileWriter(logging.RotationConfig{
        Filename:   "./logs/trading-engine.log",
        MaxSizeMB:  100,
        MaxAge:     7 * 24 * time.Hour,
        MaxBackups: 30,
    })
    if err != nil {
        panic(err)
    }

    // Multi-writer: file + stdout
    multiWriter := logging.NewMultiWriter(rotatingWriter, os.Stdout)

    // Create logger
    logger := logging.NewLogger(logging.INFO, multiWriter)

    // Enable sampling in production
    if os.Getenv("ENVIRONMENT") == "production" {
        logger.EnableSampling(0.1, true) // Keep 10% of INFO logs, all errors
    }

    // Add Sentry hook if DSN is set
    if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
        sentryHook, err := logging.NewSentryHook(dsn, os.Getenv("ENVIRONMENT"))
        if err == nil {
            logger.AddHook(sentryHook)
            logging.Info("Sentry integration enabled")
        }
    }
}
```

## Step 3: Add HTTP Middleware

Wrap all HTTP handlers with logging middleware:

```go
func main() {
    // ... initialization

    logger := logging.NewLogger(logging.INFO)
    loggingMW := logging.HTTPLoggingMiddleware(logger)
    panicMW := logging.PanicRecoveryMiddleware(logger)

    // Wrap handlers
    http.HandleFunc("/api/orders/market",
        wrapHandler(panicMW, loggingMW, apiHandler.HandlePlaceMarketOrder))

    http.HandleFunc("/api/positions/close",
        wrapHandler(panicMW, loggingMW, apiHandler.HandleClosePosition))

    // ... rest of routes
}

// Helper to apply middleware
func wrapHandler(middlewares ...func(http.Handler) http.Handler) func(http.HandlerFunc) http.HandlerFunc {
    return func(h http.HandlerFunc) http.HandlerFunc {
        handler := http.Handler(h)
        for i := len(middlewares) - 1; i >= 0; i-- {
            handler = middlewares[i](handler)
        }
        return handler.ServeHTTP
    }
}
```

## Step 4: Replace log.Printf Calls

### In Order Placement (api/server.go)

Before:
```go
log.Printf("[A-Book] Executing %s %s %.2f lots %s via LP",
    req.Side, req.Symbol, req.Volume, req.Type)
```

After:
```go
logging.WithContext(r.Context()).Info("Executing A-Book order",
    logging.Component("abook"),
    logging.String("side", req.Side),
    logging.Symbol(req.Symbol),
    logging.Float64("volume", req.Volume),
    logging.String("type", req.Type),
    logging.AccountID(req.AccountID),
)
```

### In Error Cases

Before:
```go
if err != nil {
    log.Printf("[A-Book] Order placement failed: %v", err)
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}
```

After:
```go
if err != nil {
    logging.WithContext(r.Context()).Error("Order placement failed", err,
        logging.Component("abook"),
        logging.Symbol(req.Symbol),
        logging.String("order_type", req.Type),
    )
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}
```

## Step 5: Add Audit Logging

Initialize audit logger globally:

```go
var (
    auditLogger *logging.AuditLogger
)

func main() {
    var err error
    auditLogger, err = logging.NewAuditLogger("./logs/audit")
    if err != nil {
        logging.Fatal("Failed to initialize audit logger", err)
    }
    defer auditLogger.Close()

    // ... rest of initialization
}
```

Add audit logging to critical operations:

```go
// In HandlePlaceOrder
func (s *Server) HandlePlaceOrder(w http.ResponseWriter, r *http.Request) {
    // ... existing code

    order, err := s.abookEngine.PlaceOrder(orderReq)
    if err != nil {
        logging.WithContext(r.Context()).Error("Order placement failed", err,
            logging.Component("order"),
            logging.Symbol(req.Symbol),
        )
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Audit log
    auditLogger.LogOrderPlacement(
        r.Context(),
        order.ID,
        req.Symbol,
        req.Side,
        req.Volume,
        order.OpenPrice,
        req.Type,
        req.AccountID,
    )

    // ... rest of response
}
```

## Step 6: Track Authentication Events

In `auth/service.go`:

```go
func (s *Service) Login(username, password string) (string, *User, error) {
    user, err := s.findUser(username)
    if err != nil {
        // Log failed authentication
        auditLogger.LogAuthenticationFailed(
            context.Background(),
            username,
            "", // IP address should come from request context
            "user_not_found",
        )
        return "", nil, err
    }

    if !s.verifyPassword(user, password) {
        auditLogger.LogAuthenticationFailed(
            context.Background(),
            username,
            "",
            "invalid_password",
        )
        return "", nil, errors.New("invalid credentials")
    }

    // Log successful authentication
    auditLogger.LogAuthentication(
        context.Background(),
        user.ID,
        "",
        "password",
    )

    // Generate token
    token, _ := s.generateToken(user)
    return token, user, nil
}
```

## Step 7: Add Performance Monitoring

In database query functions:

```go
func (e *Engine) GetPositions(accountID string) ([]*Position, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        if duration > 100*time.Millisecond {
            logging.LogSlowQuery(
                context.Background(),
                "GetPositions",
                duration,
            )
        }
    }()

    // ... existing query logic
}
```

## Step 8: Add Sensitive Data Masking

Before logging user input:

```go
func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logging.Error("Invalid login request", err,
            logging.Component("auth"),
        )
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Mask sensitive data before logging
    maskedData := logging.MaskSensitiveJSON(fmt.Sprintf("%+v", req))
    logging.Info("Login attempt",
        logging.Component("auth"),
        logging.String("data", maskedData),
    )

    // ... rest of login logic
}
```

## Step 9: Environment Variables

Create `.env` file:

```bash
ENVIRONMENT=production
SENTRY_DSN=https://your-sentry-dsn@sentry.io/project
LOG_LEVEL=INFO
```

## Step 10: Log Aggregation Setup

### For ELK Stack

```yaml
# filebeat.yml
filebeat.inputs:
  - type: log
    enabled: true
    paths:
      - /path/to/logs/trading-engine.log
    json.keys_under_root: true
    json.add_error_key: true

output.elasticsearch:
  hosts: ["localhost:9200"]
  index: "trading-engine-%{+yyyy.MM.dd}"
```

### For CloudWatch

```go
// Add CloudWatch writer
import "github.com/aws/aws-sdk-go/service/cloudwatchlogs"

cloudwatchWriter := newCloudWatchWriter("trading-engine", "production")
multiWriter := logging.NewMultiWriter(rotatingWriter, cloudwatchWriter)
logger := logging.NewLogger(logging.INFO, multiWriter)
```

## Step 11: Monitoring & Alerts

### Set up alerts in your monitoring system:

```yaml
# Example Datadog monitor
name: "High Error Rate - Trading Engine"
type: "metric alert"
query: "avg(last_5m):sum:trading_engine.errors{env:production} > 50"
message: |
  Error rate is above threshold
  @slack-trading-alerts
  @pagerduty-critical
```

### Example Sentry alert rules:

- Alert when error count > 10 in 5 minutes
- Alert on new error types
- Alert on errors affecting > 5 users

## Step 12: Testing

Create test to verify logging:

```go
func TestLogging(t *testing.T) {
    // Capture logs
    var buf bytes.Buffer
    logger := logging.NewLogger(logging.DEBUG, &buf)

    logger.Info("test message",
        logging.String("key", "value"),
    )

    // Verify JSON output
    var entry logging.LogEntry
    if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
        t.Fatal(err)
    }

    assert.Equal(t, "INFO", entry.Level)
    assert.Equal(t, "test message", entry.Message)
    assert.Equal(t, "value", entry.Extra["key"])
}
```

## Deployment Checklist

- [ ] Set ENVIRONMENT=production
- [ ] Set SENTRY_DSN
- [ ] Create logs directory with write permissions
- [ ] Configure log rotation
- [ ] Set up log aggregation (ELK/CloudWatch/Datadog)
- [ ] Configure alert rules
- [ ] Test audit log retention
- [ ] Verify sensitive data masking
- [ ] Test log sampling
- [ ] Monitor disk usage
- [ ] Set up log backup/archival

## Common Issues

### Issue: Logs not appearing in Sentry

**Solution**: Check SENTRY_DSN is set and Sentry hook is added:
```go
sentryHook, err := logging.NewSentryHook(os.Getenv("SENTRY_DSN"), "production")
if err != nil {
    logging.Error("Failed to initialize Sentry", err)
}
logger.AddHook(sentryHook)
```

### Issue: Disk filling up

**Solution**: Enable log rotation and compression:
```go
rotatingWriter, _ := logging.NewRotatingFileWriter(logging.RotationConfig{
    Filename:           "./logs/app.log",
    MaxSizeMB:          50,  // Smaller files
    MaxBackups:         10,  // Fewer backups
    CompressionEnabled: true,
})
```

### Issue: Too many logs

**Solution**: Enable sampling in production:
```go
logger.EnableSampling(0.05, true) // Keep 5% of INFO logs
```

## Best Practices Summary

1. Always include context (request ID, user ID, account ID)
2. Use structured fields instead of string formatting
3. Log at appropriate levels
4. Enable sampling in production
5. Use audit logger for compliance events
6. Mask sensitive data
7. Set up alerts for critical errors
8. Monitor slow queries and endpoints
9. Rotate logs to prevent disk fill
10. Test logging in development before deploying
