package addresses

type Address struct {
	ID        string   `json:"id"`
	UserID    string   `json:"user_id"`
	Label     string   `json:"label,omitempty"`
	Line1     string   `json:"line1"`
	City      string   `json:"city"`
	Lat       *float64 `json:"lat,omitempty"`
	Lng       *float64 `json:"lng,omitempty"`
	IsDefault bool     `json:"is_default"`
}
