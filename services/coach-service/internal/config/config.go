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
	UserService  string
	ProgramService string
	VideoService   string
	RabbitMQURL    string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8085"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://app_user:password@localhost:5432/powerlifting_app?sslmode=disable"),
		AuthService:    getEnv("AUTH_SERVICE", "http://auth-service:8080"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		UserService:    getEnv("USER_SERVICE", "http://user-service:8081"),
		ProgramService: getEnv("PROGRAM_SERVICE", "http://program-service:8084"),
		VideoService:   getEnv("VIDEO_SERVICE", "http://video-service:8082"),
		RabbitMQURL:    getEnv("RABBITMQ_URL", "amqp://admin:changeme123@rabbitmq:5672/"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}