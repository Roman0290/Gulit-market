package admin

type TopVendor struct {
	VendorID   string  `json:"vendor_id"`
	ShopName   string  `json:"shop_name"`
	Revenue    float64 `json:"revenue"`
	OrderCount int     `json:"order_count"`
}

type DailyStat struct {
	Date       string  `json:"date"`
	OrderCount int     `json:"order_count"`
	Revenue    float64 `json:"revenue"`
}

type AnalyticsSummary struct {
	TotalRevenue           float64        `json:"total_revenue"`
	ActiveVendors          int            `json:"active_vendors"`
	PendingVendorApprovals int            `json:"pending_vendor_approvals"`
	TotalOrders            int            `json:"total_orders"`
	OrdersByStatus         map[string]int `json:"orders_by_status"`
	TopVendors             []TopVendor    `json:"top_vendors"`
	DailyStats             []DailyStat    `json:"daily_stats"`
}
