package order

import (
	"context"

	"github.com/google/uuid"

	"order-service/internal/core/domain"
	"order-service/internal/core/ports"
)

type GetOrderCase struct {
	repo ports.OrderRepository
}

func NewGetOrderCase(repo ports.OrderRepository) *GetOrderCase {
	return &GetOrderCase{repo: repo}
}

func (uc *GetOrderCase) Execute(ctx context.Context, callerUserID, orderID string) (*OrderResult, error) {
	parsedCallerID, err := uuid.Parse(callerUserID)
	if err != nil {
		return nil, err
	}

	parsedOrderID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, err
	}

	o, err := uc.repo.GetByID(ctx, parsedOrderID)
	if err != nil {
		return nil, err
	}

	if o.UserID != parsedCallerID {
		return nil, domain.ErrNotOrderOwner
	}

	return toResult(o), nil
}
