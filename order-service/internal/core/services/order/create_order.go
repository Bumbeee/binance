package order

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"order-service/internal/core/domain"
	"order-service/internal/core/ports"
)

type CreateOrderCase struct {
	repo   ports.OrderRepository
	engine *MatchingEngine
}

func NewCreateOrderCase(repo ports.OrderRepository, engine *MatchingEngine) *CreateOrderCase {
	return &CreateOrderCase{repo: repo, engine: engine}
}

func (uc *CreateOrderCase) Execute(
	ctx context.Context,
	userID string,
	instrumentID string,
	side, orderType, price, quantity, idempotencyKey string,
) (*OrderResult, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("create_order: invalid user id in context: %w", err)
	}

	parsedInstrumentID, err := uuid.Parse(instrumentID) // TODO: sync with instrument service to check if instrument is valid and active (kafka) ?????
	if err != nil {
		return nil, err
	}

	if idempotencyKey != "" {
		if existing, err := uc.repo.FindByIdempotencyKey(ctx, parsedUserID, idempotencyKey); err != nil { // мб лучше в кэше хранить ?????
			return nil, err
		} else if existing != nil {
			return toResult(existing), nil
		}
	}

	o, err := domain.NewOrder(
		parsedUserID,
		parsedInstrumentID,
		domain.OrderSide(side),
		domain.OrderType(orderType),
		price,
		quantity,
		idempotencyKey,
	)
	if err != nil {
		return nil, err
	}

	var result *ports.MatchResult
	err = uc.repo.RunMatchingTx(ctx, func(store ports.MatchingStore) error {
		var runErr error
		result, runErr = uc.engine.Run(ctx, store, o)
		return runErr
	})
	if err != nil {
		return nil, err
	}

	return toResult(result.Order), nil
}
