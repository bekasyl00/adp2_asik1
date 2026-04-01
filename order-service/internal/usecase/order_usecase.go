package usecase

import (
	"context"

	"order-service/internal/domain"

	"github.com/google/uuid"
)

// OrderUseCase contains the business logic for order operations.
// It depends on interfaces (ports), not concrete implementations.
type OrderUseCase struct {
	repo          domain.OrderRepository
	paymentClient domain.PaymentClient
}

// NewOrderUseCase creates a new OrderUseCase with injected dependencies.
func NewOrderUseCase(repo domain.OrderRepository, paymentClient domain.PaymentClient) *OrderUseCase {
	return &OrderUseCase{
		repo:          repo,
		paymentClient: paymentClient,
	}
}

// CreateOrder handles the order creation flow:
// 1. Validates and creates order with "Pending" status
// 2. Calls Payment Service for authorization
// 3. Updates order status based on payment response
func (uc *OrderUseCase) CreateOrder(ctx context.Context, customerID, itemName string, amount int64) (*domain.Order, error) {
	// Create order with domain validation
	orderID := uuid.New().String()
	order, err := domain.NewOrder(orderID, customerID, itemName, amount)
	if err != nil {
		return nil, err
	}

	// Persist order with "Pending" status
	if err := uc.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Call Payment Service for authorization
	paymentResp, err := uc.paymentClient.AuthorizePayment(ctx, order.ID, order.Amount)
	if err != nil {
		// Payment service unavailable - mark order as Failed
		order.MarkFailed()
		_ = uc.repo.Update(ctx, order)
		return order, domain.ErrPaymentUnavailable
	}

	// Update order status based on payment response
	if paymentResp.Status == "Authorized" {
		order.MarkPaid()
	} else {
		order.MarkFailed()
	}

	if err := uc.repo.Update(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

// GetOrder retrieves an order by ID.
func (uc *OrderUseCase) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	return uc.repo.GetByID(ctx, id)
}

// CancelOrder cancels an order following business rules:
// - Only "Pending" orders can be cancelled
// - "Paid" orders cannot be cancelled
func (uc *OrderUseCase) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Domain model enforces the business rule
	if err := order.Cancel(); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}
