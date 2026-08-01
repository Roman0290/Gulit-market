package disputes

import "time"

type Status string

const (
	StatusOpen     Status = "open"
	StatusResolved Status = "resolved"
)

type Dispute struct {
	ID             string     `json:"id"`
	OrderID        string     `json:"order_id"`
	RaisedBy       string     `json:"raised_by"`
	Reason         string     `json:"reason"`
	Status         Status     `json:"status"`
	ResolutionNote string     `json:"resolution_note,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
}
