package instrument

import (
	"context"

	"spot-instrument-service/internal/core/ports"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type GetInstrumentCase struct {
	repo ports.InstrumentRepository
}

func NewGetInstrumentCase(repo ports.InstrumentRepository) *GetInstrumentCase {
	return &GetInstrumentCase{repo: repo}
}

func (uc *GetInstrumentCase) Execute(ctx context.Context, id string) (*InstrumentResult, error) {
	inst, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	trace.SpanFromContext(ctx).SetAttributes(
		attribute.String("instrument.id", inst.ID.String()),
		attribute.String("instrument.symbol", inst.Symbol),
	)

	return toResult(inst), nil
}
