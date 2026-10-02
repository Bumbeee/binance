package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"order-service/internal/core/domain"
	"order-service/internal/core/ports"
)

type pgxTradeRepository struct {
	pool *pgxpool.Pool
}

func NewTradeRepository(pool *pgxpool.Pool) ports.TradeRepository {
	return &pgxTradeRepository{pool: pool}
}

func (r *pgxTradeRepository) GetLatestByInstrumentIDs(ctx context.Context, instrumentIDs []uuid.UUID) (map[uuid.UUID]*domain.Trade, error) {
	if len(instrumentIDs) == 0 {
		return map[uuid.UUID]*domain.Trade{}, nil
	}

	query := `
		SELECT DISTINCT ON (instrument_id)
		       id, instrument_id, buy_order_id, sell_order_id, price, quantity, created_at
		FROM trades
		WHERE instrument_id = ANY($1)
		ORDER BY instrument_id, created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, instrumentIDs)
	if err != nil {
		return nil, fmt.Errorf("trade_repo.GetLatestByInstrumentIDs: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]*domain.Trade, len(instrumentIDs))
	for rows.Next() {
		trade, err := scanTrade(rows)
		if err != nil {
			return nil, fmt.Errorf("trade_repo.GetLatestByInstrumentIDs: %w", err)
		}
		result[trade.InstrumentID] = trade
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("trade_repo.GetLatestByInstrumentIDs: %w", err)
	}

	return result, nil
}

func scanTrade(row rowScanner) (*domain.Trade, error) {
	var (
		id           uuid.UUID
		instrumentID uuid.UUID
		buyOrderID   uuid.UUID
		sellOrderID  uuid.UUID
		price        string
		quantity     string
		createdAt    time.Time
	)

	if err := row.Scan(&id, &instrumentID, &buyOrderID, &sellOrderID, &price, &quantity, &createdAt); err != nil {
		return nil, err
	}

	return &domain.Trade{
		ID:           id,
		InstrumentID: instrumentID,
		BuyOrderID:   buyOrderID,
		SellOrderID:  sellOrderID,
		Price:        price,
		Quantity:     quantity,
		CreatedAt:    createdAt,
	}, nil
}
