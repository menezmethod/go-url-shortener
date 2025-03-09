package domain

import (
	"time"
)

// URL represents a stored URL in the system
type URL struct {
	ID          string    `json:"id"`
	OriginalURL string    `json:"original_url"`
	Hash        string    `json:"hash"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ShortLink represents a shortened URL
type ShortLink struct {
	ID             string     `json:"id"`
	Code           string     `json:"code"`
	CustomAlias    *string    `json:"custom_alias,omitempty"`
	URLID          string     `json:"url_id"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Embedded URL information when fetching a short link
	URL *URL `json:"url,omitempty"`
}

// LinkClick represents a click on a shortened URL
type LinkClick struct {
	ID          string    `json:"id"`
	ShortLinkID string    `json:"short_link_id"`
	Referrer    *string   `json:"referrer,omitempty"`
	UserAgent   *string   `json:"user_agent,omitempty"`
	IPAddress   *string   `json:"ip_address,omitempty"`
	Country     *string   `json:"country,omitempty"`
	City        *string   `json:"city,omitempty"`
	Device      *string   `json:"device,omitempty"`
	Browser     *string   `json:"browser,omitempty"`
	OS          *string   `json:"os,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateShortLinkRequest represents the request to create a short link
type CreateShortLinkRequest struct {
	URL            string     `json:"url"`
	CustomAlias    *string    `json:"custom_alias,omitempty"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
}

// LinkStats represents the stats for a short link
type LinkStats struct {
	TotalClicks  int            `json:"total_clicks"`
	LastClicked  *time.Time     `json:"last_clicked,omitempty"`
	TopReferrers map[string]int `json:"top_referrers,omitempty"`
	TopBrowsers  map[string]int `json:"top_browsers,omitempty"`
	TopOS        map[string]int `json:"top_os,omitempty"`
	TopDevices   map[string]int `json:"top_devices,omitempty"`
	ClicksByDay  map[string]int `json:"clicks_by_day,omitempty"`
	RecentClicks []LinkClick    `json:"recent_clicks,omitempty"`
}

// UpdateShortLinkRequest represents the request to update a short link
type UpdateShortLinkRequest struct {
	CustomAlias    *string    `json:"custom_alias,omitempty"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
	IsActive       *bool      `json:"is_active,omitempty"`
}
