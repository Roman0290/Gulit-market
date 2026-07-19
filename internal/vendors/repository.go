package vendors

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("vendor not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByUserID(ctx context.Context, userID string) (*Vendor, error) {
	const query = `
		SELECT id, user_id, shop_name, description, location, status, created_at
		FROM vendors WHERE user_id = $1
	`
	var v Vendor
	var description, location sql.NullString
	err := r.db.QueryRowContext(ctx, query, userID).
		Scan(&v.ID, &v.UserID, &v.ShopName, &description, &location, &v.Status, &v.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	v.Description = description.String
	v.Location = location.String
	return &v, nil
}
