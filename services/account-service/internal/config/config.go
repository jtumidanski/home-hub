package config

import (
	"os"
	"strings"

	"github.com/jtumidanski/home-hub/shared/go/database"
)

type Config struct {
	DB                 database.Config
	Port               string
	JWKSURL            string
	InternalToken      string
	ServiceURLs        map[string]string
	KafkaBrokers       []string
	TopicUserLifecycle string
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
		KafkaBrokers:       splitCSV(envOrDefault("BOOTSTRAP_SERVERS", "kafka-broker.kafka.svc.cluster.local:9092")),
		TopicUserLifecycle: envOrDefault("EVENT_TOPIC_USER_LIFECYCLE", "home-hub.user.lifecycle"),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// splitCSV splits a comma-separated list, trimming whitespace and dropping
// empty entries. Returns nil for an empty input so callers can distinguish
// "unset" from "set but empty" behavior if they wish.
func splitCSV(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
