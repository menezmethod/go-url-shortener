package repository

import (
	"context"

	"github.com/menezmethod/ref_go/internal/domain"
)

// URLRepository defines operations for storing and retrieving URLs
type URLRepository interface {
	// Create stores a new URL
	Create(ctx context.Context, url *domain.URL) error

	// GetByID retrieves a URL by ID
	GetByID(ctx context.Context, id string) (*domain.URL, error)

	// GetByHash retrieves a URL by hash
	GetByHash(ctx context.Context, hash string) (*domain.URL, error)
}

// ShortLinkRepository defines operations for short links
type ShortLinkRepository interface {
	// Create stores a new short link
	Create(ctx context.Context, link *domain.ShortLink) error

	// GetByID retrieves a short link by ID
	GetByID(ctx context.Context, id string) (*domain.ShortLink, error)

	// GetByCode retrieves a short link by code
	GetByCode(ctx context.Context, code string) (*domain.ShortLink, error)

	// GetByCustomAlias retrieves a short link by custom alias
	GetByCustomAlias(ctx context.Context, alias string) (*domain.ShortLink, error)

	// GetAllByURLID retrieves all short links for a URL
	GetAllByURLID(ctx context.Context, urlID string) ([]*domain.ShortLink, error)

	// Update updates a short link
	Update(ctx context.Context, link *domain.ShortLink) error

	// Delete deletes a short link
	Delete(ctx context.Context, id string) error

	// List returns a paginated list of short links
	List(ctx context.Context, offset, limit int) ([]*domain.ShortLink, error)

	// Count returns the total number of short links
	Count(ctx context.Context) (int, error)
}

// LinkClickRepository defines operations for link click analytics
type LinkClickRepository interface {
	// Create records a new link click
	Create(ctx context.Context, click *domain.LinkClick) error

	// GetByShortLinkID retrieves all clicks for a short link
	GetByShortLinkID(ctx context.Context, shortLinkID string, offset, limit int) ([]*domain.LinkClick, error)

	// GetStatsByShortLinkID retrieves statistics for a short link
	GetStatsByShortLinkID(ctx context.Context, shortLinkID string) (*domain.LinkStats, error)
}
