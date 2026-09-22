package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"order-service/internal/core/domain"
	"order-service/internal/core/ports"
)

type CreateOrderCase struct {
	repo ports.OrderRepository
}

func NewCreateOrderCase(repo ports.OrderRepository) *CreateOrderCase {
	return &CreateOrderCase{repo: repo}
}

func (uc *CreateOrderCase) Execute(
	ctx context.Context,
	userID string,
	instrumentID string,
	side, orderType, price, quantity string,
) (*OrderResult, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("create_order: invalid user id in context: %w", err)
	}

	parsedInstrumentID, err := uuid.Parse(instrumentID)
	if err != nil {
		return nil, err
	}

	o, err := domain.NewOrder(
		parsedUserID,
		parsedInstrumentID,
		domain.OrderSide(side),
		domain.OrderType(orderType),
		price,
		quantity,
	)
	if err != nil {
		return nil, err
	}

	result, err := uc.repo.SaveAndMatch(ctx, o)
	if err != nil {
		return nil, err
	}

	return toResult(result.Order), nil
}
