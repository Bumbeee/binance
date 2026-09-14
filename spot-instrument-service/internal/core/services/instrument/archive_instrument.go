package instrument

import (
	"context"

	"spot-instrument-service/internal/core/domain"
	"spot-instrument-service/internal/core/ports"
)

type ArchiveInstrumentCase struct {
	repo ports.InstrumentRepository
}

func NewArchiveInstrumentCase(repo ports.InstrumentRepository) *ArchiveInstrumentCase {
	return &ArchiveInstrumentCase{repo: repo}
}

func (uc *ArchiveInstrumentCase) Execute(ctx context.Context, id string) (*InstrumentResult, error) {
	inst, err := uc.repo.UpdateStatus(ctx, id, domain.InstrumentStatusDelisted)
	if err != nil {
		return nil, err
	}

	return toResult(inst), nil
}
