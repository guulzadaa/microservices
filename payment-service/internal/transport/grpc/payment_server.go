package grpc

import (
	"context"
	"log"
	"os"
	"strings"

	"payment-service/internal/messaging"
	"payment-service/internal/usecase"

	"github.com/google/uuid"
	paymentpb "github.com/guulzadaa/AP2_generated/paymentpb"
)

type PaymentServer struct {
	paymentpb.UnimplementedPaymentServiceServer
	usecase   *usecase.PaymentUseCase
	publisher *messaging.Publisher
}

func NewPaymentServer(uc *usecase.PaymentUseCase) *PaymentServer {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://admin:admin123@localhost:5672/"
	}

	publisher, err := messaging.NewPublisher(rabbitURL)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}

	return &PaymentServer{
		usecase:   uc,
		publisher: publisher,
	}
}

func (s *PaymentServer) ProcessPayment(
	ctx context.Context,
	req *paymentpb.PaymentRequest,
) (*paymentpb.PaymentResponse, error) {

	payment, err := s.usecase.CreatePayment(
		req.OrderId,
		req.Amount,
	)
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(payment.Status, "authorized") || strings.EqualFold(payment.Status, "completed") {
		event := messaging.PaymentCompletedEvent{
			EventID:       uuid.New().String(),
			OrderID:       req.OrderId,
			Amount:        float64(req.Amount),
			CustomerEmail: "user@example.com",
			Status:        payment.Status,
		}

		err := s.publisher.PublishPaymentCompleted(event)
		if err != nil {
			log.Println("Failed to publish payment event:", err)
		}
	}

	return &paymentpb.PaymentResponse{
		TransactionId: payment.TransactionID,
		Status:        payment.Status,
	}, nil
}

func (s *PaymentServer) GetPaymentStats(
	ctx context.Context,
	req *paymentpb.GetPaymentStatsRequest,
) (*paymentpb.PaymentStats, error) {

	stats, err := s.usecase.GetStats()
	if err != nil {
		return nil, err
	}

	return &paymentpb.PaymentStats{
		TotalCount:      stats.TotalCount,
		AuthorizedCount: stats.AuthorizedCount,
		DeclinedCount:   stats.DeclinedCount,
		TotalAmount:     stats.TotalAmount,
	}, nil
}
