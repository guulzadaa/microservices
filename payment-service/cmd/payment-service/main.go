package main

import (
	"database/sql"
	"log"
	"net"
	"os"

	"payment-service/internal/repository"
	grpcTransport "payment-service/internal/transport/grpc"
	httpdelivery "payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
	paymentpb "github.com/guulzadaa/AP2_generated/paymentpb"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
)

func initDB() *sql.DB {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "host=localhost port=5432 user=postgres password=123123 dbname=payment_db sslmode=disable"
	}

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

	paymentRepo := repository.NewPaymentRepository(db)
	paymentUC := usecase.NewPaymentUseCase(paymentRepo)
	paymentHandler := httpdelivery.NewPaymentHandler(paymentUC)
	paymentServer := grpcTransport.NewPaymentServer(paymentUC)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("failed to listen for gRPC:", err)
	}

	grpcServer := grpc.NewServer()
	paymentpb.RegisterPaymentServiceServer(grpcServer, paymentServer)

	go func() {
		log.Println("Payment gRPC Service running on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("failed to serve gRPC:", err)
		}
	}()

	router := gin.Default()
	router.POST("/payments", paymentHandler.CreatePayment)
	router.GET("/payments/:order_id", paymentHandler.GetPayment)

	log.Println("Payment REST Service running on :8081")
	if err := router.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
