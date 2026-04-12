package config

import (
	"os"
	"strconv"
	"time"

	"github.com/jtumidanski/home-hub/shared/go/database"
)

type Config struct {
	DB                    database.Config
	Port                  string
	JWKSURL               string
	PollInterval          time.Duration
	PollIntervalUrgent    time.Duration
	StaleAfterDays        int
	MaxActivePerHousehold int
	USPSClientID          string
	USPSClientSecret      string
	UPSClientID           string
	UPSClientSecret       string
	FedExAPIKey           string
	FedExSecretKey        string
	FedExSandbox          bool
	TokenEncryptionKey    string
	AccountServiceURL     string
	InternalToken         string
	RetentionInterval     time.Duration
}

func Load() Config {
	pollMin, _ := strconv.Atoi(envOrDefault("PACKAGE_POLL_INTERVAL_MINUTES", "30"))
	if pollMin < 1 {
		pollMin = 30
	}
	pollUrgentMin, _ := strconv.Atoi(envOrDefault("PACKAGE_POLL_INTERVAL_URGENT_MINUTES", "15"))
	if pollUrgentMin < 1 {
		pollUrgentMin = 15
	}
	staleDays, _ := strconv.Atoi(envOrDefault("PACKAGE_STALE_AFTER_DAYS", "14"))
	if staleDays < 1 {
		staleDays = 14
	}
	maxActive, _ := strconv.Atoi(envOrDefault("PACKAGE_MAX_ACTIVE_PER_HOUSEHOLD", "25"))
	if maxActive < 1 {
		maxActive = 25
	}

	return Config{
		DB: database.Config{
			Host:     envOrDefault("DB_HOST", "postgres.home"),
			Port:     envOrDefault("DB_PORT", "5432"),
			User:     envOrDefault("DB_USER", "home_hub"),
			Password: envOrDefault("DB_PASSWORD", ""),
			DBName:   envOrDefault("DB_NAME", "home_hub"),
			Schema:   "package",
		},
		Port:                  envOrDefault("PORT", "8080"),
		JWKSURL:               envOrDefault("JWKS_URL", "http://auth-service:8080/api/v1/auth/.well-known/jwks.json"),
		PollInterval:          time.Duration(pollMin) * time.Minute,
		PollIntervalUrgent:    time.Duration(pollUrgentMin) * time.Minute,
		StaleAfterDays:        staleDays,
		MaxActivePerHousehold: maxActive,
		USPSClientID:          envOrDefault("USPS_CLIENT_ID", ""),
		USPSClientSecret:      envOrDefault("USPS_CLIENT_SECRET", ""),
		UPSClientID:           envOrDefault("UPS_CLIENT_ID", ""),
		UPSClientSecret:       envOrDefault("UPS_CLIENT_SECRET", ""),
		FedExAPIKey:           envOrDefault("FEDEX_API_KEY", ""),
		FedExSecretKey:        envOrDefault("FEDEX_SECRET_KEY", ""),
		FedExSandbox:          envOrDefault("FEDEX_SANDBOX", "") == "true",
		TokenEncryptionKey:    envOrDefault("PACKAGE_TOKEN_ENCRYPTION_KEY", ""),
		AccountServiceURL:     envOrDefault("ACCOUNT_SERVICE_URL", "http://account-service:8080"),
		InternalToken:         os.Getenv("INTERNAL_SERVICE_TOKEN"),
		RetentionInterval:     parseDuration(os.Getenv("RETENTION_INTERVAL"), 6*time.Hour),
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
