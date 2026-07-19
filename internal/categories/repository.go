package categories

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

func (r *Repository) List(ctx context.Context) ([]Category, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, icon_url FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cats := []Category{}
	for rows.Next() {
		var c Category
		var iconURL sql.NullString
		if err := rows.Scan(&c.ID, &c.Name, &iconURL); err != nil {
			return nil, err
		}
		c.IconURL = iconURL.String
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *Repository) Create(ctx context.Context, c *Category) (*Category, error) {
	const query = `INSERT INTO categories (name, icon_url) VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, c.Name, nullableString(c.IconURL)).Scan(&c.ID)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
