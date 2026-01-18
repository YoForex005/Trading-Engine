# Logging System Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     Trading Engine Application                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   HTTP API   │  │  Order Mgmt  │  │  Auth Service│          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
│         │                  │                  │                   │
│         └──────────────────┼──────────────────┘                  │
│                            │                                      │
│                    ┌───────▼────────┐                            │
│                    │ Logging Layer  │                            │
│                    └───────┬────────┘                            │
└────────────────────────────┼─────────────────────────────────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
┌───────────────┐    ┌──────────────┐    ┌──────────────┐
│  Structured   │    │    Audit     │    │    Error     │
│  Logger       │    │    Logger    │    │   Tracker    │
└───────┬───────┘    └──────┬───────┘    └──────┬───────┘
        │                    │                    │
        │                    │                    │
        ▼                    ▼                    ▼
┌───────────────┐    ┌──────────────┐    ┌──────────────┐
│   Outputs     │    │   Outputs    │    │   Outputs    │
│               │    │              │    │              │
│ • File (JSON) │    │ • Audit File │    │ • Alerts     │
│ • Stdout      │    │ • Rotation   │    │ • Sentry     │
│ • CloudWatch  │    │ • Compliance │    │ • Metrics    │
│ • ELK Stack   │    │              │    │              │
│ • Datadog     │    │              │    │              │
│ • Sentry Hook │    │              │    │              │
└───────────────┘    └──────────────┘    └──────────────┘
```

## Component Breakdown

### 1. Structured Logger (`logger.go`)

```
┌─────────────────────────────────────────────────┐
│           Structured Logger Core                │
├─────────────────────────────────────────────────┤
│                                                  │
│  ┌──────────────┐         ┌──────────────┐     │
│  │  Log Levels  │         │   Sampling   │     │
│  │  DEBUG       │         │   Config     │     │
│  │  INFO        │         │              │     │
│  │  WARN        │◄────────┤  • Enabled   │     │
│  │  ERROR       │         │  • Rate      │     │
│  │  FATAL       │         │  • Keep Err  │     │
│  └──────┬───────┘         └──────────────┘     │
│         │                                        │
│         ▼                                        │
│  ┌──────────────────────────────────────┐      │
│  │       Log Entry Builder              │      │
│  │  • Timestamp (UTC)                   │      │
│  │  • Level                             │      │
│  │  • Message                           │      │
│  │  • Context (Request/User/Account)    │      │
│  │  • Fields (Component/Symbol/etc)     │      │
│  │  • Caller Info (File/Line/Function)  │      │
│  │  • Environment/Hostname/PID          │      │
│  └──────┬───────────────────────────────┘      │
│         │                                        │
│         ▼                                        │
│  ┌──────────────────────────────────────┐      │
│  │          Hook System                 │      │
│  │  ┌────────────┐  ┌────────────┐     │      │
│  │  │   Sentry   │  │   Custom   │     │      │
│  │  │    Hook    │  │   Hooks    │     │      │
│  │  └────────────┘  └────────────┘     │      │
│  └──────┬───────────────────────────────┘      │
│         │                                        │
│         ▼                                        │
│  ┌──────────────────────────────────────┐      │
│  │        Output Writers                │      │
│  │  ┌────────┐  ┌────────┐  ┌────────┐ │      │
│  │  │  File  │  │ Stdout │  │ Cloud  │ │      │
│  │  │Rotating│  │        │  │ Watch  │ │      │
│  │  └────────┘  └────────┘  └────────┘ │      │
│  └──────────────────────────────────────┘      │
│                                                  │
└─────────────────────────────────────────────────┘
```

### 2. HTTP Middleware (`middleware.go`)

```
HTTP Request
     │
     ▼
┌─────────────────────────────────┐
│  Panic Recovery Middleware      │
│  • Catch panics                  │
│  • Log stack trace               │
│  • Return 500 error              │
└─────────────┬───────────────────┘
              │
              ▼
┌─────────────────────────────────┐
│  Logging Middleware              │
│  • Generate Request ID           │
│  • Add to context                │
│  • Log request start             │
│  • Wrap response writer          │
└─────────────┬───────────────────┘
              │
              ▼
┌─────────────────────────────────┐
│  Business Logic Handler          │
│  • Process request               │
│  • Use logging.WithContext(ctx) │
└─────────────┬───────────────────┘
              │
              ▼
┌─────────────────────────────────┐
│  Response Capture                │
│  • Capture status code           │
│  • Capture response size         │
│  • Calculate duration            │
└─────────────┬───────────────────┘
              │
              ▼
┌─────────────────────────────────┐
│  Log Response                    │
│  • Status code                   │
│  • Duration                      │
│  • Size                          │
│  • Detect slow requests (>1s)   │
└─────────────┬───────────────────┘
              │
              ▼
HTTP Response
```

### 3. Audit Logger (`audit.go`)

```
┌─────────────────────────────────────────────────┐
│              Audit Logger                        │
├─────────────────────────────────────────────────┤
│                                                  │
│  ┌──────────────────────────────────────┐      │
│  │        Event Types                   │      │
│  │  • Order Placement                   │      │
│  │  • Order Cancellation                │      │
│  │  • Position Close                    │      │
│  │  • Authentication                    │      │
│  │  • Authentication Failed             │      │
│  │  • Admin Action                      │      │
│  │  • Deposit/Withdrawal                │      │
│  │  • LP Routing                        │      │
│  │  • Config Change                     │      │
│  └──────┬───────────────────────────────┘      │
│         │                                        │
│         ▼                                        │
│  ┌──────────────────────────────────────┐      │
│  │     Event Enrichment                 │      │
│  │  • Event ID generation               │      │
│  │  • Timestamp (UTC)                   │      │
│  │  • Context extraction                │      │
│  │  • Before/After capture              │      │
│  │  • Compliance flag                   │      │
│  └──────┬───────────────────────────────┘      │
│         │                                        │
│         ▼                                        │
│  ┌──────────────────────────────────────┐      │
│  │          Buffer                      │      │
│  │  • Buffer size: 100 events           │      │
│  │  • Auto-flush: 5 seconds             │      │
│  │  • Manual flush on buffer full       │      │
│  └──────┬───────────────────────────────┘      │
│         │                                        │
│         ▼                                        │
│  ┌──────────────────────────────────────┐      │
│  │      File Writer                     │      │
│  │  • JSON format                       │      │
│  │  • Append-only                       │      │
│  │  • Auto-rotation (100MB)             │      │
│  │  • Force sync on flush               │      │
│  └──────────────────────────────────────┘      │
│                                                  │
└─────────────────────────────────────────────────┘
```

### 4. Error Tracker (`errors.go`)

```
┌─────────────────────────────────────────────────┐
│             Error Tracker                        │
├─────────────────────────────────────────────────┤
│                                                  │
│  Error Occurs                                    │
│       │                                          │
│       ▼                                          │
│  ┌──────────────────────────────────────┐      │
│  │    Track Error                       │      │
│  │  • Generate error key                │      │
│  │  • Extract type                      │      │
│  │  • Capture context                   │      │
│  │  • Store stack trace                 │      │
│  └──────┬───────────────────────────────┘      │
│         │                                        │
│         ▼                                        │
│  ┌──────────────────────────────────────┐      │
│  │    Error Statistics                  │      │
│  │  ┌────────────────────────────┐     │      │
│  │  │ ErrorStats                 │     │      │
│  │  │ • Error Type               │     │      │
│  │  │ • Message                  │     │      │
│  │  │ • Count                    │     │      │
│  │  │ • First/Last Seen          │     │      │
│  │  │ • Occurrences Timeline     │     │      │
│  │  │ • Affected Users           │     │      │
│  │  │ • Severity                 │     │      │
│  │  │ • Alerted Flag             │     │      │
│  │  └────────────────────────────┘     │      │
│  └──────┬───────────────────────────────┘      │
│         │                                        │
│         ▼                                        │
│  ┌──────────────────────────────────────┐      │
│  │   Threshold Check                    │      │
│  │   critical:  1 occurrence            │      │
│  │   high:      5 occurrences           │      │
│  │   medium:   10 occurrences           │      │
│  │   low:      50 occurrences           │      │
│  └──────┬───────────────────────────────┘      │
│         │                                        │
│         ▼                                        │
│  ┌──────────────────────────────────────┐      │
│  │   Alert Callbacks                    │      │
│  │  • Send to PagerDuty                 │      │
│  │  • Post to Slack                     │      │
│  │  • Send email                        │      │
│  │  • Update metrics                    │      │
│  └──────────────────────────────────────┘      │
│                                                  │
│  ┌──────────────────────────────────────┐      │
│  │   Cleanup Loop (5 min)               │      │
│  │  • Remove errors older than 1 hour   │      │
│  │  • Keep recent statistics            │      │
│  └──────────────────────────────────────┘      │
│                                                  │
└─────────────────────────────────────────────────┘
```

### 5. Performance Monitoring (`performance.go`)

```
┌─────────────────────────────────────────────────┐
│       Performance Monitoring                     │
├─────────────────────────────────────────────────┤
│                                                  │
│  ┌──────────────────────────────────────┐      │
│  │   Slow Query Detection               │      │
│  │                                       │      │
│  │   Database Query Execution            │      │
│  │         │                             │      │
│  │         ├─ Start Timer                │      │
│  │         ├─ Execute Query              │      │
│  │         ├─ Calculate Duration         │      │
│  │         │                             │      │
│  │         └─ If duration > 100ms:       │      │
│  │            • Log slow query           │      │
│  │            • Store in history         │      │
│  │            • Track query pattern      │      │
│  │            • Alert if threshold met   │      │
│  └──────────────────────────────────────┘      │
│                                                  │
│  ┌──────────────────────────────────────┐      │
│  │   Slow Endpoint Detection            │      │
│  │                                       │      │
│  │   HTTP Request Processing             │      │
│  │         │                             │      │
│  │         ├─ Start Timer                │      │
│  │         ├─ Process Request            │      │
│  │         ├─ Calculate Duration         │      │
│  │         │                             │      │
│  │         └─ If duration > 1000ms:      │      │
│  │            • Log slow endpoint        │      │
│  │            • Store in history         │      │
│  │            • Track by path            │      │
│  │            • Alert if threshold met   │      │
│  └──────────────────────────────────────┘      │
│                                                  │
│  ┌──────────────────────────────────────┐      │
│  │   Performance Metrics                │      │
│  │  • Last 100 slow queries              │      │
│  │  • Last 100 slow endpoints            │      │
│  │  • Average duration by query          │      │
│  │  • Average duration by endpoint       │      │
│  │  • P95/P99 latency                    │      │
│  └──────────────────────────────────────┘      │
│                                                  │
└─────────────────────────────────────────────────┘
```

### 6. Sensitive Data Masking (`masking.go`)

```
┌─────────────────────────────────────────────────┐
│         Sensitive Data Masker                    │
├─────────────────────────────────────────────────┤
│                                                  │
│  Input String/JSON/Map                          │
│         │                                        │
│         ▼                                        │
│  ┌──────────────────────────────────────┐      │
│  │   Pattern Matching                   │      │
│  │                                       │      │
│  │   Regex Patterns:                    │      │
│  │   • Email addresses                  │      │
│  │   • Phone numbers                    │      │
│  │   • SSN                              │      │
│  │   • Credit cards                     │      │
│  │   • API keys                         │      │
│  │   • Passwords                        │      │
│  │   • Bearer tokens                    │      │
│  │   • JWT tokens                       │      │
│  └──────┬───────────────────────────────┘      │
│         │                                        │
│         ▼                                        │
│  ┌──────────────────────────────────────┐      │
│  │   Masking Strategy                   │      │
│  │                                       │      │
│  │   Email:       j***n@example.com     │      │
│  │   Phone:       XXX-XXX-XXXX          │      │
│  │   SSN:         XXX-XX-XXXX           │      │
│  │   Credit Card: XXXX-XXXX-XXXX-1234   │      │
│  │   API Key:     [REDACTED]            │      │
│  │   Password:    [REDACTED]            │      │
│  │   Token:       [REDACTED]            │      │
│  │   JWT:         [JWT_REDACTED]        │      │
│  └──────┬───────────────────────────────┘      │
│         │                                        │
│         ▼                                        │
│  Masked Output                                  │
│                                                  │
└─────────────────────────────────────────────────┘
```

### 7. Log Rotation (`rotation.go`)

```
┌─────────────────────────────────────────────────┐
│           Log Rotation System                    │
├─────────────────────────────────────────────────┤
│                                                  │
│  ┌──────────────────────────────────────┐      │
│  │   Rotation Triggers                  │      │
│  │                                       │      │
│  │   Check on Every Write:              │      │
│  │   ├─ Size > MaxSize (100MB)?         │      │
│  │   └─ Age > MaxAge (7 days)?          │      │
│  └──────┬───────────────────────────────┘      │
│         │                                        │
│         ▼                                        │
│  ┌──────────────────────────────────────┐      │
│  │   Rotation Process                   │      │
│  │   1. Close current file              │      │
│  │   2. Rename to backup with timestamp │      │
│  │   3. Compress backup (optional)      │      │
│  │   4. Open new file                   │      │
│  │   5. Reset counters                  │      │
│  └──────┬───────────────────────────────┘      │
│         │                                        │
│         ▼                                        │
│  ┌──────────────────────────────────────┐      │
│  │   Backup Management                  │      │
│  │   • Keep MaxBackups files            │      │
│  │   • Delete oldest backups            │      │
│  │   • Delete files older than MaxAge   │      │
│  │   • Cleanup runs every 1 hour        │      │
│  └──────────────────────────────────────┘      │
│                                                  │
│  File Timeline:                                 │
│  app.log              (current)                 │
│  app.log.20260118-143000 (backup)              │
│  app.log.20260117-120000 (backup)              │
│  app.log.20260116-100000 (backup, compressed)  │
│                                                  │
└─────────────────────────────────────────────────┘
```

## Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    Application Code                              │
└───────────────┬─────────────────────────────────────────────────┘
                │
                │  logging.Info("message", fields...)
                │  logging.Error("error", err, fields...)
                │  auditLogger.LogOrderPlacement(...)
                │
                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Logging Layer                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐   │
│  │   Structured   │  │     Audit      │  │     Error      │   │
│  │    Logger      │  │    Logger      │  │    Tracker     │   │
│  └───────┬────────┘  └───────┬────────┘  └───────┬────────┘   │
│          │                    │                    │             │
└──────────┼────────────────────┼────────────────────┼─────────────┘
           │                    │                    │
           │                    │                    │
     ┌─────┼────────────────────┼────────────────────┼─────┐
     │     │                    │                    │     │
     │     ▼                    ▼                    ▼     │
     │  ┌──────────┐      ┌──────────┐      ┌──────────┐ │
     │  │Sensitive │      │  Buffer  │      │  Alert   │ │
     │  │  Data    │      │  Events  │      │Threshold │ │
     │  │ Masking  │      │   (100)  │      │  Check   │ │
     │  └─────┬────┘      └─────┬────┘      └─────┬────┘ │
     │        │                  │                  │      │
     │        ▼                  ▼                  ▼      │
     │  ┌──────────┐      ┌──────────┐      ┌──────────┐ │
     │  │  Hooks   │      │  Flush   │      │Callbacks │ │
     │  │ (Sentry) │      │  (5sec)  │      │          │ │
     │  └─────┬────┘      └─────┬────┘      └─────┬────┘ │
     │        │                  │                  │      │
     └────────┼──────────────────┼──────────────────┼──────┘
              │                  │                  │
              ▼                  ▼                  ▼
        ┌──────────┐       ┌──────────┐      ┌──────────┐
        │  Sentry  │       │   File   │      │ Slack/   │
        │   API    │       │ Rotating │      │PagerDuty │
        └──────────┘       └─────┬────┘      └──────────┘
                                 │
                                 ▼
                           ┌──────────┐
                           │Rotation? │
                           └─────┬────┘
                                 │
                           Yes   │   No
                           ┌─────┴─────┐
                           ▼           ▼
                     ┌──────────┐  Continue
                     │  Rotate  │
                     │• Rename  │
                     │• Compress│
                     │• Cleanup │
                     └──────────┘
```

## Integration Points

### 1. Main Application
```go
func main() {
    setupLogging()
    setupAuditLogging()
    setupErrorTracking()
    // ... rest of application
}
```

### 2. HTTP Handlers
```go
http.Handle("/api/orders",
    panicMiddleware(
        loggingMiddleware(
            handleOrders)))
```

### 3. Business Logic
```go
func PlaceOrder(ctx context.Context, order Order) error {
    logging.WithContext(ctx).Info("Placing order", fields...)
    auditLogger.LogOrderPlacement(ctx, ...)
    // ... business logic
}
```

### 4. Error Handling
```go
if err != nil {
    logging.Error("Operation failed", err, fields...)
    logging.TrackError(ctx, err, "high", extraData)
    return err
}
```

## Deployment Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Production Environment                      │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌────────────────────────────────────────────────────┐    │
│  │           Trading Engine Application                │    │
│  │  • Logs to rotating files                          │    │
│  │  • Sends errors to Sentry                          │    │
│  │  • Writes audit trail                              │    │
│  └───────────┬────────────────────────────────────────┘    │
│              │                                               │
│              ▼                                               │
│  ┌────────────────────────────────────────────────────┐    │
│  │                  Log Files                          │    │
│  │  ./logs/production.log           (application)     │    │
│  │  ./logs/audit/audit.log          (compliance)      │    │
│  │  ./logs/production.log.20260118  (rotated)         │    │
│  └───────────┬────────────────────────────────────────┘    │
│              │                                               │
└──────────────┼───────────────────────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────────────────────┐
│              Log Aggregation Layer                            │
├──────────────────────────────────────────────────────────────┤
│                                                                │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐   │
│  │   Filebeat    │  │  CloudWatch   │  │    Datadog    │   │
│  │   Agent       │  │    Agent      │  │     Agent     │   │
│  └───────┬───────┘  └───────┬───────┘  └───────┬───────┘   │
│          │                   │                   │            │
└──────────┼───────────────────┼───────────────────┼────────────┘
           │                   │                   │
           ▼                   ▼                   ▼
    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
    │ Elasticsearch│    │  CloudWatch │    │   Datadog   │
    │ + Kibana     │    │   Insights  │    │     APM     │
    └─────────────┘    └─────────────┘    └─────────────┘
```

## Summary

The logging architecture provides:

1. **Separation of Concerns**: Structured logs, audit trail, and error tracking are separate but integrated
2. **Performance**: Buffering, sampling, and asynchronous hooks minimize performance impact
3. **Compliance**: Comprehensive audit trail with immutable logs
4. **Observability**: Integration with multiple monitoring platforms
5. **Reliability**: Automatic rotation, error tracking, and alerting
6. **Security**: Sensitive data masking throughout
7. **Scalability**: Sampling and rotation prevent resource exhaustion

Total architecture supports production-grade logging with full compliance and monitoring capabilities.
