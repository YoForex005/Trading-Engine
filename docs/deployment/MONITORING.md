# Monitoring and Observability Guide

This guide covers monitoring, metrics collection, and observability for the Trading Engine platform.

## Overview

The platform uses a modern observability stack:

- **Prometheus**: Metrics collection and time-series database
- **Grafana**: Visualization and dashboards
- **Structured Logging**: JSON logs for easy parsing and analysis

## Prometheus Metrics

### Accessing Prometheus

Prometheus is available at http://localhost:9090

### Configuration

Prometheus is configured via `monitoring/prometheus.yml`:

- **Global scrape interval**: 10 seconds
- **Evaluation interval**: 10 seconds
- **Scrape timeout**: 5 seconds

### Scrape Targets

The platform exposes metrics on the backend service:

- **Endpoint**: http://localhost:8080/metrics
- **Format**: Prometheus text format
- **Scrape interval**: 5 seconds (more frequent for trading metrics)

**Scrape configuration:**
```yaml
scrape_configs:
  - job_name: 'trading-engine-backend'
    scrape_interval: 5s
    static_configs:
      - targets: ['backend:8080']
```

### Available Metrics

The backend exposes the following metrics:

**Order Processing:**
```promql
# Total orders processed
trading_orders_processed_total{status="filled"}
trading_orders_processed_total{status="rejected"}

# Order processing rate (orders per second)
rate(trading_orders_processed_total[5m])

# Orders by type
trading_orders_processed_total{type="market"}
trading_orders_processed_total{type="limit"}
```

**Positions:**
```promql
# Active positions count
trading_positions_active{symbol="BTCUSD"}
trading_positions_active{symbol="ETHUSD"}

# Position by side
trading_positions_active{symbol="BTCUSD",side="long"}
trading_positions_active{symbol="BTCUSD",side="short"}

# Total positions opened/closed
trading_positions_opened_total
trading_positions_closed_total
```

**Order Latency:**
```promql
# Order processing duration histogram
trading_order_duration_seconds_bucket
trading_order_duration_seconds_sum
trading_order_duration_seconds_count

# P50 latency
histogram_quantile(0.50, rate(trading_order_duration_seconds_bucket[5m]))

# P95 latency
histogram_quantile(0.95, rate(trading_order_duration_seconds_bucket[5m]))

# P99 latency
histogram_quantile(0.99, rate(trading_order_duration_seconds_bucket[5m]))
```

**System Metrics:**
```promql
# HTTP requests
http_requests_total{endpoint="/api/orders",method="POST"}

# HTTP request duration
http_request_duration_seconds_bucket

# Go runtime metrics (automatic)
go_goroutines
go_memstats_alloc_bytes
go_gc_duration_seconds
```

### Useful Queries

**Trading Activity:**
```promql
# Orders per minute by status
sum(rate(trading_orders_processed_total[1m])) by (status) * 60

# Most active trading pairs
topk(5, sum(trading_positions_active) by (symbol))

# Order rejection rate
rate(trading_orders_processed_total{status="rejected"}[5m])
  / rate(trading_orders_processed_total[5m])
```

**Performance:**
```promql
# Average order latency
rate(trading_order_duration_seconds_sum[5m])
  / rate(trading_order_duration_seconds_count[5m])

# Slow orders (>100ms)
trading_order_duration_seconds_bucket{le="0.1"}

# Request throughput
rate(http_requests_total[5m])
```

**System Health:**
```promql
# Backend uptime
time() - process_start_time_seconds

# Memory usage
go_memstats_alloc_bytes / 1024 / 1024

# Goroutine count
go_goroutines

# GC pause time
rate(go_gc_duration_seconds_sum[5m])
```

### Alerting Rules (Recommended)

Add these to `monitoring/prometheus.yml` for production:

```yaml
rule_files:
  - 'alerts.yml'

# alerts.yml
groups:
  - name: trading_engine
    interval: 30s
    rules:
      - alert: HighOrderLatency
        expr: histogram_quantile(0.95, rate(trading_order_duration_seconds_bucket[5m])) > 0.1
        for: 5m
        annotations:
          summary: "High order processing latency"
          description: "P95 latency is {{ $value }}s"

      - alert: HighOrderRejectionRate
        expr: |
          rate(trading_orders_processed_total{status="rejected"}[5m])
          / rate(trading_orders_processed_total[5m]) > 0.1
        for: 5m
        annotations:
          summary: "High order rejection rate"
          description: "{{ $value }}% of orders being rejected"

      - alert: BackendDown
        expr: up{job="trading-engine-backend"} == 0
        for: 1m
        annotations:
          summary: "Backend service is down"
          description: "Backend has been down for more than 1 minute"
```

## Grafana Dashboards

### Accessing Grafana

Grafana is available at http://localhost:3001

**Default credentials:**
- Username: `admin`
- Password: `admin`

Change the password on first login.

### Data Source Configuration

Add Prometheus as a data source:

1. Navigate to Configuration → Data Sources
2. Click "Add data source"
3. Select "Prometheus"
4. Set URL: `http://prometheus:9090`
5. Click "Save & Test"

### Creating Dashboards

**Trading Overview Dashboard:**

Create a new dashboard with these panels:

1. **Order Rate (Graph)**
   ```promql
   sum(rate(trading_orders_processed_total[5m])) by (status)
   ```

2. **Active Positions (Stat)**
   ```promql
   sum(trading_positions_active)
   ```

3. **Order Latency (Graph)**
   ```promql
   histogram_quantile(0.50, rate(trading_order_duration_seconds_bucket[5m])) # P50
   histogram_quantile(0.95, rate(trading_order_duration_seconds_bucket[5m])) # P95
   histogram_quantile(0.99, rate(trading_order_duration_seconds_bucket[5m])) # P99
   ```

4. **Positions by Symbol (Bar Gauge)**
   ```promql
   sum(trading_positions_active) by (symbol)
   ```

5. **Order Rejection Rate (Graph)**
   ```promql
   rate(trading_orders_processed_total{status="rejected"}[5m])
     / rate(trading_orders_processed_total[5m])
   ```

**System Health Dashboard:**

1. **CPU & Memory (Graph)**
   ```promql
   go_memstats_alloc_bytes / 1024 / 1024  # Memory MB
   rate(process_cpu_seconds_total[5m])     # CPU usage
   ```

2. **Goroutines (Graph)**
   ```promql
   go_goroutines
   ```

3. **HTTP Request Rate (Graph)**
   ```promql
   sum(rate(http_requests_total[5m])) by (endpoint)
   ```

4. **GC Pause Time (Graph)**
   ```promql
   rate(go_gc_duration_seconds_sum[5m])
   ```

### Dashboard Provisioning

To automatically provision dashboards, create JSON files in:
```
monitoring/grafana/dashboards/
```

Grafana will automatically load these on startup.

## Structured Logging

### Log Format

The backend uses structured JSON logging:

```json
{
  "level": "info",
  "ts": "2026-01-16T10:30:45.123Z",
  "caller": "core/engine.go:123",
  "msg": "Order processed",
  "order_id": "ord_123456",
  "symbol": "BTCUSD",
  "type": "market",
  "side": "buy",
  "quantity": 1.5,
  "duration_ms": 15
}
```

### Viewing Logs

**All logs:**
```bash
docker-compose logs -f backend
```

**Formatted JSON:**
```bash
docker-compose logs backend | jq .
```

**Filter by level:**
```bash
# Errors only
docker-compose logs backend | jq 'select(.level=="error")'

# Warnings and errors
docker-compose logs backend | jq 'select(.level=="error" or .level=="warn")'

# Debug logs
docker-compose logs backend | jq 'select(.level=="debug")'
```

**Search by content:**
```bash
# Database-related logs
docker-compose logs backend | jq 'select(.msg | contains("database"))'

# Order processing
docker-compose logs backend | jq 'select(.order_id != null)'

# Specific symbol
docker-compose logs backend | jq 'select(.symbol=="BTCUSD")'
```

**Follow live logs:**
```bash
docker-compose logs -f backend | jq .
```

### Log Levels

Control log verbosity with the `LOG_LEVEL` environment variable:

- `debug`: All logs (verbose, development only)
- `info`: Informational messages (default)
- `warn`: Warning messages
- `error`: Error messages only

**Set in .env:**
```bash
LOG_LEVEL=info
```

### Common Log Patterns

**Order processing:**
```json
{"level":"info","msg":"Order received","order_id":"...","symbol":"BTCUSD"}
{"level":"debug","msg":"Validating order","order_id":"..."}
{"level":"info","msg":"Order executed","order_id":"...","price":45000.00}
```

**Position management:**
```json
{"level":"info","msg":"Position opened","position_id":"...","symbol":"BTCUSD"}
{"level":"info","msg":"Position updated","position_id":"...","pnl":150.50}
{"level":"info","msg":"Position closed","position_id":"...","reason":"take_profit"}
```

**Errors:**
```json
{"level":"error","msg":"Database query failed","error":"connection refused"}
{"level":"error","msg":"Order validation failed","order_id":"...","reason":"insufficient_margin"}
{"level":"error","msg":"Redis connection lost","error":"connection timeout"}
```

## Log Aggregation (Production)

For production deployments, consider a log aggregation solution:

### ELK Stack

**Elasticsearch + Logstash + Kibana:**
- Ship logs to Logstash
- Index in Elasticsearch
- Visualize in Kibana

**Docker Compose addition:**
```yaml
logstash:
  image: docker.elastic.co/logstash/logstash:8.11.0
  volumes:
    - ./monitoring/logstash.conf:/usr/share/logstash/pipeline/logstash.conf

elasticsearch:
  image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
  environment:
    - discovery.type=single-node

kibana:
  image: docker.elastic.co/kibana/kibana:8.11.0
  ports:
    - "5601:5601"
```

### Loki + Grafana

**Grafana Loki for log aggregation:**
```yaml
loki:
  image: grafana/loki:2.9.0
  ports:
    - "3100:3100"
  command: -config.file=/etc/loki/local-config.yaml

promtail:
  image: grafana/promtail:2.9.0
  volumes:
    - /var/lib/docker/containers:/var/lib/docker/containers:ro
```

Then query logs in Grafana alongside metrics.

## Health Checks

### Health Endpoints

The backend exposes health check endpoints:

**Liveness (process alive):**
```bash
curl http://localhost:8080/health/live
```

Returns 200 if the process is running.

**Readiness (dependencies healthy):**
```bash
curl http://localhost:8080/health/ready
```

Returns 200 if database and Redis are accessible, 503 otherwise.

**Detailed health:**
```bash
curl http://localhost:8080/health
```

Returns JSON with component status:
```json
{
  "status": "healthy",
  "checks": {
    "database": "ok",
    "redis": "ok",
    "oanda": "ok"
  },
  "timestamp": "2026-01-16T10:30:45Z"
}
```

### Monitoring Health Checks

**Prometheus query:**
```promql
up{job="trading-engine-backend"}
```

**Alert on unhealthy:**
```yaml
- alert: ServiceUnhealthy
  expr: up{job="trading-engine-backend"} == 0
  for: 1m
```

## Performance Monitoring

### Database Performance

**Slow query log:**
```sql
-- Enable slow query log
ALTER SYSTEM SET log_min_duration_statement = 100;  -- Log queries >100ms
SELECT pg_reload_conf();

-- View slow queries
SELECT * FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;
```

**Connection pool metrics:**
```promql
# If exposed by backend
database_connections_active
database_connections_idle
database_connections_max
```

### Redis Performance

**Monitor cache hits:**
```bash
docker-compose exec redis redis-cli INFO stats
```

Look for:
- `keyspace_hits`: Cache hits
- `keyspace_misses`: Cache misses
- `evicted_keys`: Keys evicted due to memory limits

**Calculate hit rate:**
```
hit_rate = keyspace_hits / (keyspace_hits + keyspace_misses)
```

Target: >90% hit rate for optimal performance

### Memory Profiling

**Go pprof endpoints:**
```bash
# Heap profile
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# CPU profile
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# Goroutine profile
curl http://localhost:8080/debug/pprof/goroutine > goroutine.prof
go tool pprof goroutine.prof
```

## Production Recommendations

### Metrics Retention

**Prometheus:**
- Default: 15 days
- Increase for production: 90 days

```yaml
# monitoring/prometheus.yml
storage:
  tsdb:
    retention.time: 90d
```

### Alerting

Configure Alertmanager for production alerts:

```yaml
alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - alertmanager:9093
```

Send alerts to:
- Email
- Slack
- PagerDuty
- Discord

### Log Retention

- Local logs: 7 days (rolling)
- Aggregated logs: 90 days minimum
- Archive critical logs: 1 year+

### Monitoring Checklist

- [ ] Prometheus scraping backend metrics
- [ ] Grafana dashboards created
- [ ] Alert rules configured
- [ ] Log aggregation set up
- [ ] Health checks monitored
- [ ] Slow query log enabled
- [ ] Redis metrics tracked
- [ ] Backup monitoring enabled
- [ ] Alerting integrated (Slack/PagerDuty)
- [ ] Runbook created for common issues

## Troubleshooting

### Prometheus Not Scraping

**Check targets:**
Navigate to http://localhost:9090/targets

**Common issues:**
- Backend not exposing /metrics endpoint
- Network connectivity (check docker-compose network)
- Firewall blocking scrape

### Grafana Can't Connect to Prometheus

**Verify data source:**
1. Test connection in data source settings
2. Check URL is correct: `http://prometheus:9090`
3. Ensure both services on same Docker network

### Missing Metrics

**Verify backend exposes metrics:**
```bash
curl http://localhost:8080/metrics
```

Should return Prometheus text format metrics.

### High Memory Usage

**Check Prometheus retention:**
```bash
docker-compose exec prometheus du -sh /prometheus
```

**Reduce retention or increase memory allocation.**

## Related Documentation

- [Local Development Guide](./LOCAL_DEV.md)
- [Operations Runbook](./OPERATIONS.md)
- [Docker Deployment](./DOCKER.md)
