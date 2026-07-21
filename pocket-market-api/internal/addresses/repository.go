package addresses

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("address not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const addressColumns = `id, user_id, label, line1, city, lat, lng, is_default`

func (r *Repository) List(ctx context.Context, userID string) ([]Address, error) {
	query := `SELECT ` + addressColumns + ` FROM addresses WHERE user_id = $1 ORDER BY is_default DESC, line1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Address{}
	for rows.Next() {
		a, err := scanAddress(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

// Exists reports whether the given address belongs to userID — used by
// checkout to validate the address_id without needing the full record.
func (r *Repository) Exists(ctx context.Context, id, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM addresses WHERE id = $1 AND user_id = $2)`, id, userID,
	).Scan(&exists)
	return exists, err
}

func (r *Repository) Create(ctx context.Context, a *Address) (*Address, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if a.IsDefault {
		if _, err := tx.ExecContext(ctx, `UPDATE addresses SET is_default = false WHERE user_id = $1`, a.UserID); err != nil {
			return nil, err
		}
	}

	query := `
		INSERT INTO addresses (user_id, label, line1, city, lat, lng, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING ` + addressColumns

	created, err := scanAddress(tx.QueryRowContext(ctx, query, a.UserID, nullableString(a.Label), a.Line1, a.City, a.Lat, a.Lng, a.IsDefault))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *Repository) Update(ctx context.Context, id, userID string, a *Address) (*Address, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if a.IsDefault {
		if _, err := tx.ExecContext(ctx, `UPDATE addresses SET is_default = false WHERE user_id = $1`, userID); err != nil {
			return nil, err
		}
	}

	query := `
		UPDATE addresses SET label = $1, line1 = $2, city = $3, lat = $4, lng = $5, is_default = $6
		WHERE id = $7 AND user_id = $8
		RETURNING ` + addressColumns

	updated, err := scanAddress(tx.QueryRowContext(ctx, query, nullableString(a.Label), a.Line1, a.City, a.Lat, a.Lng, a.IsDefault, id, userID))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &updated, nil
}

func (r *Repository) Delete(ctx context.Context, id, userID string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM addresses WHERE id = $1 AND user_id = $2`, id, userID)
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

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAddress(row rowScanner) (Address, error) {
	var a Address
	var label sql.NullString

	err := row.Scan(&a.ID, &a.UserID, &label, &a.Line1, &a.City, &a.Lat, &a.Lng, &a.IsDefault)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Address{}, ErrNotFound
		}
		return Address{}, err
	}
	a.Label = label.String
	return a, nil
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
