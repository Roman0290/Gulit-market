package banners

import "time"

type Banner struct {
	ID        string    `json:"id"`
	ImageURL  string    `json:"image_url"`
	LinkURL   string    `json:"link_url,omitempty"`
	Title     string    `json:"title,omitempty"`
	IsActive  bool      `json:"is_active"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}
