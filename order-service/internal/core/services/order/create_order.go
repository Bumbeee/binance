package order

import (
	"context"
	"fmt"
	"time"

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

type OrderResult struct {
	ID           string
	UserID       string
	InstrumentID string
	Side         string
	Type         string
	Price        string
	Quantity     string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
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

	if err := uc.repo.Save(ctx, o); err != nil {
		return nil, err
	}

	return toResult(o), nil
}
