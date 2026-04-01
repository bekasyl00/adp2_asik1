package repository

import (
	"context"
	"database/sql"

	"order-service/internal/domain"
)

// PostgresOrderRepository implements domain.OrderRepository using PostgreSQL.
type PostgresOrderRepository struct {
	db *sql.DB
}

// NewPostgresOrderRepository creates a new PostgresOrderRepository.
func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

// Create inserts a new order into the database.
func (r *PostgresOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	query := `INSERT INTO orders (id, customer_id, item_name, amount, status, created_at)
			   VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query,
		order.ID,
		order.CustomerID,
		order.ItemName,
		order.Amount,
		order.Status,
		order.CreatedAt,
	)
	return err
}

// GetByID retrieves an order by its ID.
func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	query := `SELECT id, customer_id, item_name, amount, status, created_at FROM orders WHERE id = $1`

	order := &domain.Order{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.ItemName,
		&order.Amount,
		&order.Status,
		&order.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return order, nil
}

// Update updates an existing order in the database.
func (r *PostgresOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	query := `UPDATE orders SET customer_id = $1, item_name = $2, amount = $3, status = $4 WHERE id = $5`
	result, err := r.db.ExecContext(ctx, query,
		order.CustomerID,
		order.ItemName,
		order.Amount,
		order.Status,
		order.ID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}
