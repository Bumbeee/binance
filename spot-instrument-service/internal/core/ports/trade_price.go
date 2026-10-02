package ports

import (
	"context"
	"time"
)

type PriceUpdate struct {
	InstrumentID string
	Price        string
	TradedAt     time.Time
}

type LastTradePriceSource interface {
	GetLastTradePrices(ctx context.Context, instrumentIDs []string) ([]PriceUpdate, error)
}
