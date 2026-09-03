package instrument

import (
	"context"

	"spot-instrument-service/internal/core/domain"
	"spot-instrument-service/internal/core/ports"
)

type CreateInstrumentCase struct {
	repo ports.InstrumentRepository
}

func NewCreateInstrumentCase(repo ports.InstrumentRepository) *CreateInstrumentCase {
	return &CreateInstrumentCase{repo: repo}
}

func (uc *CreateInstrumentCase) Execute(
	ctx context.Context,
	symbol, baseAsset, quoteAsset string,
	pricePrecision, quantityPrecision int32,
	minOrderSize string,
) (*InstrumentResult, error) {
	inst, err := domain.NewInstrument(symbol, baseAsset, quoteAsset, pricePrecision, quantityPrecision, minOrderSize)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, inst); err != nil {
		return nil, err
	}

	return toResult(inst), nil
}
