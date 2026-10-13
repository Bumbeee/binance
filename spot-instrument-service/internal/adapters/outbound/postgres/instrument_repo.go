package postgres

import (
	"context"
	"errors"
	"fmt"
	"market/shared/infra/pagination"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"spot-instrument-service/internal/core/domain"
	"spot-instrument-service/internal/core/ports"
)

const defaultPageSize = 20

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
	var (
		rows pgx.Rows
		err  error
	)

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

func (r *pgxInstrumentRepository) ListPage(
	ctx context.Context,
	statusFilter *domain.InstrumentStatus,
	pageSize int32,
	pageToken string,
) ([]*domain.Instrument, string, error) {

	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	var (
		cursorCreatedAt time.Time
		cursorID        uuid.UUID
		hasCursor       bool
	)

	if pageToken != "" {
		var err error
		cursorCreatedAt, cursorID, err = pagination.DecodeCursor(pageToken)
		if err != nil {
			return nil, "", fmt.Errorf("instrument_repo.ListPage: %w", err)
		}
		hasCursor = true
	}

	var (
		args       []any
		conditions []string
	)

	if statusFilter != nil {
		args = append(args, string(*statusFilter))
		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)))
	}
	if hasCursor {
		args = append(args, cursorCreatedAt, cursorID)
		conditions = append(conditions, fmt.Sprintf("(created_at, id) < ($%d, $%d)", len(args)-1, len(args)))
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	args = append(args, pageSize+1)
	query := fmt.Sprintf(`
		SELECT id, symbol, base_asset, quote_asset,
		       price_precision, quantity_precision,
		       min_order_size, current_rate, status,
		       created_at, updated_at
		FROM instruments
		%s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d
	`, where, len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("instrument_repo.ListPage: %w", err)
	}
	defer rows.Close()

	var instruments []*domain.Instrument
	for rows.Next() {
		instrument, err := scanInstrument(rows)
		if err != nil {
			return nil, "", fmt.Errorf("instrument_repo.ListPage: %w", err)
		}
		instruments = append(instruments, instrument)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("instrument_repo.ListPage: %w", err)
	}

	nextPageToken := ""
	if int32(len(instruments)) > pageSize {
		last := instruments[pageSize-1]
		nextPageToken = pagination.EncodeCursor(last.CreatedAt, last.ID)
		instruments = instruments[:pageSize]
	}

	return instruments, nextPageToken, nil
}

func (r *pgxInstrumentRepository) UpdateRates(ctx context.Context, updates []ports.PriceUpdate, now time.Time) error {
	if len(updates) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, 0, len(updates))
	rates := make([]string, 0, len(updates))
	for _, u := range updates {
		id, err := uuid.Parse(u.InstrumentID)
		if err != nil {
			return fmt.Errorf("instrument_repo.UpdateRates: bad instrument id %q: %w", u.InstrumentID, err)
		}
		ids = append(ids, id)
		rates = append(rates, u.Price)
	}

	query := `
		UPDATE instruments AS i
		SET current_rate = v.rate, updated_at = $3
		FROM unnest($1::uuid[], $2::numeric[]) AS v(id, rate)
		WHERE i.id = v.id
	`

	if _, err := r.pool.Exec(ctx, query, ids, rates, now); err != nil {
		return fmt.Errorf("instrument_repo.UpdateRates: %w", err)
	}

	return nil
}

func (r *pgxInstrumentRepository) UpdateRate(ctx context.Context, id, rate string, now time.Time) (*domain.Instrument, error) {
	query := `
		UPDATE instruments
		SET current_rate = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, symbol, base_asset, quote_asset,
		          price_precision, quantity_precision,
		          min_order_size, current_rate, status,
		          created_at, updated_at
	`

	instrument, err := scanInstrument(r.pool.QueryRow(ctx, query, rate, now, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInstrumentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("instrument_repo.UpdateRate: %w", err)
	}

	return instrument, nil
}

func (r *pgxInstrumentRepository) UpdateRateBySymbol(ctx context.Context, symbol, rate string, now time.Time) error {
	query := `UPDATE instruments SET current_rate = $1, updated_at = $2 WHERE symbol = $3`

	_, err := r.pool.Exec(ctx, query, rate, now, symbol)
	if err != nil {
		return fmt.Errorf("instrument_repo.UpdateRateBySymbol: %w", err)
	}

	return nil
}

func (r *pgxInstrumentRepository) UpdateStatus(ctx context.Context, id string, status domain.InstrumentStatus, now time.Time) (*domain.Instrument, error) {
	query := `
		UPDATE instruments
		SET status = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, symbol, base_asset, quote_asset,
		          price_precision, quantity_precision,
		          min_order_size, current_rate, status,
		          created_at, updated_at
	`

	instrument, err := scanInstrument(r.pool.QueryRow(ctx, query, string(status), now, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrInstrumentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("instrument_repo.UpdateStatus: %w", err)
	}

	return instrument, nil
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
