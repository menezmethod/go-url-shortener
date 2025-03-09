package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/menezmethod/ref_go/internal/auth"
)

// AuthService interface abstracts the token validation functionality
type AuthService interface {
	ValidateToken(token string) (*auth.TokenClaims, error)
}

// Authentication middleware checks for valid JWT token
func Authentication(authService AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get logger from context
		logger := GetLogger(c)

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Info("Missing Authorization header")
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		// Check token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			logger.Info("Invalid Authorization header format")
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			logger.Info("Invalid token", zap.Error(err))
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		// Store claims in context
		c.Set("claims", claims)

		// Continue to the next handler
		c.Next()
	}
}

// GetTokenClaims retrieves token claims from context
func GetTokenClaims(c *gin.Context) *auth.TokenClaims {
	if claims, exists := c.Get("claims"); exists {
		return claims.(*auth.TokenClaims)
	}
	return nil
}
