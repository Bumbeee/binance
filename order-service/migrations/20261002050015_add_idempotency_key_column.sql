-- +goose Up
ALTER TABLE orders ADD COLUMN idempotency_key UUID;

CREATE UNIQUE INDEX idx_orders_user_idempotency
    ON orders (user_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_orders_user_idempotency;
ALTER TABLE orders DROP COLUMN idempotency_key;