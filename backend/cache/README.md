# Multi-Tier Caching System

High-performance caching system for the trading engine backend with three-tier architecture.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Cache Manager                            │
├─────────────────────────────────────────────────────────────┤
│  L1: In-Memory LRU    │  L2: Redis Cache  │  L3: Database  │
│  (Per-Server)         │  (Shared)         │  (Persistent)  │
│  <1ms latency         │  <10ms latency    │  Variable      │
│  100MB default        │  Unlimited        │  Unlimited     │
└─────────────────────────────────────────────────────────────┘
```

## Features

- **Three-Tier Caching**: L1 (memory) → L2 (Redis) → L3 (database)
- **Automatic Cache Warming**: Pre-load frequently accessed data on startup
- **Event-Based Invalidation**: Real-time cache invalidation on data changes
- **Connection Pooling**: Redis connection pooling for high throughput
- **LRU Eviction**: Automatic eviction of least recently used items
- **CDN Integration**: Static asset caching with versioning
- **Performance Monitoring**: Real-time statistics and alerting
- **Graceful Degradation**: Falls back to lower tiers on cache miss

## Quick Start

```go
import "github.com/epic1st/rtx/backend/cache"

// 1. Create cache manager
config := cache.DefaultManagerConfig()
manager, err := cache.NewCacheManager(config, loaderFunc)
if err != nil {
    log.Fatal(err)
}

// 2. Start cache manager (warmup + monitoring)
ctx := context.Background()
manager.Start(ctx)

// 3. Use cache
manager.Set(ctx, cache.NS_Symbols, "BTCUSD", symbolData, cache.TTL_Symbol_Config)
value, err := manager.Get(ctx, cache.NS_Symbols, "BTCUSD")
```

## Cache Namespaces

```go
const (
    NS_Symbols    = "symbols"    // Symbol configurations
    NS_Accounts   = "accounts"   // User accounts
    NS_Positions  = "positions"  // Trading positions
    NS_Orders     = "orders"     // Orders
    NS_Prices     = "prices"     // Market prices
    NS_OHLC       = "ohlc"       // Historical OHLC data
    NS_API        = "api"        // API responses
    NS_LP         = "lp"         // LP configurations
    NS_Static     = "static"     // Static content
)
```

## TTL Configuration

| Data Type | TTL | Reason |
|-----------|-----|--------|
| Symbol Config | 1 hour | Rarely changes |
| User Account | 5 minutes | Moderate changes |
| LP Config | 10 minutes | Rarely changes |
| Market Price | 1 second | Frequent updates |
| Position Data | 100ms | Real-time changes |
| Order Book | 100ms | Real-time changes |
| OHLC Historical | 24 hours | Static historical data |
| API Response | 30 seconds | Configurable |

## Performance Targets

- **99% cache hit rate** for symbol configuration
- **<1ms latency** for L1 cache reads
- **<10ms latency** for L2 cache reads (Redis)
- **10,000+ requests/sec** sustained throughput
- **<100ms** cache warmup for critical data
- **50% memory reduction** vs. no caching

## Cache Strategies

### 1. Cache-Aside (Lazy Loading)
```go
value, err := cache.Get(ctx, ns, key)
if err != nil {
    // Cache miss - load from database
    value = loadFromDatabase(key)
    cache.Set(ctx, ns, key, value, ttl)
}
```

### 2. Write-Through
```go
// Update database
db.Update(key, value)

// Update cache immediately
cache.Set(ctx, ns, key, value, ttl)
```

### 3. Write-Back (Async)
```go
// Update cache
cache.Set(ctx, ns, key, value, ttl)

// Async database update (buffered)
queue.Enqueue(func() {
    db.Update(key, value)
})
```

### 4. Event-Based Invalidation
```go
// On position close
cache.InvalidatePosition(ctx, positionID)

// On balance change
cache.InvalidateAccount(ctx, accountID)

// On price update
cache.InvalidatePrice(ctx, symbol)
```

## Cache Warming

Automatically pre-load frequently accessed data on startup:

```go
config.WarmupStrategies = []cache.WarmupStrategy{
    // Warm up symbol configurations
    cache.NewSymbolConfigWarmup(loadSymbols),

    // Warm up historical OHLC data
    cache.NewOHLCHistoricalWarmup(loadOHLC, symbols),

    // Warm up active accounts
    cache.NewAccountWarmup(loadAccounts),
}
```

## Monitoring

```go
stats := manager.Stats()

// Cache performance
fmt.Printf("L1 Hit Rate: %.2f%%\n", stats["l1_hit_rate"])
fmt.Printf("L2 Hit Rate: %.2f%%\n", stats["l2_hit_rate"])
fmt.Printf("Overall Hit Rate: %.2f%%\n", stats["hit_rate"])
fmt.Printf("Avg Latency: %v\n", stats["avg_latency"])

// Automatic alerts
// - Low hit rate (<50%)
// - High latency (>10ms)
// - High eviction rate
```

## Redis Configuration

```go
config := &cache.RedisConfig{
    Address:      "localhost:6379",
    Password:     "",
    DB:           0,
    PoolSize:     100,          // Connection pool size
    MinIdleConns: 10,           // Min idle connections
    MaxRetries:   3,            // Retry failed operations
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    Prefix:       "rtx",        // Key prefix
}
```

## CDN Integration

For static assets (charts, images, CSS, JS):

```go
cdnConfig := &cache.CDNConfig{
    CDNURL:    "https://cdn.example.com",
    OriginURL: "https://api.example.com",
    PurgeURL:  "https://cdn.example.com/api",
    APIKey:    "your-api-key",
    Timeout:   10 * time.Second,
}

// Get CDN URL for asset
url := manager.GetCDNAssetURL("/assets/chart.js")

// Purge asset from CDN
manager.PurgeCDNAsset(ctx, "/assets/chart.js")
```

## Best Practices

1. **Use appropriate TTLs**: Hot data = longer TTL, volatile data = shorter TTL
2. **Invalidate on writes**: Always invalidate cache when data changes
3. **Monitor hit rates**: Target 90%+ hit rate for best performance
4. **Use batch operations**: Reduce round trips with GetMulti/SetMulti
5. **Handle cache failures gracefully**: Always have fallback to database
6. **Warm critical data**: Pre-load symbols, LP configs on startup
7. **Use namespaces**: Organize cache keys by data type
8. **Set size limits**: Prevent memory exhaustion with L1 size limits
9. **Monitor Redis**: Use Redis monitoring tools in production
10. **Test cache invalidation**: Ensure data consistency

## Integration Examples

See `integration_example.go` for complete examples:
- Basic setup
- Trading engine integration
- HTTP middleware
- Batch operations
- Performance monitoring
- Event-based invalidation

## Production Deployment

### Redis Cluster
For high availability, use Redis Cluster or Sentinel:

```go
config.RedisConfig = &cache.RedisConfig{
    Address: "redis-cluster:6379,redis-cluster:6380,redis-cluster:6381",
    // ... other settings
}
```

### Monitoring
Integrate with Prometheus/Grafana:

```go
// Expose cache metrics
http.Handle("/metrics", promhttp.Handler())

// Custom metrics
cacheHitRate.Set(stats["hit_rate"].(float64))
cacheLatency.Observe(stats["avg_latency"].(float64))
```

### Alerts
Set up alerts for:
- Hit rate < 50%
- Latency > 10ms
- Redis connection failures
- High eviction rate

## Troubleshooting

### Low Hit Rate
- Increase L1 cache size
- Increase TTLs for static data
- Check invalidation frequency
- Review access patterns

### High Latency
- Check Redis network latency
- Increase connection pool size
- Reduce serialization overhead
- Review data size

### Memory Issues
- Reduce L1 max size
- Enable more aggressive eviction
- Monitor L1 eviction rate
- Use Redis for large objects

## License

Copyright (c) 2025 RTX Trading
