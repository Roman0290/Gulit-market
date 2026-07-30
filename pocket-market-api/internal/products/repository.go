package products

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound        = errors.New("product not found")
	ErrHasOrderHistory = errors.New("product has existing orders and cannot be deleted; deactivate it instead")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type ListFilter struct {
	CategoryID string
	VendorID   string
	Query      string
}

const productColumns = `id, vendor_id, category_id, name, description, price, unit, stock_quantity, image_url, is_active, created_at, updated_at`

func (r *Repository) List(ctx context.Context, f ListFilter) ([]Product, error) {
	query := `SELECT ` + productColumns + ` FROM products WHERE is_active = true`
	args := []any{}
	argN := 1

	if f.CategoryID != "" {
		query += fmt.Sprintf(" AND category_id = $%d", argN)
		args = append(args, f.CategoryID)
		argN++
	}
	if f.VendorID != "" {
		query += fmt.Sprintf(" AND vendor_id = $%d", argN)
		args = append(args, f.VendorID)
		argN++
	}
	if f.Query != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argN)
		args = append(args, "%"+strings.TrimSpace(f.Query)+"%")
		argN++
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Product{}
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (r *Repository) GetActiveByID(ctx context.Context, id string) (*Product, error) {
	query := `SELECT ` + productColumns + ` FROM products WHERE id = $1 AND is_active = true`
	p, err := scanProduct(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *Repository) Create(ctx context.Context, p *Product) (*Product, error) {
	const query = `
		INSERT INTO products (vendor_id, category_id, name, description, price, unit, stock_quantity, image_url, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING ` + productColumns

	created, err := scanProduct(r.db.QueryRowContext(ctx, query,
		p.VendorID, nullableString(p.CategoryID), p.Name, p.Description, p.Price, p.Unit, p.StockQuantity, p.ImageURL, p.IsActive,
	))
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *Repository) Update(ctx context.Context, id, vendorID string, p *Product) (*Product, error) {
	const query = `
		UPDATE products SET
			category_id = $1, name = $2, description = $3, price = $4,
			unit = $5, stock_quantity = $6, image_url = $7, is_active = $8, updated_at = now()
		WHERE id = $9 AND vendor_id = $10
		RETURNING ` + productColumns

	updated, err := scanProduct(r.db.QueryRowContext(ctx, query,
		nullableString(p.CategoryID), p.Name, p.Description, p.Price, p.Unit, p.StockQuantity, p.ImageURL, p.IsActive, id, vendorID,
	))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &updated, nil
}

func (r *Repository) UpdateStock(ctx context.Context, id, vendorID string, quantity int) (*Product, error) {
	const query = `
		UPDATE products SET stock_quantity = $1, updated_at = now()
		WHERE id = $2 AND vendor_id = $3
		RETURNING ` + productColumns

	updated, err := scanProduct(r.db.QueryRowContext(ctx, query, quantity, id, vendorID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &updated, nil
}

func (r *Repository) Delete(ctx context.Context, id, vendorID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1 AND vendor_id = $2`, id, vendorID)
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

// SetActiveByAdmin toggles a listing's visibility regardless of which
// vendor owns it, for moderating fake/bad products without touching order
// history (a hard delete would orphan any existing order_items referencing it).
func (r *Repository) SetActiveByAdmin(ctx context.Context, id string, isActive bool) (*Product, error) {
	query := `UPDATE products SET is_active = $1, updated_at = now() WHERE id = $2 RETURNING ` + productColumns
	updated, err := scanProduct(r.db.QueryRowContext(ctx, query, isActive, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &updated, nil
}

// DeleteByAdmin removes a product regardless of vendor ownership. Fails
// with a foreign-key error if the product appears in any order_items -
// use SetActiveByAdmin instead for listings that have already sold.
func (r *Repository) DeleteByAdmin(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrHasOrderHistory
		}
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

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProduct(row rowScanner) (Product, error) {
	var p Product
	var categoryID, description, imageURL sql.NullString

	err := row.Scan(
		&p.ID, &p.VendorID, &categoryID, &p.Name, &description, &p.Price, &p.Unit,
		&p.StockQuantity, &imageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return Product{}, err
	}

	p.CategoryID = categoryID.String
	p.Description = description.String
	p.ImageURL = imageURL.String
	return p, nil
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
