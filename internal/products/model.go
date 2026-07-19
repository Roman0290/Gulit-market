package products

import "time"

type Product struct {
	ID            string    `json:"id"`
	VendorID      string    `json:"vendor_id"`
	CategoryID    string    `json:"category_id,omitempty"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	Price         float64   `json:"price"`
	Unit          string    `json:"unit"`
	StockQuantity int       `json:"stock_quantity"`
	ImageURL      string    `json:"image_url,omitempty"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
