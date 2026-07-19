package admin

import (
	"context"
	"database/sql"
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

	rows, err := r.db.QueryContext(ctx, `SELECT status, COUNT(*) FROM orders GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		s.OrdersByStatus[status] = count
		s.TotalOrders += count
	}
	return s, rows.Err()
}
