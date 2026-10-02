package trade

import (
	"context"
	"time"

	"github.com/google/uuid"

	"order-service/internal/core/ports"
)

type GetLastTradePricesCase struct {
	repo ports.TradeRepository
}

func NewGetLastTradePricesCase(repo ports.TradeRepository) *GetLastTradePricesCase {
	return &GetLastTradePricesCase{repo: repo}
}

type LastTradePriceResult struct {
	InstrumentID string
	Price        string
	TradedAt     time.Time
}

func (uc *GetLastTradePricesCase) Execute(ctx context.Context, instrumentIDs []string) ([]*LastTradePriceResult, error) {
	parsed := make([]uuid.UUID, 0, len(instrumentIDs))
	for _, id := range instrumentIDs {
		if parsedID, err := uuid.Parse(id); err == nil {
			parsed = append(parsed, parsedID)
		}
	}

	trades, err := uc.repo.GetLatestByInstrumentIDs(ctx, parsed)
	if err != nil {
		return nil, err
	}

	results := make([]*LastTradePriceResult, 0, len(trades))
	for instrumentID, t := range trades {
		results = append(results, &LastTradePriceResult{
			InstrumentID: instrumentID.String(),
			Price:        t.Price,
			TradedAt:     t.CreatedAt,
		})
	}

	return results, nil
}
