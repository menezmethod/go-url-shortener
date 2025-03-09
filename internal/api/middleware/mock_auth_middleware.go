package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/menezmethod/ref_go/internal/config"
	"go.uber.org/zap"
)

// MockAuthService defines the interface for auth service testing
type MockAuthService interface {
	ValidateToken(token string) (string, error)
	IsAdmin(userID string) bool
}

// MockAuthMiddleware handles authentication and authorization for testing
type MockAuthMiddleware struct {
	cfg     *config.Config
	logger  *zap.Logger
	authSvc MockAuthService
}

// NewMockAuthMiddleware creates a new MockAuthMiddleware for testing
func NewMockAuthMiddleware(cfg *config.Config, logger *zap.Logger, authSvc MockAuthService) *MockAuthMiddleware {
	return &MockAuthMiddleware{
		cfg:     cfg,
		logger:  logger,
		authSvc: authSvc,
	}
}

// RequireAuthForTest ensures that a valid JWT token is provided (for testing)
func (m *MockAuthMiddleware) RequireAuthForTest() gin.HandlerFunc {
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

// AdminOnlyForTest ensures that the user is an admin (for testing)
func (m *MockAuthMiddleware) AdminOnlyForTest() gin.HandlerFunc {
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
