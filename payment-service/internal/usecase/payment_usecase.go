package usecase

import (
	"context"

	"payment-service/internal/domain"

	"github.com/google/uuid"
)

// PaymentUseCase contains the business logic for payment operations.
// It depends on interfaces (ports), not concrete implementations.
type PaymentUseCase struct {
	repo domain.PaymentRepository
}

// NewPaymentUseCase creates a new PaymentUseCase with injected dependencies.
func NewPaymentUseCase(repo domain.PaymentRepository) *PaymentUseCase {
	return &PaymentUseCase{repo: repo}
}

// AuthorizePayment processes a payment authorization:
// - Generates unique IDs for the payment and transaction
// - Applies business rules (amount limits) via the domain model
// - Persists the payment record
func (uc *PaymentUseCase) AuthorizePayment(ctx context.Context, orderID string, amount int64) (*domain.Payment, error) {
	paymentID := uuid.New().String()
	transactionID := uuid.New().String()

	// Domain model handles business rules (e.g., amount > 100000 -> Declined)
	payment, err := domain.NewPayment(paymentID, orderID, transactionID, amount)
	if err != nil {
		return nil, err
	}

	// Persist payment record
	if err := uc.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

// GetPaymentByOrderID retrieves a payment record by order ID.
func (uc *PaymentUseCase) GetPaymentByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	return uc.repo.GetByOrderID(ctx, orderID)
}
