package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/menezmethod/ref_go/internal/config"
	"github.com/menezmethod/ref_go/internal/domain"
	"github.com/menezmethod/ref_go/internal/service"
	"go.uber.org/zap"
)

// MockLinkService defines the interface for link service testing
type MockLinkService interface {
	CreateLink(req service.CreateLinkRequest) (*domain.Link, error)
	GetLink(id string) (*domain.Link, error)
	GetLinkByShortURL(shortURL string) (*domain.Link, error)
	UpdateLink(id string, req service.UpdateLinkRequest) (*domain.Link, error)
	DeleteLink(id string) error
	ListLinks(userID string, page, perPage int) ([]*domain.Link, int, error)
	RecordClick(linkID, userAgent, referer, ipAddress string) error
	GetClicks(linkID string, page, perPage int) ([]*domain.Click, int, error)
}

// MockLinkHandler handles HTTP requests related to links for testing
type MockLinkHandler struct {
	cfg     *config.Config
	logger  *zap.Logger
	linkSvc MockLinkService
}

// NewMockLinkHandler creates a new MockLinkHandler for testing
func NewMockLinkHandler(cfg *config.Config, logger *zap.Logger, linkSvc MockLinkService) *MockLinkHandler {
	return &MockLinkHandler{
		cfg:     cfg,
		logger:  logger,
		linkSvc: linkSvc,
	}
}

// CreateLinkForTest handles the creation of a new link for testing
func (h *MockLinkHandler) CreateLinkForTest(c *gin.Context) {
	var req struct {
		OriginalURL string `json:"original_url" binding:"required"`
		CustomAlias string `json:"custom_alias,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	createReq := service.CreateLinkRequest{
		UserID:      userID.(string),
		OriginalURL: req.OriginalURL,
		CustomAlias: req.CustomAlias,
	}

	link, err := h.linkSvc.CreateLink(createReq)
	if err != nil {
		switch err {
		case domain.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		case domain.ErrConflict:
			c.JSON(http.StatusConflict, gin.H{"error": "Custom alias already taken"})
		default:
			h.logger.Error("Failed to create link", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, link)
}

// GetLinkForTest handles the retrieval of a link by ID for testing
func (h *MockLinkHandler) GetLinkForTest(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	link, err := h.linkSvc.GetLink(id)
	if err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		} else {
			h.logger.Error("Failed to get link", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Check if user is authorized to view this link
	if link.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to view this link"})
		return
	}

	c.JSON(http.StatusOK, link)
}
