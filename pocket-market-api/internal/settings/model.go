package settings

import "time"

type Settings struct {
	CommissionRate     float64   `json:"commission_rate"`
	TaxRate            float64   `json:"tax_rate"`
	DefaultDeliveryFee float64   `json:"default_delivery_fee"`
	UpdatedAt          time.Time `json:"updated_at"`
}
