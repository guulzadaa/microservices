package provider

import "context"

type Notification struct {
	EventID       string
	OrderID       string
	Amount        float64
	CustomerEmail string
	Status        string
}

type NotificationProvider interface {
	Send(ctx context.Context, notification Notification) error
}
