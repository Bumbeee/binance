package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"market/shared/infra/logger"
)

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("%s: %s", PanicMessage, key))
	}
	return value
}

func getEnvString(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	converted, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return converted
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	converted, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return converted
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	converted, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return converted
}

// --- Database ---

func getPGDSN() string {
	return mustGetEnv("PG_DSN")
}

func getPGMinConns() int {
	return getEnvInt("PG_MIN_CONNECTIONS", 2)
}

func getPGMaxConns() int {
	return getEnvInt("PG_MAX_CONNECTIONS", 10)
}

func getPGConnMaxIdleTime() time.Duration {
	return getEnvDuration("PG_MAX_IDLE_TIME", 30*time.Minute)
}

func getPGConnMaxLifetime() time.Duration {
	return getEnvDuration("PG_MAX_CONN_LIFETIME", time.Hour)
}

func validatePGConns(min, max int) {
	if min < 0 || max < 0 {
		panic(fmt.Sprintf("%s: PG_MIN_CONNECTIONS and PG_MAX_CONNECTIONS must be non-negative", PanicMessage))
	}
	if min > max {
		panic(fmt.Sprintf("%s: PG_MIN_CONNECTIONS must not exceed PG_MAX_CONNECTIONS", PanicMessage))
	}
}

// --- gRPC ---

func getGRPCAddr() string {
	return getEnvString("GRPC_ADDR", ":50052")
}

// --- UserService client ---

func getUserServiceAddr() string {
	return mustGetEnv("USER_SERVICE_ADDR")
}

// --- Logger ---

func getLogConfig() logger.LoggerConfig {
	return logger.LoggerConfig{
		Level:             getEnvString("LOG_LEVEL", "info"),
		EncodingJSON:      getEnvBool("LOG_ENCODING_JSON", true),
		EncodingConsole:   getEnvBool("LOG_ENCODING_CONSOLE", false),
		DisableCaller:     getEnvBool("LOG_DISABLE_CALLER", false),
		DisableStacktrace: getEnvBool("LOG_DISABLE_STACKTRACE", false),

		FilePath:       getEnvString("LOG_FILE_PATH", "logs/spot-instrument-service.log"),
		FileMaxSizeMB:  getEnvInt("LOG_FILE_MAX_SIZE_MB", 100),
		FileMaxBackups: getEnvInt("LOG_FILE_MAX_BACKUPS", 5),
		FileMaxAgeDays: getEnvInt("LOG_FILE_MAX_AGE_DAYS", 30),
	}
}
