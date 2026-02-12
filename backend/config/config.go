package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port              string
	JWTSecret         string
	AdminPasswordHash string
	FrontendURL       string
	SetupBootstrapToken string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		JWTSecret:         getEnv("JWT_SECRET", "change-me-in-production"),
		AdminPasswordHash: getEnv("ADMIN_PASSWORD_HASH", ""),
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:3000"),
		SetupBootstrapToken: getEnv("SETUP_BOOTSTRAP_TOKEN", ""),
	}
}

func (c *Config) Validate() error {
	if c.JWTSecret == "" || c.JWTSecret == "change-me-in-production" {
		return fmt.Errorf("JWT_SECRET must be set to a strong, non-default value")
	}
	if c.FrontendURL == "" {
		return fmt.Errorf("FRONTEND_URL must be set")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
