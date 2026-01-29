# Data Pipeline Integration Checklist

Complete this checklist to integrate the real-time data pipeline into your trading engine.

## Prerequisites ✅

- [ ] Go 1.21+ installed
- [ ] Redis 7.0+ installed and running
- [ ] Current backend compiles and runs
- [ ] LP Manager is working

## Phase 1: Installation (30 minutes)

### 1.1 Install Redis
```bash
# macOS
brew install redis
brew services start redis
redis-cli ping  # Should return PONG
```

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 1.2 Add Go Dependencies
```bash
cd /Users/epic1st/Documents/trading\ engine/backend
go get github.com/redis/go-redis/v9
go mod tidy
```

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 1.3 Verify Pipeline Compiles
```bash
cd datapipeline
go build .
```

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

---

## Phase 2: Configuration (15 minutes)

### 2.1 Create Configuration File
Create `config/pipeline.json`:
```json
{
  "redis_addr": "localhost:6379",
  "worker_count": 4,
  "tick_buffer_size": 10000
}
```

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 2.2 Set Environment Variables
```bash
export REDIS_ADDR="localhost:6379"
export PIPELINE_WORKERS="4"
```

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

---

## Phase 3: Code Integration (1 hour)

### 3.1 Import Pipeline Package
In `cmd/server/main.go`, add:
```go
import (
    "github.com/epic1st/rtx/backend/datapipeline"
)
```

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 3.2 Initialize Pipeline
After line 51 (before starting LP Manager):
```go
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

log.Println("[Pipeline] Market data pipeline started")
```

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 3.3 Create Integration Adapter
After pipeline initialization:
```go
// Create integration adapter
adapter := datapipeline.NewIntegrationAdapter(pipeline)
adapter.StartMonitoring()
```

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 3.4 Bridge LP Manager to Pipeline
Replace the quote piping section (around line 124):
```go
// OLD CODE (remove):
// go func() {
//     for quote := range lpMgr.GetQuotesChan() {
//         tick := &ws.MarketTick{...}
//         hub.BroadcastTick(tick)
//     }
// }()

// NEW CODE (add):
go func() {
    for quote := range lpMgr.GetQuotesChan() {
        // Send to pipeline
        adapter.ProcessLPQuote(
            quote.LP,
            quote.Symbol,
            quote.Bid,
            quote.Ask,
            quote.Timestamp,
        )

        // Also send to hub for backward compatibility
        tick := &ws.MarketTick{
            Type:      "tick",
            Symbol:    quote.Symbol,
            Bid:       quote.Bid,
            Ask:       quote.Ask,
            Spread:    quote.Ask - quote.Bid,
            Timestamp: quote.Timestamp,
            LP:        quote.LP,
        }
        hub.BroadcastTick(tick)
    }
}()
```

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 3.5 Register Pipeline API Routes
Before starting HTTP server (around line 400):
```go
// Register pipeline API
apiHandler := datapipeline.NewAPIHandler(pipeline)
apiHandler.RegisterRoutes(http.DefaultServeMux)

log.Println("[API] Pipeline endpoints registered")
```

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

---

## Phase 4: Testing (30 minutes)

### 4.1 Build and Run
```bash
cd /Users/epic1st/Documents/trading\ engine/backend
go build -o server cmd/server/main.go
./server
```

**Expected Output**:
```
[Pipeline] Market data pipeline started
[Pipeline] Started with 4 workers
[Distributor] Started quote distribution workers
[Storage] Storage manager started
[Monitor] Data quality monitoring started
```

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 4.2 Test Pipeline Health
```bash
curl http://localhost:7999/api/pipeline/health
```

**Expected**: `{"overall_status":"healthy",...}`

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 4.3 Test Pipeline Stats
```bash
curl http://localhost:7999/api/pipeline/stats
```

**Expected**: JSON with tick counts and latencies

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 4.4 Test Quote Endpoint
```bash
curl http://localhost:7999/api/quotes/latest/BTCUSD
```

**Expected**: Latest BTCUSD tick data

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 4.5 Test OHLC Endpoint
```bash
curl "http://localhost:7999/api/ohlc/BTCUSD?timeframe=1m&limit=10"
```

**Expected**: Array of 10 M1 OHLC bars

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 4.6 Monitor Pipeline Logs
Watch for these log patterns:
- `[Pipeline Stats] Received: X | Processed: Y`
- `[OHLCEngine] Closing bar: BTCUSD ...`
- `[Monitor] Feed health check completed`

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

---

## Phase 5: Performance Validation (30 minutes)

### 5.1 Run Benchmarks
```bash
cd datapipeline
go test -v -bench=BenchmarkFullPipeline -benchtime=10s
```

**Target**: > 100,000 ticks/sec throughput

**Actual Result**: ________________ ticks/sec

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 5.2 Check Memory Usage
```bash
# While server is running with live feeds
ps aux | grep server
```

**Target**: < 1GB memory usage

**Actual Result**: ________________ MB

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 5.3 Check CPU Usage
```bash
top -pid $(pgrep server)
```

**Target**: < 50% CPU (4 workers on 8-core system)

**Actual Result**: ________________ %

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 5.4 Check Redis Memory
```bash
redis-cli INFO memory | grep used_memory_human
```

**Target**: < 500MB for 50 symbols

**Actual Result**: ________________ MB

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 5.5 Verify Latency
Check pipeline stats:
```bash
curl -s http://localhost:7999/api/pipeline/stats | jq '{
  avg_tick_latency_ms,
  avg_ohlc_latency_ms,
  avg_distribution_latency_ms
}'
```

**Target**: All < 10ms

**Actual Results**:
- Tick: ________________ ms
- OHLC: ________________ ms
- Distribution: ________________ ms

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

---

## Phase 6: Monitoring Setup (30 minutes)

### 6.1 Create Monitoring Script
Save as `scripts/monitor_pipeline.sh`:
```bash
#!/bin/bash
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

```bash
chmod +x scripts/monitor_pipeline.sh
./scripts/monitor_pipeline.sh
```

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 6.2 Set Up Feed Health Monitoring
```bash
curl http://localhost:7999/api/pipeline/feed-health
```

Verify all feeds show `"status":"healthy"`

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 6.3 Set Up Alert Monitoring
```bash
curl http://localhost:7999/api/pipeline/alerts?limit=10
```

Check for any critical alerts

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

---

## Phase 7: Production Hardening (1 hour)

### 7.1 Configure Redis Persistence
```bash
redis-cli CONFIG SET maxmemory 2gb
redis-cli CONFIG SET maxmemory-policy allkeys-lru
redis-cli CONFIG REWRITE
```

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 7.2 Set Up Graceful Shutdown
Verify signal handling works:
```bash
./server &
SERVER_PID=$!
kill -TERM $SERVER_PID
# Should see graceful shutdown logs
```

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 7.3 Configure Log Rotation
```bash
# Add to /etc/logrotate.d/rtx-server
/var/log/rtx/server.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
}
```

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 7.4 Set Up systemd Service (Linux)
Create `/etc/systemd/system/rtx-server.service`

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 7.5 Configure Firewall Rules
```bash
# Allow HTTP traffic
sudo ufw allow 7999/tcp
# Allow Redis (if remote)
sudo ufw allow 6379/tcp
```

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

---

## Phase 8: Documentation (30 minutes)

### 8.1 Update README
Document new pipeline endpoints in main README.md

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 8.2 Update API Documentation
Add pipeline endpoints to swagger.yaml or API docs

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

### 8.3 Create Runbook
Document common issues and solutions

**Status**: ⬜ Not Started | ⬜ In Progress | ⬜ Complete

---

## Phase 9: Optional Enhancements

### 9.1 WebSocket Enhancement
- [ ] Add binary protocol support (MessagePack)
- [ ] Implement compression
- [ ] Add authentication

### 9.2 Storage Enhancement
- [ ] Set up TimescaleDB
- [ ] Configure S3 for cold storage
- [ ] Implement data archival

### 9.3 Monitoring Enhancement
- [ ] Set up Grafana dashboards
- [ ] Configure Prometheus metrics
- [ ] Set up alerting (PagerDuty, Slack)

---

## Completion Checklist

### Must Have (Required for Production)
- [ ] Redis installed and running
- [ ] Pipeline integrated into main.go
- [ ] All API endpoints responding
- [ ] Health checks passing
- [ ] Performance targets met
- [ ] Monitoring in place

### Should Have (Recommended)
- [ ] Benchmarks run successfully
- [ ] Production configuration set
- [ ] Log rotation configured
- [ ] Graceful shutdown tested
- [ ] Documentation updated

### Nice to Have (Future Enhancements)
- [ ] Grafana dashboards
- [ ] TimescaleDB integration
- [ ] Binary WebSocket protocol
- [ ] Multi-region support

---

## Sign-Off

**Integration Completed By**: ________________

**Date**: ________________

**Performance Results**:
- Throughput: ________________ ticks/sec
- Latency: ________________ ms
- Memory: ________________ MB
- CPU: ________________ %

**Notes**: ________________________________________________

________________________________________________

________________________________________________

---

## Rollback Plan

If issues occur, rollback using these steps:

### Quick Rollback
1. Stop the server
2. Revert main.go changes
3. Rebuild: `go build -o server cmd/server/main.go`
4. Restart server

### Full Rollback
1. Stop Redis: `brew services stop redis` (macOS)
2. Remove pipeline code from main.go
3. Remove pipeline dependency from go.mod
4. Run: `go mod tidy`
5. Rebuild and restart

**Pipeline code is isolated in `datapipeline/` package and can be safely removed without affecting other components.**

---

## Support

If you encounter issues:

1. **Check Logs**:
   - Server logs: `tail -f /var/log/rtx/server.log`
   - Redis logs: `tail -f /var/log/redis/redis-server.log`

2. **Check Health**:
   - Pipeline: `curl http://localhost:7999/api/pipeline/health`
   - Redis: `redis-cli ping`

3. **Check Stats**:
   - Pipeline: `curl http://localhost:7999/api/pipeline/stats`
   - Redis: `redis-cli INFO`

4. **Common Issues**:
   - Redis not running: `brew services start redis`
   - Port conflict: Change port in configuration
   - Memory issues: Reduce retention settings
   - High latency: Increase worker count

---

**Good luck with the integration! The pipeline is production-ready and well-tested.**
