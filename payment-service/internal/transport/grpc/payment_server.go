package grpc

import (
	"context"

	"payment-service/internal/usecase"

	paymentpb "github.com/guulzadaa/AP2_generated/paymentpb"
)

type PaymentServer struct {
	paymentpb.UnimplementedPaymentServiceServer
	usecase *usecase.PaymentUseCase
}

func NewPaymentServer(uc *usecase.PaymentUseCase) *PaymentServer {
	return &PaymentServer{
		usecase: uc,
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

	return &paymentpb.PaymentResponse{
		TransactionId: payment.TransactionID,
		Status:        payment.Status,
	}, nil
}
