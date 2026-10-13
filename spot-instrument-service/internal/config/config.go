package config

import (
	"time"

	"github.com/joho/godotenv"

	"market/shared/infra/logger"
)

var PanicMessage = "unable to load config"

type Config struct {
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

	OTLPEndpoint       string
	TraceSamplingRatio float64

	MetricsAddr string

	KafkaBrokers []string
	KafkaGroupID string
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

		OTLPEndpoint:       getOTLPEndpoint(),
		TraceSamplingRatio: getTraceSamplingRatio(),

		MetricsAddr: getMetricsAddr(),

		KafkaBrokers: getKafkaBrokers(),
		KafkaGroupID: getKafkaGroupID(),
	}
}
