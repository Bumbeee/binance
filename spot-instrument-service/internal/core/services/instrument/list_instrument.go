package instrument

import (
	"context"

	"spot-instrument-service/internal/core/domain"
	"spot-instrument-service/internal/core/ports"
)

type ListInstrumentsCase struct {
	repo ports.InstrumentRepository
}

func NewListInstrumentsCase(repo ports.InstrumentRepository) *ListInstrumentsCase {
	return &ListInstrumentsCase{repo: repo}
}

func (uc *ListInstrumentsCase) Execute(ctx context.Context, statusFilter *domain.InstrumentStatus) ([]*InstrumentResult, error) {
	instruments, err := uc.repo.List(ctx, statusFilter)
	if err != nil {
		return nil, err
	}

	results := make([]*InstrumentResult, 0, len(instruments))
	for _, inst := range instruments {
		results = append(results, toResult(inst))
	}

	return results, nil
}
