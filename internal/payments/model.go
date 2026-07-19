package payments

import "time"

type Status string

const (
	StatusPending   Status = "pending"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusRefunded  Status = "refunded"
)

type Payment struct {
	ID          string    `json:"id"`
	OrderID     string    `json:"order_id"`
	Provider    string    `json:"provider"`
	ProviderRef string    `json:"provider_ref"`
	Amount      float64   `json:"amount"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
