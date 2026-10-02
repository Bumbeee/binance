package ports

import (
	"context"
	"order-service/internal/core/domain"

	"github.com/google/uuid"
)

type TradeRepository interface {
	GetLatestByInstrumentIDs(ctx context.Context, instrumentIDs []uuid.UUID) (map[uuid.UUID]*domain.Trade, error)
}
