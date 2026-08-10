package pool

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dsn string, minConns, maxConns int, maxIdle time.Duration) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pool.ParseDSN: %w", err)
	}

	cfg.MinConns = int32(minConns)
	cfg.MaxConns = int32(maxConns)
	cfg.MaxConnIdleTime = maxIdle

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pool.CreatePool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pool.PingDB: %w", err)
	}

	return pool, nil
}
