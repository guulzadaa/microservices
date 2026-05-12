package provider

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"time"
)

type SimulatedProvider struct{}

func NewSimulatedProvider() *SimulatedProvider {
	return &SimulatedProvider{}
}

func (p *SimulatedProvider) Send(ctx context.Context, notification Notification) error {
	time.Sleep(1 * time.Second)

	if rand.Intn(100) < 30 {
		return errors.New("simulated provider temporary failure")
	}

	log.Printf("[Notification] Sent email to %s for Order #%s. Amount: $%.2f",
		notification.CustomerEmail,
		notification.OrderID,
		notification.Amount,
	)

	return nil
}
