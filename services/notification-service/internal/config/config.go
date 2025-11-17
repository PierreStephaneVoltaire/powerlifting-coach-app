package config

import (
	"os"
)

type Config struct {
	Port              string
	Environment       string
	RabbitMQURL       string
	SMTPHost          string
	SMTPPort          string
	SMTPUsername      string
	SMTPPassword      string
	EmailFromAddress  string
	UserService       string
	CoachService      string
	ProgramService    string
	VideoService      string
	AppURL            string
}

func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "8086"),
		Environment:      getEnv("ENVIRONMENT", "development"),
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://admin:changeme123@rabbitmq:5672/"),
		SMTPHost:         getEnv("SMTP_HOST", "email-smtp.us-east-1.amazonaws.com"),
		SMTPPort:         getEnv("SMTP_PORT", "587"),
		SMTPUsername:     getEnv("SMTP_USERNAME", ""),
		SMTPPassword:     getEnv("SMTP_PASSWORD", ""),
		EmailFromAddress: getEnv("EMAIL_FROM_ADDRESS", "noreply@nolift.training"),
		UserService:      getEnv("USER_SERVICE", "http://user-service:8081"),
		CoachService:     getEnv("COACH_SERVICE", "http://coach-service:8085"),
		ProgramService:   getEnv("PROGRAM_SERVICE", "http://program-service:8084"),
		VideoService:     getEnv("VIDEO_SERVICE", "http://video-service:8082"),
		AppURL:           getEnv("APP_URL", "https://app.nolift.training"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}