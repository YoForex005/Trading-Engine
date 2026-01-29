# Production-Grade Logging System

Structured JSON logging with error tracking, audit trails, and performance monitoring for the trading engine.

## Features

- **Structured JSON Logging**: Compatible with ELK Stack, Datadog, CloudWatch
- **Multiple Log Levels**: DEBUG, INFO, WARN, ERROR, FATAL
- **Request ID Tracking**: Correlation across distributed services
- **User/Account Context**: Automatic user and account tracking
- **Performance Logging**: Slow query and slow endpoint detection
- **Error Aggregation**: Track error patterns and frequencies
- **Audit Trail**: Complete audit log for compliance
- **Sentry Integration**: Automatic error reporting to Sentry
- **Sensitive Data Masking**: Automatic PII/credentials masking
- **Log Rotation**: Automatic rotation based on size and time
- **Log Sampling**: Reduce log volume in production

## Quick Start

### Basic Usage

```go
import "github.com/epic1st/rtx/backend/logging"

// Simple logging
logging.Info("Server started", logging.String("port", "7999"))
logging.Error("Failed to connect", err, logging.String("service", "database"))

// With context
ctx := logging.ContextWithRequestID(r.Context(), "req-123")
ctx = logging.ContextWithUserID(ctx, "user-456")
logging.WithContext(ctx).Info("Order placed",
    logging.OrderID(order.ID),
    logging.Symbol("EURUSD"),
    logging.Float64("volume", 1.5),
)
```

### HTTP Middleware

```go
import (
    "github.com/epic1st/rtx/backend/logging"
)

func main() {
    logger := logging.NewLogger(logging.INFO)

    // Create middleware
    loggingMiddleware := logging.HTTPLoggingMiddleware(logger)
    panicMiddleware := logging.PanicRecoveryMiddleware(logger)

    // Wrap your handlers
    http.Handle("/api/orders",
        panicMiddleware(
            loggingMiddleware(
                http.HandlerFunc(handleOrders),
            ),
        ),
    )

    http.ListenAndServe(":7999", nil)
}
```

### Audit Logging

```go
import "github.com/epic1st/rtx/backend/logging"

func main() {
    // Initialize audit logger
    auditLogger, err := logging.NewAuditLogger("./logs/audit")
    if err != nil {
        panic(err)
    }
    defer auditLogger.Close()

    // Log order placement
    auditLogger.LogOrderPlacement(
        ctx,
        order.ID,
        "EURUSD",
        "BUY",
        1.5,
        1.1234,
        "MARKET",
        account.ID,
    )

    // Log authentication
    auditLogger.LogAuthentication(ctx, user.ID, "192.168.1.1", "password")

    // Log admin action
    auditLogger.LogAdminAction(
        ctx,
        admin.ID,
        "update_leverage",
        "account",
        account.ID,
        map[string]interface{}{"leverage": 50},
        map[string]interface{}{"leverage": 100},
    )
}
```

### Sentry Integration

```go
import "github.com/epic1st/rtx/backend/logging"

func main() {
    logger := logging.NewLogger(logging.INFO)

    // Add Sentry hook
    sentryHook, err := logging.NewSentryHook(
        "https://your-sentry-dsn@sentry.io/project",
        "production",
    )
    if err != nil {
        panic(err)
    }

    logger.AddHook(sentryHook)

    // Errors will automatically be sent to Sentry
    logger.Error("Database connection failed", err,
        logging.Component("database"),
        logging.String("host", "localhost"),
    )
}
```

### Performance Monitoring

```go
import (
    "time"
    "github.com/epic1st/rtx/backend/logging"
)

func executeQuery(ctx context.Context, query string) error {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        // Automatically logs if > 100ms
        logging.LogSlowQuery(ctx, query, duration)
    }()

    // Execute query
    return db.Query(query)
}
```

### Log Rotation

```go
import "github.com/epic1st/rtx/backend/logging"

func main() {
    // Create rotating file writer
    rotatingWriter, err := logging.NewRotatingFileWriter(logging.RotationConfig{
        Filename:           "./logs/app.log",
        MaxSizeMB:          100,              // 100MB per file
        MaxAge:             7 * 24 * time.Hour, // 7 days
        MaxBackups:         10,                // Keep 10 backups
        CompressionEnabled: true,              // Gzip old logs
    })
    if err != nil {
        panic(err)
    }
    defer rotatingWriter.Close()

    // Create logger with rotating file
    logger := logging.NewLogger(logging.INFO, rotatingWriter)

    // Logs will automatically rotate
    logger.Info("Application started")
}
```

### Sensitive Data Masking

```go
import "github.com/epic1st/rtx/backend/logging"

func handleLogin(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }

    json.NewDecoder(r.Body).Decode(&req)

    // Mask sensitive data before logging
    maskedData := logging.MaskSensitiveJSON(fmt.Sprintf("%+v", req))
    logging.Info("Login attempt", logging.String("data", maskedData))
    // Output: {"password": "[REDACTED]", "username": "john"}
}
```

## Log Levels

| Level | When to Use | Production |
|-------|-------------|-----------|
| DEBUG | Development debugging, verbose output | Disabled |
| INFO  | Normal operations, state changes | Sampled |
| WARN  | Unexpected but handled situations | Enabled |
| ERROR | Errors requiring attention | Enabled + Sentry |
| FATAL | Critical errors, application exits | Enabled + Sentry |

## Production Configuration

### Recommended Settings

```go
// Production configuration
logger := logging.NewLogger(logging.INFO)

// Enable log sampling to reduce volume
logger.EnableSampling(0.1, true) // Keep 10% of INFO logs, all errors

// Add Sentry for error tracking
sentryHook, _ := logging.NewSentryHook(os.Getenv("SENTRY_DSN"), "production")
logger.AddHook(sentryHook)

// Use rotating file writer
rotatingWriter, _ := logging.NewRotatingFileWriter(logging.RotationConfig{
    Filename:   "./logs/production.log",
    MaxSizeMB:  100,
    MaxAge:     7 * 24 * time.Hour,
    MaxBackups: 30,
})

// Multi-writer: file + stdout
multiWriter := logging.NewMultiWriter(rotatingWriter, os.Stdout)
logger = logging.NewLogger(logging.INFO, multiWriter)
```

### Environment Variables

```bash
ENVIRONMENT=production   # Sets environment tag in logs
SENTRY_DSN=https://...   # Sentry integration
LOG_LEVEL=INFO           # Minimum log level
```

## Audit Trail Events

All audit events are compliance-ready and include:

- **Order Operations**: Placement, cancellation, modification
- **Position Operations**: Open, close, modify
- **Authentication**: Login, logout, failed attempts
- **Admin Actions**: All administrative changes
- **Account Operations**: Creation, modification, deposits, withdrawals
- **LP Routing**: Routing decisions and reasons
- **Configuration**: All config changes

### Audit Log Format

```json
{
  "event_id": "audit-1642512345000000000",
  "timestamp": "2026-01-18T10:30:45.123Z",
  "event_type": "order_placement",
  "user_id": "user-123",
  "account_id": "acc-456",
  "ip_address": "192.168.1.100",
  "action": "place_order",
  "resource": "order",
  "resource_id": "ord-789",
  "status": "success",
  "metadata": {
    "symbol": "EURUSD",
    "side": "BUY",
    "volume": 1.5,
    "price": 1.1234,
    "order_type": "MARKET"
  },
  "compliance": true,
  "environment": "production",
  "request_id": "req-abc-123"
}
```

## Common Log Queries

### ELK/Elasticsearch Queries

```
# Find all failed logins in last 24h
event_type:authentication_failed AND timestamp:[now-24h TO now]

# Find all orders for a specific account
event_type:order_placement AND account_id:"acc-456"

# Find slow endpoints (>1s)
duration_ms:>1000 AND level:WARN

# Find all errors from a specific component
level:ERROR AND component:"database"

# Find all admin actions
event_type:admin_action AND timestamp:[now-7d TO now]
```

### CloudWatch Insights Queries

```
# Error rate by component
fields @timestamp, component, message
| filter level = "ERROR"
| stats count() by component

# Slow queries
fields @timestamp, query, duration_ms
| filter duration_ms > 100
| sort duration_ms desc

# Authentication failures by IP
fields @timestamp, ip_address, reason
| filter event_type = "authentication_failed"
| stats count() by ip_address
```

## Alert Rules

### Recommended Alerts

1. **Error Rate**: Alert if error rate > 5% for 5 minutes
2. **Failed Logins**: Alert if > 10 failed logins from same IP in 5 minutes
3. **Slow Queries**: Alert if > 20 slow queries per minute
4. **Slow Endpoints**: Alert if P95 latency > 2s
5. **Order Failures**: Alert if order failure rate > 1%

## Integration Examples

### With existing code

```go
// Replace standard log calls
// Before:
log.Printf("[ORDER] Executing %s", order.ID)

// After:
logging.Info("Executing order",
    logging.OrderID(order.ID),
    logging.Symbol(order.Symbol),
    logging.Float64("volume", order.Volume),
)
```

### Error tracking

```go
// Before:
if err != nil {
    log.Printf("Error: %v", err)
    return err
}

// After:
if err != nil {
    logging.Error("Database query failed", err,
        logging.Component("database"),
        logging.String("operation", "insert"),
    )
    logging.TrackError(ctx, err, "high", map[string]interface{}{
        "query": query,
        "params": params,
    })
    return err
}
```

## Best Practices

1. **Always use structured fields** instead of string formatting
2. **Include context** (request ID, user ID, account ID) in all logs
3. **Log at appropriate levels** - don't log everything as ERROR
4. **Mask sensitive data** before logging
5. **Use audit logger** for compliance-critical events
6. **Monitor slow queries** and slow endpoints
7. **Set up alerts** for critical errors
8. **Sample INFO logs** in production to reduce volume
9. **Rotate logs** to prevent disk fill-up
10. **Use Sentry** for real-time error tracking

## Performance Impact

- **Overhead**: < 1ms per log entry
- **Sampling**: Reduces volume by 90% with 0.1 rate
- **Async hooks**: Sentry integration doesn't block logging
- **Buffered audit logs**: Batch writes for performance

## License

Internal use only - Trading Engine Backend
