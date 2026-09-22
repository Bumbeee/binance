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

type ListOrdersResult struct {
	Orders        []*OrderResult
	NextPageToken string
}

func (uc *ListOrdersCase) Execute(
	ctx context.Context,
	callerUserID string,
	statusFilter *domain.OrderStatus,
	pageSize int32,
	pageToken string,
) (*ListOrdersResult, error) {
	parsedCallerID, err := uuid.Parse(callerUserID)
	if err != nil {
		return nil, err
	}

	orders, nextPageToken, err := uc.repo.ListByUserID(ctx, parsedCallerID, statusFilter, pageSize, pageToken)
	if err != nil {
		return nil, err
	}

	results := make([]*OrderResult, 0, len(orders))
	for _, o := range orders {
		results = append(results, toResult(o))
	}

	return &ListOrdersResult{Orders: results, NextPageToken: nextPageToken}, nil
}
