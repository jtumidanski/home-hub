package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// RedisClient implements Client using Redis
type RedisClient struct {
	client *redis.Client
	logger logrus.FieldLogger
}

// NewRedisClient creates a new Redis-backed cache client
func NewRedisClient(redisURL string, logger logrus.FieldLogger) (*RedisClient, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	logger.Info("Connected to Redis")

	return &RedisClient{
		client: client,
		logger: logger,
	}, nil
}

// Get retrieves a value from Redis
func (r *RedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrCacheMiss{Key: key}
	}
	if err != nil {
		return nil, err
	}

	r.logger.WithField("key", key).Debug("Cache hit")
	return val, nil
}

// Set stores a value in Redis with TTL
func (r *RedisClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return err
	}

	r.logger.WithFields(logrus.Fields{
		"key": key,
		"ttl": ttl.String(),
	}).Debug("Cache set")

	return nil
}

// Delete removes a key from Redis
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	r.logger.WithField("key", key).Debug("Cache deleted")
	return nil
}

// DeletePattern removes all keys matching a pattern
func (r *RedisClient) DeletePattern(ctx context.Context, pattern string) error {
	var cursor uint64
	var keys []string

	// Scan for all matching keys
	for {
		var scanKeys []string
		var err error
		scanKeys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	// Delete all matching keys
	if len(keys) > 0 {
		err := r.client.Del(ctx, keys...).Err()
		if err != nil {
			return err
		}

		r.logger.WithFields(logrus.Fields{
			"pattern": pattern,
			"count":   len(keys),
		}).Debug("Cache pattern deleted")
	}

	return nil
}

// Ping checks if Redis is reachable
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}
