package orders

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNotFound           = errors.New("order not found")
	ErrEmptyCart          = errors.New("cart is empty")
	ErrAddressNotOwned    = errors.New("address not found for this customer")
	ErrProductUnavailable = errors.New("one or more products in the cart are no longer available")
	ErrInsufficientStock  = errors.New("insufficient stock for one or more items in the cart")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

const orderColumns = `id, customer_id, vendor_id, address_id, status, subtotal, delivery_fee, discount, total, payment_status, payment_method, created_at, updated_at`

type cartLine struct {
	productID string
	vendorID  string
	quantity  int
	price     float64
	stock     int
	isActive  bool
}

// Checkout reads the customer's cart, splits it by vendor, and atomically
// creates one order (+ order_items) per vendor, decrements stock, and
// clears the cart. The whole operation runs in a single transaction so a
// stock failure on one vendor's items rolls back the entire checkout.
func (r *Repository) Checkout(ctx context.Context, customerID, addressID string, paymentMethod PaymentMethod) ([]Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var addressOwned bool
	err = tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM addresses WHERE id = $1 AND user_id = $2)`, addressID, customerID,
	).Scan(&addressOwned)
	if err != nil {
		return nil, err
	}
	if !addressOwned {
		return nil, ErrAddressNotOwned
	}

	var cartID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM carts WHERE customer_id = $1`, customerID).Scan(&cartID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEmptyCart
	}
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT ci.product_id, p.vendor_id, ci.quantity, p.price, p.stock_quantity, p.is_active
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		ORDER BY p.vendor_id
		FOR UPDATE OF p
	`, cartID)
	if err != nil {
		return nil, err
	}

	linesByVendor := map[string][]cartLine{}
	for rows.Next() {
		var l cartLine
		if err := rows.Scan(&l.productID, &l.vendorID, &l.quantity, &l.price, &l.stock, &l.isActive); err != nil {
			rows.Close()
			return nil, err
		}
		linesByVendor[l.vendorID] = append(linesByVendor[l.vendorID], l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	if len(linesByVendor) == 0 {
		return nil, ErrEmptyCart
	}

	for _, lines := range linesByVendor {
		for _, l := range lines {
			if !l.isActive {
				return nil, ErrProductUnavailable
			}
			if l.quantity > l.stock {
				return nil, ErrInsufficientStock
			}
		}
	}

	var createdOrders []Order
	for vendorID, lines := range linesByVendor {
		var subtotal float64
		for _, l := range lines {
			subtotal += l.price * float64(l.quantity)
		}
		total := subtotal // delivery_fee and discount default to 0 in v1

		var o Order
		err := tx.QueryRowContext(ctx, `
			INSERT INTO orders (customer_id, vendor_id, address_id, status, subtotal, delivery_fee, discount, total, payment_status, payment_method)
			VALUES ($1, $2, $3, 'pending', $4, 0, 0, $5, 'pending', $6)
			RETURNING `+orderColumns,
			customerID, vendorID, addressID, subtotal, total, paymentMethod,
		).Scan(&o.ID, &o.CustomerID, &o.VendorID, &o.AddressID, &o.Status, &o.Subtotal, &o.DeliveryFee, &o.Discount, &o.Total, &o.PaymentStatus, &o.PaymentMethod, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, err
		}

		for _, l := range lines {
			lineTotal := l.price * float64(l.quantity)
			var item Item
			err := tx.QueryRowContext(ctx, `
				INSERT INTO order_items (order_id, product_id, quantity, unit_price, line_total)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id, product_id, quantity, unit_price, line_total
			`, o.ID, l.productID, l.quantity, l.price, lineTotal).Scan(&item.ID, &item.ProductID, &item.Quantity, &item.UnitPrice, &item.LineTotal)
			if err != nil {
				return nil, err
			}
			o.Items = append(o.Items, item)

			if _, err := tx.ExecContext(ctx, `UPDATE products SET stock_quantity = stock_quantity - $1 WHERE id = $2`, l.quantity, l.productID); err != nil {
				return nil, err
			}
		}

		createdOrders = append(createdOrders, o)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return createdOrders, nil
}

func (r *Repository) ListByCustomer(ctx context.Context, customerID string) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+orderColumns+` FROM orders WHERE customer_id = $1 ORDER BY created_at DESC`, customerID)
	if err != nil {
		return nil, err
	}
	orders, err := scanOrders(rows)
	if err != nil {
		return nil, err
	}

	for i := range orders {
		items, err := r.itemsForOrder(ctx, orders[i].ID)
		if err != nil {
			return nil, err
		}
		orders[i].Items = items
	}
	return orders, nil
}

// GetByIDForParticipant returns the order only if userID is either the
// customer who placed it or the user behind the vendor who owns it.
func (r *Repository) GetByIDForParticipant(ctx context.Context, orderID, userID string) (*Order, error) {
	query := `
		SELECT ` + orderColumns + ` FROM orders
		WHERE id = $1 AND (customer_id = $2 OR vendor_id IN (SELECT id FROM vendors WHERE user_id = $2))
	`
	var o Order
	err := r.db.QueryRowContext(ctx, query, orderID, userID).Scan(
		&o.ID, &o.CustomerID, &o.VendorID, &o.AddressID, &o.Status, &o.Subtotal, &o.DeliveryFee, &o.Discount, &o.Total, &o.PaymentStatus, &o.PaymentMethod, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	items, err := r.itemsForOrder(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items
	return &o, nil
}

func (r *Repository) itemsForOrder(ctx context.Context, orderID string) ([]Item, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, product_id, quantity, unit_price, line_total FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.ID, &it.ProductID, &it.Quantity, &it.UnitPrice, &it.LineTotal); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func scanOrders(rows *sql.Rows) ([]Order, error) {
	defer rows.Close()
	orders := []Order{}
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.VendorID, &o.AddressID, &o.Status, &o.Subtotal, &o.DeliveryFee, &o.Discount, &o.Total, &o.PaymentStatus, &o.PaymentMethod, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}
