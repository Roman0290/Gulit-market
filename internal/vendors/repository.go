package vendors

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound      = errors.New("vendor not found")
	ErrDuplicateUser = errors.New("this user already has a vendor profile")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const vendorColumns = `id, user_id, shop_name, description, location, status, created_at`

func (r *Repository) GetByUserID(ctx context.Context, userID string) (*Vendor, error) {
	query := `SELECT ` + vendorColumns + ` FROM vendors WHERE user_id = $1`
	return scanVendor(r.db.QueryRowContext(ctx, query, userID))
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Vendor, error) {
	query := `SELECT ` + vendorColumns + ` FROM vendors WHERE id = $1`
	return scanVendor(r.db.QueryRowContext(ctx, query, id))
}

func (r *Repository) GetApprovedByID(ctx context.Context, id string) (*Vendor, error) {
	query := `SELECT ` + vendorColumns + ` FROM vendors WHERE id = $1 AND status = 'approved'`
	return scanVendor(r.db.QueryRowContext(ctx, query, id))
}

func (r *Repository) ListApproved(ctx context.Context) ([]Vendor, error) {
	query := `SELECT ` + vendorColumns + ` FROM vendors WHERE status = 'approved' ORDER BY shop_name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Vendor{}
	for rows.Next() {
		v, err := scanVendorRow(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, rows.Err()
}

func (r *Repository) Create(ctx context.Context, v *Vendor) (*Vendor, error) {
	query := `
		INSERT INTO vendors (user_id, shop_name, description, location, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + vendorColumns

	created, err := scanVendor(r.db.QueryRowContext(ctx, query,
		v.UserID, v.ShopName, nullableString(v.Description), nullableString(v.Location), v.Status,
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateUser
		}
		return nil, err
	}
	return created, nil
}

func (r *Repository) Update(ctx context.Context, id, userID string, v *Vendor) (*Vendor, error) {
	query := `
		UPDATE vendors SET shop_name = $1, description = $2, location = $3
		WHERE id = $4 AND user_id = $5
		RETURNING ` + vendorColumns

	updated, err := scanVendor(r.db.QueryRowContext(ctx, query,
		v.ShopName, nullableString(v.Description), nullableString(v.Location), id, userID,
	))
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status Status) (*Vendor, error) {
	query := `UPDATE vendors SET status = $1 WHERE id = $2 RETURNING ` + vendorColumns
	return scanVendor(r.db.QueryRowContext(ctx, query, status, id))
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanVendor(row rowScanner) (*Vendor, error) {
	var v Vendor
	var description, location sql.NullString

	err := row.Scan(&v.ID, &v.UserID, &v.ShopName, &description, &location, &v.Status, &v.CreatedAt)
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

func scanVendorRow(rows *sql.Rows) (Vendor, error) {
	v, err := scanVendor(rows)
	if err != nil {
		return Vendor{}, err
	}
	return *v, nil
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
