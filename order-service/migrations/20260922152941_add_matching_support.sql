-- +goose Up
ALTER TABLE orders ADD COLUMN remaining_quantity NUMERIC NOT NULL DEFAULT 0;
UPDATE orders SET remaining_quantity = quantity WHERE remaining_quantity = 0;

CREATE TABLE trades (
    id             UUID PRIMARY KEY,
    instrument_id  UUID NOT NULL,
    buy_order_id   UUID NOT NULL REFERENCES orders(id),
    sell_order_id  UUID NOT NULL REFERENCES orders(id),
    price          NUMERIC NOT NULL,
    quantity       NUMERIC NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_trades_instrument_id ON trades (instrument_id);
CREATE INDEX idx_trades_buy_order_id ON trades (buy_order_id);
CREATE INDEX idx_trades_sell_order_id ON trades (sell_order_id);

CREATE INDEX idx_orders_matching ON orders (instrument_id, side, status, price, created_at) WHERE status IN ('open', 'partially_filled');

-- +goose Down
DROP INDEX IF EXISTS idx_orders_matching;
DROP TABLE trades;
ALTER TABLE orders DROP COLUMN remaining_quantity;