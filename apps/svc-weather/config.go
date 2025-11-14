package main

import (
	"fmt"
	"os"
	"time"
)

// Config holds all service configuration
type Config struct {
	// Redis
	RedisURL string

	// TTLs
	CurrentTTL  time.Duration
	ForecastTTL time.Duration
	StaleMax    time.Duration

	// Refresh
	RefreshJitter float64

	// Geohash
	GeohashPrecision int

	// Open-Meteo
	OpenMeteoBaseURL string
	OpenMeteoTimeout time.Duration

	// svc-users
	SvcUsersBaseURL    string
	SvcUsersTimeout    time.Duration
	HouseholdCacheTTL  time.Duration

	// Service
	ServicePort string
	LogLevel    string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() Config {
	return Config{
		// Redis
		RedisURL: getEnv("REDIS_URL", ""), // Empty = in-memory fallback

		// TTLs
		CurrentTTL:  parseDuration("CURRENT_TTL", 5*time.Minute),
		ForecastTTL: parseDuration("FORECAST_TTL", 1*time.Hour),
		StaleMax:    parseDuration("STALE_MAX", 24*time.Hour),

		// Refresh
		RefreshJitter: parseFloat("REFRESH_JITTER", 0.2),

		// Geohash
		GeohashPrecision: parseInt("GEOHASH_PREC", 5),

		// Open-Meteo
		OpenMeteoBaseURL: getEnv("OPENMETEO_BASE_URL", "https://api.open-meteo.com/v1/forecast"),
		OpenMeteoTimeout: parseDuration("OPENMETEO_TIMEOUT", 10*time.Second),

		// svc-users
		SvcUsersBaseURL:   getEnv("SVC_USERS_BASE_URL", "http://svc-users:8080"),
		SvcUsersTimeout:   parseDuration("SVC_USERS_TIMEOUT", 5*time.Second),
		HouseholdCacheTTL: parseDuration("HOUSEHOLD_CACHE_TTL", 24*time.Hour),

		// Service
		ServicePort: getEnv("SERVICE_PORT", "8080"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func parseFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		var f float64
		if _, err := fmt.Sscanf(value, "%f", &f); err == nil {
			return f
		}
	}
	return defaultValue
}

func parseInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var i int
		if _, err := fmt.Sscanf(value, "%d", &i); err == nil {
			return i
		}
	}
	return defaultValue
}
