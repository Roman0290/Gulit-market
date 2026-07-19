package cart

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrProductNotFound    = errors.New("product not found")
	ErrProductUnavailable = errors.New("product is not available")
	ErrItemNotFound       = errors.New("cart item not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetCart(ctx context.Context, customerID string) (*Cart, error) {
	var cartID string
	err := r.db.QueryRowContext(ctx, `SELECT id FROM carts WHERE customer_id = $1`, customerID).Scan(&cartID)
	if errors.Is(err, sql.ErrNoRows) {
		return &Cart{Items: []Item{}}, nil
	}
	if err != nil {
		return nil, err
	}

	items, err := r.listItems(ctx, cartID)
	if err != nil {
		return nil, err
	}

	c := &Cart{ID: cartID, Items: items}
	for _, it := range items {
		c.Subtotal += it.LineTotal
	}
	return c, nil
}

func (r *Repository) listItems(ctx context.Context, cartID string) ([]Item, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ci.id, ci.product_id, p.name, p.vendor_id, p.price, p.unit, ci.quantity
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		ORDER BY ci.created_at
	`, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.ProductID, &it.ProductName, &it.VendorID, &it.UnitPrice, &it.Unit, &it.Quantity); err != nil {
			return nil, err
		}
		it.LineTotal = it.UnitPrice * float64(it.Quantity)
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *Repository) AddItem(ctx context.Context, customerID, productID string, quantity int) (*Item, error) {
	var isActive bool
	err := r.db.QueryRowContext(ctx, `SELECT is_active FROM products WHERE id = $1`, productID).Scan(&isActive)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	if !isActive {
		return nil, ErrProductUnavailable
	}

	cartID, err := r.getOrCreateCartID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	var itemID string
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO cart_items (cart_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (cart_id, product_id)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity, updated_at = now()
		RETURNING id
	`, cartID, productID, quantity).Scan(&itemID)
	if err != nil {
		return nil, err
	}

	return r.getItem(ctx, itemID)
}

func (r *Repository) UpdateItemQuantity(ctx context.Context, itemID, customerID string, quantity int) (*Item, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE cart_items SET quantity = $1, updated_at = now()
		WHERE id = $2 AND cart_id IN (SELECT id FROM carts WHERE customer_id = $3)
	`, quantity, itemID, customerID)
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, ErrItemNotFound
	}
	return r.getItem(ctx, itemID)
}

func (r *Repository) RemoveItem(ctx context.Context, itemID, customerID string) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM cart_items
		WHERE id = $1 AND cart_id IN (SELECT id FROM carts WHERE customer_id = $2)
	`, itemID, customerID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrItemNotFound
	}
	return nil
}

func (r *Repository) Clear(ctx context.Context, customerID string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM cart_items WHERE cart_id IN (SELECT id FROM carts WHERE customer_id = $1)
	`, customerID)
	return err
}

func (r *Repository) getOrCreateCartID(ctx context.Context, customerID string) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO carts (customer_id) VALUES ($1)
		ON CONFLICT (customer_id) DO UPDATE SET customer_id = EXCLUDED.customer_id
		RETURNING id
	`, customerID).Scan(&id)
	return id, err
}

func (r *Repository) getItem(ctx context.Context, itemID string) (*Item, error) {
	var it Item
	err := r.db.QueryRowContext(ctx, `
		SELECT ci.id, ci.product_id, p.name, p.vendor_id, p.price, p.unit, ci.quantity
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.id = $1
	`, itemID).Scan(&it.ID, &it.ProductID, &it.ProductName, &it.VendorID, &it.UnitPrice, &it.Unit, &it.Quantity)
	if err != nil {
		return nil, err
	}
	it.LineTotal = it.UnitPrice * float64(it.Quantity)
	return &it, nil
}
