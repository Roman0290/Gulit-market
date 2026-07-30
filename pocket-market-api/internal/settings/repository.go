package settings

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

const settingsColumns = `commission_rate, tax_rate, default_delivery_fee, updated_at`

func (r *Repository) Get(ctx context.Context) (*Settings, error) {
	var s Settings
	err := r.db.QueryRowContext(ctx, `SELECT `+settingsColumns+` FROM platform_settings WHERE id = 1`).
		Scan(&s.CommissionRate, &s.TaxRate, &s.DefaultDeliveryFee, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) Update(ctx context.Context, s *Settings) (*Settings, error) {
	query := `
		UPDATE platform_settings
		SET commission_rate = $1, tax_rate = $2, default_delivery_fee = $3, updated_at = now()
		WHERE id = 1
		RETURNING ` + settingsColumns

	var updated Settings
	err := r.db.QueryRowContext(ctx, query, s.CommissionRate, s.TaxRate, s.DefaultDeliveryFee).
		Scan(&updated.CommissionRate, &updated.TaxRate, &updated.DefaultDeliveryFee, &updated.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}
