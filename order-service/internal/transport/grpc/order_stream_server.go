package grpc

import (
	"sync"

	"order-service/internal/usecase"

	orderpb "github.com/guulzadaa/AP2_generated/orderpb"
)

type OrderStreamServer struct {
	orderpb.UnimplementedOrderServiceServer

	usecase *usecase.OrderUseCase

	subscribers map[string][]chan string
	mu          sync.Mutex
}

func NewOrderStreamServer(uc *usecase.OrderUseCase) *OrderStreamServer {
	return &OrderStreamServer{
		usecase:     uc,
		subscribers: make(map[string][]chan string),
	}
}

func (s *OrderStreamServer) SubscribeToOrderUpdates(
	req *orderpb.OrderRequest,
	stream orderpb.OrderService_SubscribeToOrderUpdatesServer,
) error {

	orderID := req.OrderId

	order, err := s.usecase.GetOrderByID(orderID)
	if err == nil {

		stream.Send(&orderpb.OrderStatusUpdate{
			OrderId: order.ID,
			Status:  order.Status,
		})
	}

	ch := make(chan string)

	s.mu.Lock()
	s.subscribers[orderID] = append(s.subscribers[orderID], ch)
	s.mu.Unlock()

	for {
		status := <-ch

		err := stream.Send(&orderpb.OrderStatusUpdate{
			OrderId: orderID,
			Status:  status,
		})

		if err != nil {
			return err
		}
	}
}

func (s *OrderStreamServer) Notify(orderID string, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ch := range s.subscribers[orderID] {
		ch <- status
	}
}
