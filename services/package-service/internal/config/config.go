package config

import (
	"os"
	"strconv"
	"time"

	"github.com/jtumidanski/home-hub/shared/go/database"
)

type Config struct {
	DB                        database.Config
	Port                      string
	JWKSURL                   string
	PollInterval              time.Duration
	PollIntervalUrgent        time.Duration
	ArchiveAfterDays          int
	DeleteAfterDays           int
	StaleAfterDays            int
	MaxActivePerHousehold     int
	USPSClientID              string
	USPSClientSecret          string
	UPSClientID               string
	UPSClientSecret           string
	FedExAPIKey               string
	FedExSecretKey            string
	FedExSandbox              bool
	TokenEncryptionKey        string
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
	archiveDays, _ := strconv.Atoi(envOrDefault("PACKAGE_ARCHIVE_AFTER_DAYS", "7"))
	if archiveDays < 1 {
		archiveDays = 7
	}
	deleteDays, _ := strconv.Atoi(envOrDefault("PACKAGE_DELETE_AFTER_DAYS", "30"))
	if deleteDays < 1 {
		deleteDays = 30
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
		Port:                   envOrDefault("PORT", "8080"),
		JWKSURL:                envOrDefault("JWKS_URL", "http://auth-service:8080/api/v1/auth/.well-known/jwks.json"),
		PollInterval:           time.Duration(pollMin) * time.Minute,
		PollIntervalUrgent:     time.Duration(pollUrgentMin) * time.Minute,
		ArchiveAfterDays:       archiveDays,
		DeleteAfterDays:        deleteDays,
		StaleAfterDays:         staleDays,
		MaxActivePerHousehold:  maxActive,
		USPSClientID:           envOrDefault("USPS_CLIENT_ID", ""),
		USPSClientSecret:       envOrDefault("USPS_CLIENT_SECRET", ""),
		UPSClientID:            envOrDefault("UPS_CLIENT_ID", ""),
		UPSClientSecret:        envOrDefault("UPS_CLIENT_SECRET", ""),
		FedExAPIKey:            envOrDefault("FEDEX_API_KEY", ""),
		FedExSecretKey:         envOrDefault("FEDEX_SECRET_KEY", ""),
		FedExSandbox:           envOrDefault("FEDEX_SANDBOX", "") == "true",
		TokenEncryptionKey:     envOrDefault("PACKAGE_TOKEN_ENCRYPTION_KEY", ""),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
