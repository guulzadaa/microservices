package repository

import (
	"database/sql"
	"payment-service/internal/domain"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(payment *domain.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, transaction_id, amount, status)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(query,
		payment.ID,
		payment.OrderID,
		payment.TransactionID,
		payment.Amount,
		payment.Status,
	)

	return err
}

func (r *PaymentRepository) GetByOrderID(orderID string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, transaction_id, amount, status
		FROM payments
		WHERE order_id = $1
	`

	var payment domain.Payment

	err := r.db.QueryRow(query, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.TransactionID,
		&payment.Amount,
		&payment.Status,
	)
	if err != nil {
		return nil, err
	}

	return &payment, nil
}
