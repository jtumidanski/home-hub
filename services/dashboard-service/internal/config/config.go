package config

import (
	"os"
	"strings"
	"time"

	"github.com/jtumidanski/home-hub/shared/go/database"
)

type Config struct {
	DB                 database.Config
	Port               string
	JWKSURL            string
	AccountServiceURL  string
	InternalToken      string
	RetentionInterval  time.Duration
	KafkaBrokers       []string
	TopicUserLifecycle string
	KafkaConsumerGroup string
}

func Load() Config {
	brokers := strings.Split(envOrDefault("BOOTSTRAP_SERVERS", "kafka-broker.kafka.svc.cluster.local:9092"), ",")
	return Config{
		DB: database.Config{
			Host:     envOrDefault("DB_HOST", "postgres.home"),
			Port:     envOrDefault("DB_PORT", "5432"),
			User:     envOrDefault("DB_USER", "home_hub"),
			Password: envOrDefault("DB_PASSWORD", ""),
			DBName:   envOrDefault("DB_NAME", "home_hub"),
			Schema:   "dashboard",
		},
		Port:               envOrDefault("PORT", "8080"),
		JWKSURL:            envOrDefault("JWKS_URL", "http://auth-service:8080/api/v1/auth/.well-known/jwks.json"),
		AccountServiceURL:  envOrDefault("ACCOUNT_SERVICE_URL", "http://account-service:8080"),
		InternalToken:      os.Getenv("INTERNAL_SERVICE_TOKEN"),
		RetentionInterval:  parseDuration(os.Getenv("RETENTION_INTERVAL"), 6*time.Hour),
		KafkaBrokers:       brokers,
		TopicUserLifecycle: envOrDefault("EVENT_TOPIC_USER_LIFECYCLE", "home-hub.user.lifecycle"),
		KafkaConsumerGroup: envOrDefault("KAFKA_CONSUMER_GROUP", "dashboard-service"),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(v string, fallback time.Duration) time.Duration {
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
