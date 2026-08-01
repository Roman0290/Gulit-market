package admin

import (
	"context"
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Analytics(ctx context.Context) (*AnalyticsSummary, error) {
	s := &AnalyticsSummary{OrdersByStatus: map[string]int{}}

	if err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total), 0) FROM orders WHERE payment_status = 'paid'`,
	).Scan(&s.TotalRevenue); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM vendors WHERE status = 'approved'`,
	).Scan(&s.ActiveVendors); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM vendors WHERE status = 'pending'`,
	).Scan(&s.PendingVendorApprovals); err != nil {
		return nil, err
	}

	statusRows, err := r.db.QueryContext(ctx, `SELECT status, COUNT(*) FROM orders GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var status string
		var count int
		if err := statusRows.Scan(&status, &count); err != nil {
			return nil, err
		}
		s.OrdersByStatus[status] = count
		s.TotalOrders += count
	}
	if err := statusRows.Err(); err != nil {
		return nil, err
	}

	topVendors, err := r.topVendors(ctx)
	if err != nil {
		return nil, err
	}
	s.TopVendors = topVendors

	dailyStats, err := r.dailyStats(ctx)
	if err != nil {
		return nil, err
	}
	s.DailyStats = dailyStats

	return s, nil
}

func (r *Repository) topVendors(ctx context.Context) ([]TopVendor, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT v.id, v.shop_name, COALESCE(SUM(o.subtotal), 0), COUNT(o.id)
		FROM vendors v
		JOIN orders o ON o.vendor_id = v.id AND o.payment_status = 'paid'
		GROUP BY v.id, v.shop_name
		ORDER BY SUM(o.subtotal) DESC
		LIMIT 5
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []TopVendor{}
	for rows.Next() {
		var v TopVendor
		if err := rows.Scan(&v.VendorID, &v.ShopName, &v.Revenue, &v.OrderCount); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, rows.Err()
}

// dailyStats returns a zero-filled 30-day time series (including days with
// no orders) so the caller can plot a growth chart without gaps.
func (r *Repository) dailyStats(ctx context.Context) ([]DailyStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT gs.day::date, COUNT(o.id), COALESCE(SUM(o.total), 0)
		FROM generate_series(CURRENT_DATE - INTERVAL '29 days', CURRENT_DATE, INTERVAL '1 day') AS gs(day)
		LEFT JOIN orders o ON o.created_at::date = gs.day AND o.payment_status = 'paid'
		GROUP BY gs.day
		ORDER BY gs.day
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []DailyStat{}
	for rows.Next() {
		var d DailyStat
		var day time.Time
		if err := rows.Scan(&day, &d.OrderCount, &d.Revenue); err != nil {
			return nil, err
		}
		d.Date = day.Format("2006-01-02")
		list = append(list, d)
	}
	return list, rows.Err()
}
