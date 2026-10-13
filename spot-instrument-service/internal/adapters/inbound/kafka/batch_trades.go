package kafka

import (
	"context"
	"fmt"
	eventsv1 "market/proto/events/v1"
	"market/shared/infra/kafka"
	"spot-instrument-service/internal/core/ports"
	"spot-instrument-service/internal/core/services/instrument"

	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

func NewTradeBatchHandler(uc *instrument.ApplyTradePricesCase) kafka.BatchHandler {
	return func(ctx context.Context, recs []*kgo.Record) error {
		latest := make(map[string]ports.PriceUpdate, len(recs))
		for _, r := range recs {
			var ev eventsv1.TradeExecuted
			if err := proto.Unmarshal(r.Value, &ev); err != nil {
				return kafka.Permanent(fmt.Errorf("unmarshal trade: %w", err))
			}
			at := ev.GetExecutedAt().AsTime()
			if cur, ok := latest[ev.GetInstrumentId()]; ok && !at.After(cur.TradedAt) {
				continue
			}
			latest[ev.GetInstrumentId()] = ports.PriceUpdate{
				InstrumentID: ev.GetInstrumentId(),
				Price:        ev.GetPrice(),
				TradedAt:     at,
			}
		}
		updates := make([]ports.PriceUpdate, 0, len(latest))
		for _, u := range latest {
			updates = append(updates, u)
		}
		return uc.Execute(ctx, updates) // см. примечание
	}
}
