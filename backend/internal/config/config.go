package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all configuration for the API
type Config struct {
	// Server
	Port           string
	Environment    string
	AllowedOrigins []string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret          string
	JWTAccessTokenTTL  int // minutes
	JWTRefreshTokenTTL int // days

	// License
	LicensePrivateKey string
	LicensePublicKey  string
	LicenseIssuer     string

	// Stripe
	StripeSecretKey      string
	StripeWebhookSecret  string
	StripeProPriceID     string
	StripeEntPriceID     string

	// Email (for password reset, etc.)
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	// Downloads
	DownloadsBucket string
	DownloadsRegion string

	// Turnstile (Cloudflare)
	TurnstileSecretKey string

	// Admin/Notifications
	AdminEmail string
	ResendAPIKey string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		Environment:        getEnv("ENVIRONMENT", "development"),
		AllowedOrigins:     strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost"), ","),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://savegress:localdev123@localhost:5432/savegress?sslmode=disable"),
		RedisURL:           getEnv("REDIS_URL", "redis://:localdev123@localhost:6379/0"),
		JWTSecret:          getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		JWTAccessTokenTTL:  15,  // 15 minutes
		JWTRefreshTokenTTL: 7,   // 7 days
		LicensePrivateKey:  getEnv("LICENSE_PRIVATE_KEY", ""),
		LicensePublicKey:   getEnv("LICENSE_PUBLIC_KEY", ""),
		LicenseIssuer:      getEnv("LICENSE_ISSUER", "license.savegress.io"),
		StripeSecretKey:    getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeProPriceID:   getEnv("STRIPE_PRO_PRICE_ID", ""),
		StripeEntPriceID:   getEnv("STRIPE_ENT_PRICE_ID", ""),
		SMTPHost:           getEnv("SMTP_HOST", ""),
		SMTPPort:           getEnv("SMTP_PORT", "587"),
		SMTPUser:           getEnv("SMTP_USER", ""),
		SMTPPassword:       getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:           getEnv("SMTP_FROM", "noreply@savegress.io"),
		DownloadsBucket:    getEnv("DOWNLOADS_BUCKET", "savegress-releases"),
		DownloadsRegion:    getEnv("DOWNLOADS_REGION", "eu-central-1"),
		TurnstileSecretKey: getEnv("TURNSTILE_SECRET_KEY", "1x0000000000000000000000000000000AA"), // Test key
		AdminEmail:         getEnv("ADMIN_EMAIL", ""),
		ResendAPIKey:       getEnv("RESEND_API_KEY", ""),
	}

	// Validate required fields in production
	if cfg.Environment == "production" {
		if cfg.JWTSecret == "dev-secret-change-in-production" {
			return nil, fmt.Errorf("JWT_SECRET must be set in production")
		}
		if cfg.LicensePrivateKey == "" {
			return nil, fmt.Errorf("LICENSE_PRIVATE_KEY must be set in production")
		}
		if cfg.StripeSecretKey == "" {
			return nil, fmt.Errorf("STRIPE_SECRET_KEY must be set in production")
		}
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
