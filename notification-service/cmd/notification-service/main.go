package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentCompletedEvent struct {
	EventID       string  `json:"event_id"`
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	CustomerEmail string  `json:"customer_email"`
	Status        string  `json:"status"`
}

var (
	processedEvents = make(map[string]bool)
	mu              sync.Mutex
)

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
		"payment.completed.dlq",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to declare DLQ:", err)
	}

	queue, err := ch.QueueDeclare(
		"payment.completed",
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": "payment.completed.dlq",
		},
	)
	if err != nil {
		log.Fatal("Failed to declare queue:", err)
	}

	msgs, err := ch.Consume(
		queue.Name,
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

			mu.Lock()
			if processedEvents[event.EventID] {
				mu.Unlock()
				log.Println("Duplicate event ignored:", event.EventID)
				msg.Ack(false)
				continue
			}
			mu.Unlock()

			if event.OrderID == "00000000-0000-0000-0000-000000000000" {
				retryCount := getRetryCount(msg)

				if retryCount >= 3 {
					log.Println("Message failed 3 times. Sending to DLQ:", event.EventID)
					msg.Nack(false, false)
					continue
				}

				log.Printf("Simulated failure for event %s. Retry attempt: %d", event.EventID, retryCount+1)

				err := republishWithRetry(ch, msg, retryCount+1)
				if err != nil {
					log.Println("Failed to republish message:", err)
					msg.Nack(false, false)
					continue
				}

				msg.Ack(false)
				continue
			}

			log.Printf("[Notification] Sent email to %s for Order #%s. Amount: $%.2f",
				event.CustomerEmail,
				event.OrderID,
				event.Amount,
			)

			mu.Lock()
			processedEvents[event.EventID] = true
			mu.Unlock()

			msg.Ack(false)
		}
	}()

	<-stop
	log.Println("Notification Service shutting down...")
}

func getRetryCount(msg amqp.Delivery) int32 {
	if msg.Headers == nil {
		return 0
	}

	value, ok := msg.Headers["x-retry-count"]
	if !ok {
		return 0
	}

	switch v := value.(type) {
	case int32:
		return v
	case int:
		return int32(v)
	default:
		return 0
	}
}

func republishWithRetry(ch *amqp.Channel, msg amqp.Delivery, retryCount int32) error {
	headers := amqp.Table{}
	for key, value := range msg.Headers {
		headers[key] = value
	}
	headers["x-retry-count"] = retryCount

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return ch.PublishWithContext(
		ctx,
		"",
		"payment.completed",
		false,
		false,
		amqp.Publishing{
			ContentType:  msg.ContentType,
			DeliveryMode: amqp.Persistent,
			Body:         msg.Body,
			Headers:      headers,
		},
	)
}
