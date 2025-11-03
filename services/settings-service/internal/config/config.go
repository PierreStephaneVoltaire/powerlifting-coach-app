package config

import (
	"os"
)

type Config struct {
	Port         string
	DatabaseURL  string
	AuthService  string
	Environment  string
	JWTSecret    string
}

func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8083"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://app_user:password@localhost:5432/powerlifting_app?sslmode=disable"),
		AuthService:  getEnv("AUTH_SERVICE", "http://auth-service:8080"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}