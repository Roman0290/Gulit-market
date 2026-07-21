package cart

type Item struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	VendorID    string  `json:"vendor_id"`
	UnitPrice   float64 `json:"unit_price"`
	Unit        string  `json:"unit"`
	Quantity    int     `json:"quantity"`
	LineTotal   float64 `json:"line_total"`
}

type Cart struct {
	ID       string  `json:"id"`
	Items    []Item  `json:"items"`
	Subtotal float64 `json:"subtotal"`
}
