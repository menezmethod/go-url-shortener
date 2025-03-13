package handlers

import (
	"net/http"
	"strconv"

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

// UpdateLinkForTest handles the update of a link for testing
func (h *MockLinkHandler) UpdateLinkForTest(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// First check if the link exists and belongs to the user
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

	// Check if user is authorized to update this link
	if link.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this link"})
		return
	}

	// Parse request body
	var req struct {
		OriginalURL string `json:"original_url,omitempty"`
		CustomAlias string `json:"custom_alias,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	updateReq := service.UpdateLinkRequest{
		OriginalURL: req.OriginalURL,
		CustomAlias: req.CustomAlias,
	}

	updatedLink, err := h.linkSvc.UpdateLink(id, updateReq)
	if err != nil {
		switch err {
		case domain.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		case domain.ErrConflict:
			c.JSON(http.StatusConflict, gin.H{"error": "Custom alias already taken"})
		default:
			h.logger.Error("Failed to update link", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, updatedLink)
}

// DeleteLinkForTest handles the deletion of a link for testing
func (h *MockLinkHandler) DeleteLinkForTest(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// First check if the link exists and belongs to the user
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

	// Check if user is authorized to delete this link
	if link.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this link"})
		return
	}

	// Delete the link
	if err := h.linkSvc.DeleteLink(id); err != nil {
		h.logger.Error("Failed to delete link", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListLinksForTest handles the listing of links for testing
func (h *MockLinkHandler) ListLinksForTest(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse pagination parameters
	page := 1
	perPage := 10

	// Get page from query parameters
	pageStr := c.Query("page")
	if pageStr != "" {
		if pageInt, err := parseInt(pageStr); err == nil && pageInt > 0 {
			page = pageInt
		}
	}

	// Get per_page from query parameters
	perPageStr := c.Query("per_page")
	if perPageStr != "" {
		if perPageInt, err := parseInt(perPageStr); err == nil && perPageInt > 0 && perPageInt <= 100 {
			perPage = perPageInt
		}
	}

	// Get links
	links, total, err := h.linkSvc.ListLinks(userID.(string), page, perPage)
	if err != nil {
		h.logger.Error("Failed to list links", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Prepare response
	response := gin.H{
		"links": links,
		"meta": gin.H{
			"total":    total,
			"page":     page,
			"per_page": perPage,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetLinkStatsForTest handles the retrieval of link statistics for testing
func (h *MockLinkHandler) GetLinkStatsForTest(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// First check if the link exists and belongs to the user
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

	// Check if user is authorized to view this link's stats
	if link.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to view this link's statistics"})
		return
	}

	// Get clicks with pagination
	page := 1
	perPage := 10

	// Get page from query parameters
	pageStr := c.Query("page")
	if pageStr != "" {
		if pageInt, err := parseInt(pageStr); err == nil && pageInt > 0 {
			page = pageInt
		}
	}

	// Get per_page from query parameters
	perPageStr := c.Query("per_page")
	if perPageStr != "" {
		if perPageInt, err := parseInt(perPageStr); err == nil && perPageInt > 0 && perPageInt <= 100 {
			perPage = perPageInt
		}
	}

	// Get clicks
	clicks, total, err := h.linkSvc.GetClicks(id, page, perPage)
	if err != nil {
		h.logger.Error("Failed to get link clicks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Prepare response
	response := gin.H{
		"link_id": id,
		"clicks":  clicks,
		"meta": gin.H{
			"total":    total,
			"page":     page,
			"per_page": perPage,
		},
	}

	c.JSON(http.StatusOK, response)
}

// RedirectLinkForTest handles the redirection of a link for testing
func (h *MockLinkHandler) RedirectLinkForTest(c *gin.Context) {
	shortURL := c.Param("short_url")

	// Get link by short URL
	link, err := h.linkSvc.GetLinkByShortURL(shortURL)
	if err != nil {
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		} else {
			h.logger.Error("Failed to get link by short URL", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	// Record click
	userAgent := c.Request.UserAgent()
	referer := c.Request.Referer()
	ipAddress := c.ClientIP()

	// Record click asynchronously to not block the redirect
	go func() {
		if err := h.linkSvc.RecordClick(link.ID, userAgent, referer, ipAddress); err != nil {
			h.logger.Error("Failed to record click", zap.Error(err))
		}
	}()

	// Redirect to original URL
	c.Redirect(http.StatusMovedPermanently, link.OriginalURL)
}

// Helper function to parse integers from query parameters
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
