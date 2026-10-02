package worker

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"spot-instrument-service/internal/core/domain"
	"spot-instrument-service/internal/core/ports"
)

type RatePoller struct {
	source ports.LastTradePriceSource
	repo   ports.InstrumentRepository
	log    *zap.Logger
}

func NewRatePoller(source ports.LastTradePriceSource, repo ports.InstrumentRepository, log *zap.Logger) *RatePoller {
	return &RatePoller{source: source, repo: repo, log: log}
}

func (p *RatePoller) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.pollOnce(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (p *RatePoller) pollOnce(ctx context.Context) {
	active := domain.InstrumentStatusActive
	instruments, err := p.repo.List(ctx, &active)
	if err != nil {
		p.log.Error("rate poller: failed to list active instruments", zap.Error(err))
		return
	}
	if len(instruments) == 0 {
		return
	}

	ids := make([]string, 0, len(instruments))
	for _, inst := range instruments {
		ids = append(ids, inst.ID.String())
	}

	updates, err := p.source.GetLastTradePrices(ctx, ids)
	if err != nil {
		p.log.Warn("rate poller: failed to fetch last trade prices", zap.Error(err))
		return
	}

	now := time.Now()
	for _, u := range updates {
		if _, err := p.repo.UpdateRate(ctx, u.InstrumentID, u.Price, now); err != nil {
			if errors.Is(err, domain.ErrInstrumentNotFound) {
				continue
			}
			p.log.Warn("rate poller: failed to persist rate update",
				zap.String("instrument_id", u.InstrumentID), zap.Error(err))
		}
	}
}
