package app

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	"userservice/shared/infra/pool"
	"userservice/shared/infra/redis"
)

const shutdownTimeout = 10 * time.Second

func BuildApp() { // move inmemory to postgress + redis + goose migration
	cfg := config.Load()

	ctx := context.Background()
	dbpool, err := pool.NewPool(ctx, cfg.PGDSN, cfg.PGMinConns, cfg.PGMaxConns, cfg.PGConnMaxIdleTime)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	redisClient := redis.NewRedisClient(cfg.RedisAddr)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	passwordRequirements := cfg.PasswordConfig

	repo := postgres.NewUserRepository(dbpool)
	hasher := hasher.NewBCryptHasher(cfg.HasherCost)
	tokens := jwt.NewIssuer(cfg.JWTSecret, cfg.JWTExpiration)
	refreshTokenStore := redisadapter.NewRefreshTokenStore(redisClient)

	registerCase := auth.NewRegisterCase(repo, hasher, tokens, refreshTokenStore, cfg.RefreshTokenTTL, passwordRequirements)
	loginCase, err := auth.NewLoginCase(repo, hasher, tokens, refreshTokenStore, cfg.RefreshTokenTTL)
	if err != nil {
		log.Fatalf("failed to init login case: %v", err)
	}
	validateTokenCase := token.NewValidateTokenCase(tokens)
	logoutCase := auth.NewLogoutCase(refreshTokenStore)
	refreshTokenCase := token.NewRefreshTokenCase(refreshTokenStore, tokens, cfg.RefreshTokenTTL)
	getProfileCase := profile.NewGetProfileCase(repo)

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
	)

	user.RegisterUserServiceServer(grpcServer, userServer)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	serveErrCh := make(chan error, 1)
	go func() {
		log.Printf("UserService listening on %s", cfg.GRPCAddr)
		serveErrCh <- grpcServer.Serve(lis)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("received signal %s, shutting down...", sig)
	case err := <-serveErrCh:
		if err != nil {
			log.Printf("grpc server stopped unexpectedly: %v", err)
		}
	}

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Println("grpc server stopped gracefully")
	case <-time.After(shutdownTimeout):
		log.Println("graceful shutdown timed out, forcing stop")
		grpcServer.Stop()
	}

	dbpool.Close()
	log.Println("postgres pool closed")

	if err := redisClient.Close(); err != nil {
		log.Printf("failed to close redis client: %v", err)
	} else {
		log.Println("redis client closed")
	}

	log.Println("shutdown complete")
}
