package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound      = errors.New("user not found")
	ErrDuplicateUser = errors.New("email or phone already registered")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const userColumns = `id, name, email, phone, password_hash, role, status, created_at, updated_at`

func (r *Repository) Create(ctx context.Context, u *User) (*User, error) {
	const query = `
		INSERT INTO users (name, email, phone, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, status, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query, u.Name, u.Email, u.Phone, u.PasswordHash, u.Role).
		Scan(&u.ID, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateUser
		}
		return nil, err
	}

	return u, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE email = $1`
	return r.scanUser(r.db.QueryRowContext(ctx, query, email))
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE id = $1`
	return r.scanUser(r.db.QueryRowContext(ctx, query, id))
}

// List returns every user, for admin management. Optionally filtered by role.
func (r *Repository) List(ctx context.Context, role Role) ([]User, error) {
	query := `SELECT ` + userColumns + ` FROM users`
	args := []any{}
	if role != "" {
		query += ` WHERE role = $1`
		args = append(args, role)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status Status) (*User, error) {
	query := `UPDATE users SET status = $1, updated_at = now() WHERE id = $2 RETURNING ` + userColumns
	return r.scanUser(r.db.QueryRowContext(ctx, query, status, id))
}

func (r *Repository) scanUser(row *sql.Row) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}
