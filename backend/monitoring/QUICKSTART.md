# Quick Start Guide - Trading Engine Monitoring

Get production monitoring up and running in 5 minutes.

## Step 1: Dependencies Already Installed ✅

The required Prometheus packages are already in your `go.mod`:
```
github.com/prometheus/client_golang v1.23.2
```

## Step 2: Add Monitoring to main.go

Edit `/Users/epic1st/Documents/trading engine/backend/cmd/server/main.go`:

```go
package main

import (
    // ... existing imports ...
    "github.com/epic1st/rtx/backend/monitoring"
)

func main() {
    // STEP 1: Initialize monitoring system
    monitoring.InitializeMonitoring("v3.0.0")

    // STEP 2: Start runtime metrics collector
    runtimeCollector := monitoring.NewRuntimeMetricsCollector(30 * time.Second)
    go runtimeCollector.Start()

    // STEP 3: Register custom health checks
    hc := monitoring.GetHealthChecker()

    // Database health check
    hc.RegisterCheck("database", func() monitoring.ComponentHealth {
        // TODO: Add your database connection check
        return monitoring.ComponentHealth{
            Status:      monitoring.StatusHealthy,
            Message:     "Database connected",
            LastChecked: time.Now(),
        }
    })

    // LP connectivity health check
    hc.RegisterCheck("lp_connectivity", func() monitoring.ComponentHealth {
        lpConnected := true // TODO: Check actual LP status
        status := monitoring.StatusHealthy
        if !lpConnected {
            status = monitoring.StatusUnhealthy
        }
        return monitoring.ComponentHealth{
            Status:      status,
            Message:     "LP connection status",
            LastChecked: time.Now(),
        }
    })

    // WebSocket health check
    hc.RegisterCheck("websocket", func() monitoring.ComponentHealth {
        // TODO: Get actual connection count from hub
        activeConnections := 0
        status := monitoring.StatusHealthy
        if activeConnections == 0 {
            status = monitoring.StatusDegraded
        }
        return monitoring.ComponentHealth{
            Status:      status,
            Message:     "WebSocket running",
            LastChecked: time.Now(),
            Metadata: map[string]interface{}{
                "active_connections": activeConnections,
            },
        }
    })

    // ... your existing server setup ...

    // STEP 4: Register monitoring endpoints
    http.Handle("/metrics", monitoring.NewMetricsCollector().Handler())
    http.HandleFunc("/health", hc.HTTPHealthHandler())
    http.HandleFunc("/ready", hc.HTTPReadinessHandler())

    log.Println("🎯 Monitoring endpoints registered:")
    log.Println("  📊 Metrics:     http://localhost:7999/metrics")
    log.Println("  ❤️  Health:      http://localhost:7999/health")
    log.Println("  ✅ Readiness:   http://localhost:7999/ready")

    // ... existing code ...
    http.ListenAndServe(":7999", nil)
}
```

## Step 3: Add Monitoring to Order Execution

Find your order execution code and add:

```go
func (s *Server) HandlePlaceOrder(w http.ResponseWriter, r *http.Request) {
    // Start tracing
    span := monitoring.TraceOrderExecution("", "", "MARKET")
    defer span.Finish()

    startTime := time.Now()

    // ... your existing order execution code ...

    // Record metrics
    latencyMs := float64(time.Since(startTime).Milliseconds())
    monitoring.RecordOrderExecution(
        req.Type,           // "MARKET", "LIMIT", etc.
        req.Symbol,         // "EURUSD"
        "ABOOK",           // or "BBOOK"
        latencyMs,
        err == nil,
    )

    if err == nil {
        monitoring.RecordTradeVolume(req.Symbol, req.Side, "ABOOK", req.Volume)
    }
}
```

## Step 4: Add WebSocket Connection Tracking

In your WebSocket hub code:

```go
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
            // Track WebSocket connections
            monitoring.SetWebSocketConnections(len(h.clients))

        case client := <-h.unregister:
            delete(h.clients, client)
            // Update connection count
            monitoring.SetWebSocketConnections(len(h.clients))
        }
    }
}
```

## Step 5: Add LP Monitoring

In your LP connection code:

```go
// When LP connects
monitoring.SetLPConnected("OANDA", "FIX", true)

// When LP disconnects
monitoring.SetLPConnected("OANDA", "FIX", false)

// On quote reception
func OnQuoteReceived(lpName, symbol string) {
    startTime := time.Now()

    // ... process quote ...

    latencyMs := float64(time.Since(startTime).Milliseconds())
    monitoring.RecordLPLatency(lpName, "quote", latencyMs)
    monitoring.RecordLPQuote(lpName, symbol)
}
```

## Step 6: Test It

```bash
# Build and run
cd backend
go build -o server cmd/server/main.go
./server

# In another terminal, test endpoints:

# Check metrics
curl http://localhost:7999/metrics | head -50

# Check health
curl http://localhost:7999/health | jq

# Check readiness
curl http://localhost:7999/ready | jq
```

Expected output:

```json
{
  "status": "healthy",
  "timestamp": "2026-01-18T15:00:00Z",
  "uptime_seconds": 120,
  "version": "v3.0.0",
  "components": {
    "memory": {
      "status": "healthy",
      "message": "Memory usage normal",
      "last_checked": "2026-01-18T15:00:00Z"
    },
    "goroutines": {
      "status": "healthy",
      "message": "Goroutine count normal",
      "last_checked": "2026-01-18T15:00:00Z"
    }
  }
}
```

## Step 7: Set Up Prometheus (Optional)

Create `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'trading-engine'
    static_configs:
      - targets: ['host.docker.internal:7999']
```

Run Prometheus:

```bash
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v $(pwd)/monitoring/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus

# Access at http://localhost:9090
```

## Step 8: Set Up Grafana (Optional)

```bash
docker run -d \
  --name grafana \
  -p 3000:3000 \
  grafana/grafana

# Access at http://localhost:3000 (admin/admin)
```

Then:
1. Add Prometheus data source: `http://host.docker.internal:9090`
2. Import dashboard from `monitoring/grafana_dashboard.json`

## Verify Everything Works

### Test Order Execution Metrics

```bash
# Place a test order
curl -X POST http://localhost:7999/order \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "EURUSD",
    "side": "BUY",
    "volume": 1.0,
    "type": "MARKET"
  }'

# Check metrics
curl http://localhost:7999/metrics | grep trading_order_execution_latency
```

### Test Health Checks

```bash
# Liveness probe
curl http://localhost:7999/health

# Readiness probe
curl http://localhost:7999/ready
```

### View Logs

Logs are output in JSON format to stdout:

```json
{
  "timestamp": "2026-01-18T15:00:00.123Z",
  "level": "INFO",
  "service": "trading-engine",
  "message": "Order placed",
  "fields": {
    "order_id": "ORD-12345",
    "symbol": "EURUSD",
    "volume": 1.0,
    "execution_time_ms": 45
  }
}
```

## Common Metrics to Watch

Query these in Prometheus:

```promql
# Order execution latency (P95)
histogram_quantile(0.95, rate(trading_order_execution_latency_milliseconds_bucket[5m]))

# Order success rate
rate(trading_orders_total{status="success"}[5m]) / rate(trading_orders_total[5m])

# Active WebSocket connections
trading_websocket_connections

# LP connectivity
trading_lp_connected

# Memory usage
trading_memory_usage_bytes / (1024 * 1024)
```

## Troubleshooting

### Metrics endpoint returns empty

**Solution**: Make sure you've instrumented your code with `RecordOrderExecution()`, etc.

### Health check fails

**Solution**: Check that all registered health checks are returning valid status

### No data in Grafana

**Solution**:
1. Verify Prometheus is scraping: `http://localhost:9090/targets`
2. Check Prometheus data source in Grafana
3. Ensure time range is correct

## Next Steps

1. ✅ Basic monitoring working
2. 📧 Set up Alertmanager for notifications
3. 📊 Create custom dashboards
4. 🔍 Add more detailed tracing
5. 📝 Set up log aggregation

## Resources

- Full documentation: `monitoring/README.md`
- Installation guide: `monitoring/INSTALLATION.md`
- Alert rules: `monitoring/prometheus_alerts.yml`
- Dashboard: `monitoring/grafana_dashboard.json`
- Examples: `monitoring/integration_example.go`

## Support

If you encounter issues:
1. Check the logs for errors
2. Verify all imports are correct
3. Ensure Prometheus packages are installed
4. Review the full documentation

---

**You now have production-grade monitoring! 🎉**
