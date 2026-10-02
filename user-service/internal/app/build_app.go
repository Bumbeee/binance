package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	user "market/proto/userservice/v1"
	"market/shared/infra/logger"
	"market/shared/infra/pool"
	"market/shared/infra/redis"
	"market/shared/tracing"

	grpcadapter "userservice/internal/adapters/inbound/grpc"
	"userservice/internal/adapters/inbound/grpc/interceptor"
	"userservice/internal/adapters/outbound/hasher"
	"userservice/internal/adapters/outbound/jwt"
	"userservice/internal/adapters/outbound/postgres"
	redisadapter "userservice/internal/adapters/outbound/redis"
	"userservice/internal/config"
	"userservice/internal/core/services/auth"
	"userservice/internal/core/services/profile"
	"userservice/internal/core/services/token"
)

const serviceName = "user-service"

func BuildApp() {
	cfg := config.Load()

	log, err := logger.New(cfg.LogConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to init logger:", err)
		os.Exit(1)
	}
	defer log.Sync()

	ctx := context.Background()

	shutdownTracing, err := tracing.Init(ctx, tracing.Config{
		ServiceName:   serviceName,
		OTLPEndpoint:  cfg.OTLPEndpoint,
		SamplingRatio: cfg.TraceSamplingRatio,
	})
	if err != nil {
		log.Fatal("failed to init tracing", zap.Error(err))
	}
	defer func() {
		if err := shutdownTracing(ctx); err != nil {
			log.Error("failed to shut down tracing", zap.Error(err))
		}
	}()

	dbpool, err := pool.NewPool(ctx, log, cfg.PGDSN, cfg.PGMinConns, cfg.PGMaxConns, cfg.PGConnMaxIdleTime, cfg.PGConnMaxLifetime)
	if err != nil {
		log.Fatal("failed to connect to postgres", zap.Error(err))
	}

	redisClient, err := redis.NewRedisClient(ctx, log, cfg.RedisConfig)
	if err != nil {
		log.Fatal("failed to connect to redis", zap.Error(err))
	}

	passwordRequirements := cfg.PasswordConfig

	repo := postgres.NewUserRepository(dbpool)
	hasher := hasher.NewBCryptHasher(cfg.HasherCost)
	tokens := jwt.NewIssuer(cfg.JWTSecret, cfg.JWTExpiration)
	refreshTokenStore := redisadapter.NewRefreshTokenStore(redisClient)

	registerCase := auth.NewRegisterCase(repo, hasher, tokens, refreshTokenStore, cfg.RefreshTokenTTL, passwordRequirements)
	loginCase, err := auth.NewLoginCase(repo, hasher, tokens, refreshTokenStore, cfg.RefreshTokenTTL)
	if err != nil {
		log.Fatal("failed to init login case", zap.Error(err))
	}
	validateTokenCase := token.NewValidateTokenCase(tokens)
	logoutCase := auth.NewLogoutCase(refreshTokenStore)
	refreshTokenCase := token.NewRefreshTokenCase(refreshTokenStore, tokens, repo, cfg.RefreshTokenTTL)
	getProfileCase := profile.NewGetProfileCase(repo)
	getUserProfileCase := profile.NewGetUserProfileCase(repo)
	changePasswordCase := auth.NewChangePasswordCase(repo, hasher, passwordRequirements) // отзывать токен при сбросе?
	updateProfileCase := profile.NewUpdateProfileCase(repo)

	loggingInterceptor := interceptor.NewLoggingInterceptor(log)
	authInterceptor := interceptor.NewAuthInterceptor(validateTokenCase)

	limiterInterceptor := interceptor.NewLimiter(
		redisClient,
		log,
		cfg.RateLimitConfig.BaseDelay,
		cfg.RateLimitConfig.MaxDelay,
		cfg.RateLimitConfig.FailTTL,
	)

	validationInterceptor, err := interceptor.NewValidationInterceptor()
	if err != nil {
		log.Fatal("failed to init validation interceptor", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			loggingInterceptor.Unary(),
			validationInterceptor.Unary(),
			limiterInterceptor.Unary(),
			authInterceptor.Unary(),
		),
	)

	reflection.Register(grpcServer)

	userServer := grpcadapter.NewServer(
		registerCase,
		loginCase,
		logoutCase,
		validateTokenCase,
		refreshTokenCase,
		getProfileCase,
		getUserProfileCase,
		changePasswordCase,
		updateProfileCase,
	)

	user.RegisterUserServiceServer(grpcServer, userServer)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	serveErrCh := make(chan error, 1)
	go func() {
		log.Info("UserService listening", zap.String("addr", cfg.GRPCAddr))
		serveErrCh <- grpcServer.Serve(lis)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Info("received signal, shutting down", zap.String("signal", sig.String()))
	case err := <-serveErrCh:
		if err != nil {
			log.Error("grpc server stopped unexpectedly", zap.Error(err))
		}
	}

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Info("grpc server stopped gracefully")
	case <-time.After(cfg.ShutdownTimeout):
		log.Warn("graceful shutdown timed out, forcing stop")
		grpcServer.Stop()
	}

	dbpool.Close()
	log.Info("postgres pool closed")

	if err := redisClient.Close(); err != nil {
		log.Error("failed to close redis client", zap.Error(err))
	} else {
		log.Info("redis client closed")
	}

	log.Info("shutdown complete")
}
