package service

import (
	"github.com/menezmethod/ref_go/internal/domain"
)

// LinkService handles business logic related to links
type LinkService struct {
	linkRepo LinkRepository
}

// LinkRepository is an interface for link data access
type LinkRepository interface {
	Create(link *domain.Link) error
	GetByID(id string) (*domain.Link, error)
	GetByShortURL(shortURL string) (*domain.Link, error)
	Update(link *domain.Link) error
	Delete(id string) error
	List(userID string, limit, offset int) ([]*domain.Link, error)
	Count(userID string) (int, error)
	IncrementVisits(id string) error
	CreateClick(click *domain.Click) error
	GetClicks(linkID string, limit, offset int) ([]*domain.Click, error)
	CountClicks(linkID string) (int, error)
}

// CreateLinkRequest represents request data for creating a link
type CreateLinkRequest struct {
	UserID      string `json:"user_id"`
	OriginalURL string `json:"original_url" binding:"required"`
	CustomAlias string `json:"custom_alias,omitempty"`
}

// UpdateLinkRequest represents request data for updating a link
type UpdateLinkRequest struct {
	OriginalURL string `json:"original_url,omitempty"`
	CustomAlias string `json:"custom_alias,omitempty"`
}

// NewLinkService creates a new LinkService
func NewLinkService(linkRepo LinkRepository) *LinkService {
	return &LinkService{
		linkRepo: linkRepo,
	}
}
