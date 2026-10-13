package order

import (
	"context"
	"errors"
	"fmt"
	"market/shared/tracing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"order-service/internal/core/domain"
	"order-service/internal/core/ports"
)

type MatchingEngine struct{}

func NewMatchingEngine() *MatchingEngine {
	return &MatchingEngine{}
}

func (m *MatchingEngine) Run(ctx context.Context, store ports.MatchingStore, order *domain.Order) (res *ports.MatchResult, err error) {
	ctx, span := tracing.Start(ctx, "matching.Run",
		attribute.String("order.id", order.ID.String()),
		attribute.String("instrument.id", order.InstrumentID.String()),
		attribute.String("order.side", string(order.Side)),
		attribute.String("order.type", string(order.Type)),
		attribute.String("order.quantity", order.Quantity),
		attribute.String("order.price", order.Price),
	)
	defer span.End(&err)

	if err := store.InsertOrder(ctx, order); err != nil {
		if errors.Is(err, ports.ErrDuplicateIdempotencyKey) {
			existing, findErr := store.FindByIdempotencyKey(ctx, order.UserID, order.IdempotencyKey)
			if findErr != nil {
				return nil, fmt.Errorf("resolve concurrent idempotent insert: %w", findErr)
			}
			span.SetAttributes(attribute.Bool("matching.idempotent_replay", true))
			return &ports.MatchResult{Order: existing}, nil
		}
		return nil, fmt.Errorf("insert order: %w", err)
	}

	remaining, err := decimal.NewFromString(order.RemainingQuantity)
	if err != nil {
		return nil, fmt.Errorf("parse order quantity: %w", err)
	}
	originalQty, err := decimal.NewFromString(order.Quantity)
	if err != nil {
		return nil, fmt.Errorf("parse order quantity: %w", err)
	}

	var (
		trades  []*domain.Trade
		peers   []*domain.Order
		zero    = decimal.NewFromInt(0)
		orderPx decimal.Decimal
	)
	if order.Type == domain.OrderTypeLimit {
		orderPx, err = decimal.NewFromString(order.Price)
		if err != nil {
			return nil, fmt.Errorf("parse order price: %w", err)
		}
	}

	for remaining.GreaterThan(zero) {
		peer, err := store.LockBestMatchingOrder(ctx, order.InstrumentID, order.Side, order.Type, orderPx)
		if err != nil {
			return nil, fmt.Errorf("lock matching order: %w", err)
		}
		if peer == nil {
			break
		}

		peerRemaining, err := decimal.NewFromString(peer.RemainingQuantity)
		if err != nil {
			return nil, fmt.Errorf("parse peer quantity: %w", err)
		}

		fillQty := domain.ComputeFill(remaining, peerRemaining)

		trade := buildTrade(order, peer, fillQty)
		if err := store.InsertTrade(ctx, trade); err != nil {
			return nil, fmt.Errorf("insert trade: %w", err)
		}
		trades = append(trades, trade)
		span.AddEvent("trade", trace.WithAttributes(
			attribute.String("trade.id", trade.ID.String()),
			attribute.String("trade.price", trade.Price),
			attribute.String("trade.quantity", trade.Quantity),
			attribute.String("peer.order_id", peer.ID.String()),
		))

		remaining = remaining.Sub(fillQty)
		peerRemaining = peerRemaining.Sub(fillQty)

		peerStatus := domain.ResolveFillStatus(peerRemaining)
		if err := store.UpdateOrderFill(ctx, peer.ID, peerRemaining.String(), peerStatus); err != nil {
			return nil, fmt.Errorf("update peer order: %w", err)
		}
		peer.RemainingQuantity = peerRemaining.String()
		peer.Status = peerStatus
		peers = append(peers, peer)
	}

	order.RemainingQuantity = remaining.String()
	order.Status = domain.ResolveFinalStatus(remaining, originalQty, order.Type)
	order.UpdatedAt = time.Now()

	if err := store.UpdateOrderFill(ctx, order.ID, order.RemainingQuantity, order.Status); err != nil {
		return nil, fmt.Errorf("update new order: %w", err)
	}

	span.SetAttributes(
		attribute.Int("matching.trades", len(trades)),
		attribute.Int("matching.peers_affected", len(peers)),
		attribute.String("matching.remaining", order.RemainingQuantity),
		attribute.String("order.final_status", string(order.Status)),
	)

	return &ports.MatchResult{Order: order, Trades: trades, AffectedPeers: peers}, nil
}

func buildTrade(order, peer *domain.Order, fillQty decimal.Decimal) *domain.Trade {
	trade := &domain.Trade{
		ID:           uuid.New(),
		InstrumentID: order.InstrumentID,
		Price:        peer.Price,
		Quantity:     fillQty.String(),
		CreatedAt:    time.Now(),
	}
	if order.Side == domain.OrderSideBuy {
		trade.BuyOrderID = order.ID
		trade.SellOrderID = peer.ID
	} else {
		trade.BuyOrderID = peer.ID
		trade.SellOrderID = order.ID
	}
	return trade
}
