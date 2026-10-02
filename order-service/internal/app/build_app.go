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

	orderv1 "market/proto/orderservice/v1"
	"market/shared/infra/logger"
	"market/shared/infra/pool"
	"market/shared/tracing"

	grpcadapter "order-service/internal/adapters/inbound/grpc"
	"order-service/internal/adapters/inbound/grpc/interceptor"
	userclient "order-service/internal/adapters/outbound/client"
	"order-service/internal/adapters/outbound/postgres"
	"order-service/internal/config"
	"order-service/internal/core/services/order"
	"order-service/internal/core/services/trade"
)

const serviceName = "order-service"

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

	userClient, err := userclient.New(cfg.UserServiceAddr, cfg.UserServiceTimeout)
	if err != nil {
		log.Fatal("failed to connect to user-service", zap.Error(err))
	}

	repo := postgres.NewOrderRepository(dbpool)
	tradeRepo := postgres.NewTradeRepository(dbpool)

	matchingEngine := order.NewMatchingEngine()

	createOrderCase := order.NewCreateOrderCase(repo, matchingEngine)
	getOrderCase := order.NewGetOrderCase(repo)
	listOrdersCase := order.NewListOrdersCase(repo)
	cancelOrderCase := order.NewCancelOrderCase(repo)
	getLastTradePricesCase := trade.NewGetLastTradePricesCase(tradeRepo)

	loggingInterceptor := interceptor.NewLoggingInterceptor(log)
	authInterceptor := interceptor.NewAuthInterceptor(userClient)
	validationInterceptor, err := interceptor.NewValidationInterceptor()
	if err != nil {
		log.Fatal("failed to init validation interceptor", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			loggingInterceptor.Unary(),
			validationInterceptor.Unary(),
			authInterceptor.Unary(),
		),
	)

	reflection.Register(grpcServer)

	orderServer := grpcadapter.NewServer(
		createOrderCase,
		getOrderCase,
		listOrdersCase,
		cancelOrderCase,
		getLastTradePricesCase,
	)

	orderv1.RegisterOrderServiceServer(grpcServer, orderServer)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	serveErrCh := make(chan error, 1)
	go func() {
		log.Info("OrderService listening", zap.String("addr", cfg.GRPCAddr))
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

	if err := userClient.Close(); err != nil {
		log.Error("failed to close user-service client", zap.Error(err))
	} else {
		log.Info("user-service client closed")
	}

	log.Info("shutdown complete")
}
