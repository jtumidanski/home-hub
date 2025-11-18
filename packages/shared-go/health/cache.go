package health

import (
	"context"
	"time"
)

// CachePinger defines the interface for cache clients that support ping
type CachePinger interface {
	Ping(ctx context.Context) error
}

// CacheCheck checks the health of a cache client
type CacheCheck struct {
	client CachePinger
	name   string
}

// NewCacheCheck creates a new cache health check
func NewCacheCheck(client CachePinger) Check {
	return &CacheCheck{
		client: client,
		name:   "cache",
	}
}

// NewCacheCheckWithName creates a new cache health check with a custom name
func NewCacheCheckWithName(client CachePinger, name string) Check {
	return &CacheCheck{
		client: client,
		name:   name,
	}
}

// Name returns the name of this check
func (c *CacheCheck) Name() string {
	return c.name
}

// Execute performs the cache health check
func (c *CacheCheck) Execute(ctx context.Context) CheckResult {
	// Create a context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	// Measure execution time
	responseTime, err := measureTime(func() error {
		return c.client.Ping(checkCtx)
	})

	result := CheckResult{
		ResponseTimeMs: responseTime,
	}

	if err != nil {
		result.Status = StatusUnhealthy
		result.Error = err.Error()
	} else {
		result.Status = StatusHealthy
	}

	return result
}
