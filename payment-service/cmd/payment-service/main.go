package main

import (
	"database/sql"
	"log"
	"payment-service/internal/repository"
	httpdelivery "payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func initDB() *sql.DB {
	connStr := "host=localhost port=5432 user=postgres password=123123 dbname=payment_db sslmode=disable"

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

	router := gin.Default()

	router.POST("/payments", paymentHandler.CreatePayment)
	router.GET("/payments/:order_id", paymentHandler.GetPayment)

	log.Println("Payment Service running on :8081")
	if err := router.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}
