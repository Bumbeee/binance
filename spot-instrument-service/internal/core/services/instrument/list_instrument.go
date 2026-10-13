package instrument

import (
	"context"

	"spot-instrument-service/internal/core/domain"
	"spot-instrument-service/internal/core/ports"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ListInstrumentsResult struct {
	Instruments   []*InstrumentResult
	NextPageToken string
}

type ListInstrumentsCase struct {
	repo ports.InstrumentRepository
}

func NewListInstrumentsCase(repo ports.InstrumentRepository) *ListInstrumentsCase {
	return &ListInstrumentsCase{repo: repo}
}

func (uc *ListInstrumentsCase) Execute(
	ctx context.Context,
	statusFilter *domain.InstrumentStatus,
	pageSize int32,
	pageToken string,
) (*ListInstrumentsResult, error) {
	instruments, nextPageToken, err := uc.repo.ListPage(ctx, statusFilter, pageSize, pageToken)
	if err != nil {
		return nil, err
	}

	attrs := []attribute.KeyValue{
		attribute.Int("list.page_size", int(pageSize)),
		attribute.Int("list.returned", len(instruments)),
		attribute.Bool("list.has_next", nextPageToken != ""),
	}
	if statusFilter != nil {
		attrs = append(attrs, attribute.String("instrument.status_filter", string(*statusFilter)))
	}
	trace.SpanFromContext(ctx).SetAttributes(attrs...)

	results := make([]*InstrumentResult, 0, len(instruments))
	for _, inst := range instruments {
		results = append(results, toResult(inst))
	}

	return &ListInstrumentsResult{
		Instruments:   results,
		NextPageToken: nextPageToken,
	}, nil
}
