# Monitoring Installation Guide

## Prerequisites

Go 1.21 or higher

## Installation Steps

### 1. Install Dependencies

```bash
cd backend

# Install Prometheus client library
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promauto
go get github.com/prometheus/client_golang/prometheus/promhttp

# Update go.mod and go.sum
go mod tidy
```

### 2. Import Monitoring Package

Update your `main.go`:

```go
import (
    "github.com/epic1st/rtx/backend/monitoring"
)

func main() {
    // Initialize monitoring at startup
    monitoring.InitializeMonitoring("v3.0.0")

    // Register monitoring endpoints
    http.Handle("/metrics", monitoring.NewMetricsCollector().Handler())
    http.HandleFunc("/health", monitoring.GetHealthChecker().HTTPHealthHandler())
    http.HandleFunc("/ready", monitoring.GetHealthChecker().HTTPReadinessHandler())

    // Start runtime metrics collector
    runtimeCollector := monitoring.NewRuntimeMetricsCollector(30 * time.Second)
    go runtimeCollector.Start()

    // Register health checks
    hc := monitoring.GetHealthChecker()
    hc.RegisterCheck("memory", monitoring.MemoryHealthCheck(80))
    hc.RegisterCheck("goroutines", monitoring.GoroutineHealthCheck(10000))

    // Your existing server code...
    log.Println("Server with monitoring started on :7999")
    http.ListenAndServe(":7999", nil)
}
```

### 3. Wrap API Handlers

Wrap your existing handlers with monitoring:

```go
// Before
http.HandleFunc("/api/orders", handleOrders)

// After
http.HandleFunc("/api/orders", monitoring.WrapHandlerWithMonitoring("/api/orders", handleOrders))
```

### 4. Add Metrics to Order Execution

In your order execution code:

```go
import "github.com/epic1st/rtx/backend/monitoring"

func PlaceOrder(order *Order) error {
    // Start tracing
    span := monitoring.TraceOrderExecution(order.ID, order.Symbol, order.Type)
    defer span.Finish()

    startTime := time.Now()

    // Your order execution logic...
    err := executeOrder(order)

    // Record metrics
    latencyMs := float64(time.Since(startTime).Milliseconds())
    monitoring.RecordOrderExecution(
        order.Type,
        order.Symbol,
        "ABOOK",
        latencyMs,
        err == nil,
    )

    if err == nil {
        monitoring.RecordTradeVolume(order.Symbol, order.Side, "ABOOK", order.Volume)
    }

    return err
}
```

### 5. Add LP Monitoring

In your LP connection code:

```go
func ConnectLP(lpName string) error {
    // Track connection
    monitoring.SetLPConnected(lpName, "FIX", false)

    err := connectToLP(lpName)

    if err == nil {
        monitoring.SetLPConnected(lpName, "FIX", true)
    }

    return err
}

func OnLPQuote(lpName, symbol string, quote *Quote) {
    startTime := time.Now()

    // Process quote...

    latencyMs := float64(time.Since(startTime).Milliseconds())
    monitoring.RecordLPLatency(lpName, "quote", latencyMs)
    monitoring.RecordLPQuote(lpName, symbol)
}
```

### 6. Add WebSocket Monitoring

In your WebSocket hub:

```go
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
            monitoring.SetWebSocketConnections(len(h.clients))

        case client := <-h.unregister:
            delete(h.clients, client)
            monitoring.SetWebSocketConnections(len(h.clients))
        }
    }
}
```

## Verification

### 1. Check Metrics Endpoint

```bash
curl http://localhost:7999/metrics
```

You should see Prometheus-formatted metrics:

```
# HELP trading_order_execution_latency_milliseconds Order execution latency in milliseconds (p50, p95, p99)
# TYPE trading_order_execution_latency_milliseconds histogram
trading_order_execution_latency_milliseconds_bucket{execution_mode="ABOOK",order_type="MARKET",symbol="EURUSD",le="1"} 0
trading_order_execution_latency_milliseconds_bucket{execution_mode="ABOOK",order_type="MARKET",symbol="EURUSD",le="5"} 15
...
```

### 2. Check Health Endpoint

```bash
curl http://localhost:7999/health
```

Response:

```json
{
  "status": "healthy",
  "timestamp": "2024-01-18T10:30:00Z",
  "uptime_seconds": 3600,
  "version": "v3.0.0",
  "components": {
    "memory": {
      "status": "healthy",
      "message": "Memory usage normal",
      "last_checked": "2024-01-18T10:30:00Z",
      "metadata": {
        "used_mb": 150.5,
        "total_mb": 512,
        "usage_percent": 29.4
      }
    }
  }
}
```

### 3. Check Readiness Endpoint

```bash
curl http://localhost:7999/ready
```

## Docker Integration

Update your `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o trading-engine cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/trading-engine .

# Expose ports
EXPOSE 7999

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:7999/health || exit 1

CMD ["./trading-engine"]
```

## Kubernetes Integration

Create `k8s-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: trading-engine
  labels:
    app: trading-engine
spec:
  replicas: 3
  selector:
    matchLabels:
      app: trading-engine
  template:
    metadata:
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
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 7999
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
```

## Prometheus Setup

Create `prometheus.yml`:

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
  - '/etc/prometheus/prometheus_alerts.yml'

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['localhost:9093']
```

Start Prometheus:

```bash
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v $(pwd)/monitoring/prometheus.yml:/etc/prometheus/prometheus.yml \
  -v $(pwd)/monitoring/prometheus_alerts.yml:/etc/prometheus/prometheus_alerts.yml \
  prom/prometheus
```

Access Prometheus UI at `http://localhost:9090`

## Grafana Setup

Start Grafana:

```bash
docker run -d \
  --name grafana \
  -p 3000:3000 \
  grafana/grafana
```

1. Access Grafana at `http://localhost:3000` (default: admin/admin)
2. Add Prometheus data source: `http://prometheus:9090`
3. Import dashboard from `monitoring/grafana_dashboard.json`

## Testing

Run the monitoring integration test:

```bash
go test ./monitoring/... -v
```

## Troubleshooting

### Metrics Not Showing

1. Check if `/metrics` endpoint is accessible
2. Verify Prometheus is scraping the target
3. Check Prometheus targets page: `http://localhost:9090/targets`

### Health Check Failing

1. Check logs for errors
2. Verify all components are initialized
3. Test individual health checks

### High Memory Usage Alerts

1. Check goroutine leaks: `curl http://localhost:7999/metrics | grep goroutines`
2. Enable pprof for profiling
3. Review memory allocation patterns

## Next Steps

1. Set up Alertmanager for notifications
2. Configure alert routing (email, Slack, PagerDuty)
3. Create custom dashboards for specific metrics
4. Set up log aggregation (ELK, Loki)
5. Enable distributed tracing (Jaeger, Zipkin)

## Support

For issues or questions, refer to:
- Prometheus documentation: https://prometheus.io/docs/
- Grafana documentation: https://grafana.com/docs/
- Project README: `monitoring/README.md`
