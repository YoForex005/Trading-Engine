# Cache Integration Guide

Step-by-step guide to integrate the multi-tier caching system into the trading engine.

## Step 1: Update cmd/server/main.go

Add cache initialization at startup:

```go
package main

import (
    // ... existing imports
    "github.com/epic1st/rtx/backend/cache"
)

func main() {
    log.Println("╔═══════════════════════════════════════════════════════════╗")
    log.Printf("║          %s - Backend v3.0                ║", brokerConfig.BrokerName)
    log.Println("╚═══════════════════════════════════════════════════════════╝")

    // ===== STEP 1: Initialize Cache Manager =====
    cacheManager, err := initializeCacheManager()
    if err != nil {
        log.Printf("[Cache] Warning: Failed to initialize cache: %v. Running without cache.", err)
    }

    // Initialize tick storage
    tickStore := tickstore.NewTickStore("default", brokerConfig.MaxTicksPerSymbol)

    // ... rest of initialization
}

func initializeCacheManager() (*cache.CacheManager, error) {
    // Configure cache
    config := &cache.ManagerConfig{
        L1Size:     100 * 1024 * 1024, // 100MB in-memory cache
        L1MaxItems: 10000,
        RedisConfig: &cache.RedisConfig{
            Address:      getEnv("REDIS_ADDR", "localhost:6379"),
            Password:     getEnv("REDIS_PASSWORD", ""),
            DB:           0,
            PoolSize:     100,
            MinIdleConns: 10,
            Prefix:       "rtx",
        },
        EnableWarmup: true,
    }

    // Define loader function (fallback to database)
    loader := func(ctx context.Context, key string) (interface{}, error) {
        // This is called when cache misses occur
        // You can load from database, API, etc.
        log.Printf("[Cache] Cache miss for key: %s, loading from source...", key)
        return nil, cache.ErrNotFound
    }

    // Add warmup strategies
    config.WarmupStrategies = []cache.WarmupStrategy{
        cache.NewSymbolConfigWarmup(loadAllSymbols),
        cache.NewAccountWarmup(loadActiveAccounts),
    }

    // Create cache manager
    manager, err := cache.NewCacheManager(config, loader)
    if err != nil {
        return nil, err
    }

    // Start cache manager
    ctx := context.Background()
    if err := manager.Start(ctx); err != nil {
        return nil, err
    }

    log.Println("[Cache] Cache manager initialized successfully")
    return manager, nil
}

func loadAllSymbols(ctx context.Context) (map[string]interface{}, error) {
    // Load symbols from database/config
    // This is called during cache warmup
    symbols := make(map[string]interface{})

    // Example: Load from JSON config or database
    // For now, return empty map
    // In production, load from bbookEngine.GetSymbols() or similar

    return symbols, nil
}

func loadActiveAccounts(ctx context.Context) (map[string]interface{}, error) {
    // Load active accounts from database
    accounts := make(map[string]interface{})

    // Example: Load from bbookEngine.GetAccounts() or database

    return accounts, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

## Step 2: Update api/server.go

Add cache to the Server struct:

```go
type Server struct {
    authService *auth.Service
    bbookAPI    *handlers.APIHandler

    // ... existing fields

    // Cache manager
    cacheManager *cache.CacheManager
}

func NewServer(authService *auth.Service, bbookAPI *handlers.APIHandler, lpMgr *lpmanager.Manager, cacheManager *cache.CacheManager) *Server {
    // ... existing initialization

    return &Server{
        authService:     authService,
        bbookAPI:        bbookAPI,
        // ... other fields
        cacheManager:    cacheManager,
    }
}
```

## Step 3: Add Cached Symbol Lookup

Create a helper method for cached symbol access:

```go
// GetSymbolCached retrieves symbol configuration with caching
func (s *Server) GetSymbolCached(ctx context.Context, symbol string) (interface{}, error) {
    if s.cacheManager == nil {
        // Cache disabled, fetch directly
        return s.getSymbolFromDB(symbol)
    }

    // Try cache first
    value, err := s.cacheManager.GetSymbolConfig(ctx, symbol)
    if err == nil {
        return value, nil
    }

    // Cache miss - load from database
    symbolData, err := s.getSymbolFromDB(symbol)
    if err != nil {
        return nil, err
    }

    // Cache for 1 hour
    s.cacheManager.Set(ctx, cache.NS_Symbols, symbol, symbolData, cache.TTL_Symbol_Config)

    return symbolData, nil
}

func (s *Server) getSymbolFromDB(symbol string) (interface{}, error) {
    // Load from database or bbookEngine
    return nil, fmt.Errorf("not implemented")
}
```

## Step 4: Cache Market Prices

Update price handling to use cache:

```go
// HandleGetTicks with caching
func (s *Server) HandleGetTicks(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")

    symbol := r.URL.Query().Get("symbol")
    if symbol == "" {
        http.Error(w, "Missing symbol parameter", http.StatusBadRequest)
        return
    }

    // Try cache first (1 second TTL)
    ctx := context.Background()
    cacheKey := fmt.Sprintf("ticks:%s:latest", symbol)

    if s.cacheManager != nil {
        if cached, err := s.cacheManager.GetWithTTL(ctx, cache.NS_Prices, cacheKey, cache.TTL_Market_Price); err == nil {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(cached)
            return
        }
    }

    // Cache miss - get from tick store
    limit := 500
    if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
        if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
            limit = l
        }
    }

    ticks := s.tickStore.GetHistory(symbol, limit)

    // Cache the result
    if s.cacheManager != nil {
        s.cacheManager.Set(ctx, cache.NS_Prices, cacheKey, ticks, cache.TTL_Market_Price)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ticks)
}
```

## Step 5: Cache Account Data

Add caching to account endpoints:

```go
func (s *Server) HandleGetAccount(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")

    // Get account ID from auth token or query param
    accountID := "demo_001" // In production, extract from JWT

    ctx := context.Background()

    // Try cache first (5 minute TTL)
    if s.cacheManager != nil {
        if cached, err := s.cacheManager.GetAccount(ctx, accountID); err == nil {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(cached)
            return
        }
    }

    // Cache miss - load from database
    account := s.loadAccountFromDB(accountID)

    // Cache for 5 minutes
    if s.cacheManager != nil {
        s.cacheManager.Set(ctx, cache.NS_Accounts, accountID, account, cache.TTL_User_Account)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(account)
}

// Invalidate account cache on balance change
func (s *Server) OnBalanceChange(accountID string) {
    if s.cacheManager != nil {
        ctx := context.Background()
        s.cacheManager.InvalidateAccount(ctx, accountID)
    }
}
```

## Step 6: Cache OHLC Data

Update OHLC endpoint with caching:

```go
func (s *Server) HandleGetOHLC(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")

    symbol := r.URL.Query().Get("symbol")
    timeframe := r.URL.Query().Get("timeframe")

    ctx := context.Background()
    cacheKey := fmt.Sprintf("%s:%s", symbol, timeframe)

    // Try cache first (24 hour TTL for historical data)
    if s.cacheManager != nil {
        if cached, err := s.cacheManager.GetWithTTL(ctx, cache.NS_OHLC, cacheKey, cache.TTL_OHLC_Historical); err == nil {
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(cached)
            return
        }
    }

    // Cache miss - get from tick store
    var tf int64 = 60
    switch timeframe {
    case "5m": tf = 300
    case "15m": tf = 900
    case "1h": tf = 3600
    case "4h": tf = 14400
    case "1d": tf = 86400
    }

    limit := 500
    ohlc := s.tickStore.GetOHLC(symbol, tf, limit)

    // Cache for 24 hours
    if s.cacheManager != nil {
        s.cacheManager.Set(ctx, cache.NS_OHLC, cacheKey, ohlc, cache.TTL_OHLC_Historical)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ohlc)
}
```

## Step 7: Add Cache Stats Endpoint

Add monitoring endpoint:

```go
// Add to main.go HTTP routes
http.HandleFunc("/admin/cache/stats", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Content-Type", "application/json")

    if cacheManager == nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "enabled": false,
            "message": "Cache not enabled",
        })
        return
    }

    stats := cacheManager.Stats()
    json.NewEncoder(w).Encode(stats)
})

// Add cache clear endpoint
http.HandleFunc("/admin/cache/clear", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")

    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    if cacheManager == nil {
        http.Error(w, "Cache not enabled", http.StatusServiceUnavailable)
        return
    }

    ctx := context.Background()
    namespace := r.URL.Query().Get("namespace")

    if namespace != "" {
        cacheManager.InvalidateNamespace(ctx, namespace)
    } else {
        // Clear all
        cacheManager.multiTier.Clear(ctx)
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Cache cleared",
    })
})
```

## Step 8: Environment Variables

Add to `.env` or environment configuration:

```bash
# Redis Configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Cache Configuration
CACHE_L1_SIZE_MB=100
CACHE_L1_MAX_ITEMS=10000
CACHE_ENABLE_WARMUP=true
```

## Step 9: Docker Compose (Optional)

If using Docker, add Redis to docker-compose.yml:

```yaml
version: '3.8'

services:
  backend:
    build: .
    ports:
      - "7999:7999"
    environment:
      - REDIS_ADDR=redis:6379
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes

volumes:
  redis_data:
```

## Step 10: Testing

Test the cache integration:

```bash
# 1. Start Redis (if using Docker)
docker-compose up -d redis

# 2. Run tests
cd backend
go test ./cache -v

# 3. Start server
go run cmd/server/main.go

# 4. Test cache stats endpoint
curl http://localhost:7999/admin/cache/stats

# 5. Test cached endpoint
curl http://localhost:7999/api/symbols  # First call - cache miss
curl http://localhost:7999/api/symbols  # Second call - cache hit

# 6. Clear cache
curl -X POST http://localhost:7999/admin/cache/clear
```

## Performance Validation

Expected improvements after integration:

1. **Symbol Lookups**:
   - Before: 10-50ms (database)
   - After: <1ms (L1 cache), 99% hit rate

2. **Account Queries**:
   - Before: 20-100ms (database)
   - After: <1ms (L1 cache), 90% hit rate

3. **Price Queries**:
   - Before: 5-20ms (tick store)
   - After: <1ms (L1 cache), 95% hit rate

4. **OHLC Data**:
   - Before: 50-200ms (tick store processing)
   - After: <1ms (L1 cache), 99% hit rate

5. **Overall Throughput**:
   - Before: 1,000 req/sec
   - After: 10,000+ req/sec

## Monitoring Checklist

After integration, monitor:

- [ ] Cache hit rate (target: >90%)
- [ ] L1 hit rate (target: >70%)
- [ ] Cache latency (target: <1ms for L1)
- [ ] Memory usage
- [ ] Eviction rate
- [ ] Redis connection pool usage

## Troubleshooting

### Cache Not Working
1. Check Redis connection: `redis-cli ping`
2. Check environment variables
3. Check logs for cache errors
4. Verify cache manager initialization

### Low Hit Rate
1. Increase L1 cache size
2. Increase TTLs for static data
3. Add more warmup strategies
4. Check invalidation frequency

### High Memory Usage
1. Reduce L1 max size
2. Enable more aggressive eviction
3. Reduce TTLs
4. Use Redis for large objects

## Next Steps

1. Monitor cache performance for 1 week
2. Tune cache sizes and TTLs based on metrics
3. Add more warmup strategies
4. Implement distributed cache invalidation
5. Set up Prometheus/Grafana monitoring

## Support

For issues or questions:
- See `cache/README.md` for detailed documentation
- See `cache/integration_example.go` for code examples
- Check `cache/IMPLEMENTATION_SUMMARY.md` for architecture details
