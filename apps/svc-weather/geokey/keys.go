package geokey

import (
	"fmt"
)

// KeyBuilder constructs cache keys from geokeys
type KeyBuilder struct{}

// NewKeyBuilder creates a new KeyBuilder
func NewKeyBuilder() KeyBuilder {
	return KeyBuilder{}
}

// CurrentKey builds a cache key for current weather
// Pattern: weather:current:{geokey}
func (k KeyBuilder) CurrentKey(geokey string) string {
	return fmt.Sprintf("weather:current:%s", geokey)
}

// ForecastKey builds a cache key for forecast weather
// Pattern: weather:forecast:{geokey}:7d
func (k KeyBuilder) ForecastKey(geokey string, days int) string {
	return fmt.Sprintf("weather:forecast:%s:%dd", geokey, days)
}

// Pattern returns the cache key pattern for listing/deletion
func (k KeyBuilder) Pattern() string {
	return "weather:*"
}
