package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS middleware configuration
type CORSConfig struct {
	AllowedOrigins []string
}

// CORSMiddleware creates a CORS middleware with the provided configuration
func CORSMiddleware(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if isOriginAllowed(origin, config.AllowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// DefaultCORSMiddleware creates a CORS middleware with default allowed origins
func DefaultCORSMiddleware() gin.HandlerFunc {
	allowedOrigins := []string{
		"https://app.nolift.training",
		"http://localhost:3000",
		"http://localhost:5173",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:5173",
	}

	// Allow additional origins from environment variable
	if extraOrigins := os.Getenv("ALLOWED_ORIGINS"); extraOrigins != "" {
		origins := strings.Split(extraOrigins, ",")
		for _, origin := range origins {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowedOrigins = append(allowedOrigins, origin)
			}
		}
	}

	return CORSMiddleware(CORSConfig{
		AllowedOrigins: allowedOrigins,
	})
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == origin {
			return true
		}
		// Support wildcard matching for development
		if strings.Contains(allowed, "*") {
			pattern := strings.ReplaceAll(allowed, "*", "")
			if strings.Contains(origin, pattern) {
				return true
			}
		}
	}

	return false
}
