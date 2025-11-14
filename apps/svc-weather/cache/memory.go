package cache

import (
	"context"
	"strings"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

// MemoryClient implements Client using in-memory cache
type MemoryClient struct {
	cache  *gocache.Cache
	logger logrus.FieldLogger
}

// NewMemoryClient creates a new in-memory cache client
func NewMemoryClient(logger logrus.FieldLogger) *MemoryClient {
	logger.Warn("Using in-memory cache (not suitable for multi-instance deployments)")

	return &MemoryClient{
		cache:  gocache.New(5*time.Minute, 10*time.Minute),
		logger: logger,
	}
}

// Get retrieves a value from in-memory cache
func (m *MemoryClient) Get(ctx context.Context, key string) ([]byte, error) {
	val, found := m.cache.Get(key)
	if !found {
		return nil, ErrCacheMiss{Key: key}
	}

	m.logger.WithField("key", key).Debug("Cache hit (memory)")
	return val.([]byte), nil
}

// Set stores a value in in-memory cache with TTL
func (m *MemoryClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.cache.Set(key, value, ttl)

	m.logger.WithFields(logrus.Fields{
		"key": key,
		"ttl": ttl.String(),
	}).Debug("Cache set (memory)")

	return nil
}

// Delete removes a key from in-memory cache
func (m *MemoryClient) Delete(ctx context.Context, key string) error {
	m.cache.Delete(key)

	m.logger.WithField("key", key).Debug("Cache deleted (memory)")
	return nil
}

// DeletePattern removes all keys matching a pattern
func (m *MemoryClient) DeletePattern(ctx context.Context, pattern string) error {
	// Convert Redis pattern to simple prefix matching
	prefix := strings.TrimSuffix(pattern, "*")

	deleted := 0
	for key := range m.cache.Items() {
		if strings.HasPrefix(key, prefix) {
			m.cache.Delete(key)
			deleted++
		}
	}

	m.logger.WithFields(logrus.Fields{
		"pattern": pattern,
		"count":   deleted,
	}).Debug("Cache pattern deleted (memory)")

	return nil
}

// Ping always succeeds for in-memory cache
func (m *MemoryClient) Ping(ctx context.Context) error {
	return nil
}
