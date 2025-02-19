package cache

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisJsonCache struct {
	client *redis.Client
}

func NewRedisJsonCache(client *redis.Client) (*RedisJsonCache, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisJsonCache{
		client: client,
	}, nil
}

func (cache *RedisJsonCache) Set(ctx context.Context, key string, value any, opts ...Option) error {
	valueJson, err := json.Marshal(value)
	if err != nil {
		return err
	}

	var cacheOpts cacheOptions
	for _, opt := range opts {
		opt(&cacheOpts)
	}

	return cache.client.Set(ctx, key, valueJson, cacheOpts.exp).Err()
}

func (cache *RedisJsonCache) Get(ctx context.Context, key string, target any) error {
	valueJson, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(valueJson, target)
}

func (cache *RedisJsonCache) Del(ctx context.Context, key string) error {
	return cache.client.Del(ctx, key).Err()
}
