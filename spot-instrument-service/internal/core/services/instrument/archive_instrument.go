package instrument

import (
	"context"
	"time"

	"spot-instrument-service/internal/core/domain"
	"spot-instrument-service/internal/core/ports"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ArchiveInstrumentCase struct {
	repo ports.InstrumentRepository
}

func NewArchiveInstrumentCase(repo ports.InstrumentRepository) *ArchiveInstrumentCase {
	return &ArchiveInstrumentCase{repo: repo}
}

func (uc *ArchiveInstrumentCase) Execute(ctx context.Context, id string) (*InstrumentResult, error) {
	inst, err := uc.repo.UpdateStatus(ctx, id, domain.InstrumentStatusDelisted, time.Now())
	if err != nil {
		return nil, err
	}

	trace.SpanFromContext(ctx).SetAttributes(
		attribute.String("instrument.id", inst.ID.String()),
		attribute.String("instrument.symbol", inst.Symbol),
	)

	return toResult(inst), nil
}
