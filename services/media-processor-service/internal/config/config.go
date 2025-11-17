package config

import (
	"os"
	"strconv"
)

type Config struct {
	RabbitMQURL      string
	SpacesAccessKey  string
	SpacesSecretKey  string
	SpacesBucket     string
	SpacesRegion     string
	SpacesEndpoint   string
	CDNUrl           string
	FFmpegPath       string
	ThumbnailSize    int
}

func Load() *Config {
	thumbnailSize, _ := strconv.Atoi(getEnv("THUMBNAIL_SIZE", "640"))

	return &Config{
		RabbitMQURL:     getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		SpacesAccessKey: getEnv("SPACES_KEY", ""),
		SpacesSecretKey: getEnv("SPACES_SECRET", ""),
		SpacesBucket:    getEnv("SPACES_BUCKET", "nolift-videos"),
		SpacesRegion:    getEnv("SPACES_REGION", "us-east-1"),
		SpacesEndpoint:  getEnv("SPACES_ENDPOINT", ""),
		CDNUrl:          getEnv("CDN_URL", ""),
		FFmpegPath:      getEnv("FFMPEG_PATH", "ffmpeg"),
		ThumbnailSize:   thumbnailSize,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
