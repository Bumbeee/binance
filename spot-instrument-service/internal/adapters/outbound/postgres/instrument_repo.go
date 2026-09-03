package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"spot-instrument-service/internal/core/domain"
	"spot-instrument-service/internal/core/ports"
)

type pgxInstrumentRepository struct {
	pool *pgxpool.Pool
}

func NewInstrumentRepository(pool *pgxpool.Pool) ports.InstrumentRepository {
	return &pgxInstrumentRepository{pool: pool}
}

func (r *pgxInstrumentRepository) Save(ctx context.Context, instrument *domain.Instrument) error {
	query := `
		INSERT INTO instruments (
			id, symbol, base_asset, quote_asset,
			price_precision, quantity_precision,
			min_order_size, current_rate, status,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.pool.Exec(ctx, query,
		instrument.ID,
		instrument.Symbol,
		instrument.BaseAsset,
		instrument.QuoteAsset,
		instrument.PricePrecision,
		instrument.QuantityPrecision,
		instrument.MinOrderSize,
		instrument.CurrentRate,
		string(instrument.Status),
		instrument.CreatedAt,
		instrument.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrInstrumentAlreadyExists
		}
		return fmt.Errorf("instrument_repo.Save: %w", err)
	}

	return nil
}

func (r *pgxInstrumentRepository) GetByID(ctx context.Context, id string) (*domain.Instrument, error) {
	query := `
		SELECT id, symbol, base_asset, quote_asset,
		       price_precision, quantity_precision,
		       min_order_size, current_rate, status,
		       created_at, updated_at
		FROM instruments
		WHERE id = $1
	`

	instrument, err := scanInstrument(r.pool.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInstrumentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("instrument_repo.GetByID: %w", err)
	}

	return instrument, nil
}

func (r *pgxInstrumentRepository) List(ctx context.Context, statusFilter *domain.InstrumentStatus) ([]*domain.Instrument, error) {
	var rows pgx.Rows
	var err error

	if statusFilter != nil {
		query := `
			SELECT id, symbol, base_asset, quote_asset,
			       price_precision, quantity_precision,
			       min_order_size, current_rate, status,
			       created_at, updated_at
			FROM instruments
			WHERE status = $1
			ORDER BY symbol
		`
		rows, err = r.pool.Query(ctx, query, string(*statusFilter))
	} else {
		query := `
			SELECT id, symbol, base_asset, quote_asset,
			       price_precision, quantity_precision,
			       min_order_size, current_rate, status,
			       created_at, updated_at
			FROM instruments
			ORDER BY symbol
		`
		rows, err = r.pool.Query(ctx, query)
	}
	if err != nil {
		return nil, fmt.Errorf("instrument_repo.List: %w", err)
	}
	defer rows.Close()

	var instruments []*domain.Instrument
	for rows.Next() {
		instrument, err := scanInstrument(rows)
		if err != nil {
			return nil, fmt.Errorf("instrument_repo.List: %w", err)
		}
		instruments = append(instruments, instrument)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("instrument_repo.List: %w", err)
	}

	return instruments, nil
}

func (r *pgxInstrumentRepository) UpdateRate(ctx context.Context, id, rate string) error {
	query := `UPDATE instruments SET current_rate = $1, updated_at = $2 WHERE id = $3`

	tag, err := r.pool.Exec(ctx, query, rate, time.Now(), id)
	if err != nil {
		return fmt.Errorf("instrument_repo.UpdateRate: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrInstrumentNotFound
	}

	return nil
}

func (r *pgxInstrumentRepository) UpdateStatus(ctx context.Context, id string, status domain.InstrumentStatus) error {
	query := `UPDATE instruments SET status = $1, updated_at = $2 WHERE id = $3`

	tag, err := r.pool.Exec(ctx, query, string(status), time.Now(), id)
	if err != nil {
		return fmt.Errorf("instrument_repo.UpdateStatus: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrInstrumentNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanInstrument(row rowScanner) (*domain.Instrument, error) {
	var (
		id                uuid.UUID
		symbol            string
		baseAsset         string
		quoteAsset        string
		pricePrecision    int32
		quantityPrecision int32
		minOrderSize      string
		currentRate       string
		status            string
		createdAt         time.Time
		updatedAt         time.Time
	)

	if err := row.Scan(
		&id, &symbol, &baseAsset, &quoteAsset,
		&pricePrecision, &quantityPrecision,
		&minOrderSize, &currentRate, &status,
		&createdAt, &updatedAt,
	); err != nil {
		return nil, err
	}

	return &domain.Instrument{
		ID:                id,
		Symbol:            symbol,
		BaseAsset:         baseAsset,
		QuoteAsset:        quoteAsset,
		PricePrecision:    pricePrecision,
		QuantityPrecision: quantityPrecision,
		MinOrderSize:      minOrderSize,
		CurrentRate:       currentRate,
		Status:            domain.InstrumentStatus(status),
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}, nil
}
