package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/menezmethod/ref_go/internal/config"
	"go.uber.org/zap"
)

// AuthService defines the interface for auth service
type AuthService interface {
	ValidateToken(token string) (string, error)
	IsAdmin(userID string) bool
}

// AuthMiddleware handles authentication and authorization
type AuthMiddleware struct {
	cfg     *config.Config
	logger  *zap.Logger
	authSvc AuthService
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(cfg *config.Config, logger *zap.Logger, authSvc AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		cfg:     cfg,
		logger:  logger,
		authSvc: authSvc,
	}
}

// RequireAuth ensures that a valid JWT token is provided
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// Check if Authorization header exists
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check for Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		// Validate token
		token := parts[1]
		userID, err := m.authSvc.ValidateToken(token)
		if err != nil {
			m.logger.Error("Failed to validate token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("user_id", userID)
		c.Next()
	}
}

// AdminOnly ensures that the user is an admin
func (m *AuthMiddleware) AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		if !m.authSvc.IsAdmin(userID.(string)) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
