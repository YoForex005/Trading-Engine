package cache

import (
	"context"
	"time"
)

// Cache defines the interface for all cache implementations
type Cache interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (interface{}, error)

	// Set stores a value in cache with TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists
	Exists(ctx context.Context, key string) (bool, error)

	// Clear removes all entries
	Clear(ctx context.Context) error

	// GetMulti retrieves multiple values at once
	GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error)

	// SetMulti stores multiple values at once
	SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error

	// Stats returns cache statistics
	Stats() CacheStats
}

// CacheStats holds cache performance metrics
type CacheStats struct {
	Hits       int64
	Misses     int64
	Sets       int64
	Deletes    int64
	Evictions  int64
	Size       int64
	HitRate    float64
	AvgGetTime time.Duration
	AvgSetTime time.Duration
}

// CacheKey generates a cache key with namespace
func CacheKey(namespace, key string) string {
	if namespace == "" {
		return key
	}
	return namespace + ":" + key
}

// CacheTTL constants for different data types
const (
	// Hot data - frequently accessed, rarely changes
	TTL_Symbol_Config    = 1 * time.Hour
	TTL_User_Account     = 5 * time.Minute
	TTL_LP_Config        = 10 * time.Minute

	// Warm data - moderate access, moderate changes
	TTL_Market_Price     = 1 * time.Second
	TTL_Position_Data    = 100 * time.Millisecond
	TTL_Order_Book       = 100 * time.Millisecond

	// Cold data - historical, rarely changes
	TTL_OHLC_Historical  = 24 * time.Hour
	TTL_Static_Content   = 365 * 24 * time.Hour

	// API Response caching
	TTL_API_Response     = 30 * time.Second
	TTL_API_List         = 10 * time.Second

	// No expiration
	TTL_Permanent        = 0
)

// Cache namespaces
const (
	NS_Symbols    = "symbols"
	NS_Accounts   = "accounts"
	NS_Positions  = "positions"
	NS_Orders     = "orders"
	NS_Prices     = "prices"
	NS_OHLC       = "ohlc"
	NS_API        = "api"
	NS_LP         = "lp"
	NS_Static     = "static"
)
