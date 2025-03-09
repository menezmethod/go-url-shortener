package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/menezmethod/ref_go/internal/cache"
	"github.com/menezmethod/ref_go/internal/domain"
)

// CachedURLShortenerService wraps the base URL shortener service with caching
type CachedURLShortenerService struct {
	base   *URLShortenerService
	cache  cache.CacheInterface
	logger *zap.Logger
}

// NewCachedURLShortenerService creates a new cached URL shortener service
func NewCachedURLShortenerService(base *URLShortenerService, cache cache.CacheInterface, logger *zap.Logger) *CachedURLShortenerService {
	return &CachedURLShortenerService{
		base:   base,
		cache:  cache,
		logger: logger,
	}
}

// CreateShortLink creates a new short link (delegated to base service, updates cache)
func (s *CachedURLShortenerService) CreateShortLink(ctx context.Context, req *domain.CreateShortLinkRequest) (*domain.ShortLink, error) {
	// Create link using the base service
	link, err := s.base.CreateShortLink(ctx, req)
	if err != nil {
		return nil, err
	}

	// Add link to cache
	s.cache.Set(link.Code, link, 0)

	return link, nil
}

// GetShortLink gets a short link by ID (with caching)
func (s *CachedURLShortenerService) GetShortLink(ctx context.Context, id string) (*domain.ShortLink, error) {
	// Try to get link from cache by ID
	if cachedLink, found := s.cache.Get("id:" + id); found {
		s.logger.Debug("Cache hit for link ID", zap.String("id", id))
		return cachedLink.(*domain.ShortLink), nil
	}

	// Get link from the base service
	link, err := s.base.GetShortLink(ctx, id)
	if err != nil {
		return nil, err
	}

	// Add link to cache
	s.cache.Set("id:"+id, link, 0)
	s.cache.Set(link.Code, link, 0)

	return link, nil
}

// GetShortLinkByCode gets a short link by code (with caching)
func (s *CachedURLShortenerService) GetShortLinkByCode(ctx context.Context, code string) (*domain.ShortLink, error) {
	// Try to get link from cache by code
	if cachedLink, found := s.cache.Get(code); found {
		s.logger.Debug("Cache hit for link code", zap.String("code", code))
		return cachedLink.(*domain.ShortLink), nil
	}

	// Get link from the base service
	link, err := s.base.GetShortLinkByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Add link to cache
	s.cache.Set(code, link, 0)
	s.cache.Set("id:"+link.ID, link, 0)

	return link, nil
}

// UpdateShortLink updates a short link (invalidates cache)
func (s *CachedURLShortenerService) UpdateShortLink(ctx context.Context, id string, req *domain.UpdateShortLinkRequest) (*domain.ShortLink, error) {
	// Get the current link to know what to invalidate
	oldLink, err := s.base.GetShortLink(ctx, id)
	if err == nil {
		// Invalidate the old code in the cache
		s.cache.Delete(oldLink.Code)
	}

	// Update link using the base service
	link, err := s.base.UpdateShortLink(ctx, id, req)
	if err != nil {
		return nil, err
	}

	// Invalidate cache entries
	s.cache.Delete("id:" + id)

	// Add updated link to cache
	s.cache.Set("id:"+id, link, 0)
	s.cache.Set(link.Code, link, 0)

	return link, nil
}

// DeleteShortLink deletes a short link (invalidates cache)
func (s *CachedURLShortenerService) DeleteShortLink(ctx context.Context, id string) error {
	// Get the current link to know what to invalidate
	oldLink, err := s.base.GetShortLink(ctx, id)
	if err == nil {
		// Invalidate the old code in the cache
		s.cache.Delete(oldLink.Code)
	}

	// Delete link using the base service
	err = s.base.DeleteShortLink(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate cache entry
	s.cache.Delete("id:" + id)

	return nil
}

// ListShortLinks lists short links (not cached)
func (s *CachedURLShortenerService) ListShortLinks(ctx context.Context, page, pageSize int) ([]*domain.ShortLink, int, error) {
	// List links using the base service (not cached due to pagination)
	return s.base.ListShortLinks(ctx, page, pageSize)
}

// RecordClick records a click on a short link
func (s *CachedURLShortenerService) RecordClick(ctx context.Context, shortLinkID string, referrer, userAgent, ipAddress string) error {
	// Record click using the base service
	return s.base.RecordClick(ctx, shortLinkID, referrer, userAgent, ipAddress)
}

// GetLinkStats gets statistics for a short link
func (s *CachedURLShortenerService) GetLinkStats(ctx context.Context, shortLinkID string) (*domain.LinkStats, error) {
	// Get stats using the base service (not cached as they change frequently)
	return s.base.GetLinkStats(ctx, shortLinkID)
}

// GetCacheStats gets statistics about the cache
func (s *CachedURLShortenerService) GetCacheStats() cache.Stats {
	return s.cache.GetStats()
}
