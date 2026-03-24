package config

import (
	"os"

	"github.com/jtumidanski/home-hub/shared/go/database"
)

// Config holds all configuration for the auth service.
type Config struct {
	DB            database.Config
	Port          string
	JWTPrivateKey string
	JWTKeyID      string
}

// Load reads configuration from environment variables.
func Load() Config {
	return Config{
		DB: database.Config{
			Host:     envOrDefault("DB_HOST", "postgres.home"),
			Port:     envOrDefault("DB_PORT", "5432"),
			User:     envOrDefault("DB_USER", "home_hub"),
			Password: envOrDefault("DB_PASSWORD", ""),
			DBName:   envOrDefault("DB_NAME", "home_hub"),
			Schema:   "auth",
		},
		Port:          envOrDefault("PORT", "8080"),
		JWTPrivateKey: os.Getenv("JWT_PRIVATE_KEY"),
		JWTKeyID:      envOrDefault("JWT_KEY_ID", "home-hub-1"),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
