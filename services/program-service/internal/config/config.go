package config

import (
	"os"
)

type Config struct {
	Port            string
	DatabaseURL     string
	AuthService     string
	Environment     string
	JWTSecret       string
	LiteLLMEndpoint string
	RabbitMQURL     string
	UserService     string
	SettingsService string
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8084"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://app_user:password@localhost:5432/powerlifting_app?sslmode=disable"),
		AuthService:     getEnv("AUTH_SERVICE", "http://auth-service:8080"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		JWTSecret:       getEnv("JWT_SECRET", "your-secret-key"),
		LiteLLMEndpoint: getEnv("LITELLM_ENDPOINT", "http://litellm:4000"),
		RabbitMQURL:     getEnv("RABBITMQ_URL", "amqp://admin:changeme123@rabbitmq:5672/"),
		UserService:     getEnv("USER_SERVICE", "http://user-service:8081"),
		SettingsService: getEnv("SETTINGS_SERVICE", "http://settings-service:8082"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}