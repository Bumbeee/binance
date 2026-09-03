package pool

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dsn string, minConns, maxConns int32, maxIdleTime time.Duration) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pool.ParseConfig: %w", err)
	}

	cfg.MinConns = minConns
	cfg.MaxConns = maxConns
	cfg.MaxConnIdleTime = maxIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pool.NewWithConfig: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("pool.PingDB: %w", err)
	}

	return pool, nil
}
