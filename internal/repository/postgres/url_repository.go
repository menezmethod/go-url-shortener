package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/menezmethod/ref_go/internal/db"
	"github.com/menezmethod/ref_go/internal/domain"
)

// URLRepository implements the repository.URLRepository interface
type URLRepository struct {
	db *db.DB
}

// NewURLRepository creates a new URL repository
func NewURLRepository(db *db.DB) *URLRepository {
	return &URLRepository{
		db: db,
	}
}

// Create stores a new URL
func (r *URLRepository) Create(ctx context.Context, url *domain.URL) error {
	query := `
		INSERT INTO urls (id, original_url, hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		url.ID,
		url.OriginalURL,
		url.Hash,
		url.CreatedAt,
		url.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("creating url: %w", err)
	}

	return nil
}

// GetByID retrieves a URL by ID
func (r *URLRepository) GetByID(ctx context.Context, id string) (*domain.URL, error) {
	query := `
		SELECT id, original_url, hash, created_at, updated_at
		FROM urls
		WHERE id = $1
	`

	var url domain.URL
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.Hash,
		&url.CreatedAt,
		&url.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("url not found: %w", err)
		}
		return nil, fmt.Errorf("getting url by id: %w", err)
	}

	return &url, nil
}

// GetByHash retrieves a URL by hash
func (r *URLRepository) GetByHash(ctx context.Context, hash string) (*domain.URL, error) {
	query := `
		SELECT id, original_url, hash, created_at, updated_at
		FROM urls
		WHERE hash = $1
	`

	var url domain.URL
	err := r.db.QueryRowContext(ctx, query, hash).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.Hash,
		&url.CreatedAt,
		&url.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("url not found: %w", err)
		}
		return nil, fmt.Errorf("getting url by hash: %w", err)
	}

	return &url, nil
}
