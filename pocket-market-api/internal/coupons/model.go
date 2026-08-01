package coupons

import "time"

type DiscountType string

const (
	DiscountPercent DiscountType = "percent"
	DiscountFixed   DiscountType = "fixed"
)

type Coupon struct {
	ID            string       `json:"id"`
	Code          string       `json:"code"`
	DiscountType  DiscountType `json:"discount_type"`
	DiscountValue float64      `json:"discount_value"`
	IsActive      bool         `json:"is_active"`
	ExpiresAt     *time.Time   `json:"expires_at,omitempty"`
	UsageLimit    *int         `json:"usage_limit,omitempty"`
	TimesUsed     int          `json:"times_used"`
	CreatedAt     time.Time    `json:"created_at"`
}
