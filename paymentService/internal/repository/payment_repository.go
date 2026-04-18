package repository

import (
	"database/sql"
	"fmt"
	"paymentService/internal/domain"
)

type PaymentRepository interface {
	Create(payment *domain.Payment) error
	GetByOrderID(orderID string) (*domain.Payment, error)
	FindByAmountRange(min, max int64) ([]*domain.Payment, error)
}
type paymentRepo struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Create(p *domain.Payment) error {
	_, err := r.db.Exec(
		"INSERT INTO payments (id, order_id, transaction_id, amount, status) VALUES ($1,$2,$3,$4,$5)",
		p.ID, p.OrderID, p.TransactionID, p.Amount, p.Status,
	)
	return err
}

func (r *paymentRepo) GetByOrderID(orderID string) (*domain.Payment, error) {
	row := r.db.QueryRow("SELECT id, order_id, transaction_id, amount, status FROM payments WHERE order_id=$1", orderID)

	var p domain.Payment
	err := row.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
func (r *paymentRepo) FindByAmountRange(min, max int64) ([]*domain.Payment, error) {
	query := `SELECT id, order_id, transaction_id, amount, status FROM payments WHERE 1=1`
	args := []interface{}{}
	i := 1

	if min > 0 {
		query += fmt.Sprintf(" AND amount >= $%d", i)
		args = append(args, min)
		i++
	}
	if max > 0 {
		query += fmt.Sprintf(" AND amount <= $%d", i)
		args = append(args, max)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		p := &domain.Payment{}
		if err := rows.Scan(&p.ID, &p.OrderID, &p.TransactionID, &p.Amount, &p.Status); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}
