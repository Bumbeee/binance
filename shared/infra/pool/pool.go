package pool

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewPool(ctx context.Context, log *zap.Logger, dsn string, minConns, maxConns int, maxIdle, maxLifetime time.Duration) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pool.ParseDSN: %w", err)
	}

	cfg.MinConns = int32(minConns)
	cfg.MaxConns = int32(maxConns)
	cfg.MaxConnIdleTime = maxIdle
	cfg.MaxConnLifetime = maxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pool.CreatePool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pool.PingDB: %w", err)
	}

	log.Info("connected to postgres",
		zap.Int("min_conns", minConns),
		zap.Int("max_conns", maxConns),
		zap.Duration("max_conn_idle_time", maxIdle),
		zap.Duration("max_conn_lifetime", maxLifetime),
	)

	return pool, nil
}
