package cache

import (
	"context"
	"time"
)

// JsonCache is a cache with the JSON serialization
type JsonCache interface {
	// Set serializes value to the JSON and sets the value for the given key with the given options
	Set(ctx context.Context, key string, value any, opts ...Option) error

	// Get returns the deserialized value for the given key
	Get(ctx context.Context, key string, target any) error

	// Del deletes the value for the given key
	Del(ctx context.Context, key string) error
}

type Option func(*cacheOptions)

type cacheOptions struct {
	exp time.Duration
}

func WithExpiration(exp time.Duration) Option {
	return func(opts *cacheOptions) {
		opts.exp = exp
	}
}
