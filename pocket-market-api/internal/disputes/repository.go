package disputes

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNotFound        = errors.New("dispute not found")
	ErrNotParticipant  = errors.New("you are not a participant on this order")
	ErrAlreadyResolved = errors.New("dispute is already resolved")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const disputeColumns = `id, order_id, raised_by, reason, status, COALESCE(resolution_note, ''), created_at, resolved_at`

// Create opens a dispute, but only if raisedBy is actually a participant
// (the customer who placed the order, or the user behind the owning vendor).
func (r *Repository) Create(ctx context.Context, orderID, raisedBy, reason string) (*Dispute, error) {
	var isParticipant bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM orders
			WHERE id = $1 AND (customer_id = $2 OR vendor_id IN (SELECT id FROM vendors WHERE user_id = $2))
		)
	`, orderID, raisedBy).Scan(&isParticipant)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, ErrNotParticipant
	}

	query := `
		INSERT INTO disputes (order_id, raised_by, reason)
		VALUES ($1, $2, $3)
		RETURNING ` + disputeColumns

	return scanDispute(r.db.QueryRowContext(ctx, query, orderID, raisedBy, reason))
}

func (r *Repository) ListAll(ctx context.Context, status Status) ([]Dispute, error) {
	query := `SELECT ` + disputeColumns + ` FROM disputes`
	args := []any{}
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Dispute{}
	for rows.Next() {
		d, err := scanDispute(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *d)
	}
	return list, rows.Err()
}

func (r *Repository) Resolve(ctx context.Context, id, resolutionNote string) (*Dispute, error) {
	query := `
		UPDATE disputes SET status = 'resolved', resolution_note = $1, resolved_at = now()
		WHERE id = $2 AND status = 'open'
		RETURNING ` + disputeColumns

	d, err := scanDispute(r.db.QueryRowContext(ctx, query, resolutionNote, id))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			var exists bool
			r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM disputes WHERE id = $1)`, id).Scan(&exists)
			if exists {
				return nil, ErrAlreadyResolved
			}
			return nil, ErrNotFound
		}
		return nil, err
	}
	return d, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDispute(row rowScanner) (*Dispute, error) {
	var d Dispute
	err := row.Scan(&d.ID, &d.OrderID, &d.RaisedBy, &d.Reason, &d.Status, &d.ResolutionNote, &d.CreatedAt, &d.ResolvedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &d, nil
}
