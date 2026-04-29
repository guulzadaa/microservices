package usecase

import (
	"errors"
	"order-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

type OrderUseCase struct {
	repo          OrderRepository
	paymentClient PaymentClient
}

func NewOrderUseCase(repo OrderRepository, paymentClient PaymentClient) *OrderUseCase {
	return &OrderUseCase{
		repo:          repo,
		paymentClient: paymentClient,
	}
}

func (uc *OrderUseCase) CreateOrder(customerID, itemName string, amount int64) (*domain.Order, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	order := &domain.Order{
		ID:         uuid.New().String(),
		CustomerID: customerID,
		ItemName:   itemName,
		Amount:     amount,
		Status:     "Pending",
		CreatedAt:  time.Now(),
	}

	if err := uc.repo.Create(order); err != nil {
		return nil, err
	}

	_, paymentStatus, err := uc.paymentClient.CreatePayment(order.ID, order.Amount)
	if err != nil {
		_ = uc.repo.UpdateStatus(order.ID, "Failed")
		return nil, err
	}

	if paymentStatus == "Authorized" {
		order.Status = "Paid"
	} else {
		order.Status = "Failed"
	}

	if err := uc.repo.UpdateStatus(order.ID, order.Status); err != nil {
		return nil, err
	}

	return order, nil
}

func (uc *OrderUseCase) GetOrderByID(id string) (*domain.Order, error) {
	return uc.repo.GetByID(id)
}

func (uc *OrderUseCase) CancelOrder(id string) error {
	order, err := uc.repo.GetByID(id)
	if err != nil {
		return err
	}

	if order.Status != "Pending" {
		return errors.New("only pending orders can be cancelled")
	}

	return uc.repo.UpdateStatus(id, "Cancelled")
}

func (uc *OrderUseCase) GetPaymentStats() (*PaymentStats, error) {
	return uc.paymentClient.GetPaymentStats()
}
