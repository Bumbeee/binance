package app

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	user "market/proto/userservice/v1"
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
	"userservice/shared/infra/logger"
	"userservice/shared/infra/pool"
	"userservice/shared/infra/redis"
)

const shutdownTimeout = 10 * time.Second

func BuildApp() {
	cfg := config.Load()

	log := logger.New(cfg.LogConfig)
	defer log.Sync()

	ctx := context.Background()
	dbpool, err := pool.NewPool(ctx, cfg.PGDSN, cfg.PGMinConns, cfg.PGMaxConns, cfg.PGConnMaxIdleTime)
	if err != nil {
		log.Fatal("failed to connect to postgres", zap.Error(err))
	}

	redisClient := redis.NewRedisClient(cfg.RedisAddr)
	if err := redisClient.Ping(ctx).Err(); err != nil {
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
	changePasswordCase := auth.NewChangePasswordCase(repo, hasher, passwordRequirements)

	authInterceptor := interceptor.NewAuthInterceptor(validateTokenCase)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
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
	case <-time.After(shutdownTimeout):
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
