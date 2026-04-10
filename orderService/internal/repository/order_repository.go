package repository

import (
	"database/sql"
	"orderService/internal/domain"
)

type CustomerRevenue struct {
	CustomerID  string
	TotalAmount int64
	OrdersCount int
}

type OrderRepository interface {
	Create(order *domain.Order, idempotencyKey string) error
	GetByID(id string) (*domain.Order, error)
	UpdateStatus(id string, status string) error
	GetByIdempotencyKey(key string) (*domain.Order, error)
	GetRevenueByCustomerID(customerID string) (*CustomerRevenue, error)
}

type orderRepo struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) Create(order *domain.Order, idempotencyKey string) error {
	var key *string
	if idempotencyKey != "" {
		key = &idempotencyKey
	}

	_, err := r.db.Exec(
		`INSERT INTO orders (id, customer_id, item_name, amount, status, created_at, idempotency_key)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		order.ID, order.CustomerID, order.ItemName, order.Amount, order.Status, order.CreatedAt, key,
	)
	return err
}

func (r *orderRepo) GetByID(id string) (*domain.Order, error) {
	row := r.db.QueryRow(
		`SELECT id, customer_id, item_name, amount, status, created_at FROM orders WHERE id=$1`,
		id,
	)

	var o domain.Order
	if err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepo) UpdateStatus(id string, status string) error {
	_, err := r.db.Exec(`UPDATE orders SET status=$1 WHERE id=$2`, status, id)
	return err
}

func (r *orderRepo) GetByIdempotencyKey(key string) (*domain.Order, error) {
	if key == "" {
		return nil, nil
	}

	row := r.db.QueryRow(
		`SELECT id, customer_id, item_name, amount, status, created_at FROM orders WHERE idempotency_key=$1`,
		key,
	)

	var o domain.Order
	err := row.Scan(&o.ID, &o.CustomerID, &o.ItemName, &o.Amount, &o.Status, &o.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepo) GetRevenueByCustomerID(customerID string) (*CustomerRevenue, error) {
	row := r.db.QueryRow(
		`SELECT COALESCE(SUM(amount), 0), COUNT(*) FROM orders WHERE customer_id=$1 AND status='Paid'`,
		customerID,
	)

	rev := &CustomerRevenue{CustomerID: customerID}
	if err := row.Scan(&rev.TotalAmount, &rev.OrdersCount); err != nil {
		return nil, err
	}
	return rev, nil
}
