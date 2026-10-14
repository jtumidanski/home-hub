package config

import (
	"os"
	"strconv"
	"time"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore"
	"github.com/jtumidanski/home-hub/shared/go/database"
)

type Config struct {
	DB                 database.Config
	Port               string
	JWKSURL            string
	CategoryServiceURL string
	AccountServiceURL  string
	InternalToken      string
	RetentionInterval  time.Duration
	ObjectStore        objectstore.Config
}

func Load() Config {
	return Config{
		DB: database.Config{
			Host:     envOrDefault("DB_HOST", "postgres.home"),
			Port:     envOrDefault("DB_PORT", "5432"),
			User:     envOrDefault("DB_USER", "home_hub"),
			Password: envOrDefault("DB_PASSWORD", ""),
			DBName:   envOrDefault("DB_NAME", "home_hub"),
			Schema:   "recipe",
		},
		Port:               envOrDefault("PORT", "8080"),
		JWKSURL:            envOrDefault("JWKS_URL", "http://auth-service:8080/api/v1/auth/.well-known/jwks.json"),
		CategoryServiceURL: envOrDefault("CATEGORY_SERVICE_URL", "http://category-service:8080"),
		AccountServiceURL:  envOrDefault("ACCOUNT_SERVICE_URL", "http://account-service:8080"),
		InternalToken:      os.Getenv("INTERNAL_SERVICE_TOKEN"),
		RetentionInterval:  parseDuration(os.Getenv("RETENTION_INTERVAL"), 6*time.Hour),
		ObjectStore: objectstore.Config{
			Endpoint:  os.Getenv("S3_ENDPOINT"),
			Bucket:    envOrDefault("S3_BUCKET", "home-hub-recipes"),
			AccessKey: os.Getenv("S3_ACCESS_KEY"),
			SecretKey: os.Getenv("S3_SECRET_KEY"),
			UseSSL:    parseBool(os.Getenv("S3_USE_SSL"), false),
			Region:    envOrDefault("S3_REGION", "us-east-1"),
		},
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

func parseBool(v string, fallback bool) bool {
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
