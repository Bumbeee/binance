-- +goose Up
CREATE TABLE instruments (
    id                  UUID PRIMARY KEY,
    symbol              VARCHAR(20) NOT NULL UNIQUE,
    base_asset          VARCHAR(10) NOT NULL,
    quote_asset         VARCHAR(10) NOT NULL,
    price_precision     INT NOT NULL,
    quantity_precision  INT NOT NULL,
    min_order_size      NUMERIC NOT NULL,
    current_rate        NUMERIC NOT NULL DEFAULT 0,
    status              VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_instruments_status ON instruments (status);

-- +goose Down
DROP TABLE instruments;