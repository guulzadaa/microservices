package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"
	orderpb "github.com/guulzadaa/AP2_generated/orderpb"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"order-service/internal/repository"
	grpcTransport "order-service/internal/transport/grpc"
	httpdelivery "order-service/internal/transport/http"
	"order-service/internal/usecase"
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func initDB() *sql.DB {
	connStr := getEnv("DATABASE_URL", "host=localhost port=5432 user=postgres password=123123 dbname=order_db sslmode=disable")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("failed to open db:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("failed to connect db:", err)
	}

	return db
}

func main() {
	db := initDB()
	defer db.Close()

	paymentGRPCAddr := getEnv("PAYMENT_GRPC_ADDR", "localhost:50051")

	conn, err := grpc.Dial(
		paymentGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal("failed to connect to payment gRPC service:", err)
	}
	defer conn.Close()

	orderRepo := repository.NewOrderRepository(db)
	paymentClient := repository.NewPaymentGRPCClient(conn)
	orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient)

	streamServer := grpcTransport.NewOrderStreamServer(orderUC)
	orderHandler := httpdelivery.NewOrderHandler(orderUC, streamServer)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal("failed to listen for order gRPC service:", err)
	}

	grpcServer := grpc.NewServer()
	orderpb.RegisterOrderServiceServer(grpcServer, streamServer)

	go func() {
		log.Println("Order gRPC Streaming Service running on :50052")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("failed to serve order gRPC:", err)
		}
	}()

	router := gin.Default()

	router.POST("/orders", orderHandler.CreateOrder)
	router.GET("/orders/:id", orderHandler.GetOrder)
	router.PATCH("/orders/:id/cancel", orderHandler.CancelOrder)
	router.GET("/payments/stats", orderHandler.GetPaymentStats)

	log.Println("Order REST Service running on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
