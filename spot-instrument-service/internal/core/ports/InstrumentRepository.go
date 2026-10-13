package ports

import (
	"context"
	"time"

	"spot-instrument-service/internal/core/domain"
)

type PriceUpdate struct {
	InstrumentID string
	Price        string
	TradedAt     time.Time
}

type InstrumentRepository interface {
	Save(ctx context.Context, instrument *domain.Instrument) error
	GetByID(ctx context.Context, id string) (*domain.Instrument, error)
	List(ctx context.Context, statusFilter *domain.InstrumentStatus) ([]*domain.Instrument, error)
	ListPage(ctx context.Context, statusFilter *domain.InstrumentStatus, pageSize int32, pageToken string) ([]*domain.Instrument, string, error)
	UpdateRate(ctx context.Context, id, rate string, now time.Time) (*domain.Instrument, error)
	UpdateRates(ctx context.Context, updates []PriceUpdate, now time.Time) error
	UpdateRateBySymbol(ctx context.Context, symbol, rate string, now time.Time) error
	UpdateStatus(ctx context.Context, id string, status domain.InstrumentStatus, now time.Time) (*domain.Instrument, error)
}
