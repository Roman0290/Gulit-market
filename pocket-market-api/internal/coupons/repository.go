package coupons

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound      = errors.New("coupon not found")
	ErrDuplicateCode = errors.New("a coupon with this code already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const couponColumns = `id, code, discount_type, discount_value, is_active, expires_at, usage_limit, times_used, created_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCoupon(row rowScanner) (*Coupon, error) {
	var c Coupon
	err := row.Scan(&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.IsActive, &c.ExpiresAt, &c.UsageLimit, &c.TimesUsed, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *Repository) List(ctx context.Context) ([]Coupon, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+couponColumns+` FROM coupons ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Coupon{}
	for rows.Next() {
		c, err := scanCoupon(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *c)
	}
	return list, rows.Err()
}

func (r *Repository) Create(ctx context.Context, c *Coupon) (*Coupon, error) {
	query := `
		INSERT INTO coupons (code, discount_type, discount_value, is_active, expires_at, usage_limit)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + couponColumns

	created, err := scanCoupon(r.db.QueryRowContext(ctx, query, c.Code, c.DiscountType, c.DiscountValue, c.IsActive, c.ExpiresAt, c.UsageLimit))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateCode
		}
		return nil, err
	}
	return created, nil
}

func (r *Repository) Update(ctx context.Context, id string, c *Coupon) (*Coupon, error) {
	query := `
		UPDATE coupons SET discount_type = $1, discount_value = $2, is_active = $3, expires_at = $4, usage_limit = $5
		WHERE id = $6
		RETURNING ` + couponColumns

	return scanCoupon(r.db.QueryRowContext(ctx, query, c.DiscountType, c.DiscountValue, c.IsActive, c.ExpiresAt, c.UsageLimit, id))
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM coupons WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
