package orders

import "time"

type Status string

const (
	StatusPending        Status = "pending"
	StatusAccepted       Status = "accepted"
	StatusPreparing      Status = "preparing"
	StatusOutForDelivery Status = "out_for_delivery"
	StatusDelivered      Status = "delivered"
	StatusCancelled      Status = "cancelled"
)

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusFailed  PaymentStatus = "failed"
)

type PaymentMethod string

const (
	PaymentMethodCard PaymentMethod = "card"
	PaymentMethodCOD  PaymentMethod = "cod"
)

func (m PaymentMethod) Valid() bool {
	return m == PaymentMethodCard || m == PaymentMethodCOD
}

type Item struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	LineTotal float64 `json:"line_total"`
}

type Order struct {
	ID            string        `json:"id"`
	CustomerID    string        `json:"customer_id"`
	VendorID      string        `json:"vendor_id"`
	AddressID     string        `json:"address_id"`
	Status        Status        `json:"status"`
	Subtotal      float64       `json:"subtotal"`
	DeliveryFee   float64       `json:"delivery_fee"`
	Discount      float64       `json:"discount"`
	Total         float64       `json:"total"`
	PaymentStatus PaymentStatus `json:"payment_status"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	Items         []Item        `json:"items,omitempty"`
}
