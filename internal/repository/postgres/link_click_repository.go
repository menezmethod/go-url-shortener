package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/menezmethod/ref_go/internal/db"
	"github.com/menezmethod/ref_go/internal/domain"
)

// LinkClickRepository implements the repository.LinkClickRepository interface
type LinkClickRepository struct {
	db *db.DB
}

// NewLinkClickRepository creates a new link click repository
func NewLinkClickRepository(db *db.DB) *LinkClickRepository {
	return &LinkClickRepository{
		db: db,
	}
}

// Create records a new link click
func (r *LinkClickRepository) Create(ctx context.Context, click *domain.LinkClick) error {
	query := `
		INSERT INTO link_clicks (
			id, short_link_id, referrer, user_agent, ip_address, 
			country, city, device, browser, os, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		click.ID,
		click.ShortLinkID,
		click.Referrer,
		click.UserAgent,
		click.IPAddress,
		click.Country,
		click.City,
		click.Device,
		click.Browser,
		click.OS,
		click.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("creating link click: %w", err)
	}

	return nil
}

// GetByShortLinkID retrieves all clicks for a short link
func (r *LinkClickRepository) GetByShortLinkID(
	ctx context.Context,
	shortLinkID string,
	offset,
	limit int,
) ([]*domain.LinkClick, error) {
	query := `
		SELECT id, short_link_id, referrer, user_agent, ip_address, 
               country, city, device, browser, os, created_at
		FROM link_clicks
		WHERE short_link_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, shortLinkID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("getting link clicks by short link id: %w", err)
	}
	defer rows.Close()

	var clicks []*domain.LinkClick

	for rows.Next() {
		var click domain.LinkClick

		err := rows.Scan(
			&click.ID,
			&click.ShortLinkID,
			&click.Referrer,
			&click.UserAgent,
			&click.IPAddress,
			&click.Country,
			&click.City,
			&click.Device,
			&click.Browser,
			&click.OS,
			&click.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scanning link click row: %w", err)
		}

		clicks = append(clicks, &click)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating link click rows: %w", err)
	}

	return clicks, nil
}

// GetStatsByShortLinkID retrieves statistics for a short link
func (r *LinkClickRepository) GetStatsByShortLinkID(ctx context.Context, shortLinkID string) (*domain.LinkStats, error) {
	// Get total clicks
	countQuery := `
		SELECT COUNT(*)
		FROM link_clicks
		WHERE short_link_id = $1
	`

	var totalClicks int
	err := r.db.QueryRowContext(ctx, countQuery, shortLinkID).Scan(&totalClicks)
	if err != nil {
		return nil, fmt.Errorf("counting link clicks: %w", err)
	}

	// If no clicks, return empty stats
	if totalClicks == 0 {
		return &domain.LinkStats{
			TotalClicks:  0,
			TopReferrers: make(map[string]int),
			TopBrowsers:  make(map[string]int),
			TopOS:        make(map[string]int),
			TopDevices:   make(map[string]int),
			ClicksByDay:  make(map[string]int),
		}, nil
	}

	// Get last clicked time
	lastClickedQuery := `
		SELECT created_at
		FROM link_clicks
		WHERE short_link_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var lastClicked time.Time
	err = r.db.QueryRowContext(ctx, lastClickedQuery, shortLinkID).Scan(&lastClicked)
	if err != nil {
		return nil, fmt.Errorf("getting last clicked time: %w", err)
	}

	// Get top referrers
	topReferrersQuery := `
		SELECT referrer, COUNT(*) as count
		FROM link_clicks
		WHERE short_link_id = $1 AND referrer IS NOT NULL
		GROUP BY referrer
		ORDER BY count DESC
		LIMIT 5
	`

	referrerRows, err := r.db.QueryContext(ctx, topReferrersQuery, shortLinkID)
	if err != nil {
		return nil, fmt.Errorf("getting top referrers: %w", err)
	}
	defer referrerRows.Close()

	topReferrers := make(map[string]int)
	for referrerRows.Next() {
		var referrer string
		var count int
		if err := referrerRows.Scan(&referrer, &count); err != nil {
			return nil, fmt.Errorf("scanning referrer row: %w", err)
		}
		topReferrers[referrer] = count
	}

	// Get top browsers
	topBrowsersQuery := `
		SELECT browser, COUNT(*) as count
		FROM link_clicks
		WHERE short_link_id = $1 AND browser IS NOT NULL
		GROUP BY browser
		ORDER BY count DESC
		LIMIT 5
	`

	browserRows, err := r.db.QueryContext(ctx, topBrowsersQuery, shortLinkID)
	if err != nil {
		return nil, fmt.Errorf("getting top browsers: %w", err)
	}
	defer browserRows.Close()

	topBrowsers := make(map[string]int)
	for browserRows.Next() {
		var browser string
		var count int
		if err := browserRows.Scan(&browser, &count); err != nil {
			return nil, fmt.Errorf("scanning browser row: %w", err)
		}
		topBrowsers[browser] = count
	}

	// Get top operating systems
	topOSQuery := `
		SELECT os, COUNT(*) as count
		FROM link_clicks
		WHERE short_link_id = $1 AND os IS NOT NULL
		GROUP BY os
		ORDER BY count DESC
		LIMIT 5
	`

	osRows, err := r.db.QueryContext(ctx, topOSQuery, shortLinkID)
	if err != nil {
		return nil, fmt.Errorf("getting top operating systems: %w", err)
	}
	defer osRows.Close()

	topOS := make(map[string]int)
	for osRows.Next() {
		var os string
		var count int
		if err := osRows.Scan(&os, &count); err != nil {
			return nil, fmt.Errorf("scanning os row: %w", err)
		}
		topOS[os] = count
	}

	// Get top devices
	topDevicesQuery := `
		SELECT device, COUNT(*) as count
		FROM link_clicks
		WHERE short_link_id = $1 AND device IS NOT NULL
		GROUP BY device
		ORDER BY count DESC
		LIMIT 5
	`

	deviceRows, err := r.db.QueryContext(ctx, topDevicesQuery, shortLinkID)
	if err != nil {
		return nil, fmt.Errorf("getting top devices: %w", err)
	}
	defer deviceRows.Close()

	topDevices := make(map[string]int)
	for deviceRows.Next() {
		var device string
		var count int
		if err := deviceRows.Scan(&device, &count); err != nil {
			return nil, fmt.Errorf("scanning device row: %w", err)
		}
		topDevices[device] = count
	}

	// Get clicks by day for the last 30 days
	clicksByDayQuery := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM link_clicks
		WHERE short_link_id = $1 AND created_at >= NOW() - INTERVAL '30 days'
		GROUP BY date
		ORDER BY date
	`

	dayRows, err := r.db.QueryContext(ctx, clicksByDayQuery, shortLinkID)
	if err != nil {
		return nil, fmt.Errorf("getting clicks by day: %w", err)
	}
	defer dayRows.Close()

	clicksByDay := make(map[string]int)
	for dayRows.Next() {
		var date time.Time
		var count int
		if err := dayRows.Scan(&date, &count); err != nil {
			return nil, fmt.Errorf("scanning day row: %w", err)
		}
		clicksByDay[date.Format("2006-01-02")] = count
	}

	// Get recent clicks
	recentClicksQuery := `
		SELECT id, short_link_id, referrer, user_agent, ip_address, 
               country, city, device, browser, os, created_at
		FROM link_clicks
		WHERE short_link_id = $1
		ORDER BY created_at DESC
		LIMIT 10
	`

	recentRows, err := r.db.QueryContext(ctx, recentClicksQuery, shortLinkID)
	if err != nil {
		return nil, fmt.Errorf("getting recent clicks: %w", err)
	}
	defer recentRows.Close()

	var recentClicks []domain.LinkClick
	for recentRows.Next() {
		var click domain.LinkClick
		if err := recentRows.Scan(
			&click.ID,
			&click.ShortLinkID,
			&click.Referrer,
			&click.UserAgent,
			&click.IPAddress,
			&click.Country,
			&click.City,
			&click.Device,
			&click.Browser,
			&click.OS,
			&click.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning recent click row: %w", err)
		}
		recentClicks = append(recentClicks, click)
	}

	return &domain.LinkStats{
		TotalClicks:  totalClicks,
		LastClicked:  &lastClicked,
		TopReferrers: topReferrers,
		TopBrowsers:  topBrowsers,
		TopOS:        topOS,
		TopDevices:   topDevices,
		ClicksByDay:  clicksByDay,
		RecentClicks: recentClicks,
	}, nil
}
