package config

import (
	"market/shared/infra/logger"
	"market/shared/infra/redis"
	"market/shared/validator"
	"time"

	"github.com/joho/godotenv"
)

var PanicMessage = "unabled to load config"

type RateLimitConfig struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
	FailTTL   time.Duration
}

type Config struct {
	PGDSN             string
	PGMinConns        int
	PGMaxConns        int
	PGConnMaxIdleTime time.Duration
	PGConnMaxLifetime time.Duration

	RedisConfig redis.RedisConfig

	HasherCost int

	JWTSecret     string
	JWTExpiration time.Duration

	RefreshTokenTTL time.Duration

	PasswordConfig validator.PasswordRequirements

	GRPCAddr string

	LogConfig logger.LoggerConfig

	RateLimitConfig RateLimitConfig
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

		RedisConfig: getRedisConfig(),

		HasherCost: getHasherCost(),

		JWTSecret:     getJWTSecret(),
		JWTExpiration: getJWTExpiration(),

		RefreshTokenTTL: getRefreshTTL(),

		PasswordConfig: getPasswordConfig(),

		GRPCAddr: getGRPCAddr(),

		LogConfig: getLogConfig(),

		RateLimitConfig: RateLimitConfig{
			BaseDelay: getRateLimitBaseDelay(),
			MaxDelay:  getRateLimitMaxDelay(),
			FailTTL:   getRateLimitFailTTL(),
		},
	}
}
