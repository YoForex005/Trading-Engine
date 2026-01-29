package datapipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// StorageManager handles multi-tier data storage
type StorageManager struct {
	redis          *redis.Client
	config         *PipelineConfig

	// Storage tiers
	// Hot: Redis (last N ticks per symbol)
	// Warm: Could be TimescaleDB (not implemented in this version)
	// Cold: Could be S3/file storage (not implemented)

	ctx            context.Context
}

// NewStorageManager creates a new storage manager
func NewStorageManager(redis *redis.Client, config *PipelineConfig) *StorageManager {
	return &StorageManager{
		redis:  redis,
		config: config,
	}
}

// Start starts the storage manager
func (s *StorageManager) Start(ctx context.Context) error {
	s.ctx = ctx

	log.Println("[Storage] Storage manager started")
	return nil
}

// StoreTick stores a tick in hot storage (Redis)
func (s *StorageManager) StoreTick(tick *NormalizedTick) error {
	// Store in Redis sorted set by timestamp
	key := fmt.Sprintf("ticks:%s", tick.Symbol)

	// Serialize tick
	data, err := json.Marshal(tick)
	if err != nil {
		return err
	}

	// Add to sorted set (score = timestamp)
	score := float64(tick.Timestamp.Unix())
	if err := s.redis.ZAdd(s.ctx, key, redis.Z{
		Score:  score,
		Member: data,
	}).Err(); err != nil {
		return err
	}

	// Trim to keep only last N ticks (hot data retention)
	if err := s.redis.ZRemRangeByRank(s.ctx, key, 0, int64(-s.config.HotDataRetention-1)).Err(); err != nil {
		log.Printf("[Storage] Failed to trim ticks for %s: %v", tick.Symbol, err)
	}

	// Also store latest tick in a separate key for fast access
	latestKey := fmt.Sprintf("tick:latest:%s", tick.Symbol)
	if err := s.redis.Set(s.ctx, latestKey, data, 1*time.Hour).Err(); err != nil {
		log.Printf("[Storage] Failed to store latest tick for %s: %v", tick.Symbol, err)
	}

	return nil
}

// StoreOHLC stores an OHLC bar in Redis
func (s *StorageManager) StoreOHLC(bar *OHLCBar) error {
	// Store in Redis sorted set by timeframe
	key := fmt.Sprintf("ohlc:%s:%d", bar.Symbol, bar.Timeframe)

	// Serialize bar
	data, err := json.Marshal(bar)
	if err != nil {
		return err
	}

	// Add to sorted set (score = open time unix timestamp)
	score := float64(bar.OpenTime.Unix())
	if err := s.redis.ZAdd(s.ctx, key, redis.Z{
		Score:  score,
		Member: data,
	}).Err(); err != nil {
		return err
	}

	// Trim to keep only last 500 bars per timeframe
	if err := s.redis.ZRemRangeByRank(s.ctx, key, 0, -501).Err(); err != nil {
		log.Printf("[Storage] Failed to trim OHLC for %s: %v", bar.Symbol, err)
	}

	// Set expiry on the key (30 days)
	s.redis.Expire(s.ctx, key, 30*24*time.Hour)

	return nil
}

// GetLatestTick retrieves the latest tick for a symbol
func (s *StorageManager) GetLatestTick(symbol string) (*NormalizedTick, error) {
	key := fmt.Sprintf("tick:latest:%s", symbol)

	data, err := s.redis.Get(s.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("no tick found for %s", symbol)
		}
		return nil, err
	}

	var tick NormalizedTick
	if err := json.Unmarshal(data, &tick); err != nil {
		return nil, err
	}

	return &tick, nil
}

// GetRecentTicks retrieves recent ticks for a symbol
func (s *StorageManager) GetRecentTicks(symbol string, limit int) ([]*NormalizedTick, error) {
	if limit <= 0 {
		limit = 100
	}

	key := fmt.Sprintf("ticks:%s", symbol)

	// Get last N ticks (descending order)
	results, err := s.redis.ZRevRange(s.ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	ticks := make([]*NormalizedTick, 0, len(results))
	for _, data := range results {
		var tick NormalizedTick
		if err := json.Unmarshal([]byte(data), &tick); err != nil {
			log.Printf("[Storage] Failed to unmarshal tick: %v", err)
			continue
		}
		ticks = append(ticks, &tick)
	}

	return ticks, nil
}

// GetOHLCBars retrieves OHLC bars for a symbol and timeframe
func (s *StorageManager) GetOHLCBars(symbol string, timeframe Timeframe, limit int) ([]*OHLCBar, error) {
	if limit <= 0 {
		limit = 100
	}

	key := fmt.Sprintf("ohlc:%s:%d", symbol, timeframe)

	// Get last N bars (descending order)
	results, err := s.redis.ZRevRange(s.ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	bars := make([]*OHLCBar, 0, len(results))
	for _, data := range results {
		var bar OHLCBar
		if err := json.Unmarshal([]byte(data), &bar); err != nil {
			log.Printf("[Storage] Failed to unmarshal OHLC: %v", err)
			continue
		}
		bars = append(bars, &bar)
	}

	return bars, nil
}

// GetOHLCBarsByTimeRange retrieves OHLC bars within a time range
func (s *StorageManager) GetOHLCBarsByTimeRange(symbol string, timeframe Timeframe, start, end time.Time) ([]*OHLCBar, error) {
	key := fmt.Sprintf("ohlc:%s:%d", symbol, timeframe)

	// Get bars in time range
	results, err := s.redis.ZRangeByScore(s.ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", start.Unix()),
		Max: fmt.Sprintf("%d", end.Unix()),
	}).Result()
	if err != nil {
		return nil, err
	}

	bars := make([]*OHLCBar, 0, len(results))
	for _, data := range results {
		var bar OHLCBar
		if err := json.Unmarshal([]byte(data), &bar); err != nil {
			log.Printf("[Storage] Failed to unmarshal OHLC: %v", err)
			continue
		}
		bars = append(bars, &bar)
	}

	return bars, nil
}

// GetStorageStats returns storage statistics
func (s *StorageManager) GetStorageStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get total keys
	totalKeys := 0
	iter := s.redis.Scan(s.ctx, 0, "ticks:*", 0).Iterator()
	for iter.Next(s.ctx) {
		totalKeys++
	}

	stats["total_tick_keys"] = totalKeys

	// Get total OHLC keys
	totalOHLCKeys := 0
	ohlcIter := s.redis.Scan(s.ctx, 0, "ohlc:*", 0).Iterator()
	for ohlcIter.Next(s.ctx) {
		totalOHLCKeys++
	}

	stats["total_ohlc_keys"] = totalOHLCKeys

	// Get memory usage (Redis INFO)
	info := s.redis.Info(s.ctx, "memory").Val()
	stats["redis_info"] = info

	return stats, nil
}

// HealthCheck performs health check on storage
func (s *StorageManager) HealthCheck() ComponentHealth {
	// Check Redis connection
	if err := s.redis.Ping(s.ctx).Err(); err != nil {
		return ComponentHealth{
			Status:    "unhealthy",
			LastError: fmt.Sprintf("Redis ping failed: %v", err),
		}
	}

	// Get storage stats
	stats, err := s.GetStorageStats()
	if err != nil {
		return ComponentHealth{
			Status:    "degraded",
			LastError: fmt.Sprintf("Failed to get stats: %v", err),
		}
	}

	return ComponentHealth{
		Status:  "healthy",
		Metrics: stats,
	}
}

// CleanupOldData removes data older than retention period
func (s *StorageManager) CleanupOldData() error {
	log.Println("[Storage] Starting cleanup of old data...")

	// Cleanup ticks older than warm retention
	cutoffTime := time.Now().AddDate(0, 0, -s.config.WarmDataRetentionDays)
	cutoffScore := float64(cutoffTime.Unix())

	// Scan all tick keys
	iter := s.redis.Scan(s.ctx, 0, "ticks:*", 0).Iterator()
	cleanedKeys := 0

	for iter.Next(s.ctx) {
		key := iter.Val()

		// Remove old entries
		removed, err := s.redis.ZRemRangeByScore(s.ctx, key, "-inf", fmt.Sprintf("%f", cutoffScore)).Result()
		if err != nil {
			log.Printf("[Storage] Failed to cleanup %s: %v", key, err)
			continue
		}

		if removed > 0 {
			cleanedKeys++
			log.Printf("[Storage] Cleaned %d old ticks from %s", removed, key)
		}
	}

	// Cleanup OHLC older than retention
	ohlcIter := s.redis.Scan(s.ctx, 0, "ohlc:*", 0).Iterator()
	for ohlcIter.Next(s.ctx) {
		key := ohlcIter.Val()

		removed, err := s.redis.ZRemRangeByScore(s.ctx, key, "-inf", fmt.Sprintf("%f", cutoffScore)).Result()
		if err != nil {
			log.Printf("[Storage] Failed to cleanup %s: %v", key, err)
			continue
		}

		if removed > 0 {
			log.Printf("[Storage] Cleaned %d old OHLC bars from %s", removed, key)
		}
	}

	log.Printf("[Storage] Cleanup complete, processed %d keys", cleanedKeys)
	return nil
}
