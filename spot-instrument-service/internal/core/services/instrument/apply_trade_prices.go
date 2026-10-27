package instrument

import (
	"context"
	"spot-instrument-service/internal/core/ports"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const maxTracedUpdates = 10 // сколько сообщений должны попасть в спан?

type ApplyTradePricesCase struct {
	repo ports.InstrumentRepository
}

func NewApplyTradePricesCase(repo ports.InstrumentRepository) *ApplyTradePricesCase {
	return &ApplyTradePricesCase{repo: repo}
}

func (uc *ApplyTradePricesCase) Execute(ctx context.Context, updates []ports.PriceUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int("rates.updates", len(updates)))

	n := min(len(updates), maxTracedUpdates) // сколько сообщений должны попасть в спан?
	for _, u := range updates[:n] {
		span.AddEvent("rate_update", trace.WithAttributes(
			attribute.String("instrument.id", u.InstrumentID),
			attribute.String("price", u.Price),
		))
	}
	if len(updates) > n {
		span.SetAttributes(attribute.Int("rates.updates_omitted", len(updates)-n))
	}

	return uc.repo.UpdateRates(ctx, updates, time.Now())
}
