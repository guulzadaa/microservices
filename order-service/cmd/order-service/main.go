package main

import (
	"database/sql"
	"log"
	"net/http"
	"order-service/internal/repository"
	httpdelivery "order-service/internal/transport/http"
	"order-service/internal/usecase"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func initDB() *sql.DB {
	connStr := "host=localhost port=5432 user=postgres password=123123 dbname=order_db sslmode=disable"

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

	httpClient := &http.Client{
		Timeout: 2 * time.Second,
	}

	orderRepo := repository.NewOrderRepository(db)
	paymentClient := repository.NewPaymentHTTPClient("http://localhost:8081", httpClient)
	orderUC := usecase.NewOrderUseCase(orderRepo, paymentClient)
	orderHandler := httpdelivery.NewOrderHandler(orderUC)

	router := gin.Default()

	router.POST("/orders", orderHandler.CreateOrder)
	router.GET("/orders/:id", orderHandler.GetOrder)
	router.PATCH("/orders/:id/cancel", orderHandler.CancelOrder)

	log.Println("Order Service running on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
