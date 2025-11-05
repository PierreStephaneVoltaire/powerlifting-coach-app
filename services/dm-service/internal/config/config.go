package config

import "os"

type Config struct {
	Port         string
	DatabaseURL  string
	RabbitMQURL  string
	AuthService  string
	JWTSecret    string
}

func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8084"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/dm_service?sslmode=disable"),
		RabbitMQURL:  getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		AuthService:  getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
