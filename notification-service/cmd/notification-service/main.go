package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"notification-service/internal/idempotency"
	"notification-service/internal/provider"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentCompletedEvent struct {
	EventID       string  `json:"event_id"`
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	CustomerEmail string  `json:"customer_email"`
	Status        string  `json:"status"`
}

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://admin:admin123@localhost:5672/"
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open channel:", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		"payment.completed",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to declare queue:", err)
	}

	msgs, err := ch.Consume(
		"payment.completed",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to register consumer:", err)
	}

	redisStore := idempotency.NewRedisStore()

	var notificationProvider provider.NotificationProvider

	providerMode := os.Getenv("PROVIDER_MODE")

	switch providerMode {
	case "SIMULATED":
		notificationProvider = provider.NewSimulatedProvider()
	default:
		notificationProvider = provider.NewSimulatedProvider()
	}

	log.Println("Notification Service started. Waiting for messages...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		for msg := range msgs {
			var event PaymentCompletedEvent

			err := json.Unmarshal(msg.Body, &event)
			if err != nil {
				log.Println("Invalid message:", err)
				msg.Nack(false, false)
				continue
			}

			ctx := context.Background()

			processed, err := redisStore.IsProcessed(ctx, event.EventID)
			if err != nil {
				log.Println("Redis idempotency check failed:", err)
				msg.Nack(false, true)
				continue
			}

			if processed {
				log.Println("Duplicate event ignored:", event.EventID)
				msg.Ack(false)
				continue
			}

			notification := provider.Notification{
				EventID:       event.EventID,
				OrderID:       event.OrderID,
				Amount:        event.Amount,
				CustomerEmail: event.CustomerEmail,
				Status:        event.Status,
			}

			maxRetries := getEnvAsInt("NOTIFICATION_MAX_RETRIES", 3)
			baseBackoff := getEnvAsInt("NOTIFICATION_BACKOFF_SECONDS", 2)

			success := false

			for attempt := 0; attempt <= maxRetries; attempt++ {
				err = notificationProvider.Send(ctx, notification)

				if err == nil {
					success = true
					break
				}

				log.Printf(
					"Notification send failed. Attempt %d/%d. Error: %v",
					attempt+1,
					maxRetries+1,
					err,
				)

				backoff := time.Duration(baseBackoff*(1<<attempt)) * time.Second

				log.Printf("Retrying in %v...", backoff)

				time.Sleep(backoff)
			}

			if !success {
				log.Println("Notification permanently failed:", event.EventID)

				msg.Nack(false, false)
				continue
			}

			err = redisStore.MarkProcessed(ctx, event.EventID)
			if err != nil {
				log.Println("Failed to save idempotency key:", err)
				msg.Nack(false, true)
				continue
			}

			msg.Ack(false)
		}
	}()

	<-stop

	log.Println("Notification Service shutting down...")
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)

	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
