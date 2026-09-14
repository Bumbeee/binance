package order

import (
	"context"

	"github.com/google/uuid"

	"order-service/internal/core/domain"
	"order-service/internal/core/ports"
)

type CancelOrderCase struct {
	repo ports.OrderRepository
}

func NewCancelOrderCase(repo ports.OrderRepository) *CancelOrderCase {
	return &CancelOrderCase{repo: repo}
}

func (uc *CancelOrderCase) Execute(ctx context.Context, callerUserID, orderID string) (*OrderResult, error) {
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

	if err := o.Cancel(); err != nil {
		return nil, err
	}

	if err := uc.repo.UpdateStatus(ctx, parsedOrderID, o.Status); err != nil {
		return nil, err
	}

	return toResult(o), nil
}
