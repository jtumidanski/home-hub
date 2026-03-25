package config

import (
	"os"
	"strings"

	"github.com/jtumidanski/home-hub/shared/go/database"
)

// Config holds all configuration for the auth service.
type Config struct {
	DB            database.Config
	Port          string
	JWTPrivateKey string
	JWTKeyID      string
	OIDC          OIDCConfig
}

// OIDCConfig holds OIDC provider configuration.
type OIDCConfig struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
}

// CallbackPath is the fixed OAuth callback path registered in the router.
const CallbackPath = "/api/v1/auth/callback/google"

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
		JWTPrivateKey: strings.ReplaceAll(os.Getenv("JWT_PRIVATE_KEY"), `\n`, "\n"),
		JWTKeyID:      envOrDefault("JWT_KEY_ID", "home-hub-1"),
		OIDC: OIDCConfig{
			IssuerURL:    envOrDefault("OIDC_ISSUER_URL", "https://accounts.google.com"),
			ClientID:     os.Getenv("OIDC_CLIENT_ID"),
			ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		},
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
