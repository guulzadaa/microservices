package usecase

import (
	"payment-service/internal/domain"
	"payment-service/internal/messaging"

	"github.com/google/uuid"
)

type PaymentUseCase struct {
	repo      PaymentRepository
	publisher *messaging.Publisher
}

func NewPaymentUseCase(repo PaymentRepository, publisher *messaging.Publisher) *PaymentUseCase {
	return &PaymentUseCase{
		repo:      repo,
		publisher: publisher,
	}
}

func (uc *PaymentUseCase) CreatePayment(orderID string, amount int64) (*domain.Payment, error) {
	status := "Authorized"
	if amount > 100000 {
		status = "Declined"
	}

	payment := &domain.Payment{
		ID:            uuid.New().String(),
		OrderID:       orderID,
		TransactionID: uuid.New().String(),
		Amount:        amount,
		Status:        status,
	}

	if err := uc.repo.Create(payment); err != nil {
		return nil, err
	}

	if status == "Authorized" && uc.publisher != nil {
		event := messaging.PaymentCompletedEvent{
			EventID:       uuid.New().String(),
			OrderID:       orderID,
			Amount:        float64(amount) / 100,
			CustomerEmail: "user@example.com",
			Status:        "completed",
		}

		if err := uc.publisher.PublishPaymentCompleted(event); err != nil {
			return nil, err
		}
	}

	return payment, nil
}

func (uc *PaymentUseCase) GetByOrderID(orderID string) (*domain.Payment, error) {
	return uc.repo.GetByOrderID(orderID)
}

func (uc *PaymentUseCase) GetStats() (*PaymentStats, error) {
	return uc.repo.GetStats()
}
