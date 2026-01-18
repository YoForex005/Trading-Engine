package cache

import (
	"container/list"
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("key not found in cache")
	ErrExpired  = errors.New("cache entry expired")
)

// memoryEntry represents a cached item
type memoryEntry struct {
	key       string
	value     interface{}
	expiresAt time.Time
	size      int64
}

// MemoryCache implements an in-memory LRU cache with TTL support
type MemoryCache struct {
	mu sync.RWMutex

	// LRU tracking
	items    map[string]*list.Element
	lruList  *list.List

	// Configuration
	maxSize  int64 // Maximum cache size in bytes
	maxItems int   // Maximum number of items

	// Statistics
	stats struct {
		hits      int64
		misses    int64
		sets      int64
		deletes   int64
		evictions int64
		size      int64
		getTime   time.Duration
		setTime   time.Duration
		getCalls  int64
		setCalls  int64
	}

	// Eviction callback
	onEvict func(key string, value interface{})
}

// NewMemoryCache creates a new in-memory cache
// maxSize: maximum cache size in bytes (0 = unlimited)
// maxItems: maximum number of items (0 = unlimited)
func NewMemoryCache(maxSize int64, maxItems int) *MemoryCache {
	c := &MemoryCache{
		items:    make(map[string]*list.Element),
		lruList:  list.New(),
		maxSize:  maxSize,
		maxItems: maxItems,
	}

	// Start cleanup goroutine for expired entries
	go c.cleanupExpired()

	return c
}

// Get retrieves a value from cache
func (c *MemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	start := time.Now()
	c.mu.Lock()
	defer func() {
		c.stats.getTime += time.Since(start)
		c.stats.getCalls++
		c.mu.Unlock()
	}()

	elem, exists := c.items[key]
	if !exists {
		c.stats.misses++
		return nil, ErrNotFound
	}

	entry := elem.Value.(*memoryEntry)

	// Check expiration
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		c.removeLocked(key)
		c.stats.misses++
		return nil, ErrExpired
	}

	// Move to front (most recently used)
	c.lruList.MoveToFront(elem)
	c.stats.hits++

	return entry.value, nil
}

// Set stores a value in cache with TTL
func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	c.mu.Lock()
	defer func() {
		c.stats.setTime += time.Since(start)
		c.stats.setCalls++
		c.mu.Unlock()
	}()

	// Calculate size
	size := c.estimateSize(value)

	// Remove existing entry if present
	if elem, exists := c.items[key]; exists {
		c.removeLocked(key)
		c.lruList.Remove(elem)
	}

	// Evict if necessary
	for c.needsEviction(size) {
		c.evictOldest()
	}

	// Create entry
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	entry := &memoryEntry{
		key:       key,
		value:     value,
		expiresAt: expiresAt,
		size:      size,
	}

	// Add to cache
	elem := c.lruList.PushFront(entry)
	c.items[key] = elem
	c.stats.size += size
	c.stats.sets++

	return nil
}

// Delete removes a value from cache
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.removeLocked(key)
	c.stats.deletes++
	return nil
}

// Exists checks if a key exists
func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	elem, exists := c.items[key]
	if !exists {
		return false, nil
	}

	entry := elem.Value.(*memoryEntry)
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		return false, nil
	}

	return true, nil
}

// Clear removes all entries
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.lruList = list.New()
	c.stats.size = 0

	return nil
}

// GetMulti retrieves multiple values at once
func (c *MemoryCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, key := range keys {
		if value, err := c.Get(ctx, key); err == nil {
			result[key] = value
		}
	}

	return result, nil
}

// SetMulti stores multiple values at once
func (c *MemoryCache) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	for key, value := range items {
		if err := c.Set(ctx, key, value, ttl); err != nil {
			return err
		}
	}
	return nil
}

// Stats returns cache statistics
func (c *MemoryCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.stats.hits + c.stats.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.stats.hits) / float64(total)
	}

	avgGetTime := time.Duration(0)
	if c.stats.getCalls > 0 {
		avgGetTime = c.stats.getTime / time.Duration(c.stats.getCalls)
	}

	avgSetTime := time.Duration(0)
	if c.stats.setCalls > 0 {
		avgSetTime = c.stats.setTime / time.Duration(c.stats.setCalls)
	}

	return CacheStats{
		Hits:       c.stats.hits,
		Misses:     c.stats.misses,
		Sets:       c.stats.sets,
		Deletes:    c.stats.deletes,
		Evictions:  c.stats.evictions,
		Size:       c.stats.size,
		HitRate:    hitRate,
		AvgGetTime: avgGetTime,
		AvgSetTime: avgSetTime,
	}
}

// SetEvictionCallback sets a callback for eviction events
func (c *MemoryCache) SetEvictionCallback(callback func(key string, value interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEvict = callback
}

// Private methods

func (c *MemoryCache) removeLocked(key string) {
	if elem, exists := c.items[key]; exists {
		entry := elem.Value.(*memoryEntry)
		c.stats.size -= entry.size
		delete(c.items, key)
		c.lruList.Remove(elem)

		if c.onEvict != nil {
			c.onEvict(key, entry.value)
		}
	}
}

func (c *MemoryCache) needsEviction(newSize int64) bool {
	if c.maxItems > 0 && len(c.items) >= c.maxItems {
		return true
	}
	if c.maxSize > 0 && c.stats.size+newSize > c.maxSize {
		return true
	}
	return false
}

func (c *MemoryCache) evictOldest() {
	elem := c.lruList.Back()
	if elem == nil {
		return
	}

	entry := elem.Value.(*memoryEntry)
	c.removeLocked(entry.key)
	c.stats.evictions++
}

func (c *MemoryCache) estimateSize(value interface{}) int64 {
	// Estimate size based on JSON serialization
	data, err := json.Marshal(value)
	if err != nil {
		return 1024 // Default 1KB if can't serialize
	}
	return int64(len(data))
}

func (c *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		var toRemove []string

		for elem := c.lruList.Front(); elem != nil; elem = elem.Next() {
			entry := elem.Value.(*memoryEntry)
			if !entry.expiresAt.IsZero() && now.After(entry.expiresAt) {
				toRemove = append(toRemove, entry.key)
			}
		}

		for _, key := range toRemove {
			c.removeLocked(key)
		}

		c.mu.Unlock()
	}
}
