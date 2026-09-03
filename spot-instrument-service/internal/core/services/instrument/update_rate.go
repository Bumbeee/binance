package instrument

import (
	"context"

	"spot-instrument-service/internal/core/ports"
)

type UpdateRateCase struct {
	repo ports.InstrumentRepository
}

func NewUpdateRateCase(repo ports.InstrumentRepository) *UpdateRateCase {
	return &UpdateRateCase{repo: repo}
}

func (uc *UpdateRateCase) Execute(ctx context.Context, id, rate string) (*InstrumentResult, error) {
	if err := uc.repo.UpdateRate(ctx, id, rate); err != nil {
		return nil, err
	}

	inst, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return toResult(inst), nil
}
