package config

import (
	"os"
	"strconv"
	"time"

	"github.com/jtumidanski/home-hub/shared/go/database"
)

const CallbackPath = "/api/v1/calendar/connections/google/callback"

type Config struct {
	DB                       database.Config
	Port                     string
	JWKSURL                  string
	GoogleCalendarClientID   string
	GoogleCalendarSecret     string
	TokenEncryptionKey       string
	SyncInterval             time.Duration
}

func Load() Config {
	intervalMin, _ := strconv.Atoi(envOrDefault("SYNC_INTERVAL_MINUTES", "15"))
	if intervalMin < 1 {
		intervalMin = 15
	}

	return Config{
		DB: database.Config{
			Host:     envOrDefault("DB_HOST", "postgres.home"),
			Port:     envOrDefault("DB_PORT", "5432"),
			User:     envOrDefault("DB_USER", "home_hub"),
			Password: envOrDefault("DB_PASSWORD", ""),
			DBName:   envOrDefault("DB_NAME", "home_hub"),
			Schema:   "calendar",
		},
		Port:                     envOrDefault("PORT", "8080"),
		JWKSURL:                  envOrDefault("JWKS_URL", "http://auth-service:8080/api/v1/auth/.well-known/jwks.json"),
		GoogleCalendarClientID:   envOrDefault("GOOGLE_CALENDAR_CLIENT_ID", ""),
		GoogleCalendarSecret:     envOrDefault("GOOGLE_CALENDAR_CLIENT_SECRET", ""),
		TokenEncryptionKey:       envOrDefault("CALENDAR_TOKEN_ENCRYPTION_KEY", ""),
		SyncInterval:             time.Duration(intervalMin) * time.Minute,
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
