package config

import (
	"time"

	"github.com/joho/godotenv"

	"spot-instrument-service/shared/logger"
)

var PanicMessage = "unable to load config"

type Config struct {
	PGDSN             string
	PGMinConns        int
	PGMaxConns        int
	PGConnMaxIdleTime time.Duration

	GRPCAddr string

	UserServiceAddr string

	LogConfig logger.LoggerStruct
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

		GRPCAddr: getGRPCAddr(),

		UserServiceAddr: getUserServiceAddr(),

		LogConfig: getLogConfig(),
	}
}
