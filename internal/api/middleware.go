package api

import (
	"net/http"
	"strings"

	"github.com/forfire912/machineServer/internal/config"
	"github.com/forfire912/machineServer/internal/service"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware provides authentication middleware
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Auth.Enabled {
			c.Next()
			return
		}

		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Check if it's a Bearer token or API key
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if !validateJWT(token, cfg.Auth.JWTSecret) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				c.Abort()
				return
			}
		} else if strings.HasPrefix(authHeader, "ApiKey ") {
			apiKey := strings.TrimPrefix(authHeader, "ApiKey ")
			if !validateAPIKey(apiKey, cfg.Auth.APIKeys) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
				c.Abort()
				return
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func validateJWT(token string, secret string) bool {
	// Simplified JWT validation
	// In production, use a proper JWT library
	return token != ""
}

func validateAPIKey(apiKey string, validKeys []string) bool {
	for _, key := range validKeys {
		if key == apiKey {
			return true
		}
	}
	return false
}

// AuditMiddleware logs all API requests
func AuditMiddleware(svc *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userID := c.GetString("user_id")
		if userID == "" {
			userID = "anonymous"
		}

		// Log the request
		resource := c.Request.URL.Path
		action := c.Request.Method
		ip := c.ClientIP()

		c.Next()

		// Log after request is processed
		svc.LogAudit(userID, action, resource, "", ip)
	}
}

// CORSMiddleware adds CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware provides basic rate limiting
func RateLimitMiddleware() gin.HandlerFunc {
	// This is a placeholder - in production use a proper rate limiting solution
	return func(c *gin.Context) {
		c.Next()
	}
}
