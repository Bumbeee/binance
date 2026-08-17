package config

import (
	"time"
	"userservice/shared/infra/logger"
	"userservice/shared/validator"

	"github.com/joho/godotenv"
)

var PanicMessage = "unabled to load config"

type Config struct {
	PGDSN             string
	PGMinConns        int
	PGMaxConns        int
	PGConnMaxIdleTime time.Duration

	RedisAddr string

	HasherCost int

	JWTSecret     string
	JWTExpiration time.Duration

	RefreshTokenTTL time.Duration

	PasswordConfig validator.PasswordRequirements

	GRPCAddr string

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
		// TODO: add life time

		RedisAddr: getRedisAddr(),

		HasherCost: getHasherCost(),

		JWTSecret:     getJWTSecret(),
		JWTExpiration: getJWTExpiration(),

		RefreshTokenTTL: getRefreshTTL(),

		PasswordConfig: getPasswordConfig(),

		GRPCAddr: getGRPCAddr(),

		LogConfig: getLogConfig(),
	}
}
