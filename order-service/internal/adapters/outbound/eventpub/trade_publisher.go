package eventpub

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	eventsv1 "market/proto/events/v1"
	"market/shared/events"
	"market/shared/infra/kafka"

	"order-service/internal/core/domain"
)

type TradePublisher struct {
	producer *kafka.Producer
}

func NewTradePublisher(p *kafka.Producer) *TradePublisher {
	return &TradePublisher{producer: p}
}

func (p *TradePublisher) PublishTrades(ctx context.Context, trades []*domain.Trade) error {
	msgs := make([]kafka.Message, 0, len(trades))
	for _, t := range trades {
		value, err := proto.Marshal(&eventsv1.TradeExecuted{
			TradeId:      t.ID.String(),
			InstrumentId: t.InstrumentID.String(),
			Price:        t.Price,
			Quantity:     t.Quantity,
			ExecutedAt:   timestamppb.New(t.CreatedAt),
		})
		if err != nil {
			return fmt.Errorf("marshal trade %s: %w", t.ID, err)
		}
		msgs = append(msgs, kafka.Message{
			Topic: events.TopicTradesExecuted,
			Key:   t.InstrumentID.String(),
			Value: value,
		})
	}
	return p.producer.PublishBatch(ctx, msgs)
}
