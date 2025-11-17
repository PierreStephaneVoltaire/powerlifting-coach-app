package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                   string
	KeycloakURL            string
	KeycloakRealm          string
	KeycloakClientID       string
	KeycloakSecret         string
	KeycloakAdminUser      string
	KeycloakAdminPassword  string
	JWTSecret              string
	DatabaseURL            string
	Environment            string
}

func Load() *Config {
	return &Config{
		Port:                  getEnv("PORT", "8080"),
		KeycloakURL:           getEnv("KEYCLOAK_URL", "http://keycloak:8080"),
		KeycloakRealm:         getEnv("KEYCLOAK_REALM", "powerlifting"),
		KeycloakClientID:      getEnv("KEYCLOAK_CLIENT_ID", "powerlifting-app"),
		KeycloakSecret:        getEnv("KEYCLOAK_CLIENT_SECRET", ""),
		KeycloakAdminUser:     getEnv("KEYCLOAK_ADMIN_USER", "admin"),
		KeycloakAdminPassword: getEnv("KEYCLOAK_ADMIN_PASSWORD", ""),
		JWTSecret:             getEnv("JWT_SECRET", "your-secret-key"),
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		Environment:           getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}