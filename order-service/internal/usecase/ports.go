package usecase

import "order-service/internal/domain"

type OrderRepository interface {
	Create(order *domain.Order) error
	GetByID(id string) (*domain.Order, error)
	UpdateStatus(id string, status string) error
}

type PaymentStats struct {
	TotalCount      int64
	AuthorizedCount int64
	DeclinedCount   int64
	TotalAmount     int64
}

type PaymentClient interface {
	CreatePayment(orderID string, amount int64) (transactionID string, paymentStatus string, err error)
	GetPaymentStats() (*PaymentStats, error)
}
