package cache

import (
	"context"
	"fmt"
	"time"
)

// Example integration showing how to use the multi-tier caching system

// Example 1: Basic cache manager setup
func ExampleBasicSetup() error {
	// Configure cache manager
	config := &ManagerConfig{
		L1Size:     50 * 1024 * 1024, // 50MB L1 cache
		L1MaxItems: 5000,
		RedisConfig: &RedisConfig{
			Address:  "localhost:6379",
			Password: "",
			DB:       0,
			PoolSize: 50,
			Prefix:   "rtx",
		},
		EnableWarmup: true,
	}

	// Define loader function (L3 - database)
	loader := func(ctx context.Context, key string) (interface{}, error) {
		// This would fetch from database
		return nil, ErrNotFound
	}

	// Create cache manager
	manager, err := NewCacheManager(config, loader)
	if err != nil {
		return err
	}

	// Start cache manager (warmup + monitoring)
	ctx := context.Background()
	if err := manager.Start(ctx); err != nil {
		return err
	}

	// Use cache
	if err := manager.Set(ctx, NS_Symbols, "BTCUSD", map[string]interface{}{
		"symbol":     "BTCUSD",
		"tickSize":   0.01,
		"lotSize":    0.001,
		"commission": 0.1,
	}, TTL_Symbol_Config); err != nil {
		return err
	}

	value, err := manager.Get(ctx, NS_Symbols, "BTCUSD")
	if err != nil {
		return err
	}

	fmt.Printf("Symbol config: %+v\n", value)
	return nil
}

// Example 2: Trading engine integration
type TradingEngineCache struct {
	cache *CacheManager
}

func NewTradingEngineCache() (*TradingEngineCache, error) {
	config := DefaultManagerConfig()

	// Custom loader for trading data
	loader := func(ctx context.Context, key string) (interface{}, error) {
		// Parse key to determine data type
		// Load from database
		return nil, ErrNotFound
	}

	// Add warmup strategies
	config.WarmupStrategies = []WarmupStrategy{
		NewSymbolConfigWarmup(func(ctx context.Context) (map[string]interface{}, error) {
			// Load all symbols from database
			symbols := map[string]interface{}{
				"BTCUSD": map[string]interface{}{"symbol": "BTCUSD", "enabled": true},
				"ETHUSD": map[string]interface{}{"symbol": "ETHUSD", "enabled": true},
			}
			return symbols, nil
		}),
	}

	manager, err := NewCacheManager(config, loader)
	if err != nil {
		return nil, err
	}

	return &TradingEngineCache{cache: manager}, nil
}

// GetSymbol retrieves symbol configuration with automatic caching
func (t *TradingEngineCache) GetSymbol(ctx context.Context, symbol string) (map[string]interface{}, error) {
	value, err := t.cache.GetSymbolConfig(ctx, symbol)
	if err != nil {
		return nil, err
	}

	config, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid symbol config format")
	}

	return config, nil
}

// UpdateSymbol updates symbol configuration and invalidates cache
func (t *TradingEngineCache) UpdateSymbol(ctx context.Context, symbol string, config map[string]interface{}) error {
	// Update database first
	// ...

	// Invalidate cache
	return t.cache.Delete(ctx, NS_Symbols, symbol)
}

// GetPrice retrieves latest price with 1-second cache
func (t *TradingEngineCache) GetPrice(ctx context.Context, symbol string) (float64, error) {
	value, err := t.cache.GetPrice(ctx, symbol)
	if err == nil {
		if price, ok := value.(float64); ok {
			return price, nil
		}
	}

	// Cache miss - fetch from price source
	// This would typically come from WebSocket/FIX feed
	price := 50000.0 // Example price

	// Cache for 1 second
	t.cache.Set(ctx, NS_Prices, symbol, price, TTL_Market_Price)

	return price, nil
}

// UpdatePrice updates price and invalidates cache
func (t *TradingEngineCache) UpdatePrice(ctx context.Context, symbol string, bid, ask float64) error {
	price := map[string]float64{
		"bid": bid,
		"ask": ask,
	}

	// Set with short TTL (1 second)
	return t.cache.Set(ctx, NS_Prices, symbol, price, TTL_Market_Price)
}

// GetAccount retrieves account with automatic caching
func (t *TradingEngineCache) GetAccount(ctx context.Context, accountID string) (map[string]interface{}, error) {
	value, err := t.cache.GetAccount(ctx, accountID)
	if err == nil {
		if account, ok := value.(map[string]interface{}); ok {
			return account, nil
		}
	}

	// Cache miss - load from database
	account := map[string]interface{}{
		"accountID": accountID,
		"balance":   10000.0,
		"equity":    10000.0,
	}

	// Cache for 5 minutes
	t.cache.Set(ctx, NS_Accounts, accountID, account, TTL_User_Account)

	return account, nil
}

// OnBalanceChange invalidates account cache when balance changes
func (t *TradingEngineCache) OnBalanceChange(ctx context.Context, accountID string) error {
	return t.cache.InvalidateAccount(ctx, accountID)
}

// GetOHLC retrieves OHLC data with caching
func (t *TradingEngineCache) GetOHLC(ctx context.Context, symbol, timeframe string, limit int) ([]interface{}, error) {
	key := fmt.Sprintf("%s:%s:%d", symbol, timeframe, limit)

	value, err := t.cache.GetWithTTL(ctx, NS_OHLC, key, TTL_OHLC_Historical)
	if err == nil {
		if ohlc, ok := value.([]interface{}); ok {
			return ohlc, nil
		}
	}

	// Cache miss - load from tick store or database
	ohlc := []interface{}{
		map[string]interface{}{"open": 50000.0, "high": 51000.0, "low": 49000.0, "close": 50500.0},
	}

	// Cache for 24 hours (historical data rarely changes)
	t.cache.Set(ctx, NS_OHLC, key, ohlc, TTL_OHLC_Historical)

	return ohlc, nil
}

// Example 3: HTTP middleware for API response caching
func CacheMiddleware(cache *CacheManager, ttl time.Duration) func(next func() (interface{}, error)) (interface{}, error) {
	return func(next func() (interface{}, error)) (interface{}, error) {
		ctx := context.Background()

		// Generate cache key from request
		cacheKey := "api:endpoint:params" // In real implementation, generate from request

		// Try cache first
		value, err := cache.GetWithTTL(ctx, NS_API, cacheKey, ttl)
		if err == nil {
			return value, nil
		}

		// Cache miss - execute handler
		result, err := next()
		if err != nil {
			return nil, err
		}

		// Cache result
		cache.Set(ctx, NS_API, cacheKey, result, ttl)

		return result, nil
	}
}

// Example 4: Batch operations for high performance
func ExampleBatchOperations(cache *CacheManager) error {
	ctx := context.Background()

	// Batch set multiple symbols
	symbols := map[string]interface{}{
		CacheKey(NS_Symbols, "BTCUSD"): map[string]interface{}{"symbol": "BTCUSD"},
		CacheKey(NS_Symbols, "ETHUSD"): map[string]interface{}{"symbol": "ETHUSD"},
		CacheKey(NS_Symbols, "XRPUSD"): map[string]interface{}{"symbol": "XRPUSD"},
	}

	if err := cache.multiTier.SetMulti(ctx, symbols, TTL_Symbol_Config); err != nil {
		return err
	}

	// Batch get multiple symbols
	keys := []string{
		CacheKey(NS_Symbols, "BTCUSD"),
		CacheKey(NS_Symbols, "ETHUSD"),
		CacheKey(NS_Symbols, "XRPUSD"),
	}

	results, err := cache.multiTier.GetMulti(ctx, keys)
	if err != nil {
		return err
	}

	fmt.Printf("Retrieved %d symbols from cache\n", len(results))
	return nil
}

// Example 5: Monitoring cache performance
func ExampleMonitoring(cache *CacheManager) {
	stats := cache.Stats()

	multiTierStats := stats["multi_tier"].(map[string]interface{})

	fmt.Printf("Cache Performance:\n")
	fmt.Printf("  L1 Hits: %d\n", multiTierStats["l1_hits"])
	fmt.Printf("  L2 Hits: %d\n", multiTierStats["l2_hits"])
	fmt.Printf("  L3 Hits: %d\n", multiTierStats["l3_hits"])
	fmt.Printf("  Misses: %d\n", multiTierStats["misses"])
	fmt.Printf("  Hit Rate: %.2f%%\n", multiTierStats["hit_rate"].(float64)*100)
	fmt.Printf("  Avg Latency: %v\n", multiTierStats["avg_latency"])
	fmt.Printf("  L1 Size: %d bytes\n", multiTierStats["l1_size"])
}

// Example 6: Event-based cache invalidation
func SetupEventBasedInvalidation(cache *CacheManager) {
	// Set invalidation handler
	cache.SetOnInvalidate(func(key string) {
		// Log invalidation event
		fmt.Printf("Cache invalidated: %s\n", key)

		// Optionally publish to message queue for distributed invalidation
		// publishInvalidationEvent(key)
	})
}

// Example 7: Graceful shutdown
func ExampleGracefulShutdown(cache *CacheManager) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return cache.Shutdown(ctx)
}
