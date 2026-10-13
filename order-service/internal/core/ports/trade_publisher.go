package ports

import (
	"context"
	"order-service/internal/core/domain"
)

type TradePublisher interface {
	PublishTrades(ctx context.Context, trades []*domain.Trade) error
}
