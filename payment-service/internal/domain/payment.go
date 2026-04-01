package domain

import "errors"

// Payment statuses
const (
	StatusAuthorized = "Authorized"
	StatusDeclined   = "Declined"
)

// Payment limits
const (
	MaxPaymentAmount int64 = 100000 // Maximum payment amount in cents ($1000.00)
)

// Payment represents the payment entity in the domain layer.
// It has no dependency on HTTP, JSON transport concerns, or framework-specific logic.
type Payment struct {
	ID            string
	OrderID       string
	TransactionID string
	Amount        int64  // Amount in cents
	Status        string // "Authorized" or "Declined"
}

// Domain errors
var (
	ErrPaymentNotFound  = errors.New("payment not found")
	ErrInvalidAmount    = errors.New("amount must be greater than 0")
	ErrAmountExceeded   = errors.New("payment amount exceeds maximum limit")
)

// NewPayment creates a new Payment with business rule validation.
// If amount > 100000 (1000 units), the payment is Declined.
func NewPayment(id, orderID, transactionID string, amount int64) (*Payment, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	status := StatusAuthorized
	if amount > MaxPaymentAmount {
		status = StatusDeclined
	}

	return &Payment{
		ID:            id,
		OrderID:       orderID,
		TransactionID: transactionID,
		Amount:        amount,
		Status:        status,
	}, nil
}
