# Redis Architecture & Configuration

## Overview

Redis serves as the high-performance caching and real-time data layer for the RTX Trading Engine. It provides sub-millisecond latency for hot data access and powers real-time market data distribution via Pub/Sub.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
└─────────────────────────────────────────────────────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
        ▼                  ▼                  ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  Cache Layer │  │  Pub/Sub     │  │ Rate Limiting│
│              │  │              │  │              │
│ • Accounts   │  │ • Market Data│  │ • API Limits │
│ • Positions  │  │ • Executions │  │ • User Limits│
│ • Prices     │  │ • Alerts     │  │              │
└──────────────┘  └──────────────┘  └──────────────┘
        │                  │                  │
        └──────────────────┼──────────────────┘
                           │
                           ▼
                   ┌──────────────┐
                   │ Redis Server │
                   │              │
                   │ DB 0: Cache  │
                   │ DB 1: Pub/Sub│
                   │ DB 2: Limits │
                   └──────────────┘
```

## Use Cases

### 1. Cache Layer (DB 0)

#### Account Balance Cache
```
Key Pattern: account:{account_id}
TTL: 60 seconds
Purpose: Reduce database load for frequent balance queries
```

**Data Structure**:
```json
{
  "id": 1,
  "account_number": "RTX-000001",
  "balance": 10000.00,
  "equity": 10250.00,
  "margin": 500.00,
  "free_margin": 9750.00,
  "margin_level": 2050.00,
  "open_positions": 3
}
```

#### Latest Prices Cache
```
Key Pattern: price:{symbol}
TTL: 1 second
Purpose: Real-time price delivery
```

**Data Structure**:
```json
{
  "symbol": "BTCUSD",
  "time": "2026-01-18T10:30:45.123Z",
  "bid": 95000.50,
  "ask": 95001.50,
  "mid": 95001.00,
  "spread": 1.00
}
```

#### Open Positions Cache
```
Key Pattern: positions:{account_id}
TTL: 30 seconds
Purpose: Quick position lookup for validation
```

**Data Structure**:
```json
[
  {
    "id": 12345,
    "symbol": "EURUSD",
    "side": "BUY",
    "volume": 1.0,
    "open_price": 1.0950,
    "current_price": 1.0955,
    "unrealized_pnl": 50.00
  }
]
```

### 2. Pub/Sub Layer (DB 1)

#### Market Data Channel
```
Channel: market_data:{symbol}
Purpose: Distribute real-time ticks to subscribers
```

**Message Format**:
```json
{
  "type": "tick",
  "symbol": "BTCUSD",
  "time": 1705575045123,
  "bid": 95000.50,
  "ask": 95001.50,
  "lp": "OANDA"
}
```

#### Execution Notifications
```
Channel: executions:{account_id}
Purpose: Notify clients of order fills
```

**Message Format**:
```json
{
  "type": "execution",
  "order_id": 67890,
  "position_id": 12345,
  "symbol": "EURUSD",
  "side": "BUY",
  "volume": 1.0,
  "price": 1.0950,
  "timestamp": 1705575045123
}
```

### 3. Rate Limiting (DB 2)

#### API Rate Limits
```
Key Pattern: ratelimit:api:{user_id}
Algorithm: Sliding window counter
Window: 60 seconds
Limit: 100 requests/minute
```

**Implementation**:
```lua
-- rate_limit.lua
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local current_time = tonumber(ARGV[3])

local start_time = current_time - window

-- Remove old entries
redis.call('ZREMRANGEBYSCORE', key, 0, start_time)

-- Count requests in window
local count = redis.call('ZCARD', key)

if count < limit then
    redis.call('ZADD', key, current_time, current_time)
    redis.call('EXPIRE', key, window)
    return 1
else
    return 0
end
```

#### Order Rate Limits
```
Key Pattern: ratelimit:orders:{account_id}
Window: 1 second
Limit: 10 orders/second
```

## Configuration

### redis.conf

```ini
# Network
bind 0.0.0.0
port 6379
protected-mode yes
requirepass your_secure_password_here
tcp-backlog 511
timeout 0
tcp-keepalive 300

# General
daemonize no
supervised systemd
pidfile /var/run/redis_6379.pid
loglevel notice
logfile /var/log/redis/redis.log
databases 16

# Snapshotting (Persistence)
save 900 1      # After 900 sec if at least 1 key changed
save 300 10     # After 300 sec if at least 10 keys changed
save 60 10000   # After 60 sec if at least 10000 keys changed
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir /var/lib/redis

# Replication
# slaveof <masterip> <masterport>
# masterauth <master-password>
slave-serve-stale-data yes
slave-read-only yes
repl-diskless-sync no
repl-diskless-sync-delay 5
repl-ping-slave-period 10
repl-timeout 60
repl-disable-tcp-nodelay no
repl-backlog-size 64mb
repl-backlog-ttl 3600

# AOF (Append Only File)
appendonly yes
appendfilename "appendonly.aof"
appendfsync everysec
no-appendfsync-on-rewrite no
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
aof-load-truncated yes

# Lua Scripting
lua-time-limit 5000

# Slow Log
slowlog-log-slower-than 10000
slowlog-max-len 128

# Latency Monitor
latency-monitor-threshold 100

# Event Notification
notify-keyspace-events ""

# Advanced
hash-max-ziplist-entries 512
hash-max-ziplist-value 64
list-max-ziplist-size -2
list-compress-depth 0
set-max-intset-entries 512
zset-max-ziplist-entries 128
zset-max-ziplist-value 64
hll-sparse-max-bytes 3000
activerehashing yes
client-output-buffer-limit normal 0 0 0
client-output-buffer-limit slave 256mb 64mb 60
client-output-buffer-limit pubsub 32mb 8mb 60
hz 10
aof-rewrite-incremental-fsync yes

# Memory Management
maxmemory 16gb
maxmemory-policy allkeys-lru
maxmemory-samples 5
```

### Docker Deployment

```yaml
# docker-compose.yml
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    container_name: trading-redis
    restart: always
    ports:
      - "6379:6379"
    volumes:
      - ./redis.conf:/usr/local/etc/redis/redis.conf
      - redis-data:/data
    command: redis-server /usr/local/etc/redis/redis.conf
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 3
    networks:
      - trading-network

  redis-sentinel:
    image: redis:7-alpine
    container_name: trading-redis-sentinel
    restart: always
    ports:
      - "26379:26379"
    volumes:
      - ./sentinel.conf:/usr/local/etc/redis/sentinel.conf
    command: redis-sentinel /usr/local/etc/redis/sentinel.conf
    depends_on:
      - redis
    networks:
      - trading-network

volumes:
  redis-data:

networks:
  trading-network:
    driver: bridge
```

## Go Client Implementation

### Connection Pool

```go
// backend/internal/cache/redis.go
package cache

import (
    "context"
    "encoding/json"
    "time"

    "github.com/go-redis/redis/v8"
)

type RedisClient struct {
    client *redis.Client
    ctx    context.Context
}

func NewRedisClient(addr, password string) *RedisClient {
    rdb := redis.NewClient(&redis.Options{
        Addr:         addr,
        Password:     password,
        DB:           0, // Cache DB
        PoolSize:     100,
        MinIdleConns: 10,
        MaxConnAge:   5 * time.Minute,
        PoolTimeout:  4 * time.Second,
        IdleTimeout:  5 * time.Minute,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
    })

    return &RedisClient{
        client: rdb,
        ctx:    context.Background(),
    }
}

// Account Cache
func (c *RedisClient) SetAccount(accountID int64, data interface{}, ttl time.Duration) error {
    key := fmt.Sprintf("account:%d", accountID)
    value, err := json.Marshal(data)
    if err != nil {
        return err
    }

    return c.client.Set(c.ctx, key, value, ttl).Err()
}

func (c *RedisClient) GetAccount(accountID int64) (map[string]interface{}, error) {
    key := fmt.Sprintf("account:%d", accountID)
    val, err := c.client.Get(c.ctx, key).Result()
    if err != nil {
        return nil, err
    }

    var data map[string]interface{}
    if err := json.Unmarshal([]byte(val), &data); err != nil {
        return nil, err
    }

    return data, nil
}

// Latest Price Cache
func (c *RedisClient) SetPrice(symbol string, bid, ask float64, ttl time.Duration) error {
    key := fmt.Sprintf("price:%s", symbol)
    data := map[string]interface{}{
        "symbol": symbol,
        "time":   time.Now().UnixMilli(),
        "bid":    bid,
        "ask":    ask,
        "mid":    (bid + ask) / 2,
        "spread": ask - bid,
    }

    value, err := json.Marshal(data)
    if err != nil {
        return err
    }

    return c.client.Set(c.ctx, key, value, ttl).Err()
}

func (c *RedisClient) GetPrice(symbol string) (map[string]interface{}, error) {
    key := fmt.Sprintf("price:%s", symbol)
    val, err := c.client.Get(c.ctx, key).Result()
    if err != nil {
        return nil, err
    }

    var data map[string]interface{}
    if err := json.Unmarshal([]byte(val), &data); err != nil {
        return nil, err
    }

    return data, nil
}

// Positions Cache
func (c *RedisClient) SetPositions(accountID int64, positions interface{}, ttl time.Duration) error {
    key := fmt.Sprintf("positions:%d", accountID)
    value, err := json.Marshal(positions)
    if err != nil {
        return err
    }

    return c.client.Set(c.ctx, key, value, ttl).Err()
}

func (c *RedisClient) InvalidatePositions(accountID int64) error {
    key := fmt.Sprintf("positions:%d", accountID)
    return c.client.Del(c.ctx, key).Err()
}
```

### Pub/Sub Implementation

```go
// backend/internal/pubsub/redis_pubsub.go
package pubsub

import (
    "context"
    "encoding/json"
    "log"

    "github.com/go-redis/redis/v8"
)

type RedisPubSub struct {
    client *redis.Client
    ctx    context.Context
}

func NewRedisPubSub(addr, password string) *RedisPubSub {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       1, // Pub/Sub DB
    })

    return &RedisPubSub{
        client: rdb,
        ctx:    context.Background(),
    }
}

// Publish market data tick
func (p *RedisPubSub) PublishTick(symbol string, bid, ask float64, lp string) error {
    channel := fmt.Sprintf("market_data:%s", symbol)
    data := map[string]interface{}{
        "type":   "tick",
        "symbol": symbol,
        "time":   time.Now().UnixMilli(),
        "bid":    bid,
        "ask":    ask,
        "lp":     lp,
    }

    msg, err := json.Marshal(data)
    if err != nil {
        return err
    }

    return p.client.Publish(p.ctx, channel, msg).Err()
}

// Subscribe to market data
func (p *RedisPubSub) SubscribeMarketData(symbol string, handler func(map[string]interface{})) {
    channel := fmt.Sprintf("market_data:%s", symbol)
    pubsub := p.client.Subscribe(p.ctx, channel)
    defer pubsub.Close()

    ch := pubsub.Channel()
    for msg := range ch {
        var data map[string]interface{}
        if err := json.Unmarshal([]byte(msg.Payload), &data); err != nil {
            log.Printf("Failed to unmarshal message: %v", err)
            continue
        }

        handler(data)
    }
}

// Publish execution notification
func (p *RedisPubSub) PublishExecution(accountID int64, execution map[string]interface{}) error {
    channel := fmt.Sprintf("executions:%d", accountID)
    msg, err := json.Marshal(execution)
    if err != nil {
        return err
    }

    return p.client.Publish(p.ctx, channel, msg).Err()
}
```

### Rate Limiting

```go
// backend/internal/ratelimit/redis_limiter.go
package ratelimit

import (
    "context"
    "fmt"
    "time"

    "github.com/go-redis/redis/v8"
)

type RedisRateLimiter struct {
    client *redis.Client
    ctx    context.Context
    script *redis.Script
}

var rateLimitScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local current_time = tonumber(ARGV[3])

local start_time = current_time - window

redis.call('ZREMRANGEBYSCORE', key, 0, start_time)

local count = redis.call('ZCARD', key)

if count < limit then
    redis.call('ZADD', key, current_time, current_time)
    redis.call('EXPIRE', key, window)
    return 1
else
    return 0
end
`

func NewRedisRateLimiter(addr, password string) *RedisRateLimiter {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       2, // Rate limit DB
    })

    return &RedisRateLimiter{
        client: rdb,
        ctx:    context.Background(),
        script: redis.NewScript(rateLimitScript),
    }
}

// Check API rate limit
func (r *RedisRateLimiter) CheckAPILimit(userID int64, limit int, window time.Duration) (bool, error) {
    key := fmt.Sprintf("ratelimit:api:%d", userID)
    currentTime := time.Now().UnixMilli()
    windowMs := int64(window / time.Millisecond)

    result, err := r.script.Run(r.ctx, r.client, []string{key}, limit, windowMs, currentTime).Result()
    if err != nil {
        return false, err
    }

    return result.(int64) == 1, nil
}

// Check order rate limit
func (r *RedisRateLimiter) CheckOrderLimit(accountID int64, limit int, window time.Duration) (bool, error) {
    key := fmt.Sprintf("ratelimit:orders:%d", accountID)
    currentTime := time.Now().UnixMilli()
    windowMs := int64(window / time.Millisecond)

    result, err := r.script.Run(r.ctx, r.client, []string{key}, limit, windowMs, currentTime).Result()
    if err != nil {
        return false, err
    }

    return result.(int64) == 1, nil
}

// Get remaining requests
func (r *RedisRateLimiter) GetRemaining(userID int64, limit int, window time.Duration) (int, error) {
    key := fmt.Sprintf("ratelimit:api:%d", userID)
    currentTime := time.Now().UnixMilli()
    startTime := currentTime - int64(window/time.Millisecond)

    // Remove old entries
    r.client.ZRemRangeByScore(r.ctx, key, "0", fmt.Sprintf("%d", startTime))

    // Count current requests
    count, err := r.client.ZCard(r.ctx, key).Result()
    if err != nil {
        return 0, err
    }

    remaining := limit - int(count)
    if remaining < 0 {
        remaining = 0
    }

    return remaining, nil
}
```

## High Availability

### Redis Sentinel

```conf
# sentinel.conf
port 26379
sentinel monitor trading-redis 127.0.0.1 6379 2
sentinel auth-pass trading-redis your_secure_password_here
sentinel down-after-milliseconds trading-redis 5000
sentinel parallel-syncs trading-redis 1
sentinel failover-timeout trading-redis 10000
```

### Client Failover

```go
// backend/internal/cache/redis_ha.go
package cache

import (
    "github.com/go-redis/redis/v8"
)

func NewRedisHAClient() *redis.Client {
    return redis.NewFailoverClient(&redis.FailoverOptions{
        MasterName:    "trading-redis",
        SentinelAddrs: []string{"sentinel1:26379", "sentinel2:26379", "sentinel3:26379"},
        Password:      "your_secure_password_here",
        DB:            0,
        PoolSize:      100,
        MaxRetries:    3,
    })
}
```

## Monitoring

### Key Metrics

```bash
# Monitor commands/sec
redis-cli INFO stats | grep instantaneous_ops_per_sec

# Monitor memory usage
redis-cli INFO memory | grep used_memory_human

# Monitor connected clients
redis-cli INFO clients | grep connected_clients

# Monitor keyspace
redis-cli INFO keyspace

# Monitor latency
redis-cli --latency
```

### Prometheus Exporter

```yaml
# docker-compose.yml
redis-exporter:
  image: oliver006/redis_exporter:latest
  container_name: redis-exporter
  environment:
    - REDIS_ADDR=redis:6379
    - REDIS_PASSWORD=your_secure_password_here
  ports:
    - "9121:9121"
  depends_on:
    - redis
```

### Grafana Dashboard

Import dashboard ID: **11835** (Redis Dashboard for Prometheus)

Key panels:
- Commands/sec
- Hit rate
- Memory usage
- Connected clients
- Network I/O
- Keyspace

## Best Practices

### 1. Key Naming Convention

```
{namespace}:{entity}:{id}:{field}
```

Examples:
- `account:1:balance`
- `price:BTCUSD`
- `positions:1`
- `ratelimit:api:123`

### 2. TTL Strategy

| Data Type | TTL | Reason |
|-----------|-----|--------|
| Account balance | 60s | Frequently updated |
| Latest price | 1s | Real-time data |
| Open positions | 30s | Moderate update frequency |
| Symbol specs | 1h | Rarely changes |
| User sessions | 24h | Long-lived |

### 3. Cache Invalidation

```go
// Invalidate on write
func (s *Service) UpdateAccountBalance(accountID int64, newBalance float64) error {
    // Update database
    if err := s.db.UpdateBalance(accountID, newBalance); err != nil {
        return err
    }

    // Invalidate cache
    s.cache.InvalidateAccount(accountID)

    return nil
}
```

### 4. Error Handling

```go
func (c *RedisClient) GetWithFallback(key string) (interface{}, error) {
    // Try cache
    val, err := c.client.Get(c.ctx, key).Result()
    if err == redis.Nil {
        // Cache miss - fetch from DB
        return c.fetchFromDB(key)
    } else if err != nil {
        // Redis error - fallback to DB
        log.Printf("Redis error: %v, falling back to DB", err)
        return c.fetchFromDB(key)
    }

    return val, nil
}
```

---

**Document Version**: 1.0
**Last Updated**: 2026-01-18
**Status**: Production Ready
