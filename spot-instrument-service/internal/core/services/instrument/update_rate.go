package instrument

import (
	"context"
	"time"

	"spot-instrument-service/internal/core/ports"
)

type UpdateRateCase struct {
	repo ports.InstrumentRepository
}

func NewUpdateRateCase(repo ports.InstrumentRepository) *UpdateRateCase {
	return &UpdateRateCase{repo: repo}
}

func (uc *UpdateRateCase) Execute(ctx context.Context, id, rate string) (*InstrumentResult, error) {
	inst, err := uc.repo.UpdateRate(ctx, id, rate, time.Now())
	if err != nil {
		return nil, err
	}

	return toResult(inst), nil
}
