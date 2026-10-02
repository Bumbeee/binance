package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"order-service/internal/core/domain"
)

type MatchingStore interface {
	InsertOrder(ctx context.Context, order *domain.Order) error
	FindByIdempotencyKey(ctx context.Context, userID uuid.UUID, key string) (*domain.Order, error)

	LockBestMatchingOrder(
		ctx context.Context,
		instrumentID uuid.UUID,
		side domain.OrderSide,
		orderType domain.OrderType,
		limitPrice decimal.Decimal,
	) (*domain.Order, error)

	UpdateOrderFill(ctx context.Context, orderID uuid.UUID, remainingQuantity string, status domain.OrderStatus) error
	InsertTrade(ctx context.Context, trade *domain.Trade) error
}

type MatchResult struct {
	Order         *domain.Order
	Trades        []*domain.Trade
	AffectedPeers []*domain.Order
}

type OrderRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)

	ListByUserID(
		ctx context.Context,
		userID uuid.UUID,
		statusFilter *domain.OrderStatus,
		pageSize int32,
		pageToken string,
	) (orders []*domain.Order, nextPageToken string, err error)

	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
	RunMatchingTx(ctx context.Context, fn func(MatchingStore) error) error
	FindByIdempotencyKey(ctx context.Context, userID uuid.UUID, key string) (*domain.Order, error)
}
