package repository

import (
	"context"
	"time"

	paymentpb "github.com/guulzadaa/AP2_generated/paymentpb"
	"google.golang.org/grpc"
)

type PaymentGRPCClient struct {
	client paymentpb.PaymentServiceClient
}

func NewPaymentGRPCClient(conn *grpc.ClientConn) *PaymentGRPCClient {
	return &PaymentGRPCClient{
		client: paymentpb.NewPaymentServiceClient(conn),
	}
}

func (p *PaymentGRPCClient) CreatePayment(orderID string, amount int64) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := p.client.ProcessPayment(ctx, &paymentpb.PaymentRequest{
		OrderId: orderID,
		Amount:  amount,
	})
	if err != nil {
		return "", "", err
	}

	return resp.TransactionId, resp.Status, nil
}
