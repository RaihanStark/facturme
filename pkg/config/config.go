// Package config provides configuration management for the application.
// It loads configuration from environment variables and .env files.
package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	JWTSecret          string
	Port               string
	SMTPHost           string
	SMTPPort           string
	SMTPUsername       string
	SMTPPassword       string
	SenderEmail        string
	SenderName         string
	AppURL             string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/worklio?sslmode=disable"),
		JWTSecret:           getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
		Port:                getEnv("PORT", "8080"),
		SMTPHost:            getEnv("SMTP_HOST", ""),
		SMTPPort:            getEnv("SMTP_PORT", "465"),
		SMTPUsername:        getEnv("SMTP_USERNAME", ""),
		SMTPPassword:        getEnv("SMTP_PASSWORD", ""),
		SenderEmail:         getEnv("SENDER_EMAIL", "noreply@yourdomain.com"),
		SenderName:          getEnv("SENDER_NAME", "FacturMe"),
		AppURL:              getEnv("APP_URL", "http://localhost:5173"),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
