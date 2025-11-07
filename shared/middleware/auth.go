package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthConfig struct {
	JWTSecret    string
	AuthService  string // URL of auth service
	SkipPaths    []string
	RequiredRole string
}

type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	UserType string `json:"user_type"`
	Roles    []string `json:"roles"`
}

func AuthMiddleware(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for certain paths
		for _, path := range config.SkipPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		// Validate token with auth service
		claims, err := validateToken(c.Request.Context(), tokenString, config.AuthService)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Check required role if specified
		if config.RequiredRole != "" && !hasRole(claims.Roles, config.RequiredRole) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		// Add claims to context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_type", claims.UserType)
		c.Set("user_roles", claims.Roles)

		c.Next()
	}
}

func validateToken(ctx context.Context, token, authServiceURL string) (*Claims, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", authServiceURL+"/api/v1/auth/validate", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token validation failed: %s", string(body))
	}

	var result struct {
		Valid    bool     `json:"valid"`
		UserID   string   `json:"user_id"`
		Email    string   `json:"email"`
		UserType string   `json:"user_type"`
		Roles    []string `json:"roles"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return &Claims{
		UserID:   result.UserID,
		Email:    result.Email,
		UserType: result.UserType,
		Roles:    result.Roles,
	}, nil
}

func hasRole(roles []string, requiredRole string) bool {
	for _, role := range roles {
		if role == requiredRole {
			return true
		}
	}
	return false
}

// GetUserID extracts user ID from gin context
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(string)
	}
	return ""
}

// GetUserType extracts user type from gin context
func GetUserType(c *gin.Context) string {
	if userType, exists := c.Get("user_type"); exists {
		return userType.(string)
	}
	return ""
}