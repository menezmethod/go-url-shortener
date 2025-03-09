package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/menezmethod/ref_go/internal/domain"
	"github.com/menezmethod/ref_go/internal/repository"
)

// URLShortenerService handles URL shortening operations
type URLShortenerService struct {
	urlRepo       repository.URLRepository
	linkRepo      repository.ShortLinkRepository
	clickRepo     repository.LinkClickRepository
	logger        *zap.Logger
	baseURL       string
	defaultExpiry time.Duration
}

// NewURLShortenerService creates a new URL shortener service
func NewURLShortenerService(
	urlRepo repository.URLRepository,
	linkRepo repository.ShortLinkRepository,
	clickRepo repository.LinkClickRepository,
	logger *zap.Logger,
	baseURL string,
	defaultExpiry time.Duration,
) *URLShortenerService {
	return &URLShortenerService{
		urlRepo:       urlRepo,
		linkRepo:      linkRepo,
		clickRepo:     clickRepo,
		logger:        logger,
		baseURL:       baseURL,
		defaultExpiry: defaultExpiry,
	}
}

// CreateShortLink creates a new short link
func (s *URLShortenerService) CreateShortLink(ctx context.Context, req *domain.CreateShortLinkRequest) (*domain.ShortLink, error) {
	// Validate URL
	if err := s.validateURL(req.URL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Generate hash for the URL
	hash := s.generateHash(req.URL)

	// Check if URL already exists
	existingURL, err := s.urlRepo.GetByHash(ctx, hash)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("checking existing URL: %w", err)
	}

	var urlID string
	if existingURL != nil {
		// URL already exists, use existing URL ID
		urlID = existingURL.ID
	} else {
		// Create new URL
		urlID = uuid.New().String()
		now := time.Now().UTC()
		newURL := &domain.URL{
			ID:          urlID,
			OriginalURL: req.URL,
			Hash:        hash,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := s.urlRepo.Create(ctx, newURL); err != nil {
			return nil, fmt.Errorf("creating URL: %w", err)
		}
	}

	// Generate short code or use custom alias
	var code string
	if req.CustomAlias != nil && *req.CustomAlias != "" {
		code = *req.CustomAlias

		// Check if custom alias is already in use
		existingLink, err := s.linkRepo.GetByCustomAlias(ctx, code)
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("checking existing custom alias: %w", err)
		}

		if existingLink != nil {
			return nil, fmt.Errorf("custom alias already in use")
		}
	} else {
		// Generate short code
		code = s.generateCode(hash)

		// Check for collisions and regenerate if necessary
		attempts := 0
		for attempts < 5 {
			existingLink, err := s.linkRepo.GetByCode(ctx, code)
			if err != nil && !strings.Contains(err.Error(), "not found") {
				return nil, fmt.Errorf("checking existing code: %w", err)
			}

			if existingLink == nil {
				// Code is available
				break
			}

			// Code collision, try with a different variation
			attempts++
			code = s.generateCode(hash + fmt.Sprintf("-%d", attempts))
		}

		if attempts >= 5 {
			return nil, fmt.Errorf("failed to generate unique code after %d attempts", attempts)
		}
	}

	// Set expiration date if provided or use default
	var expirationDate *time.Time
	if req.ExpirationDate != nil {
		expirationDate = req.ExpirationDate
	} else if s.defaultExpiry > 0 {
		expiry := time.Now().UTC().Add(s.defaultExpiry)
		expirationDate = &expiry
	}

	// Create short link
	now := time.Now().UTC()
	shortLink := &domain.ShortLink{
		ID:             uuid.New().String(),
		Code:           code,
		CustomAlias:    req.CustomAlias,
		URLID:          urlID,
		ExpirationDate: expirationDate,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.linkRepo.Create(ctx, shortLink); err != nil {
		return nil, fmt.Errorf("creating short link: %w", err)
	}

	// Retrieve URL data to include in response
	url, err := s.urlRepo.GetByID(ctx, urlID)
	if err != nil {
		return nil, fmt.Errorf("retrieving URL data: %w", err)
	}

	shortLink.URL = url
	return shortLink, nil
}

// GetShortLink retrieves a short link by ID
func (s *URLShortenerService) GetShortLink(ctx context.Context, id string) (*domain.ShortLink, error) {
	return s.linkRepo.GetByID(ctx, id)
}

// GetShortLinkByCode retrieves a short link by code
func (s *URLShortenerService) GetShortLinkByCode(ctx context.Context, code string) (*domain.ShortLink, error) {
	// Try to find by custom alias first
	link, err := s.linkRepo.GetByCustomAlias(ctx, code)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("checking custom alias: %w", err)
	}

	if link != nil {
		return link, nil
	}

	// Then try by code
	return s.linkRepo.GetByCode(ctx, code)
}

// UpdateShortLink updates a short link
func (s *URLShortenerService) UpdateShortLink(ctx context.Context, id string, req *domain.UpdateShortLinkRequest) (*domain.ShortLink, error) {
	// Get existing link
	link, err := s.linkRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("retrieving short link: %w", err)
	}

	// Update fields if provided
	if req.CustomAlias != nil {
		// Check if custom alias is already in use by another link
		if *req.CustomAlias != "" {
			existingLink, err := s.linkRepo.GetByCustomAlias(ctx, *req.CustomAlias)
			if err != nil && !strings.Contains(err.Error(), "not found") {
				return nil, fmt.Errorf("checking existing custom alias: %w", err)
			}

			if existingLink != nil && existingLink.ID != id {
				return nil, fmt.Errorf("custom alias already in use")
			}
		}
		link.CustomAlias = req.CustomAlias
	}

	if req.ExpirationDate != nil {
		link.ExpirationDate = req.ExpirationDate
	}

	if req.IsActive != nil {
		link.IsActive = *req.IsActive
	}

	link.UpdatedAt = time.Now().UTC()

	// Save updates
	if err := s.linkRepo.Update(ctx, link); err != nil {
		return nil, fmt.Errorf("updating short link: %w", err)
	}

	// Retrieve URL data
	url, err := s.urlRepo.GetByID(ctx, link.URLID)
	if err != nil {
		return nil, fmt.Errorf("retrieving URL data: %w", err)
	}

	link.URL = url
	return link, nil
}

// DeleteShortLink deletes a short link
func (s *URLShortenerService) DeleteShortLink(ctx context.Context, id string) error {
	return s.linkRepo.Delete(ctx, id)
}

// ListShortLinks lists all short links with pagination
func (s *URLShortenerService) ListShortLinks(ctx context.Context, page, pageSize int) ([]*domain.ShortLink, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// Get total count
	total, err := s.linkRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("counting short links: %w", err)
	}

	// Get links
	links, err := s.linkRepo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("listing short links: %w", err)
	}

	return links, total, nil
}

// RecordClick records a click on a short link
func (s *URLShortenerService) RecordClick(ctx context.Context, shortLinkID string, referrer, userAgent, ipAddress string) error {
	// Extract useful information from user agent
	browser, os, device := parseUserAgent(userAgent)

	// Create click record
	click := &domain.LinkClick{
		ID:          uuid.New().String(),
		ShortLinkID: shortLinkID,
		CreatedAt:   time.Now().UTC(),
	}

	// Set optional fields
	if referrer != "" {
		click.Referrer = &referrer
	}

	if userAgent != "" {
		click.UserAgent = &userAgent
	}

	if ipAddress != "" {
		click.IPAddress = &ipAddress
	}

	if browser != "" {
		click.Browser = &browser
	}

	if os != "" {
		click.OS = &os
	}

	if device != "" {
		click.Device = &device
	}

	// Save click asynchronously to not block redirection
	go func() {
		// Create a new context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.clickRepo.Create(ctx, click); err != nil {
			s.logger.Error("Failed to record click",
				zap.String("short_link_id", shortLinkID),
				zap.Error(err),
			)
		}
	}()

	return nil
}

// GetLinkStats gets statistics for a short link
func (s *URLShortenerService) GetLinkStats(ctx context.Context, shortLinkID string) (*domain.LinkStats, error) {
	return s.clickRepo.GetStatsByShortLinkID(ctx, shortLinkID)
}

// generateHash creates a hash for a URL
func (s *URLShortenerService) generateHash(originalURL string) string {
	hasher := sha256.New()
	hasher.Write([]byte(originalURL))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// generateCode creates a short code from a hash
func (s *URLShortenerService) generateCode(hash string) string {
	// Generate a short code of 6 characters
	bytes := []byte(hash)[:6]
	code := base64.URLEncoding.EncodeToString(bytes)

	// Remove padding and limit to 6 characters
	code = strings.TrimRight(code, "=")
	if len(code) > 6 {
		code = code[:6]
	}

	return code
}

// validateURL validates a URL
func (s *URLShortenerService) validateURL(rawURL string) error {
	// Check if URL is not empty
	if rawURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Parse URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use HTTP or HTTPS protocol")
	}

	// Check host
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a host")
	}

	return nil
}

// parseUserAgent extracts browser, OS and device information from user agent
func parseUserAgent(userAgent string) (browser, os, device string) {
	// This is a simple implementation - in a real project, you might use a proper
	// user agent parsing library like https://github.com/mssola/user_agent

	userAgent = strings.ToLower(userAgent)

	// Extract browser
	switch {
	case strings.Contains(userAgent, "chrome") && !strings.Contains(userAgent, "chromium"):
		browser = "Chrome"
	case strings.Contains(userAgent, "firefox"):
		browser = "Firefox"
	case strings.Contains(userAgent, "safari") && !strings.Contains(userAgent, "chrome"):
		browser = "Safari"
	case strings.Contains(userAgent, "edge"):
		browser = "Edge"
	case strings.Contains(userAgent, "opera"):
		browser = "Opera"
	default:
		browser = "Other"
	}

	// Extract OS
	switch {
	case strings.Contains(userAgent, "windows"):
		os = "Windows"
	case strings.Contains(userAgent, "mac os") || strings.Contains(userAgent, "macos"):
		os = "macOS"
	case strings.Contains(userAgent, "linux"):
		os = "Linux"
	case strings.Contains(userAgent, "android"):
		os = "Android"
	case strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad"):
		os = "iOS"
	default:
		os = "Other"
	}

	// Extract device type
	switch {
	case strings.Contains(userAgent, "mobile"):
		device = "Mobile"
	case strings.Contains(userAgent, "tablet"):
		device = "Tablet"
	case strings.Contains(userAgent, "ipad"):
		device = "Tablet"
	default:
		device = "Desktop"
	}

	return browser, os, device
}
