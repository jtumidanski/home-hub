package cache

import (
	"context"
	"time"
)

// Client defines the interface for cache operations
type Client interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	Ping(ctx context.Context) error
}

// ErrCacheMiss is returned when a key is not found in the cache
type ErrCacheMiss struct {
	Key string
}

func (e ErrCacheMiss) Error() string {
	return "cache miss: " + e.Key
}

// IsCacheMiss returns true if the error is a cache miss
func IsCacheMiss(err error) bool {
	_, ok := err.(ErrCacheMiss)
	return ok
}
