package repository

import (
	"database/sql"
	"payment-service/internal/domain"
	"payment-service/internal/usecase"
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

	_, err := r.db.Exec(
		query,
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

func (r *PaymentRepository) GetStats() (*usecase.PaymentStats, error) {
	query := `
		SELECT
			COUNT(*) AS total_count,
			COUNT(*) FILTER (WHERE status = 'Authorized') AS authorized_count,
			COUNT(*) FILTER (WHERE status = 'Declined') AS declined_count,
			COALESCE(SUM(amount), 0) AS total_amount
		FROM payments
	`

	var stats usecase.PaymentStats

	err := r.db.QueryRow(query).Scan(
		&stats.TotalCount,
		&stats.AuthorizedCount,
		&stats.DeclinedCount,
		&stats.TotalAmount,
	)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
