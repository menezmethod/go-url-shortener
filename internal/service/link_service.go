package service

import (
	"crypto/rand"
	"encoding/base64"
	"net/url"

	"github.com/menezmethod/ref_go/internal/domain"
)

// CreateLink creates a new short link
func (s *LinkService) CreateLink(req CreateLinkRequest) (*domain.Link, error) {
	// Validate URL
	if _, err := url.ParseRequestURI(req.OriginalURL); err != nil {
		return nil, domain.ErrValidation
	}

	// Generate a short URL if not provided
	shortURL := req.CustomAlias
	if shortURL == "" {
		// Generate a random short code
		shortURL = generateShortCode(6)
	} else {
		// If custom alias is provided, check if it's available
		existing, err := s.linkRepo.GetByShortURL(shortURL)
		if err == nil && existing != nil {
			return nil, domain.ErrConflict
		}
	}

	// Create link object
	link := &domain.Link{
		UserID:      req.UserID,
		OriginalURL: req.OriginalURL,
		ShortURL:    shortURL,
	}

	// Save to repository
	err := s.linkRepo.Create(link)
	if err != nil {
		return nil, err
	}

	return link, nil
}

// Generate a random short code of specified length
func generateShortCode(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}

// GetLink retrieves a link by ID
func (s *LinkService) GetLink(id string) (*domain.Link, error) {
	return s.linkRepo.GetByID(id)
}

// GetLinkByShortURL retrieves a link by short URL
func (s *LinkService) GetLinkByShortURL(shortURL string) (*domain.Link, error) {
	return s.linkRepo.GetByShortURL(shortURL)
}

// UpdateLink updates an existing link
func (s *LinkService) UpdateLink(id string, req UpdateLinkRequest) (*domain.Link, error) {
	link, err := s.linkRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.OriginalURL != "" {
		link.OriginalURL = req.OriginalURL
	}

	if req.CustomAlias != "" {
		link.ShortURL = req.CustomAlias
	}

	err = s.linkRepo.Update(link)
	if err != nil {
		return nil, err
	}

	return link, nil
}

// DeleteLink deletes a link
func (s *LinkService) DeleteLink(id string) error {
	return s.linkRepo.Delete(id)
}

// ListLinks lists links for a user
func (s *LinkService) ListLinks(userID string, page, perPage int) ([]*domain.Link, int, error) {
	// Calculate offset
	offset := (page - 1) * perPage

	// Get links
	links, err := s.linkRepo.List(userID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	count, err := s.linkRepo.Count(userID)
	if err != nil {
		return nil, 0, err
	}

	return links, count, nil
}

// RecordClick records a click on a link
func (s *LinkService) RecordClick(linkID, userAgent, referer, ipAddress string) error {
	// Increment visits count
	err := s.linkRepo.IncrementVisits(linkID)
	if err != nil {
		return err
	}

	// Create click record
	click := &domain.Click{
		LinkID:    linkID,
		UserAgent: userAgent,
		Referer:   referer,
		IPAddress: ipAddress,
	}

	return s.linkRepo.CreateClick(click)
}

// GetClicks gets click data for a link
func (s *LinkService) GetClicks(linkID string, page, perPage int) ([]*domain.Click, int, error) {
	// Calculate offset
	offset := (page - 1) * perPage

	// Get clicks
	clicks, err := s.linkRepo.GetClicks(linkID, perPage, offset)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	count, err := s.linkRepo.CountClicks(linkID)
	if err != nil {
		return nil, 0, err
	}

	return clicks, count, nil
}
