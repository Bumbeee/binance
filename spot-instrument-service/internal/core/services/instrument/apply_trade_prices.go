package instrument

import (
	"context"
	"spot-instrument-service/internal/core/ports"
	"time"
)

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
	return uc.repo.UpdateRates(ctx, updates, time.Now())
}
