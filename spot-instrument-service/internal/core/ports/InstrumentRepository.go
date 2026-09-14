package ports

import (
	"context"

	"spot-instrument-service/internal/core/domain"
)

type InstrumentRepository interface {
	Save(ctx context.Context, instrument *domain.Instrument) error
	GetByID(ctx context.Context, id string) (*domain.Instrument, error)
	List(ctx context.Context, statusFilter *domain.InstrumentStatus) ([]*domain.Instrument, error)
	UpdateRate(ctx context.Context, id, rate string) (*domain.Instrument, error)
	UpdateRateBySymbol(ctx context.Context, symbol, rate string) error
	UpdateStatus(ctx context.Context, id string, status domain.InstrumentStatus) (*domain.Instrument, error)
}
