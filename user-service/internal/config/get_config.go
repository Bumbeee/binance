package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"userservice/shared/infra/logger"
	"userservice/shared/validator"
)

// mustGetEnv возвращает значение обязательной переменной или паникует
// с указанием, какой именно переменной не хватает.
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

// validatePGConns проверяет консистентность пула соединений.
func validatePGConns(min, max int) {
	if min < 0 || max < 0 {
		panic(fmt.Sprintf("%s: PG_MIN_CONNECTIONS and PG_MAX_CONNECTIONS must be non-negative", PanicMessage))
	}
	if min > max {
		panic(fmt.Sprintf("%s: PG_MIN_CONNECTIONS must not exceed PG_MAX_CONNECTIONS", PanicMessage))
	}
}

// --- Redis ---

func getRedisAddr() string {
	return mustGetEnv("REDIS_ADDR")
}

// --- Hasher ---

func getHasherCost() int {
	cost := getEnvInt("HASHER_COST", 12)
	// bcrypt допускает диапазон [4, 31]
	if cost < 4 || cost > 31 {
		panic(fmt.Sprintf("%s: HASHER_COST must be in range [4, 31], got %d", PanicMessage, cost))
	}
	return cost
}

// --- JWT ---

func getJWTSecret() string {
	secret := mustGetEnv("JWT_SECRET")
	if len(secret) < 16 {
		panic(fmt.Sprintf("%s: JWT_SECRET must be at least 16 characters long", PanicMessage))
	}
	return secret
}

func getJWTExpiration() time.Duration {
	return getEnvDuration("JWT_EXPIRATION", 1*time.Hour)
}

// --- Refresh Token ---

func getRefreshTTL() time.Duration {
	return getEnvDuration("REFRESH_TTL", 7*24*time.Hour)
}

// --- Password ---

func getPasswordConfig() validator.PasswordRequirements {
	cfg := validator.PasswordRequirements{
		MinLength:      getEnvInt("PASSWORD_MIN_LENGTH", 8),
		MaxLength:      getEnvInt("PASSWORD_MAX_LENGTH", 64),
		RequireUpper:   getEnvBool("PASSWORD_REQUIRES_UPPERCASE", true),
		RequireLower:   getEnvBool("PASSWORD_REQUIRES_LOWERCASE", true),
		RequireDigit:   getEnvBool("PASSWORD_REQUIRES_DIGIT", true),
		RequireSpecial: getEnvBool("PASSWORD_REQUIRES_SPECIAL", true),
		DisallowSpaces: getEnvBool("PASSWORD_DISALLOW_SPACES", true),
	}

	if cfg.MinLength <= 0 {
		panic(fmt.Sprintf("%s: PASSWORD_MIN_LENGTH must be positive", PanicMessage))
	}
	if cfg.MaxLength < cfg.MinLength {
		panic(fmt.Sprintf("%s: PASSWORD_MAX_LENGTH must be >= PASSWORD_MIN_LENGTH", PanicMessage))
	}

	return cfg
}

// --- GRPC ---

func getGRPCAddr() string {
	return getEnvString("GRPC_ADDR", ":50051")
}

// --- Logger ---

func getLogConfig() logger.LoggerStruct {
	return logger.LoggerStruct{
		Level:             getEnvString("LOG_LEVEL", "info"),
		Encoding:          getEnvString("LOG_ENCODING", "json"),
		DisableCaller:     getEnvBool("LOG_DISABLE_CALLER", false),
		DisableStacktrace: getEnvBool("LOG_DISABLE_STACKTRACE", false),

		FilePath:       getEnvString("LOG_FILE_PATH", "logs/user-service.log"),
		FileMaxSizeMB:  getEnvInt("LOG_FILE_MAX_SIZE_MB", 100),
		FileMaxBackups: getEnvInt("LOG_FILE_MAX_BACKUPS", 5),
		FileMaxAgeDays: getEnvInt("LOG_FILE_MAX_AGE_DAYS", 30),
	}
}
