package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/menezmethod/ref_go/internal/db"
	"github.com/menezmethod/ref_go/internal/domain"
)

// ShortLinkRepository implements the repository.ShortLinkRepository interface
type ShortLinkRepository struct {
	db *db.DB
}

// NewShortLinkRepository creates a new short link repository
func NewShortLinkRepository(db *db.DB) *ShortLinkRepository {
	return &ShortLinkRepository{
		db: db,
	}
}

// Create stores a new short link
func (r *ShortLinkRepository) Create(ctx context.Context, link *domain.ShortLink) error {
	query := `
		INSERT INTO short_links (id, code, custom_alias, url_id, expiration_date, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		link.ID,
		link.Code,
		link.CustomAlias,
		link.URLID,
		link.ExpirationDate,
		link.IsActive,
		link.CreatedAt,
		link.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("creating short link: %w", err)
	}

	return nil
}

// GetByID retrieves a short link by ID
func (r *ShortLinkRepository) GetByID(ctx context.Context, id string) (*domain.ShortLink, error) {
	query := `
		SELECT s.id, s.code, s.custom_alias, s.url_id, s.expiration_date, s.is_active, s.created_at, s.updated_at,
               u.id, u.original_url, u.hash, u.created_at, u.updated_at
		FROM short_links s
		JOIN urls u ON s.url_id = u.id
		WHERE s.id = $1
	`

	var link domain.ShortLink
	var url domain.URL

	// Nullable fields
	var customAlias sql.NullString
	var expirationDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&link.ID,
		&link.Code,
		&customAlias,
		&link.URLID,
		&expirationDate,
		&link.IsActive,
		&link.CreatedAt,
		&link.UpdatedAt,
		&url.ID,
		&url.OriginalURL,
		&url.Hash,
		&url.CreatedAt,
		&url.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("short link not found: %w", err)
		}
		return nil, fmt.Errorf("getting short link by id: %w", err)
	}

	// Handle nullable fields
	if customAlias.Valid {
		link.CustomAlias = &customAlias.String
	}

	if expirationDate.Valid {
		link.ExpirationDate = &expirationDate.Time
	}

	// Set the URL object
	link.URL = &url

	return &link, nil
}

// GetByCode retrieves a short link by code
func (r *ShortLinkRepository) GetByCode(ctx context.Context, code string) (*domain.ShortLink, error) {
	query := `
		SELECT s.id, s.code, s.custom_alias, s.url_id, s.expiration_date, s.is_active, s.created_at, s.updated_at,
               u.id, u.original_url, u.hash, u.created_at, u.updated_at
		FROM short_links s
		JOIN urls u ON s.url_id = u.id
		WHERE s.code = $1
	`

	var link domain.ShortLink
	var url domain.URL

	// Nullable fields
	var customAlias sql.NullString
	var expirationDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&link.ID,
		&link.Code,
		&customAlias,
		&link.URLID,
		&expirationDate,
		&link.IsActive,
		&link.CreatedAt,
		&link.UpdatedAt,
		&url.ID,
		&url.OriginalURL,
		&url.Hash,
		&url.CreatedAt,
		&url.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("short link not found: %w", err)
		}
		return nil, fmt.Errorf("getting short link by code: %w", err)
	}

	// Handle nullable fields
	if customAlias.Valid {
		link.CustomAlias = &customAlias.String
	}

	if expirationDate.Valid {
		link.ExpirationDate = &expirationDate.Time
	}

	// Set the URL object
	link.URL = &url

	return &link, nil
}

// GetByCustomAlias retrieves a short link by custom alias
func (r *ShortLinkRepository) GetByCustomAlias(ctx context.Context, alias string) (*domain.ShortLink, error) {
	query := `
		SELECT s.id, s.code, s.custom_alias, s.url_id, s.expiration_date, s.is_active, s.created_at, s.updated_at,
               u.id, u.original_url, u.hash, u.created_at, u.updated_at
		FROM short_links s
		JOIN urls u ON s.url_id = u.id
		WHERE s.custom_alias = $1
	`

	var link domain.ShortLink
	var url domain.URL

	// Nullable fields
	var customAlias sql.NullString
	var expirationDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, alias).Scan(
		&link.ID,
		&link.Code,
		&customAlias,
		&link.URLID,
		&expirationDate,
		&link.IsActive,
		&link.CreatedAt,
		&link.UpdatedAt,
		&url.ID,
		&url.OriginalURL,
		&url.Hash,
		&url.CreatedAt,
		&url.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("short link not found: %w", err)
		}
		return nil, fmt.Errorf("getting short link by custom alias: %w", err)
	}

	// Handle nullable fields
	if customAlias.Valid {
		link.CustomAlias = &customAlias.String
	}

	if expirationDate.Valid {
		link.ExpirationDate = &expirationDate.Time
	}

	// Set the URL object
	link.URL = &url

	return &link, nil
}

// GetAllByURLID retrieves all short links for a URL
func (r *ShortLinkRepository) GetAllByURLID(ctx context.Context, urlID string) ([]*domain.ShortLink, error) {
	query := `
		SELECT id, code, custom_alias, url_id, expiration_date, is_active, created_at, updated_at
		FROM short_links
		WHERE url_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, urlID)
	if err != nil {
		return nil, fmt.Errorf("getting short links by url id: %w", err)
	}
	defer rows.Close()

	var links []*domain.ShortLink

	for rows.Next() {
		var link domain.ShortLink
		var customAlias sql.NullString
		var expirationDate sql.NullTime

		err := rows.Scan(
			&link.ID,
			&link.Code,
			&customAlias,
			&link.URLID,
			&expirationDate,
			&link.IsActive,
			&link.CreatedAt,
			&link.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scanning short link row: %w", err)
		}

		// Handle nullable fields
		if customAlias.Valid {
			link.CustomAlias = &customAlias.String
		}

		if expirationDate.Valid {
			link.ExpirationDate = &expirationDate.Time
		}

		links = append(links, &link)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating short link rows: %w", err)
	}

	return links, nil
}

// Update updates a short link
func (r *ShortLinkRepository) Update(ctx context.Context, link *domain.ShortLink) error {
	query := `
		UPDATE short_links
		SET custom_alias = $1, expiration_date = $2, is_active = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		link.CustomAlias,
		link.ExpirationDate,
		link.IsActive,
		time.Now().UTC(),
		link.ID,
	)

	if err != nil {
		return fmt.Errorf("updating short link: %w", err)
	}

	return nil
}

// Delete deletes a short link
func (r *ShortLinkRepository) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM short_links
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting short link: %w", err)
	}

	// Check if any rows were affected
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking affected rows: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("short link not found")
	}

	return nil
}

// List returns a paginated list of short links
func (r *ShortLinkRepository) List(ctx context.Context, offset, limit int) ([]*domain.ShortLink, error) {
	query := `
		SELECT s.id, s.code, s.custom_alias, s.url_id, s.expiration_date, s.is_active, s.created_at, s.updated_at,
               u.id, u.original_url, u.hash, u.created_at, u.updated_at
		FROM short_links s
		JOIN urls u ON s.url_id = u.id
		ORDER BY s.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("listing short links: %w", err)
	}
	defer rows.Close()

	var links []*domain.ShortLink

	for rows.Next() {
		var link domain.ShortLink
		var url domain.URL
		var customAlias sql.NullString
		var expirationDate sql.NullTime

		err := rows.Scan(
			&link.ID,
			&link.Code,
			&customAlias,
			&link.URLID,
			&expirationDate,
			&link.IsActive,
			&link.CreatedAt,
			&link.UpdatedAt,
			&url.ID,
			&url.OriginalURL,
			&url.Hash,
			&url.CreatedAt,
			&url.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scanning short link row: %w", err)
		}

		// Handle nullable fields
		if customAlias.Valid {
			link.CustomAlias = &customAlias.String
		}

		if expirationDate.Valid {
			link.ExpirationDate = &expirationDate.Time
		}

		// Set the URL object
		link.URL = &url

		links = append(links, &link)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating short link rows: %w", err)
	}

	return links, nil
}

// Count returns the total number of short links
func (r *ShortLinkRepository) Count(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM short_links
	`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting short links: %w", err)
	}

	return count, nil
}
