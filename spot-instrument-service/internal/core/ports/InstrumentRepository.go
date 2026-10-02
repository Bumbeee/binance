package ports

import (
	"context"
	"time"

	"spot-instrument-service/internal/core/domain"
)

type InstrumentRepository interface {
	Save(ctx context.Context, instrument *domain.Instrument) error
	GetByID(ctx context.Context, id string) (*domain.Instrument, error)
	List(ctx context.Context, statusFilter *domain.InstrumentStatus) ([]*domain.Instrument, error)
	UpdateRate(ctx context.Context, id, rate string, now time.Time) (*domain.Instrument, error)
	UpdateRateBySymbol(ctx context.Context, symbol, rate string, now time.Time) error
	UpdateStatus(ctx context.Context, id string, status domain.InstrumentStatus, now time.Time) (*domain.Instrument, error)
}
