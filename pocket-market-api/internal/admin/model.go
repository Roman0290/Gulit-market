package admin

type AnalyticsSummary struct {
	TotalRevenue           float64        `json:"total_revenue"`
	ActiveVendors          int            `json:"active_vendors"`
	PendingVendorApprovals int            `json:"pending_vendor_approvals"`
	TotalOrders            int            `json:"total_orders"`
	OrdersByStatus         map[string]int `json:"orders_by_status"`
}
