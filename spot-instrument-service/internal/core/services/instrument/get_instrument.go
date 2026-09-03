package instrument

import (
	"context"

	"spot-instrument-service/internal/core/ports"
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
	return toResult(inst), nil
}
