# Trading Engine Monitoring & Observability

Production-grade monitoring, logging, tracing, and alerting for the trading engine.

## Components

### 1. Prometheus Metrics (`prometheus.go`)

Exports comprehensive metrics at `/metrics` endpoint for Prometheus scraping.

**Key Metrics:**

- **Order Execution:**
  - `trading_order_execution_latency_milliseconds` - Histogram (p50, p95, p99)
  - `trading_orders_total` - Counter by type, status, mode
  - `trading_order_errors_total` - Counter by error type

- **WebSocket:**
  - `trading_websocket_connections` - Gauge
  - `trading_websocket_messages_total` - Counter by message type

- **Positions:**
  - `trading_active_positions` - Gauge by symbol, side
  - `trading_position_pnl_usd` - Gauge by account, symbol

- **Trading Volume & P&L:**
  - `trading_volume_lots_total` - Counter by symbol, side, mode
  - `trading_pnl_usd_total` - Counter by account, symbol

- **API Performance:**
  - `trading_api_requests_total` - Counter by endpoint, method, status
  - `trading_api_request_duration_milliseconds` - Histogram

- **Database:**
  - `trading_db_query_duration_milliseconds` - Histogram
  - `trading_db_connections_active` - Gauge

- **LP Connectivity:**
  - `trading_lp_connected` - Gauge (1=connected, 0=disconnected)
  - `trading_lp_latency_milliseconds` - Histogram
  - `trading_lp_quotes_received_total` - Counter

- **Runtime:**
  - `trading_memory_usage_bytes` - Gauge
  - `trading_goroutines_count` - Gauge

- **Account Metrics:**
  - `trading_account_balance_usd` - Gauge
  - `trading_account_equity_usd` - Gauge
  - `trading_account_margin_used_usd` - Gauge

- **SLO Tracking:**
  - `trading_slo_order_execution_success_total` - Counter
  - `trading_slo_order_execution_within_target_total` - Counter (<100ms)

### 2. Structured Logging (`logger.go`)

JSON-formatted structured logging to stdout.

**Features:**
- Multiple log levels: DEBUG, INFO, WARN, ERROR, FATAL
- Structured fields for filtering and analysis
- Automatic source file/line for errors
- Stack traces for fatal errors
- Specialized log methods for orders, trades, performance, security

**Example:**
```go
logger := monitoring.GetLogger()
logger.Info("Order placed", map[string]interface{}{
    "order_id": "12345",
    "symbol": "EURUSD",
    "volume": 1.0,
})
```

### 3. Distributed Tracing (`tracer.go`)

Lightweight distributed tracing with trace and span IDs.

**Features:**
- Trace ID and Span ID generation
- Parent-child span relationships
- Span tags and logs
- Context propagation
- Duration tracking

**Example:**
```go
span := monitoring.TraceOrderExecution("12345", "EURUSD", "MARKET")
defer span.Finish()

span.SetTag("lp_name", "OANDA")
span.LogFields(map[string]interface{}{
    "execution_price": 1.1050,
})
```

### 4. Health Checks (`health.go`)

Liveness and readiness probes for Kubernetes/Docker.

**Endpoints:**
- `GET /health` - Liveness probe (overall health)
- `GET /ready` - Readiness probe (ready to serve traffic)

**Health Statuses:**
- `healthy` - Component functioning normally
- `degraded` - Component working but with issues
- `unhealthy` - Component not working

**Built-in Checks:**
- Memory usage monitoring
- Goroutine count monitoring
- Uptime tracking

**Example:**
```go
hc := monitoring.GetHealthChecker()
hc.RegisterCheck("database", func() monitoring.ComponentHealth {
    // Check database connection
    return monitoring.ComponentHealth{
        Status: monitoring.StatusHealthy,
        Message: "Database connected",
        Metadata: map[string]interface{}{
            "connection_pool": 10,
        },
    }
})
```

### 5. Alerting (`alerts.go`)

Alert rules and thresholds for proactive monitoring.

**Default Alert Rules:**

| Alert | Threshold | Duration | Severity |
|-------|-----------|----------|----------|
| HighOrderLatency | >500ms (p95) | 2min | Warning |
| CriticalOrderLatency | >2000ms (p95) | 1min | Critical |
| HighOrderErrorRate | >5% | 5min | Warning |
| LPDisconnected | 0 connections | 30s | Critical |
| HighLPLatency | >1000ms (p95) | 2min | Warning |
| HighMemoryUsage | >80% | 5min | Warning |
| HighGoroutineCount | >10000 | 5min | Warning |
| HighAPIErrorRate | >5% | 5min | Warning |
| SlowDatabaseQueries | >100ms (p95) | 5min | Warning |
| SLOViolation | <95% within 100ms | 10min | Critical |

### 6. Runtime Metrics Collector (`runtime.go`)

Automatic collection of Go runtime metrics.

**Collects:**
- Memory allocation and usage
- Goroutine count
- Automatic alerting on thresholds

## Integration

### Server Setup

```go
package main

import (
    "net/http"
    "time"
    "github.com/epic1st/rtx/backend/monitoring"
)

func main() {
    // Initialize monitoring
    logger := monitoring.GetLogger()
    logger.SetMinLevel(monitoring.INFO)

    healthChecker := monitoring.GetHealthChecker()
    alertManager := monitoring.GetAlertManager()

    // Register default alert rules
    for _, rule := range monitoring.GetDefaultAlertRules() {
        alertManager.RegisterRule(rule)
    }

    // Register health checks
    healthChecker.RegisterCheck("memory", monitoring.MemoryHealthCheck(80))
    healthChecker.RegisterCheck("goroutines", monitoring.GoroutineHealthCheck(10000))

    // Start runtime metrics collector
    runtimeCollector := monitoring.NewRuntimeMetricsCollector(30 * time.Second)
    go runtimeCollector.Start()

    // Register monitoring endpoints
    http.Handle("/metrics", monitoring.NewMetricsCollector().Handler())
    http.HandleFunc("/health", healthChecker.HTTPHealthHandler())
    http.HandleFunc("/ready", healthChecker.HTTPReadinessHandler())

    // Wrap API endpoints with metrics middleware
    http.HandleFunc("/api/orders", monitoring.APIRequestMiddleware("/api/orders", handleOrders))

    logger.Info("Server started", map[string]interface{}{
        "port": 7999,
    })

    http.ListenAndServe(":7999", nil)
}
```

### Recording Metrics

```go
// Order execution
start := time.Now()
err := placeOrder(order)
latencyMs := float64(time.Since(start).Milliseconds())

monitoring.RecordOrderExecution(
    "MARKET",
    "EURUSD",
    "ABOOK",
    latencyMs,
    err == nil,
)

// WebSocket connection tracking
monitoring.SetWebSocketConnections(hub.ClientCount())

// Position tracking
monitoring.SetActivePositions("EURUSD", "BUY", 5)
monitoring.SetPositionPnL("demo_001", "EURUSD", 150.50)

// LP metrics
monitoring.SetLPConnected("OANDA", "FIX", true)
monitoring.RecordLPLatency("OANDA", "quote", 25.5)
monitoring.RecordLPQuote("OANDA", "EURUSD")

// Account metrics
monitoring.SetAccountBalance("demo_001", "demo", 5000.0)
monitoring.SetAccountEquity("demo_001", 5150.50)
monitoring.SetAccountMarginUsed("demo_001", 500.0)
```

## Prometheus Configuration

**prometheus.yml:**
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'trading-engine'
    static_configs:
      - targets: ['localhost:7999']
    metrics_path: '/metrics'

rule_files:
  - 'alerts.yml'

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['localhost:9093']
```

**alerts.yml:**
```yaml
groups:
  - name: trading_engine
    interval: 30s
    rules:
      - alert: HighOrderLatency
        expr: trading_order_execution_latency_milliseconds{quantile="0.95"} > 500
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High order execution latency"
          description: "P95 latency is {{ $value }}ms"

      - alert: LPDisconnected
        expr: trading_lp_connected == 0
        for: 30s
        labels:
          severity: critical
        annotations:
          summary: "LP connection lost"
          description: "{{ $labels.lp_name }} disconnected"
```

## Grafana Dashboards

Import pre-built dashboards for visualization:

1. **Order Execution Dashboard**
   - Order latency histograms (p50, p95, p99)
   - Order success/failure rates
   - Order volume by symbol
   - Execution mode distribution

2. **System Health Dashboard**
   - Memory usage
   - Goroutine count
   - API request rates
   - Database query performance

3. **LP Connectivity Dashboard**
   - Connection status by LP
   - LP latency metrics
   - Quote reception rates
   - Failover events

4. **SLO Dashboard**
   - Order execution SLO (95% within 100ms)
   - API availability
   - Error budgets

## Kubernetes Deployment

```yaml
apiVersion: v1
kind: Service
metadata:
  name: trading-engine
  labels:
    app: trading-engine
spec:
  ports:
  - port: 7999
    name: http
  - port: 7999
    name: metrics
  selector:
    app: trading-engine

---
apiVersion: v1
kind: Pod
metadata:
  name: trading-engine
  labels:
    app: trading-engine
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "7999"
    prometheus.io/path: "/metrics"
spec:
  containers:
  - name: trading-engine
    image: trading-engine:latest
    ports:
    - containerPort: 7999
      name: http
    livenessProbe:
      httpGet:
        path: /health
        port: 7999
      initialDelaySeconds: 30
      periodSeconds: 10
    readinessProbe:
      httpGet:
        path: /ready
        port: 7999
      initialDelaySeconds: 10
      periodSeconds: 5
```

## Performance Impact

- **Metrics collection:** <1ms per operation
- **Logging:** <0.5ms per log entry
- **Tracing:** <0.1ms per span
- **Health checks:** <5ms per check
- **Memory overhead:** ~10MB

## Best Practices

1. **Use structured logging:** Always include relevant fields
2. **Set appropriate log levels:** DEBUG in dev, INFO in production
3. **Monitor SLOs:** Track order execution latency SLOs
4. **Alert on thresholds:** Set up Prometheus alerting
5. **Regular health checks:** Monitor system health continuously
6. **Trace critical paths:** Use tracing for order execution flows
7. **Dashboard visibility:** Create Grafana dashboards for real-time monitoring

## License

Proprietary - RTX Trading Engine
