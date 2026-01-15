package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Tick struct {
	Symbol string
	Bid    string // Decimal as string
	Ask    string // Decimal as string
	Time   int64
}

type TickCache struct {
	client *redis.Client
}

func NewTickCache(client *redis.Client) *TickCache {
	return &TickCache{client: client}
}

// SetTick caches latest tick with 60s TTL
func (c *TickCache) SetTick(ctx context.Context, symbol string, tick Tick) error {
	key := fmt.Sprintf("tick:%s:latest", symbol)
	data, err := json.Marshal(tick)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, 60*time.Second).Err()
}

// GetTick retrieves cached tick
func (c *TickCache) GetTick(ctx context.Context, symbol string) (*Tick, error) {
	key := fmt.Sprintf("tick:%s:latest", symbol)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var tick Tick
	if err := json.Unmarshal(data, &tick); err != nil {
		return nil, err
	}
	return &tick, nil
}
