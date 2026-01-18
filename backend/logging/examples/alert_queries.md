# Common Log Search Queries and Alert Rules

## ELK Stack / Elasticsearch Queries

### Security & Authentication

```
# Failed login attempts from single IP
event_type:authentication_failed AND ip_address:*
| stats count by ip_address
| where count > 10

# Multiple failed logins in short time
event_type:authentication_failed AND timestamp:[now-5m TO now]

# Successful login after multiple failures
(event_type:authentication_failed OR event_type:authentication)
AND user_id:*
| sort timestamp
```

### Trading Operations

```
# All orders for specific account in last hour
event_type:order_placement AND account_id:"acc-456" AND timestamp:[now-1h TO now]

# Failed orders with reasons
level:ERROR AND component:order AND timestamp:[now-24h TO now]
| stats count by error

# High volume orders (> 10 lots)
event_type:order_placement AND metadata.volume:>10

# Orders by symbol
event_type:order_placement
| stats count by metadata.symbol
| sort count desc
```

### Performance Monitoring

```
# Slow database queries
duration_ms:>100 AND component:database
| stats avg(duration_ms), max(duration_ms), count by query

# Slow API endpoints
duration_ms:>1000 AND path:*
| stats avg(duration_ms), p95(duration_ms), count by path, method

# Error rate by endpoint
level:ERROR AND path:* AND timestamp:[now-1h TO now]
| stats count by path
| sort count desc
```

### Error Analysis

```
# Top errors in last 24h
level:ERROR AND timestamp:[now-24h TO now]
| stats count by message
| sort count desc
| limit 10

# Errors by component
level:ERROR AND component:*
| stats count by component

# New error types (not seen before)
level:ERROR AND timestamp:[now-1h TO now]
NOT message:[* TO now-24h]
```

### Admin Actions Audit

```
# All admin actions today
event_type:admin_action AND timestamp:[now-1d TO now]
| table timestamp, user_id, action, resource, resource_id

# Account balance modifications
event_type:admin_action AND resource:account
| table timestamp, user_id, action, before, after

# Configuration changes
event_type:config_change
| sort timestamp desc
```

## CloudWatch Insights Queries

### Error Rate Analysis

```
# Error rate per minute
fields @timestamp, level, message
| filter level = "ERROR"
| stats count(*) as error_count by bin(1m)

# Error rate by component
fields @timestamp, component, message
| filter level = "ERROR"
| stats count(*) by component
| sort count desc

# P95 error response time
fields @timestamp, duration_ms, path
| filter level = "ERROR"
| stats pct(duration_ms, 95) as p95, avg(duration_ms) as avg by path
```

### Performance Metrics

```
# Average response time by endpoint
fields @timestamp, path, method, duration_ms
| stats avg(duration_ms) as avg_duration,
        pct(duration_ms, 50) as p50,
        pct(duration_ms, 95) as p95,
        pct(duration_ms, 99) as p99
  by path, method
| sort p95 desc

# Slow query detection
fields @timestamp, query, duration_ms
| filter duration_ms > 100
| stats count(*) as slow_count, avg(duration_ms) as avg_duration by query
| sort slow_count desc

# Request rate per minute
fields @timestamp
| stats count(*) as requests by bin(1m)
```

### Authentication & Security

```
# Failed login attempts by IP
fields @timestamp, ip_address, event_type, reason
| filter event_type = "authentication_failed"
| stats count(*) as attempts by ip_address
| sort attempts desc

# Suspicious login patterns
fields @timestamp, user_id, ip_address
| filter event_type = "authentication"
| stats count(distinct ip_address) as ip_count by user_id
| filter ip_count > 3

# Account access timeline
fields @timestamp, event_type, action, account_id
| filter account_id = "acc-456"
| sort @timestamp desc
```

### Trading Activity

```
# Order volume by symbol
fields @timestamp, metadata.symbol, metadata.volume
| filter event_type = "order_placement"
| stats sum(metadata.volume) as total_volume by metadata.symbol
| sort total_volume desc

# Orders per account
fields @timestamp, account_id
| filter event_type = "order_placement"
| stats count(*) as order_count by account_id
| sort order_count desc

# P&L by account
fields @timestamp, account_id, metadata.pnl
| filter event_type = "position_close"
| stats sum(metadata.pnl) as total_pnl by account_id
| sort total_pnl desc
```

## Datadog Queries

### APM Traces

```
# Slow database queries
service:trading-engine resource_name:database operation:query @duration:>100ms
group by:resource_name

# Error rate by endpoint
service:trading-engine status:error
group by:resource_name
rollup:count

# Request latency percentiles
service:trading-engine
measure:@duration
percentiles:p50,p75,p95,p99
```

### Custom Metrics

```
# Order placement rate
trading_engine.orders.placed{*}
.as_rate()

# Active positions
trading_engine.positions.active{*}
group by:symbol

# Error rate by component
trading_engine.errors{*}
.as_rate()
group by:component
```

## Alert Rules

### Critical Alerts (PagerDuty)

```yaml
# High error rate
name: "Critical: High Error Rate"
query: "level:ERROR"
threshold: "> 100 errors in 5 minutes"
severity: critical
notification: pagerduty

# Database connection failures
name: "Critical: Database Down"
query: "component:database AND error:connection"
threshold: "> 5 errors in 1 minute"
severity: critical
notification: pagerduty

# Order placement failures
name: "Critical: Order Placement Failing"
query: "event_type:order_placement AND status:failed"
threshold: "> 10% failure rate in 5 minutes"
severity: critical
notification: pagerduty
```

### High Priority Alerts (Slack)

```yaml
# Slow endpoints
name: "High: Slow API Endpoints"
query: "duration_ms:>2000"
threshold: "> 20 slow requests in 5 minutes"
severity: high
notification: slack-trading-alerts

# Failed logins
name: "High: Multiple Failed Logins"
query: "event_type:authentication_failed"
threshold: "> 10 from same IP in 5 minutes"
severity: high
notification: slack-security

# Large order volume
name: "High: Large Order Detected"
query: "event_type:order_placement AND metadata.volume:>100"
threshold: "> 0 in 1 minute"
severity: high
notification: slack-trading-alerts
```

### Medium Priority Alerts (Email)

```yaml
# Slow queries
name: "Medium: Slow Database Queries"
query: "duration_ms:>500 AND component:database"
threshold: "> 50 queries in 10 minutes"
severity: medium
notification: email-devops

# High API latency
name: "Medium: High P95 Latency"
query: "duration_ms:*"
threshold: "p95 > 1000ms for 10 minutes"
severity: medium
notification: email-devops

# Unusual trading volume
name: "Medium: Unusual Trading Volume"
query: "event_type:order_placement"
threshold: "> 1000 orders in 1 minute"
severity: medium
notification: email-trading-ops
```

### Anomaly Detection

```yaml
# Sudden spike in errors
name: "Anomaly: Error Rate Spike"
query: "level:ERROR"
threshold: "> 2x baseline in 10 minutes"
algorithm: anomaly_detection
notification: slack-trading-alerts

# Drop in order volume
name: "Anomaly: Order Volume Drop"
query: "event_type:order_placement"
threshold: "< 0.5x baseline in 15 minutes"
algorithm: anomaly_detection
notification: slack-trading-ops
```

## Compliance Queries

### Regulatory Reporting

```
# All trades for specific period (MiFID II)
event_type:order_placement AND timestamp:[2026-01-01 TO 2026-01-31]
| table timestamp, user_id, account_id, metadata.symbol, metadata.volume, metadata.price

# Position modifications
event_type:position_modify
| table timestamp, user_id, account_id, resource_id, before, after

# Account withdrawals
event_type:withdrawal
| table timestamp, account_id, metadata.amount, metadata.method
```

### Audit Trail

```
# Complete audit trail for account
(event_type:order_placement OR event_type:position_close OR
 event_type:deposit OR event_type:withdrawal OR event_type:admin_action)
AND account_id:"acc-456"
| sort timestamp desc

# Admin actions requiring approval
event_type:admin_action AND action:(update_leverage OR adjust OR bonus)
| table timestamp, user_id, action, resource_id, before, after
```

## Dashboard Queries

### Real-time Operations Dashboard

```
# Active orders per second
event_type:order_placement | timechart count span=1s

# Error rate vs success rate
(level:ERROR OR level:INFO) | timechart count by level

# Top active symbols
event_type:order_placement | stats count by metadata.symbol | top 10

# Average response time
duration_ms:* | timechart avg(duration_ms) span=1m
```

### Business Metrics Dashboard

```
# Daily order volume
event_type:order_placement | timechart span=1d count

# P&L by day
event_type:position_close | timechart span=1d sum(metadata.pnl)

# Active accounts
user_id:* | dedup user_id | stats count

# Most traded symbols
event_type:order_placement | stats sum(metadata.volume) by metadata.symbol
```

## Performance Optimization Queries

### Identify Bottlenecks

```
# Slowest endpoints
duration_ms:* | stats avg(duration_ms), max(duration_ms) by path | sort avg desc

# Most frequent slow queries
duration_ms:>100 | stats count by query | sort count desc

# High memory usage patterns
level:WARN AND message:*memory* | stats count by component
```

## Export Examples

### Export to CSV

```bash
# Export failed orders to CSV
curl -X POST "http://elk:9200/trading-engine-*/_search" \
  -H 'Content-Type: application/json' \
  -d '{
    "query": {
      "bool": {
        "must": [
          {"match": {"level": "ERROR"}},
          {"match": {"component": "order"}}
        ]
      }
    }
  }' | jq -r '.hits.hits[]._source | [.timestamp, .account_id, .order_id, .error] | @csv'
```

### Export audit trail

```bash
# Export compliance data
curl -X POST "http://elk:9200/trading-engine-*/_search" \
  -H 'Content-Type: application/json' \
  -d '{
    "query": {
      "bool": {
        "must": [
          {"match": {"compliance": true}},
          {"range": {"timestamp": {"gte": "2026-01-01", "lte": "2026-01-31"}}}
        ]
      }
    }
  }' > compliance_report_january_2026.json
```
