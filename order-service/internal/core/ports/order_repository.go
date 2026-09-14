package ports

import (
	"context"

	"github.com/google/uuid"

	"order-service/internal/core/domain"
)

type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, statusFilter *domain.OrderStatus) ([]*domain.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
}
