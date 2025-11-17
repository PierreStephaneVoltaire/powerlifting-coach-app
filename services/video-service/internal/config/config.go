package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                string
	DatabaseURL         string
	AuthService         string
	Environment         string
	JWTSecret           string
	SpacesAccessKey     string
	SpacesSecretKey     string
	SpacesBucket        string
	SpacesRegion        string
	SpacesEndpoint      string
	CDNUrl              string
	RabbitMQURL         string
	MaxFileSize         int64
	AllowedExtensions   []string
	FFmpegPath          string
	ThumbnailSize       int
}

func Load() *Config {
	maxFileSize, _ := strconv.ParseInt(getEnv("MAX_FILE_SIZE", "104857600"), 10, 64) // 100MB default
	thumbnailSize, _ := strconv.Atoi(getEnv("THUMBNAIL_SIZE", "320"))

	return &Config{
		Port:              getEnv("PORT", "8082"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://app_user:password@localhost:5432/nolift_app?sslmode=disable"),
		AuthService:       getEnv("AUTH_SERVICE", "http://auth-service:8080"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		JWTSecret:         getEnv("JWT_SECRET", "your-secret-key"),
		SpacesAccessKey:   getEnv("SPACES_ACCESS_KEY", ""),
		SpacesSecretKey:   getEnv("SPACES_SECRET_KEY", ""),
		SpacesBucket:      getEnv("SPACES_BUCKET", "nolift-videos"),
		SpacesRegion:      getEnv("SPACES_REGION", "us-east-1"),
		SpacesEndpoint:    getEnv("SPACES_ENDPOINT", ""),
		CDNUrl:            getEnv("CDN_URL", ""),
		RabbitMQURL:       getEnv("RABBITMQ_URL", "amqp://admin:changeme123@rabbitmq:5672/"),
		MaxFileSize:       maxFileSize,
		AllowedExtensions: []string{".mp4", ".mov", ".avi", ".mkv", ".webm"},
		FFmpegPath:        getEnv("FFMPEG_PATH", "ffmpeg"),
		ThumbnailSize:     thumbnailSize,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}