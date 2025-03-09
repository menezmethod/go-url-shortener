package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/menezmethod/ref_go/internal/api/middleware"
	"github.com/menezmethod/ref_go/internal/domain"
	"github.com/menezmethod/ref_go/internal/metrics"
)

// LinkService defines the interface for link-related operations
type LinkService interface {
	CreateShortLink(ctx context.Context, req *domain.CreateShortLinkRequest) (*domain.ShortLink, error)
	GetShortLink(ctx context.Context, id string) (*domain.ShortLink, error)
	GetShortLinkByCode(ctx context.Context, code string) (*domain.ShortLink, error)
	UpdateShortLink(ctx context.Context, id string, req *domain.UpdateShortLinkRequest) (*domain.ShortLink, error)
	DeleteShortLink(ctx context.Context, id string) error
	ListShortLinks(ctx context.Context, page, pageSize int) ([]*domain.ShortLink, int, error)
	RecordClick(ctx context.Context, shortLinkID string, referrer, userAgent, ipAddress string) error
	GetLinkStats(ctx context.Context, shortLinkID string) (*domain.LinkStats, error)
}

// LinkHandler handles link-related routes
type LinkHandler struct {
	linkService LinkService
	baseURL     string
	metrics     *metrics.Metrics
}

// NewLinkHandler creates a new link handler
func NewLinkHandler(linkService LinkService, baseURL string, metrics *metrics.Metrics) *LinkHandler {
	return &LinkHandler{
		linkService: linkService,
		baseURL:     baseURL,
		metrics:     metrics,
	}
}

// CreateLink handles link creation
// @Summary Create a new short link
// @Description Create a new short link for a URL, optionally with a custom alias
// @Tags links
// @Accept json
// @Produce json
// @Param request body domain.CreateShortLinkRequest true "Link creation request"
// @Success 201 {object} domain.ShortLink "Link created successfully"
// @Failure 400 {object} map[string]string "Invalid request or URL"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /links [post]
func (h *LinkHandler) CreateLink(c *gin.Context) {
	logger := middleware.GetLogger(c)

	// Parse request body
	var req domain.CreateShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Info("Failed to decode request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create link
	link, err := h.linkService.CreateShortLink(c.Request.Context(), &req)
	if err != nil {
		logger.Info("Failed to create short link", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusCreated, link)
}

// GetLink handles link retrieval
// @Summary Get a short link by code
// @Description Get details of a short link using its code
// @Tags links
// @Accept json
// @Produce json
// @Param code path string true "Short link code"
// @Success 200 {object} domain.ShortLink "Link details"
// @Failure 400 {object} map[string]string "Invalid code"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Link not found"
// @Security BearerAuth
// @Router /links/{code} [get]
func (h *LinkHandler) GetLink(c *gin.Context) {
	logger := middleware.GetLogger(c)

	// Extract code from URL
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Link code is required"})
		return
	}

	// Get link by code
	link, err := h.linkService.GetShortLinkByCode(c.Request.Context(), code)
	if err != nil {
		logger.Info("Failed to get short link", zap.String("code", code), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, link)
}

// UpdateLink handles link updates
// @Summary Update a short link
// @Description Update properties of an existing short link
// @Tags links
// @Accept json
// @Produce json
// @Param code path string true "Short link code"
// @Param request body domain.UpdateShortLinkRequest true "Update request"
// @Success 200 {object} domain.ShortLink "Updated link"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Link not found"
// @Security BearerAuth
// @Router /links/{code} [put]
func (h *LinkHandler) UpdateLink(c *gin.Context) {
	logger := middleware.GetLogger(c)

	// Extract code from URL
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Link code is required"})
		return
	}

	// Get link by code first to get its ID
	link, err := h.linkService.GetShortLinkByCode(c.Request.Context(), code)
	if err != nil {
		logger.Info("Failed to get short link", zap.String("code", code), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	// Parse request body
	var req domain.UpdateShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Info("Failed to decode request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Update link using its ID
	updatedLink, err := h.linkService.UpdateShortLink(c.Request.Context(), link.ID, &req)
	if err != nil {
		logger.Info("Failed to update short link", zap.String("id", link.ID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return response
	c.JSON(http.StatusOK, updatedLink)
}

// DeleteLink handles link deletion
// @Summary Delete a short link
// @Description Delete a short link by its code
// @Tags links
// @Accept json
// @Produce json
// @Param code path string true "Short link code"
// @Success 204 "No content"
// @Failure 400 {object} map[string]string "Invalid code"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Link not found"
// @Security BearerAuth
// @Router /links/{code} [delete]
func (h *LinkHandler) DeleteLink(c *gin.Context) {
	logger := middleware.GetLogger(c)

	// Extract code from URL
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Link code is required"})
		return
	}

	// Get link by code first to get its ID
	link, err := h.linkService.GetShortLinkByCode(c.Request.Context(), code)
	if err != nil {
		logger.Info("Failed to get short link", zap.String("code", code), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	// Delete link using its ID
	if err := h.linkService.DeleteShortLink(c.Request.Context(), link.ID); err != nil {
		logger.Info("Failed to delete short link", zap.String("id", link.ID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete link"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListLinks handles listing links
func (h *LinkHandler) ListLinks(c *gin.Context) {
	logger := middleware.GetLogger(c)

	// Parse query parameters
	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")

	page := 1
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
	}

	pageSize := 10
	if pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 || pageSize > 100 {
			pageSize = 10
		}
	}

	// Get links
	links, total, err := h.linkService.ListShortLinks(c.Request.Context(), page, pageSize)
	if err != nil {
		logger.Error("Failed to list short links", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list links"})
		return
	}

	// Prepare response
	response := struct {
		Links []*domain.ShortLink `json:"links"`
		Meta  struct {
			Total   int `json:"total"`
			Page    int `json:"page"`
			PerPage int `json:"per_page"`
		} `json:"meta"`
	}{
		Links: links,
		Meta: struct {
			Total   int `json:"total"`
			Page    int `json:"page"`
			PerPage int `json:"per_page"`
		}{
			Total:   total,
			Page:    page,
			PerPage: pageSize,
		},
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// GetLinkStats handles retrieving link statistics
// @Summary Get link statistics
// @Description Get usage statistics for a short link
// @Tags links
// @Accept json
// @Produce json
// @Param code path string true "Short link code"
// @Success 200 {object} domain.LinkStats "Link statistics"
// @Failure 400 {object} map[string]string "Invalid code"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Link not found"
// @Security BearerAuth
// @Router /links/{code}/stats [get]
func (h *LinkHandler) GetLinkStats(c *gin.Context) {
	logger := middleware.GetLogger(c)

	// Extract code from URL
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Link code is required"})
		return
	}

	// Get link by code first to get its ID
	link, err := h.linkService.GetShortLinkByCode(c.Request.Context(), code)
	if err != nil {
		logger.Info("Failed to get short link", zap.String("code", code), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	// Get link stats using its ID
	stats, err := h.linkService.GetLinkStats(c.Request.Context(), link.ID)
	if err != nil {
		logger.Error("Failed to get link stats", zap.String("id", link.ID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get link statistics"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, stats)
}

// RedirectLink handles redirection for short links
func (h *LinkHandler) RedirectLink(c *gin.Context) {
	logger := middleware.GetLogger(c)

	logger.Debug("Starting redirect process")

	// Extract code from URL
	code := c.Param("code")
	if code == "" {
		logger.Info("Empty code parameter received")
		c.Status(http.StatusNotFound)
		return
	}

	logger.Info("Redirect request received",
		zap.String("code", code))

	// Get link by code
	link, err := h.linkService.GetShortLinkByCode(c.Request.Context(), code)
	if err != nil {
		logger.Info("Failed to get short link by code",
			zap.String("code", code),
			zap.Error(err),
		)
		c.Status(http.StatusNotFound)
		return
	}

	logger.Info("Link found for redirect",
		zap.String("link_id", link.ID),
		zap.String("original_url", link.URL.OriginalURL))

	// Check if link is active
	if !link.IsActive {
		logger.Info("Attempt to access inactive link", zap.String("code", code))
		c.Status(http.StatusNotFound)
		return
	}

	// Check if link is expired
	if link.ExpirationDate != nil && time.Now().UTC().After(*link.ExpirationDate) {
		logger.Info("Attempt to access expired link",
			zap.String("code", code),
			zap.Time("expiration", *link.ExpirationDate),
		)
		c.Status(http.StatusNotFound)
		return
	}

	// Record click asynchronously
	go func() {
		referrer := c.GetHeader("Referer")
		userAgent := c.GetHeader("User-Agent")
		ipAddress := c.ClientIP()

		// Create a new context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.linkService.RecordClick(ctx, link.ID, referrer, userAgent, ipAddress); err != nil {
			logger.Error("Failed to record click",
				zap.String("link_id", link.ID),
				zap.Error(err),
			)
		} else {
			logger.Info("Click recorded successfully",
				zap.String("link_id", link.ID))
		}
	}()

	// Log before redirect
	logger.Info("About to perform redirect",
		zap.String("link_id", link.ID),
		zap.String("original_url", link.URL.OriginalURL),
		zap.String("code", code))

	// Record redirect in metrics
	if h.metrics != nil {
		logger.Info("Recording redirect in metrics", zap.String("link_id", link.ID))
		h.metrics.RecordRedirect(link.ID)
	} else {
		logger.Error("Metrics collector is nil, cannot record redirect")
	}

	// Redirect to original URL
	c.Redirect(http.StatusMovedPermanently, link.URL.OriginalURL)

	// Log after redirect
	logger.Info("Redirect completed",
		zap.String("link_id", link.ID),
		zap.String("destination", link.URL.OriginalURL))
}
