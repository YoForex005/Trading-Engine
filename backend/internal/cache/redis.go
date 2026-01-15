package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr string) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr, // e.g., "localhost:6379"
		DB:       0,    // Default DB
		PoolSize: 10,   // Connection pool
	})

	return &RedisClient{client: rdb}
}

func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
