package config

import (
	"os"

	"github.com/jtumidanski/home-hub/shared/go/database"
)

type Config struct {
	DB            database.Config
	Port          string
	JWKSURL       string
	InternalToken string
	ServiceURLs   map[string]string
}

func Load() Config {
	return Config{
		DB: database.Config{
			Host:     envOrDefault("DB_HOST", "postgres.home"),
			Port:     envOrDefault("DB_PORT", "5432"),
			User:     envOrDefault("DB_USER", "home_hub"),
			Password: envOrDefault("DB_PASSWORD", ""),
			DBName:   envOrDefault("DB_NAME", "home_hub"),
			Schema:   "account",
		},
		Port:          envOrDefault("PORT", "8080"),
		JWKSURL:       envOrDefault("JWKS_URL", "http://auth-service:8080/api/v1/auth/.well-known/jwks.json"),
		InternalToken: os.Getenv("INTERNAL_SERVICE_TOKEN"),
		ServiceURLs: map[string]string{
			"productivity-service": envOrDefault("PRODUCTIVITY_URL", "http://productivity-service:8080"),
			"recipe-service":       envOrDefault("RECIPE_URL", "http://recipe-service:8080"),
			"tracker-service":      envOrDefault("TRACKER_URL", "http://tracker-service:8080"),
			"workout-service":      envOrDefault("WORKOUT_URL", "http://workout-service:8080"),
			"calendar-service":     envOrDefault("CALENDAR_URL", "http://calendar-service:8080"),
			"package-service":      envOrDefault("PACKAGE_URL", "http://package-service:8080"),
		},
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
