package payments

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrOrderNotFound    = errors.New("order not found for this customer")
	ErrOrderAlreadyPaid = errors.New("order is already paid")
	ErrPaymentNotFound  = errors.New("payment not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type OrderForPayment struct {
	Total         float64
	PaymentStatus string
}

func (r *Repository) GetOrderForPayment(ctx context.Context, orderID, customerID string) (*OrderForPayment, error) {
	var o OrderForPayment
	err := r.db.QueryRowContext(ctx,
		`SELECT total, payment_status FROM orders WHERE id = $1 AND customer_id = $2`,
		orderID, customerID,
	).Scan(&o.Total, &o.PaymentStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return &o, nil
}

func (r *Repository) Create(ctx context.Context, p *Payment) (*Payment, error) {
	const query = `
		INSERT INTO payments (order_id, provider, provider_ref, amount, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, order_id, provider, provider_ref, amount, status, created_at
	`
	err := r.db.QueryRowContext(ctx, query, p.OrderID, p.Provider, p.ProviderRef, p.Amount, p.Status).
		Scan(&p.ID, &p.OrderID, &p.Provider, &p.ProviderRef, &p.Amount, &p.Status, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// MarkSucceeded transactionally flips both the payment and its order to
// their "paid" states in response to a Stripe webhook event.
func (r *Repository) MarkSucceeded(ctx context.Context, providerRef string) error {
	return r.markPaymentAndOrder(ctx, providerRef, StatusSucceeded, "paid")
}

func (r *Repository) MarkFailed(ctx context.Context, providerRef string) error {
	return r.markPaymentAndOrder(ctx, providerRef, StatusFailed, "failed")
}

func (r *Repository) markPaymentAndOrder(ctx context.Context, providerRef string, paymentStatus Status, orderPaymentStatus string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var orderID string
	err = tx.QueryRowContext(ctx,
		`UPDATE payments SET status = $1 WHERE provider_ref = $2 RETURNING order_id`,
		paymentStatus, providerRef,
	).Scan(&orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPaymentNotFound
		}
		return err
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE orders SET payment_status = $1, updated_at = now() WHERE id = $2`,
		orderPaymentStatus, orderID,
	); err != nil {
		return err
	}

	return tx.Commit()
}
