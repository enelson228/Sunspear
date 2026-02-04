package config

import "os"

type Config struct {
	Port              string
	JWTSecret         string
	AdminPasswordHash string
	FrontendURL       string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		JWTSecret:         getEnv("JWT_SECRET", "change-me-in-production"),
		AdminPasswordHash: getEnv("ADMIN_PASSWORD_HASH", ""),
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
