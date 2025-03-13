package repository

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/menezmethod/ref_go/internal/common"
	"github.com/menezmethod/ref_go/internal/domain"
)

// Scanner is an interface for the Scan method that both sql.Row and sql.Rows implement
type Scanner interface {
	Scan(dest ...interface{}) error
}

// DB is an interface for database operations
type DB interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) Scanner
	Begin() (*sql.Tx, error)
	Prepare(query string) (*sql.Stmt, error)
	Ping() error
	Close() error
	SetMaxOpenConns(n int)
	SetMaxIdleConns(n int)
	SetConnMaxLifetime(d interface{})
}

// PostgresLinkRepository implements LinkRepository for PostgreSQL
type PostgresLinkRepository struct {
	db common.DB
}

// NewPostgresLinkRepository creates a new PostgresLinkRepository
func NewPostgresLinkRepository(db common.DB) *PostgresLinkRepository {
	return &PostgresLinkRepository{
		db: db,
	}
}

// Create creates a new link in the database
func (r *PostgresLinkRepository) Create(link *domain.Link) error {
	// Set created and updated time
	now := time.Now()
	link.CreatedAt = now
	link.UpdatedAt = now

	// Execute SQL
	_, err := r.db.Exec(
		"INSERT INTO links (id, user_id, original_url, short_url, visits, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		link.ID, link.UserID, link.OriginalURL, link.ShortURL, link.Visits, link.CreatedAt, link.UpdatedAt,
	)

	// Check for postgres-specific errors
	if err, ok := err.(*pq.Error); ok {
		if err.Code == "23505" { // Unique violation
			return domain.ErrConflict
		}
	}

	return err
}

// GetByID gets a link by ID
func (r *PostgresLinkRepository) GetByID(id string) (*domain.Link, error) {
	link := &domain.Link{}
	err := r.db.QueryRow(
		"SELECT id, user_id, original_url, short_url, visits, created_at, updated_at FROM links WHERE id = $1",
		id,
	).Scan(
		&link.ID, &link.UserID, &link.OriginalURL, &link.ShortURL, &link.Visits, &link.CreatedAt, &link.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}

	return link, err
}

// GetByShortURL gets a link by short URL
func (r *PostgresLinkRepository) GetByShortURL(shortURL string) (*domain.Link, error) {
	link := &domain.Link{}
	err := r.db.QueryRow(
		"SELECT id, user_id, original_url, short_url, visits, created_at, updated_at FROM links WHERE short_url = $1",
		shortURL,
	).Scan(
		&link.ID, &link.UserID, &link.OriginalURL, &link.ShortURL, &link.Visits, &link.CreatedAt, &link.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}

	return link, err
}

// Update updates a link
func (r *PostgresLinkRepository) Update(link *domain.Link) error {
	// Update updated time
	link.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		"UPDATE links SET original_url = $1, short_url = $2, visits = $3, updated_at = $4 WHERE id = $5",
		link.OriginalURL, link.ShortURL, link.Visits, link.UpdatedAt, link.ID,
	)

	return err
}

// Delete deletes a link
func (r *PostgresLinkRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM links WHERE id = $1", id)
	return err
}

// List lists links for a user with pagination
func (r *PostgresLinkRepository) List(userID string, limit, offset int) ([]*domain.Link, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, original_url, short_url, visits, created_at, updated_at FROM links WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := []*domain.Link{}

	for rows.Next() {
		link := &domain.Link{}
		if err := rows.Scan(&link.ID, &link.UserID, &link.OriginalURL, &link.ShortURL, &link.Visits, &link.CreatedAt, &link.UpdatedAt); err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

// Count counts the number of links for a user
func (r *PostgresLinkRepository) Count(userID string) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM links WHERE user_id = $1", userID).Scan(&count)
	return count, err
}

// IncrementVisits increments the visits count for a link
func (r *PostgresLinkRepository) IncrementVisits(id string) error {
	_, err := r.db.Exec("UPDATE links SET visits = visits + 1 WHERE id = $1", id)
	return err
}

// CreateClick creates a new click record
func (r *PostgresLinkRepository) CreateClick(click *domain.Click) error {
	click.CreatedAt = time.Now()
	_, err := r.db.Exec(
		"INSERT INTO clicks (id, link_id, user_agent, referer, ip_address, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		click.ID, click.LinkID, click.UserAgent, click.Referer, click.IPAddress, click.CreatedAt,
	)
	return err
}

// GetClicks gets clicks for a link with pagination
func (r *PostgresLinkRepository) GetClicks(linkID string, limit, offset int) ([]*domain.Click, error) {
	rows, err := r.db.Query(
		"SELECT id, link_id, user_agent, referer, ip_address, created_at FROM clicks WHERE link_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		linkID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clicks := []*domain.Click{}

	for rows.Next() {
		click := &domain.Click{}
		if err := rows.Scan(&click.ID, &click.LinkID, &click.UserAgent, &click.Referer, &click.IPAddress, &click.CreatedAt); err != nil {
			return nil, err
		}
		clicks = append(clicks, click)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clicks, nil
}

// CountClicks counts the number of clicks for a link
func (r *PostgresLinkRepository) CountClicks(linkID string) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM clicks WHERE link_id = $1", linkID).Scan(&count)
	return count, err
}
