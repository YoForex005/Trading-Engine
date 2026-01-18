# Data Pipeline Installation Guide

## Prerequisites

1. **Go 1.21+** - Required for building
2. **Redis 7.0+** - Required for pub/sub and storage
3. **Existing Trading Engine** - Your current RTX backend

## Step 1: Install Redis

### macOS
```bash
brew install redis
brew services start redis
```

### Ubuntu/Debian
```bash
sudo apt update
sudo apt install redis-server
sudo systemctl start redis-server
sudo systemctl enable redis-server
```

### Docker
```bash
docker run -d --name redis -p 6379:6379 redis:7-alpine
```

## Step 2: Verify Redis

```bash
redis-cli ping
# Should return: PONG
```

## Step 3: Add Dependencies

Add to your `go.mod`:

```bash
cd /Users/epic1st/Documents/trading\ engine/backend
go get github.com/redis/go-redis/v9
```

## Step 4: Integration Options

### Option A: Full Integration (Recommended)

Modify `/Users/epic1st/Documents/trading engine/backend/cmd/server/main.go`:

```go
import (
    "github.com/epic1st/rtx/backend/datapipeline"
    // ... existing imports
)

func main() {
    // ... existing initialization ...

    // Initialize data pipeline
    pipelineConfig := datapipeline.DefaultPipelineConfig()
    pipelineConfig.RedisAddr = "localhost:6379"

    pipeline, err := datapipeline.NewPipeline(pipelineConfig)
    if err != nil {
        log.Fatalf("Failed to create pipeline: %v", err)
    }

    if err := pipeline.Start(); err != nil {
        log.Fatalf("Failed to start pipeline: %v", err)
    }

    // Create integration adapter
    adapter := datapipeline.NewIntegrationAdapter(pipeline)
    adapter.StartMonitoring()

    // Bridge LP Manager to Pipeline
    go func() {
        for quote := range lpMgr.GetQuotesChan() {
            adapter.ProcessLPQuote(
                quote.LP,
                quote.Symbol,
                quote.Bid,
                quote.Ask,
                quote.Timestamp,
            )
        }
    }()

    // Register pipeline API
    apiHandler := datapipeline.NewAPIHandler(pipeline)
    apiHandler.RegisterRoutes(http.DefaultServeMux)

    // ... rest of your code ...
}
```

### Option B: Standalone Mode

Run the pipeline as a separate service:

```bash
cd /Users/epic1st/Documents/trading\ engine/backend/examples
go run pipeline_integration_example.go
```

## Step 5: Configuration

Create `/Users/epic1st/Documents/trading engine/backend/config/pipeline.json`:

```json
{
  "redis_addr": "localhost:6379",
  "redis_password": "",
  "redis_db": 0,
  "tick_buffer_size": 10000,
  "ohlc_buffer_size": 1000,
  "distribution_buffer_size": 5000,
  "worker_count": 4,
  "enable_deduplication": true,
  "enable_out_of_order_check": true,
  "max_tick_age_seconds": 60,
  "price_sanity_threshold": 0.10,
  "hot_data_retention": 1000,
  "warm_data_retention_days": 30,
  "enable_health_checks": true,
  "health_check_interval": "30s",
  "stale_quote_threshold": "10s"
}
```

## Step 6: Test the Installation

### Test 1: Pipeline Health Check

```bash
curl http://localhost:7999/api/pipeline/health
```

Expected response:
```json
{
  "timestamp": "2026-01-18T...",
  "overall_status": "healthy",
  "components": {
    "ingester": {"status": "healthy"},
    "ohlc_engine": {"status": "healthy"},
    "distributor": {"status": "healthy"},
    "storage": {"status": "healthy"}
  }
}
```

### Test 2: Pipeline Statistics

```bash
curl http://localhost:7999/api/pipeline/stats
```

### Test 3: Latest Quote

```bash
curl http://localhost:7999/api/quotes/latest/EURUSD
```

### Test 4: OHLC Data

```bash
curl "http://localhost:7999/api/ohlc/EURUSD?timeframe=1m&limit=10"
```

## Step 7: Monitoring Setup

### Add to your monitoring dashboard:

```bash
# Pipeline stats endpoint
/api/pipeline/stats

# Feed health endpoint
/api/pipeline/feed-health

# Alerts endpoint
/api/pipeline/alerts
```

### Example monitoring script:

```bash
#!/bin/bash
# monitor_pipeline.sh

while true; do
    curl -s http://localhost:7999/api/pipeline/stats | jq '{
        ticks_received,
        ticks_processed,
        ticks_dropped,
        avg_tick_latency_ms,
        clients_connected
    }'
    sleep 5
done
```

## Step 8: Performance Tuning

### For High Throughput (100k+ ticks/sec):

```go
config := datapipeline.DefaultPipelineConfig()
config.WorkerCount = 8
config.TickBufferSize = 50000
config.OHLCBufferSize = 5000
config.DistributionBufferSize = 20000
```

### For Low Latency (< 10ms):

```go
config := datapipeline.DefaultPipelineConfig()
config.WorkerCount = 16
config.EnableDeduplication = false // If not needed
config.EnableOutOfOrderCheck = false // If not needed
```

### Redis Optimization:

```bash
# Edit /etc/redis/redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru
appendonly no  # For hot data only
```

## Step 9: Production Deployment

### Use systemd service (Linux):

Create `/etc/systemd/system/rtx-pipeline.service`:

```ini
[Unit]
Description=RTX Market Data Pipeline
After=network.target redis.service

[Service]
Type=simple
User=rtx
WorkingDirectory=/opt/rtx
ExecStart=/opt/rtx/server
Restart=always
RestartSec=5
Environment="REDIS_ADDR=localhost:6379"

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable rtx-pipeline
sudo systemctl start rtx-pipeline
sudo systemctl status rtx-pipeline
```

### Docker Deployment:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 7999
CMD ["./server"]
```

```bash
docker build -t rtx-pipeline .
docker run -d \
  --name rtx-pipeline \
  -p 7999:7999 \
  -e REDIS_ADDR=redis:6379 \
  --link redis \
  rtx-pipeline
```

## Step 10: Verify Everything Works

Run the integration test:

```bash
cd /Users/epic1st/Documents/trading\ engine/backend/datapipeline
go test -v -bench=BenchmarkFullPipeline -benchtime=10s
```

Expected output:
```
=== Pipeline Benchmark Results ===
Total Ticks: 1000000
Duration: 10s
Throughput: 100000 ticks/sec
Processed: 999950
Dropped: 50
Avg Tick Latency: 0.8ms
Avg OHLC Latency: 3.2ms
Avg Distribution Latency: 5.1ms
```

## Troubleshooting

### Issue: Pipeline won't start
```bash
# Check Redis
redis-cli ping

# Check logs
journalctl -u rtx-pipeline -f
```

### Issue: High latency
```bash
# Check Redis performance
redis-cli --latency

# Increase worker count
config.WorkerCount = 16
```

### Issue: Ticks being dropped
```bash
# Increase buffer sizes
config.TickBufferSize = 50000
config.DistributionBufferSize = 20000
```

### Issue: Memory usage high
```bash
# Reduce retention
config.HotDataRetention = 500
config.WarmDataRetentionDays = 7

# Enable Redis eviction
redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

## Next Steps

1. **WebSocket Integration** - Connect WebSocket clients
2. **Monitoring** - Set up Grafana dashboards
3. **Alerting** - Configure alert notifications
4. **Backup** - Set up Redis backups
5. **Scaling** - Add more Redis instances for horizontal scaling

## Support

For issues or questions:
- Check logs: `/var/log/rtx/pipeline.log`
- Monitor Redis: `redis-cli MONITOR`
- Check metrics: `curl http://localhost:7999/api/pipeline/stats`
