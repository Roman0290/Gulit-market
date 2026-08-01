package banners

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("banner not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const bannerColumns = `id, image_url, COALESCE(link_url, ''), COALESCE(title, ''), is_active, sort_order, created_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanBanner(row rowScanner) (*Banner, error) {
	var b Banner
	err := row.Scan(&b.ID, &b.ImageURL, &b.LinkURL, &b.Title, &b.IsActive, &b.SortOrder, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &b, nil
}

func (r *Repository) ListActive(ctx context.Context) ([]Banner, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+bannerColumns+` FROM banners WHERE is_active = true ORDER BY sort_order, created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Banner{}
	for rows.Next() {
		b, err := scanBanner(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *b)
	}
	return list, rows.Err()
}

func (r *Repository) ListAll(ctx context.Context) ([]Banner, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+bannerColumns+` FROM banners ORDER BY sort_order, created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []Banner{}
	for rows.Next() {
		b, err := scanBanner(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *b)
	}
	return list, rows.Err()
}

func (r *Repository) Create(ctx context.Context, b *Banner) (*Banner, error) {
	query := `
		INSERT INTO banners (image_url, link_url, title, is_active, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + bannerColumns

	return scanBanner(r.db.QueryRowContext(ctx, query, b.ImageURL, nullableString(b.LinkURL), nullableString(b.Title), b.IsActive, b.SortOrder))
}

func (r *Repository) Update(ctx context.Context, id string, b *Banner) (*Banner, error) {
	query := `
		UPDATE banners SET image_url = $1, link_url = $2, title = $3, is_active = $4, sort_order = $5
		WHERE id = $6
		RETURNING ` + bannerColumns

	return scanBanner(r.db.QueryRowContext(ctx, query, b.ImageURL, nullableString(b.LinkURL), nullableString(b.Title), b.IsActive, b.SortOrder, id))
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM banners WHERE id = $1`, id)
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

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
