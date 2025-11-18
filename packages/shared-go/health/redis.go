package health

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCheck checks the health of a Redis connection
type RedisCheck struct {
	client *redis.Client
	name   string
}

// NewRedisCheck creates a new Redis health check
func NewRedisCheck(client *redis.Client) Check {
	return &RedisCheck{
		client: client,
		name:   "redis",
	}
}

// NewRedisCheckWithName creates a new Redis health check with a custom name
func NewRedisCheckWithName(client *redis.Client, name string) Check {
	return &RedisCheck{
		client: client,
		name:   name,
	}
}

// Name returns the name of this check
func (c *RedisCheck) Name() string {
	return c.name
}

// Execute performs the Redis health check
func (c *RedisCheck) Execute(ctx context.Context) CheckResult {
	// Create a context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	// Measure execution time
	responseTime, err := measureTime(func() error {
		return c.client.Ping(checkCtx).Err()
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

	// Add Redis pool stats as metadata if available
	if stats := c.client.PoolStats(); stats != nil {
		result.Metadata = map[string]interface{}{
			"hits":        stats.Hits,
			"misses":      stats.Misses,
			"timeouts":    stats.Timeouts,
			"total_conns": stats.TotalConns,
			"idle_conns":  stats.IdleConns,
			"stale_conns": stats.StaleConns,
		}
	}

	return result
}
