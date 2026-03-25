package config

import (
	"os"
	"strconv"
	"time"

	"github.com/jtumidanski/home-hub/shared/go/database"
)

type Config struct {
	DB              database.Config
	Port            string
	JWKSURL         string
	RefreshInterval time.Duration
	CacheTTL        time.Duration
}

func Load() Config {
	return Config{
		DB: database.Config{
			Host:     envOrDefault("DB_HOST", "postgres.home"),
			Port:     envOrDefault("DB_PORT", "5432"),
			User:     envOrDefault("DB_USER", "home_hub"),
			Password: envOrDefault("DB_PASSWORD", ""),
			DBName:   envOrDefault("DB_NAME", "home_hub"),
			Schema:   "weather",
		},
		Port:            envOrDefault("PORT", "8080"),
		JWKSURL:         envOrDefault("JWKS_URL", "http://auth-service:8080/api/v1/auth/.well-known/jwks.json"),
		RefreshInterval: durationOrDefault("REFRESH_INTERVAL_MINUTES", 15*time.Minute),
		CacheTTL:        durationOrDefault("CACHE_TTL_MINUTES", 15*time.Minute),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func durationOrDefault(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if mins, err := strconv.Atoi(v); err == nil {
			return time.Duration(mins) * time.Minute
		}
	}
	return fallback
}
