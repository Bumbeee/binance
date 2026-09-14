-- +goose Up
CREATE TABLE orders (
    id            UUID PRIMARY KEY,
    user_id       UUID NOT NULL,
    instrument_id UUID NOT NULL,
    side          VARCHAR(10) NOT NULL,
    type          VARCHAR(10) NOT NULL,
    price         NUMERIC,
    quantity      NUMERIC NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'open',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_orders_user_id_status ON orders (user_id, status);

-- +goose Down
DROP TABLE orders;