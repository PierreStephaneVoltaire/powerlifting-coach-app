package config

import "os"

type Config struct {
	Port            string
	LiteLLMEndpoint string
	LiteLLMAPIKey   string
	Environment     string
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8090"),
		LiteLLMEndpoint: getEnv("LITELLM_ENDPOINT", "http://litellm:4000"),
		LiteLLMAPIKey:   getEnv("LITELLM_API_KEY", ""),
		Environment:     getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
