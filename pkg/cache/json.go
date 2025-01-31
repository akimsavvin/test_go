package cache

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
)

type JsonCache interface {
	Set(ctx context.Context, key string, value any, opts ...Option) error
	Get(ctx context.Context, key string, target any) error
	Del(ctx context.Context, key string) error
}

type Option func(*cacheOptions)

type redisJsonCache struct {
	client *redis.Client
}

type cacheOptions struct {
	exp time.Duration
}

func WithExpiration(exp time.Duration) Option {
	return func(opts *cacheOptions) {
		opts.exp = exp
	}
}

func NewRedisJsonCache(addr string, user string, pass string) (JsonCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: user,
		Password: pass,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &redisJsonCache{
		client: client,
	}, nil
}

func (cache *redisJsonCache) Set(ctx context.Context, key string, value any, opts ...Option) error {
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

func (cache *redisJsonCache) Get(ctx context.Context, key string, target any) error {
	valueJson, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(valueJson, target)
}

func (cache *redisJsonCache) Del(ctx context.Context, key string) error {
	return cache.client.Del(ctx, key).Err()
}
