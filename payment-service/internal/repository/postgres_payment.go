package repository

import (
	"context"
	"database/sql"

	"payment-service/internal/domain"
)

// PostgresPaymentRepository implements domain.PaymentRepository using PostgreSQL.
type PostgresPaymentRepository struct {
	db *sql.DB
}

// NewPostgresPaymentRepository creates a new PostgresPaymentRepository.
func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

// Create inserts a new payment into the database.
func (r *PostgresPaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	query := `INSERT INTO payments (id, order_id, transaction_id, amount, status)
			   VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query,
		payment.ID,
		payment.OrderID,
		payment.TransactionID,
		payment.Amount,
		payment.Status,
	)
	return err
}

// GetByOrderID retrieves a payment by its order ID.
func (r *PostgresPaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	query := `SELECT id, order_id, transaction_id, amount, status FROM payments WHERE order_id = $1`

	payment := &domain.Payment{}
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.TransactionID,
		&payment.Amount,
		&payment.Status,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}
	return payment, nil
}
