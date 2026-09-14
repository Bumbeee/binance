package order

import (
	"context"

	"github.com/google/uuid"

	"order-service/internal/core/domain"
	"order-service/internal/core/ports"
)

type ListOrdersCase struct {
	repo ports.OrderRepository
}

func NewListOrdersCase(repo ports.OrderRepository) *ListOrdersCase {
	return &ListOrdersCase{repo: repo}
}

func (uc *ListOrdersCase) Execute(ctx context.Context, callerUserID string, statusFilter *domain.OrderStatus) ([]*OrderResult, error) {
	parsedCallerID, err := uuid.Parse(callerUserID)
	if err != nil {
		return nil, err
	}

	orders, err := uc.repo.ListByUserID(ctx, parsedCallerID, statusFilter)
	if err != nil {
		return nil, err
	}

	results := make([]*OrderResult, 0, len(orders))
	for _, o := range orders {
		results = append(results, toResult(o))
	}

	return results, nil
}
