package payouts

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

func (r *Repository) Balance(ctx context.Context, vendorID string) (*Balance, error) {
	b := &Balance{VendorID: vendorID}

	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(subtotal - commission_amount), 0)
		FROM orders WHERE vendor_id = $1 AND payment_status = 'paid'
	`, vendorID).Scan(&b.TotalEarned)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0) FROM payouts WHERE vendor_id = $1
	`, vendorID).Scan(&b.TotalPaidOut)
	if err != nil {
		return nil, err
	}

	b.BalanceDue = b.TotalEarned - b.TotalPaidOut
	return b, nil
}

func (r *Repository) Create(ctx context.Context, p *Payout) (*Payout, error) {
	const query = `
		INSERT INTO payouts (vendor_id, amount, note)
		VALUES ($1, $2, $3)
		RETURNING id, vendor_id, amount, COALESCE(note, ''), created_at
	`
	err := r.db.QueryRowContext(ctx, query, p.VendorID, p.Amount, nullableString(p.Note)).
		Scan(&p.ID, &p.VendorID, &p.Amount, &p.Note, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *Repository) ListByVendor(ctx context.Context, vendorID string) ([]Payout, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, vendor_id, amount, COALESCE(note, ''), created_at
		FROM payouts WHERE vendor_id = $1 ORDER BY created_at DESC
	`, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Payout{}
	for rows.Next() {
		var p Payout
		if err := rows.Scan(&p.ID, &p.VendorID, &p.Amount, &p.Note, &p.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
