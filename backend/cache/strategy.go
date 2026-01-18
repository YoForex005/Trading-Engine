package cache

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// MultiTierCache implements L1 (memory) -> L2 (Redis) -> L3 (database) caching strategy
type MultiTierCache struct {
	l1 *MemoryCache  // In-memory cache (fastest)
	l2 *RedisCache   // Redis cache (shared)

	// Loader function to fetch from source (L3 - database)
	loader LoaderFunc

	// Statistics
	mu    sync.RWMutex
	stats struct {
		l1Hits    int64
		l2Hits    int64
		l3Hits    int64
		misses    int64
		totalTime time.Duration
		calls     int64
	}

	// Configuration
	writeThrough bool // If true, writes go to all layers
	writeBack    bool // If true, writes are buffered and flushed periodically
}

// LoaderFunc is called when data is not found in cache
type LoaderFunc func(ctx context.Context, key string) (interface{}, error)

// NewMultiTierCache creates a multi-tier cache
func NewMultiTierCache(l1Size int64, l1MaxItems int, redisConfig *RedisConfig, loader LoaderFunc) (*MultiTierCache, error) {
	l1 := NewMemoryCache(l1Size, l1MaxItems)

	var l2 *RedisCache
	var err error
	if redisConfig != nil {
		l2, err = NewRedisCache(redisConfig)
		if err != nil {
			log.Printf("[Cache] Warning: Failed to connect to Redis: %v. Running with L1 cache only.", err)
		}
	}

	c := &MultiTierCache{
		l1:           l1,
		l2:           l2,
		loader:       loader,
		writeThrough: true,
		writeBack:    false,
	}

	// Set L1 eviction callback to write to L2
	l1.SetEvictionCallback(func(key string, value interface{}) {
		if c.l2 != nil && c.writeBack {
			ctx := context.Background()
			c.l2.Set(ctx, key, value, TTL_Permanent)
		}
	})

	return c, nil
}

// Get retrieves a value using cache-aside pattern (L1 -> L2 -> L3)
func (c *MultiTierCache) Get(ctx context.Context, key string) (interface{}, error) {
	start := time.Now()
	defer func() {
		c.mu.Lock()
		c.stats.totalTime += time.Since(start)
		c.stats.calls++
		c.mu.Unlock()
	}()

	// Try L1 (in-memory)
	value, err := c.l1.Get(ctx, key)
	if err == nil {
		c.mu.Lock()
		c.stats.l1Hits++
		c.mu.Unlock()
		return value, nil
	}

	// Try L2 (Redis)
	if c.l2 != nil {
		value, err = c.l2.Get(ctx, key)
		if err == nil {
			c.mu.Lock()
			c.stats.l2Hits++
			c.mu.Unlock()

			// Promote to L1
			c.l1.Set(ctx, key, value, TTL_Permanent)
			return value, nil
		}
	}

	// Try L3 (loader - database)
	if c.loader != nil {
		value, err = c.loader(ctx, key)
		if err == nil {
			c.mu.Lock()
			c.stats.l3Hits++
			c.mu.Unlock()

			// Warm cache
			c.l1.Set(ctx, key, value, TTL_Permanent)
			if c.l2 != nil {
				c.l2.Set(ctx, key, value, TTL_Permanent)
			}
			return value, nil
		}
	}

	// Not found anywhere
	c.mu.Lock()
	c.stats.misses++
	c.mu.Unlock()

	return nil, ErrNotFound
}

// GetWithTTL retrieves a value with specific TTL
func (c *MultiTierCache) GetWithTTL(ctx context.Context, key string, ttl time.Duration) (interface{}, error) {
	start := time.Now()

	// Try L1
	value, err := c.l1.Get(ctx, key)
	if err == nil {
		c.mu.Lock()
		c.stats.l1Hits++
		c.stats.totalTime += time.Since(start)
		c.stats.calls++
		c.mu.Unlock()
		return value, nil
	}

	// Try L2
	if c.l2 != nil {
		value, err = c.l2.Get(ctx, key)
		if err == nil {
			c.mu.Lock()
			c.stats.l2Hits++
			c.stats.totalTime += time.Since(start)
			c.stats.calls++
			c.mu.Unlock()

			// Promote to L1 with TTL
			c.l1.Set(ctx, key, value, ttl)
			return value, nil
		}
	}

	// Try loader
	if c.loader != nil {
		value, err = c.loader(ctx, key)
		if err == nil {
			c.mu.Lock()
			c.stats.l3Hits++
			c.stats.totalTime += time.Since(start)
			c.stats.calls++
			c.mu.Unlock()

			// Warm both caches with TTL
			c.l1.Set(ctx, key, value, ttl)
			if c.l2 != nil {
				c.l2.Set(ctx, key, value, ttl)
			}
			return value, nil
		}
	}

	c.mu.Lock()
	c.stats.misses++
	c.stats.totalTime += time.Since(start)
	c.stats.calls++
	c.mu.Unlock()

	return nil, ErrNotFound
}

// Set stores a value in all cache tiers (write-through)
func (c *MultiTierCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Write to L1
	if err := c.l1.Set(ctx, key, value, ttl); err != nil {
		return err
	}

	// Write-through to L2
	if c.writeThrough && c.l2 != nil {
		if err := c.l2.Set(ctx, key, value, ttl); err != nil {
			log.Printf("[Cache] Warning: Failed to write to L2: %v", err)
		}
	}

	return nil
}

// Delete removes a value from all cache tiers
func (c *MultiTierCache) Delete(ctx context.Context, key string) error {
	// Delete from all tiers
	c.l1.Delete(ctx, key)
	if c.l2 != nil {
		c.l2.Delete(ctx, key)
	}
	return nil
}

// Exists checks if a key exists in any tier
func (c *MultiTierCache) Exists(ctx context.Context, key string) (bool, error) {
	exists, _ := c.l1.Exists(ctx, key)
	if exists {
		return true, nil
	}

	if c.l2 != nil {
		return c.l2.Exists(ctx, key)
	}

	return false, nil
}

// Clear removes all entries from all tiers
func (c *MultiTierCache) Clear(ctx context.Context) error {
	c.l1.Clear(ctx)
	if c.l2 != nil {
		c.l2.Clear(ctx)
	}
	return nil
}

// GetMulti retrieves multiple values
func (c *MultiTierCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	missing := make([]string, 0)

	// Try L1 first
	for _, key := range keys {
		if value, err := c.l1.Get(ctx, key); err == nil {
			result[key] = value
		} else {
			missing = append(missing, key)
		}
	}

	if len(missing) == 0 {
		return result, nil
	}

	// Try L2 for missing keys
	if c.l2 != nil {
		l2Results, _ := c.l2.GetMulti(ctx, missing)
		for key, value := range l2Results {
			result[key] = value
			c.l1.Set(ctx, key, value, TTL_Permanent) // Promote to L1
		}

		// Update missing list
		newMissing := make([]string, 0)
		for _, key := range missing {
			if _, ok := l2Results[key]; !ok {
				newMissing = append(newMissing, key)
			}
		}
		missing = newMissing
	}

	// Try loader for remaining keys
	if c.loader != nil && len(missing) > 0 {
		for _, key := range missing {
			if value, err := c.loader(ctx, key); err == nil {
				result[key] = value
				c.l1.Set(ctx, key, value, TTL_Permanent)
				if c.l2 != nil {
					c.l2.Set(ctx, key, value, TTL_Permanent)
				}
			}
		}
	}

	return result, nil
}

// SetMulti stores multiple values
func (c *MultiTierCache) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if err := c.l1.SetMulti(ctx, items, ttl); err != nil {
		return err
	}

	if c.writeThrough && c.l2 != nil {
		if err := c.l2.SetMulti(ctx, items, ttl); err != nil {
			log.Printf("[Cache] Warning: Failed to write multiple items to L2: %v", err)
		}
	}

	return nil
}

// InvalidatePattern invalidates all keys matching a pattern
func (c *MultiTierCache) InvalidatePattern(ctx context.Context, pattern string) error {
	// For memory cache, we need to iterate (not efficient)
	// For Redis, we can use SCAN with pattern
	if c.l2 != nil {
		// This is a simplified implementation
		// In production, you'd want to use Redis SCAN
		return fmt.Errorf("pattern invalidation not fully implemented")
	}
	return nil
}

// Stats returns aggregated statistics
func (c *MultiTierCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	l1Stats := c.l1.Stats()

	totalHits := c.stats.l1Hits + c.stats.l2Hits + c.stats.l3Hits
	total := totalHits + c.stats.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(totalHits) / float64(total)
	}

	avgTime := time.Duration(0)
	if c.stats.calls > 0 {
		avgTime = c.stats.totalTime / time.Duration(c.stats.calls)
	}

	stats := map[string]interface{}{
		"l1_hits":      c.stats.l1Hits,
		"l2_hits":      c.stats.l2Hits,
		"l3_hits":      c.stats.l3Hits,
		"misses":       c.stats.misses,
		"total":        total,
		"hit_rate":     hitRate,
		"l1_hit_rate":  float64(c.stats.l1Hits) / float64(total),
		"avg_latency":  avgTime,
		"l1_size":      l1Stats.Size,
		"l1_evictions": l1Stats.Evictions,
	}

	if c.l2 != nil {
		l2Stats := c.l2.Stats()
		stats["l2_stats"] = l2Stats
	}

	return stats
}

// SetWriteThrough enables/disables write-through mode
func (c *MultiTierCache) SetWriteThrough(enabled bool) {
	c.writeThrough = enabled
}

// SetWriteBack enables/disables write-back mode
func (c *MultiTierCache) SetWriteBack(enabled bool) {
	c.writeBack = enabled
}
