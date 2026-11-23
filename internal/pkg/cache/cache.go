package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func New(client *redis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) Get(ctx context.Context, key string) (string, bool) {
	if c == nil || c.client == nil {
		return "", false
	}
	value, err := c.client.Get(ctx, key).Result()
	return value, err == nil
}

func (c *Cache) Set(ctx context.Context, key, value string, ttl time.Duration) {
	if c == nil || c.client == nil {
		return
	}
	_ = c.client.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, key string) {
	if c == nil || c.client == nil {
		return
	}
	_ = c.client.Del(ctx, key).Err()
}
