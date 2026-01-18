package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements distributed caching with Redis
type RedisCache struct {
	client *redis.Client
	prefix string

	// Connection pool configuration
	poolSize     int
	minIdleConns int

	// Statistics
	mu    sync.RWMutex
	stats struct {
		hits      int64
		misses    int64
		sets      int64
		deletes   int64
		getTime   time.Duration
		setTime   time.Duration
		getCalls  int64
		setCalls  int64
		errors    int64
	}

	// Pipelining support
	pipelineSize int

	// Lua scripts
	scripts map[string]*redis.Script
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Address      string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Prefix       string
}

// DefaultRedisConfig returns default Redis configuration
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Address:      "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     100,
		MinIdleConns: 10,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		Prefix:       "rtx",
	}
}

// NewRedisCache creates a new Redis cache client
func NewRedisCache(config *RedisConfig) (*RedisCache, error) {
	if config == nil {
		config = DefaultRedisConfig()
	}

	client := redis.NewClient(&redis.Options{
		Addr:         config.Address,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	c := &RedisCache{
		client:       client,
		prefix:       config.Prefix,
		poolSize:     config.PoolSize,
		minIdleConns: config.MinIdleConns,
		pipelineSize: 100,
		scripts:      make(map[string]*redis.Script),
	}

	// Load Lua scripts
	c.loadScripts()

	return c, nil
}

// Get retrieves a value from Redis cache
func (c *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	start := time.Now()
	defer func() {
		c.mu.Lock()
		c.stats.getTime += time.Since(start)
		c.stats.getCalls++
		c.mu.Unlock()
	}()

	fullKey := c.makeKey(key)
	data, err := c.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			c.mu.Lock()
			c.stats.misses++
			c.mu.Unlock()
			return nil, ErrNotFound
		}
		c.mu.Lock()
		c.stats.errors++
		c.mu.Unlock()
		return nil, err
	}

	// Deserialize
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.stats.hits++
	c.mu.Unlock()

	return value, nil
}

// Set stores a value in Redis cache with TTL
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		c.mu.Lock()
		c.stats.setTime += time.Since(start)
		c.stats.setCalls++
		c.mu.Unlock()
	}()

	// Serialize
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	fullKey := c.makeKey(key)
	err = c.client.Set(ctx, fullKey, data, ttl).Err()
	if err != nil {
		c.mu.Lock()
		c.stats.errors++
		c.mu.Unlock()
		return err
	}

	c.mu.Lock()
	c.stats.sets++
	c.mu.Unlock()

	return nil
}

// Delete removes a value from Redis cache
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := c.makeKey(key)
	err := c.client.Del(ctx, fullKey).Err()
	if err != nil {
		c.mu.Lock()
		c.stats.errors++
		c.mu.Unlock()
		return err
	}

	c.mu.Lock()
	c.stats.deletes++
	c.mu.Unlock()

	return nil
}

// Exists checks if a key exists in Redis
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.makeKey(key)
	count, err := c.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Clear removes all entries with the prefix
func (c *RedisCache) Clear(ctx context.Context) error {
	pattern := c.makeKey("*")
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()

	pipe := c.client.Pipeline()
	count := 0

	for iter.Next(ctx) {
		pipe.Del(ctx, iter.Val())
		count++

		if count >= c.pipelineSize {
			if _, err := pipe.Exec(ctx); err != nil {
				return err
			}
			pipe = c.client.Pipeline()
			count = 0
		}
	}

	if count > 0 {
		if _, err := pipe.Exec(ctx); err != nil {
			return err
		}
	}

	return iter.Err()
}

// GetMulti retrieves multiple values at once using pipelining
func (c *RedisCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	pipe := c.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, key := range keys {
		fullKey := c.makeKey(key)
		cmds[key] = pipe.Get(ctx, fullKey)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	result := make(map[string]interface{})
	for key, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			continue // Skip missing keys
		}

		var value interface{}
		if err := json.Unmarshal(data, &value); err == nil {
			result[key] = value
		}
	}

	return result, nil
}

// SetMulti stores multiple values at once using pipelining
func (c *RedisCache) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	pipe := c.client.Pipeline()

	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}

		fullKey := c.makeKey(key)
		pipe.Set(ctx, fullKey, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Stats returns cache statistics
func (c *RedisCache) Stats() CacheStats {
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
		Size:       0, // Redis doesn't track this locally
		HitRate:    hitRate,
		AvgGetTime: avgGetTime,
		AvgSetTime: avgSetTime,
	}
}

// IncrementAtomic atomically increments a counter
func (c *RedisCache) IncrementAtomic(ctx context.Context, key string, delta int64) (int64, error) {
	fullKey := c.makeKey(key)
	return c.client.IncrBy(ctx, fullKey, delta).Result()
}

// SetNX sets a key only if it doesn't exist (useful for locks)
func (c *RedisCache) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, err
	}

	fullKey := c.makeKey(key)
	return c.client.SetNX(ctx, fullKey, data, ttl).Result()
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// Private methods

func (c *RedisCache) makeKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return c.prefix + ":" + key
}

func (c *RedisCache) loadScripts() {
	// Atomic operation scripts
	c.scripts["get_and_expire"] = redis.NewScript(`
		local value = redis.call('GET', KEYS[1])
		if value then
			redis.call('EXPIRE', KEYS[1], ARGV[1])
		end
		return value
	`)

	c.scripts["set_if_higher"] = redis.NewScript(`
		local current = redis.call('GET', KEYS[1])
		local new_value = tonumber(ARGV[1])
		if not current or tonumber(current) < new_value then
			redis.call('SET', KEYS[1], ARGV[1])
			if ARGV[2] then
				redis.call('EXPIRE', KEYS[1], ARGV[2])
			end
			return 1
		end
		return 0
	`)
}

// RunScript executes a Lua script
func (c *RedisCache) RunScript(ctx context.Context, scriptName string, keys []string, args ...interface{}) (interface{}, error) {
	script, ok := c.scripts[scriptName]
	if !ok {
		return nil, fmt.Errorf("script not found: %s", scriptName)
	}

	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = c.makeKey(key)
	}

	return script.Run(ctx, c.client, fullKeys, args...).Result()
}
