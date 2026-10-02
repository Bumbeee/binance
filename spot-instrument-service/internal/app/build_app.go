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

	spotv1 "market/proto/spotinstrumentservice/v1"
	"market/shared/infra/logger"
	"market/shared/infra/pool"
	"market/shared/tracing"

	grpcadapter "spot-instrument-service/internal/adapters/inbound/grpc"
	"spot-instrument-service/internal/adapters/inbound/grpc/interceptor"
	"spot-instrument-service/internal/adapters/outbound/orderclient"
	"spot-instrument-service/internal/adapters/outbound/postgres"
	"spot-instrument-service/internal/adapters/outbound/userclient"
	"spot-instrument-service/internal/config"
	"spot-instrument-service/internal/core/services/instrument"
	"spot-instrument-service/internal/worker"
)

const serviceName = "spot-instrument-service"

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

	orderClient, err := orderclient.New(cfg.OrderServiceAddr, cfg.OrderServiceTimeout)
	if err != nil {
		log.Fatal("failed to connect to order-service", zap.Error(err))
	}

	repo := postgres.NewInstrumentRepository(dbpool)

	createInstrumentCase := instrument.NewCreateInstrumentCase(repo)
	getInstrumentCase := instrument.NewGetInstrumentCase(repo)
	listInstrumentsCase := instrument.NewListInstrumentsCase(repo)
	archiveInstrumentCase := instrument.NewArchiveInstrumentCase(repo)
	updateRateCase := instrument.NewUpdateRateCase(repo)

	ratePoller := worker.NewRatePoller(orderClient, repo, log)
	pollerCtx, cancelPoller := context.WithCancel(context.Background())
	go ratePoller.Run(pollerCtx, cfg.RatePollInterval)

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

	instrumentServer := grpcadapter.NewServer(
		createInstrumentCase,
		getInstrumentCase,
		listInstrumentsCase,
		archiveInstrumentCase,
		updateRateCase,
	)

	spotv1.RegisterSpotInstrumentServiceServer(grpcServer, instrumentServer)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	serveErrCh := make(chan error, 1)
	go func() {
		log.Info("SpotInstrumentService listening", zap.String("addr", cfg.GRPCAddr))
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

	cancelPoller()

	dbpool.Close()
	log.Info("postgres pool closed")

	if err := userClient.Close(); err != nil {
		log.Error("failed to close user-service client", zap.Error(err))
	} else {
		log.Info("user-service client closed")
	}

	if err := orderClient.Close(); err != nil {
		log.Error("failed to close order-service client", zap.Error(err))
	} else {
		log.Info("order-service client closed")
	}

	log.Info("shutdown complete")
}
