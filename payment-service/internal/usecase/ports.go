package usecase

import "payment-service/internal/domain"

type PaymentStats struct {
	TotalCount      int64
	AuthorizedCount int64
	DeclinedCount   int64
	TotalAmount     int64
}

type PaymentRepository interface {
	Create(payment *domain.Payment) error
	GetByOrderID(orderID string) (*domain.Payment, error)
	GetStats() (*PaymentStats, error)
}
