package domain

import (
	"errors"
	"time"
)

// Order statuses
const (
	StatusPending   = "Pending"
	StatusPaid      = "Paid"
	StatusFailed    = "Failed"
	StatusCancelled = "Cancelled"
)

// Order represents the order entity in the domain layer.
// It has no dependency on HTTP, JSON, or any framework-specific logic.
type Order struct {
	ID         string
	CustomerID string
	ItemName   string
	Amount     int64 // Amount in cents (e.g., 1000 = $10.00)
	Status     string
	CreatedAt  time.Time
}

// Domain errors
var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidAmount      = errors.New("amount must be greater than 0")
	ErrCannotCancelPaid   = errors.New("paid orders cannot be cancelled")
	ErrCannotCancelOrder  = errors.New("only pending orders can be cancelled")
	ErrPaymentUnavailable = errors.New("payment service unavailable")
)

// NewOrder creates a new Order with validation.
func NewOrder(id, customerID, itemName string, amount int64) (*Order, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	return &Order{
		ID:         id,
		CustomerID: customerID,
		ItemName:   itemName,
		Amount:     amount,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
	}, nil
}

// Cancel transitions the order to Cancelled status.
// Only "Pending" orders can be cancelled.
func (o *Order) Cancel() error {
	if o.Status == StatusPaid {
		return ErrCannotCancelPaid
	}
	if o.Status != StatusPending {
		return ErrCannotCancelOrder
	}
	o.Status = StatusCancelled
	return nil
}

// MarkPaid transitions the order to Paid status.
func (o *Order) MarkPaid() {
	o.Status = StatusPaid
}

// MarkFailed transitions the order to Failed status.
func (o *Order) MarkFailed() {
	o.Status = StatusFailed
}
