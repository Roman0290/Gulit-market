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

func (r *Repository) Create(ctx context.Context, u *User) (*User, error) {
	const query = `
		INSERT INTO users (name, email, phone, password_hash, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query, u.Name, u.Email, u.Phone, u.PasswordHash, u.Role).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
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
	const query = `
		SELECT id, name, email, phone, password_hash, role, created_at, updated_at
		FROM users WHERE email = $1
	`
	return r.scanUser(r.db.QueryRowContext(ctx, query, email))
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	const query = `
		SELECT id, name, email, phone, password_hash, role, created_at, updated_at
		FROM users WHERE id = $1
	`
	return r.scanUser(r.db.QueryRowContext(ctx, query, id))
}

func (r *Repository) scanUser(row *sql.Row) (*User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}
