package domain

import "context"

// OrderRepository is a port (interface) for persistence operations.
// The use case layer depends on this interface, not on a concrete implementation.
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id string) (*Order, error)
	Update(ctx context.Context, order *Order) error
}

// PaymentResponse represents the response from the Payment Service.
type PaymentResponse struct {
	OrderID       string
	TransactionID string
	Amount        int64
	Status        string // "Authorized" or "Declined"
}

// PaymentClient is a port (interface) for inter-service communication.
// The use case layer depends on this interface for calling the Payment Service.
type PaymentClient interface {
	AuthorizePayment(ctx context.Context, orderID string, amount int64) (*PaymentResponse, error)
}
