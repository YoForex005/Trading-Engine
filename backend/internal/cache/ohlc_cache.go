package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type OHLC struct {
	Symbol    string
	Interval  string
	Timestamp int64
	Open      string // Decimal as string
	High      string
	Low       string
	Close     string
	Volume    int64
}

type OHLCCache struct {
	client *redis.Client
}

func NewOHLCCache(client *redis.Client) *OHLCCache {
	return &OHLCCache{client: client}
}

// SetOHLC caches OHLC candle with interval-based TTL
func (c *OHLCCache) SetOHLC(ctx context.Context, symbol, interval string, candle OHLC) error {
	key := fmt.Sprintf("ohlc:%s:%s:%d", symbol, interval, candle.Timestamp)
	data, err := json.Marshal(candle)
	if err != nil {
		return err
	}

	// TTL based on interval
	ttl := map[string]time.Duration{
		"M1":  1 * time.Hour,
		"M5":  3 * time.Hour,
		"M15": 6 * time.Hour,
		"H1":  24 * time.Hour,
		"H4":  3 * 24 * time.Hour,
		"D1":  7 * 24 * time.Hour,
	}[interval]

	if ttl == 0 {
		ttl = 1 * time.Hour // Default fallback
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// GetOHLC retrieves cached candle
func (c *OHLCCache) GetOHLC(ctx context.Context, symbol, interval string, timestamp int64) (*OHLC, error) {
	key := fmt.Sprintf("ohlc:%s:%s:%d", symbol, interval, timestamp)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var candle OHLC
	if err := json.Unmarshal(data, &candle); err != nil {
		return nil, err
	}
	return &candle, nil
}
