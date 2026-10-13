package order

import (
	"context"
	"fmt"
	"market/shared/infra/logger"
	"market/shared/tracing"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"order-service/internal/core/domain"
	"order-service/internal/core/ports"
)

type CreateOrderCase struct {
	repo      ports.OrderRepository
	engine    *MatchingEngine
	publisher ports.TradePublisher
	log       *zap.Logger
}

func NewCreateOrderCase(repo ports.OrderRepository, engine *MatchingEngine,
	publisher ports.TradePublisher, log *zap.Logger) *CreateOrderCase {
	return &CreateOrderCase{repo: repo, engine: engine, publisher: publisher, log: log}
}

func (uc *CreateOrderCase) Execute(
	ctx context.Context,
	userID string,
	instrumentID string,
	side, orderType, price, quantity, idempotencyKey string,
) (res *OrderResult, err error) {
	ctx, span := tracing.Start(ctx, "order.Create",
		attribute.String("instrument.id", instrumentID),
		attribute.String("order.side", side),
		attribute.String("order.type", orderType),
		attribute.Bool("order.idempotency_key_provided", idempotencyKey != ""),
	)

	defer span.End(&err,
		domain.ErrInvalidSide,
		domain.ErrInvalidType,
		domain.ErrEmptyQuantity,
		domain.ErrPriceRequiredForLimit,
		domain.ErrPriceNotAllowedForMarket,
	)

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("create_order: invalid user id in context: %w", err)
	}

	parsedInstrumentID, err := uuid.Parse(instrumentID) // TODO: sync with instrument service to check if instrument is valid and active (kafka) ?????
	if err != nil {
		return nil, err
	}

	if idempotencyKey != "" {
		if existing, err := uc.repo.FindByIdempotencyKey(ctx, parsedUserID, idempotencyKey); err != nil {
			return nil, err
		} else if existing != nil {
			span.SetAttributes(
				attribute.Bool("order.idempotent_replay", true),
				attribute.String("order.id", existing.ID.String()),
				attribute.String("order.status", string(existing.Status)),
			)
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
	span.SetAttributes(attribute.String("order.id", o.ID.String()))

	var result *ports.MatchResult
	err = uc.repo.RunMatchingTx(ctx, func(store ports.MatchingStore) error {
		var runErr error
		result, runErr = uc.engine.Run(ctx, store, o)
		return runErr
	})
	if err != nil {
		return nil, err
	}

	span.SetAttributes(
		attribute.String("order.status", string(result.Order.Status)),
		attribute.Int("order.trades", len(result.Trades)),
	)

	if len(result.Trades) > 0 {
		pubCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
		defer cancel()

		if perr := uc.publisher.PublishTrades(pubCtx, result.Trades); perr != nil {
			// TODO: outbox??
			span.RecordError(perr)
			span.SetAttributes(attribute.Bool("order.trades_publish_failed", true))
			logger.WithTrace(ctx, uc.log).Error("publish trades failed", zap.Error(perr))
		}
	}

	return toResult(result.Order), nil
}
