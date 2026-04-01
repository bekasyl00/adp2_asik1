-- Payment Service Database Migration
-- Creates the payments table in the payment_db database

CREATE TABLE IF NOT EXISTS payments (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    transaction_id VARCHAR(36) NOT NULL UNIQUE,
    amount BIGINT NOT NULL CHECK (amount > 0),
    status VARCHAR(20) NOT NULL DEFAULT 'Authorized'
);

-- Index for faster lookups by order_id
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);

-- Index for transaction_id lookups
CREATE INDEX IF NOT EXISTS idx_payments_transaction_id ON payments(transaction_id);
