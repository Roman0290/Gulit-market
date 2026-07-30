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

const orderColumns = `id, customer_id, vendor_id, address_id, status, subtotal, delivery_fee, discount, commission_amount, tax_amount, total, payment_status, payment_method, created_at, updated_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanOrder(row rowScanner) (Order, error) {
	var o Order
	err := row.Scan(
		&o.ID, &o.CustomerID, &o.VendorID, &o.AddressID, &o.Status, &o.Subtotal, &o.DeliveryFee, &o.Discount,
		&o.CommissionAmount, &o.TaxAmount, &o.Total, &o.PaymentStatus, &o.PaymentMethod, &o.CreatedAt, &o.UpdatedAt,
	)
	return o, err
}

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
//
// commissionRate and taxRate are percentages (e.g. 10 = 10%), applied to
// each vendor group's subtotal independently. deliveryFee is charged once
// per vendor group (each vendor's order is its own delivery).
func (r *Repository) Checkout(ctx context.Context, customerID, addressID string, paymentMethod PaymentMethod, commissionRate, taxRate, deliveryFee float64) ([]Order, error) {
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
		commissionAmount := round2(subtotal * commissionRate / 100)
		taxAmount := round2(subtotal * taxRate / 100)
		total := subtotal + taxAmount + deliveryFee

		o, err := scanOrder(tx.QueryRowContext(ctx, `
			INSERT INTO orders (customer_id, vendor_id, address_id, status, subtotal, delivery_fee, discount, commission_amount, tax_amount, total, payment_status, payment_method)
			VALUES ($1, $2, $3, 'pending', $4, $5, 0, $6, $7, $8, 'pending', $9)
			RETURNING `+orderColumns,
			customerID, vendorID, addressID, subtotal, deliveryFee, commissionAmount, taxAmount, total, paymentMethod,
		))
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

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
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

// ListAll returns every order platform-wide, for admin use.
func (r *Repository) ListAll(ctx context.Context) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+orderColumns+` FROM orders ORDER BY created_at DESC`)
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

// GetByID returns an order regardless of participant, for admin use.
func (r *Repository) GetByID(ctx context.Context, orderID string) (*Order, error) {
	o, err := scanOrder(r.db.QueryRowContext(ctx, `SELECT `+orderColumns+` FROM orders WHERE id = $1`, orderID))
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

func (r *Repository) ListByVendor(ctx context.Context, vendorID string) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+orderColumns+` FROM orders WHERE vendor_id = $1 ORDER BY created_at DESC`, vendorID)
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

// UpdateStatusForVendor updates an order's status only if it belongs to the
// given vendor, returning ErrNotFound otherwise (no existence leak).
func (r *Repository) UpdateStatusForVendor(ctx context.Context, orderID, vendorID string, status Status) (*Order, error) {
	query := `
		UPDATE orders SET status = $1, updated_at = now()
		WHERE id = $2 AND vendor_id = $3
		RETURNING ` + orderColumns

	o, err := scanOrder(r.db.QueryRowContext(ctx, query, status, orderID, vendorID))
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

// GetStatusForVendor returns the current status of an order owned by the
// given vendor, used to validate status transitions before writing.
func (r *Repository) GetStatusForVendor(ctx context.Context, orderID, vendorID string) (Status, error) {
	var status Status
	err := r.db.QueryRowContext(ctx,
		`SELECT status FROM orders WHERE id = $1 AND vendor_id = $2`, orderID, vendorID,
	).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return status, nil
}

// GetByIDForParticipant returns the order only if userID is either the
// customer who placed it or the user behind the vendor who owns it.
func (r *Repository) GetByIDForParticipant(ctx context.Context, orderID, userID string) (*Order, error) {
	query := `
		SELECT ` + orderColumns + ` FROM orders
		WHERE id = $1 AND (customer_id = $2 OR vendor_id IN (SELECT id FROM vendors WHERE user_id = $2))
	`
	o, err := scanOrder(r.db.QueryRowContext(ctx, query, orderID, userID))
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
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}
