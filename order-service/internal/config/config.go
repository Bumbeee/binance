package config

import (
	"time"

	"github.com/joho/godotenv"

	"market/shared/infra/logger"
)

var PanicMessage = "unable to load config"

type Config struct { // TODO: check uber.go/fx - если делать, то там многа, есть резон?
	PGDSN             string
	PGMinConns        int
	PGMaxConns        int
	PGConnMaxIdleTime time.Duration
	PGConnMaxLifetime time.Duration

	GRPCAddr string

	UserServiceAddr    string
	UserServiceTimeout time.Duration

	LogConfig logger.LoggerConfig

	ShutdownTimeout time.Duration

	DefaultPageSize int

	OTLPEndpoint       string
	TraceSamplingRatio float64
}

func Load() *Config {
	_ = godotenv.Load()

	pgMinConns := getPGMinConns()
	pgMaxConns := getPGMaxConns()
	validatePGConns(pgMinConns, pgMaxConns)

	return &Config{
		PGDSN:             getPGDSN(),
		PGMinConns:        pgMinConns,
		PGMaxConns:        pgMaxConns,
		PGConnMaxIdleTime: getPGConnMaxIdleTime(),
		PGConnMaxLifetime: getPGConnMaxLifetime(),

		GRPCAddr: getGRPCAddr(),

		UserServiceAddr:    getUserServiceAddr(),
		UserServiceTimeout: getUserServiceTimeout(),

		LogConfig: getLogConfig(),

		ShutdownTimeout: getShutdownTimeout(),

		DefaultPageSize: getDefaultPageSize(),

		OTLPEndpoint:       getOTLPEndpoint(),
		TraceSamplingRatio: getTraceSamplingRatio(),
	}
}
