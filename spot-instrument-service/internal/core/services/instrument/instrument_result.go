package instrument

import (
	"spot-instrument-service/internal/core/domain"
	"time"
)

type InstrumentResult struct {
	ID                string
	Symbol            string
	BaseAsset         string
	QuoteAsset        string
	PricePrecision    int32
	QuantityPrecision int32
	MinOrderSize      string
	CurrentRate       string
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func toResult(i *domain.Instrument) *InstrumentResult {
	return &InstrumentResult{
		ID:                i.ID.String(),
		Symbol:            i.Symbol,
		BaseAsset:         i.BaseAsset,
		QuoteAsset:        i.QuoteAsset,
		PricePrecision:    i.PricePrecision,
		QuantityPrecision: i.QuantityPrecision,
		MinOrderSize:      i.MinOrderSize,
		CurrentRate:       i.CurrentRate,
		Status:            string(i.Status),
		CreatedAt:         i.CreatedAt,
		UpdatedAt:         i.UpdatedAt,
	}
}
