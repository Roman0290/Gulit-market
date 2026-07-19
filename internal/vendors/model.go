package vendors

import "time"

type Status string

const (
	StatusPending   Status = "pending"
	StatusApproved  Status = "approved"
	StatusRejected  Status = "rejected"
	StatusSuspended Status = "suspended"
)

type Vendor struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	ShopName    string    `json:"shop_name"`
	Description string    `json:"description,omitempty"`
	Location    string    `json:"location,omitempty"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
