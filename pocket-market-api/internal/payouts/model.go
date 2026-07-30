package payouts

import "time"

type Payout struct {
	ID        string    `json:"id"`
	VendorID  string    `json:"vendor_id"`
	Amount    float64   `json:"amount"`
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Balance is what a vendor has earned from paid orders (subtotal minus
// platform commission) minus what's already been recorded as paid out.
// This is a bookkeeping ledger only - creating a Payout does not move real
// money, since that would require Stripe Connect (separate vendor bank
// account onboarding), which is out of scope here.
type Balance struct {
	VendorID     string  `json:"vendor_id"`
	TotalEarned  float64 `json:"total_earned"`
	TotalPaidOut float64 `json:"total_paid_out"`
	BalanceDue   float64 `json:"balance_due"`
}
