package ports

import (
	"context"

	"github.com/google/uuid"

	"order-service/internal/core/domain"
)

type MatchResult struct {
	Order         *domain.Order
	Trades        []*domain.Trade
	AffectedPeers []*domain.Order
}

type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error

	ListByUserID(
		ctx context.Context,
		userID uuid.UUID,
		statusFilter *domain.OrderStatus,
		pageSize int32,
		pageToken string,
	) (orders []*domain.Order, nextPageToken string, err error)

	SaveAndMatch(ctx context.Context, order *domain.Order) (*MatchResult, error)
}
