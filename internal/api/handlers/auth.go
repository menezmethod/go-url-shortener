package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/menezmethod/ref_go/internal/api/middleware"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	ValidateMasterPassword(password string) bool
	GenerateToken() (string, error)
}

// AuthHandler handles authentication-related routes
type AuthHandler struct {
	authService AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// TokenRequest represents the token request payload
type TokenRequest struct {
	MasterPassword string `json:"master_password" binding:"required" example:"your_master_password"`
}

// TokenResponse represents the token response
type TokenResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// GenerateToken handles token generation
// @Summary Generate authentication token
// @Description Generate a JWT token using the master password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body TokenRequest true "Token request with master password"
// @Success 200 {object} TokenResponse "Token generated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized - Invalid master password"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/token [post]
func (h *AuthHandler) GenerateToken(c *gin.Context) {
	logger := middleware.GetLogger(c)

	// Parse request body
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Info("Failed to decode request body", zap.Error(err))
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Validate master password
	if !h.authService.ValidateMasterPassword(req.MasterPassword) {
		logger.Info("Invalid master password")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// Generate token
	token, err := h.authService.GenerateToken()
	if err != nil {
		logger.Error("Failed to generate token", zap.Error(err))
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	// Return response
	c.JSON(200, TokenResponse{Token: token})
}
