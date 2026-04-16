package main

import (
	"context"
	"io"
	"log"
	"os"

	orderpb "github.com/guulzadaa/AP2_generated/orderpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("please provide order_id: go run ./cmd/stream-client <order_id>")
	}

	orderID := os.Args[1]

	conn, err := grpc.Dial(
		"localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("failed to connect to order gRPC service:", err)
	}
	defer conn.Close()

	client := orderpb.NewOrderServiceClient(conn)

	stream, err := client.SubscribeToOrderUpdates(context.Background(), &orderpb.OrderRequest{
		OrderId: orderID,
	})
	if err != nil {
		log.Fatal("failed to subscribe to order updates:", err)
	}

	log.Println("subscribed to order updates for order:", orderID)

	for {
		update, err := stream.Recv()
		if err == io.EOF {
			log.Println("stream closed")
			return
		}
		if err != nil {
			log.Fatal("error receiving stream update:", err)
		}

		log.Printf("order update received: order_id=%s status=%s\n", update.OrderId, update.Status)
	}
}
