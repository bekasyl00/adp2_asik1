package domain

import "context"

// PaymentRepository is a port (interface) for persistence operations.
// The use case layer depends on this interface, not on a concrete implementation.
type PaymentRepository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByOrderID(ctx context.Context, orderID string) (*Payment, error)
}
