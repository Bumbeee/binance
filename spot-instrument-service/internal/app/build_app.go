package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	spotv1 "market/proto/spotinstrumentservice/v1"
	"market/shared/infra/logger"
	pool "market/shared/infra/pool"
	grpcadapter "spot-instrument-service/internal/adapters/inbound/grpc"
	"spot-instrument-service/internal/adapters/inbound/grpc/interceptor"
	"spot-instrument-service/internal/adapters/outbound/postgres"
	"spot-instrument-service/internal/adapters/outbound/userclient"
	"spot-instrument-service/internal/config"
	"spot-instrument-service/internal/core/services/instrument"
)

const shutdownTimeout = 10 * time.Second

func BuildApp() {
	cfg := config.Load()

	log, err := logger.New(cfg.LogConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to init logger:", err)
		os.Exit(1)
	}
	defer log.Sync()

	ctx := context.Background()

	dbpool, err := pool.NewPool(ctx, log, cfg.PGDSN, cfg.PGMinConns, cfg.PGMaxConns, cfg.PGConnMaxIdleTime, cfg.PGConnMaxLifetime)
	if err != nil {
		log.Fatal("failed to connect to postgres", zap.Error(err))
	}

	userClient, err := userclient.New(cfg.UserServiceAddr)
	if err != nil {
		log.Fatal("failed to connect to user-service", zap.Error(err))
	}

	repo := postgres.NewInstrumentRepository(dbpool)

	createInstrumentCase := instrument.NewCreateInstrumentCase(repo)
	getInstrumentCase := instrument.NewGetInstrumentCase(repo)
	listInstrumentsCase := instrument.NewListInstrumentsCase(repo)
	archiveInstrumentCase := instrument.NewArchiveInstrumentCase(repo)
	updateRateCase := instrument.NewUpdateRateCase(repo)

	loggingInterceptor := interceptor.NewLoggingInterceptor(log)
	authInterceptor := interceptor.NewAuthInterceptor(userClient)

	validationInterceptor, err := interceptor.NewValidationInterceptor()
	if err != nil {
		log.Fatal("failed to init validation interceptor", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
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
	case <-time.After(shutdownTimeout):
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
