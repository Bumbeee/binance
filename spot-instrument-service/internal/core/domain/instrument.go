package domain

import (
	"time"

	"github.com/google/uuid"
)

type InstrumentStatus string

const (
	InstrumentStatusActive   InstrumentStatus = "active"
	InstrumentStatusPaused   InstrumentStatus = "paused"
	InstrumentStatusDelisted InstrumentStatus = "delisted"
)

func (s InstrumentStatus) IsValid() bool {
	switch s {
	case InstrumentStatusActive, InstrumentStatusPaused, InstrumentStatusDelisted:
		return true
	}
	return false
}

type Instrument struct {
	ID                uuid.UUID
	Symbol            string
	BaseAsset         string
	QuoteAsset        string
	PricePrecision    int32
	QuantityPrecision int32
	MinOrderSize      string
	CurrentRate       string
	Status            InstrumentStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NewInstrument(symbol, baseAsset, quoteAsset string, pricePrecision, quantityPrecision int32, minOrderSize string) (*Instrument, error) {
	if symbol == "" {
		return nil, ErrEmptySymbol
	}
	if baseAsset == "" {
		return nil, ErrEmptyBaseAsset
	}
	if quoteAsset == "" {
		return nil, ErrEmptyQuoteAsset
	}
	if pricePrecision < 0 {
		return nil, ErrInvalidPricePrecision
	}
	if quantityPrecision < 0 {
		return nil, ErrInvalidQuantityPrecision
	}
	if minOrderSize == "" {
		return nil, ErrEmptyMinOrderSize
	}

	now := time.Now()

	return &Instrument{
		ID:                uuid.New(),
		Symbol:            symbol,
		BaseAsset:         baseAsset,
		QuoteAsset:        quoteAsset,
		PricePrecision:    pricePrecision,
		QuantityPrecision: quantityPrecision,
		MinOrderSize:      minOrderSize,
		CurrentRate:       "0",
		Status:            InstrumentStatusActive,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

func (i *Instrument) UpdateRate(rate string) error {
	if rate == "" {
		return ErrEmptyRate
	}
	i.CurrentRate = rate
	i.UpdatedAt = time.Now()
	return nil
}

func (i *Instrument) Archive() {
	i.Status = InstrumentStatusDelisted
	i.UpdatedAt = time.Now()
}
