-- Order Service Database Migration
-- Creates the orders table in the order_db database

CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(36) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL,
    item_name VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'Pending',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for faster lookups by customer
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);

-- Index for filtering by status
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
