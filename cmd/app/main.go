package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/kazah/golang-task-03/internal/calculator"
	"github.com/kazah/golang-task-03/internal/client/binance"
	"github.com/kazah/golang-task-03/internal/config"
	internalgrpc "github.com/kazah/golang-task-03/internal/grpc"
	"github.com/kazah/golang-task-03/internal/logger"
	"github.com/kazah/golang-task-03/internal/repository/postgres"
	"github.com/kazah/golang-task-03/internal/service"
	"github.com/kazah/golang-task-03/internal/telemetry"
)

// main wires all parts of the service together and waits for a stop signal.
func main() {
	// read config from env and flags
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = log.Sync() }()

	log.Info("starting rate service",
		zap.Int("grpc_port", cfg.GRPCPort),
		zap.Int("metrics_port", cfg.MetricsPort),
		zap.String("db_host", cfg.DB.Host),
		zap.String("exchange_url", cfg.Binance.BaseURL),
		zap.String("symbol", cfg.Symbol),
	)

	// tracing is optional, the service works without it
	shutdownTracer, err := telemetry.InitTracer(context.Background(), "rate-service")
	if err != nil {
		log.Warn("failed to init tracer provider", zap.Error(err))
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = shutdownTracer(ctx)
		}()
	}

	// connect to the database, 15 seconds is enough for the container to start
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 15*time.Second)
	pool, err := postgres.NewPool(dbCtx, cfg.DB)
	dbCancel()
	if err != nil {
		log.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer pool.Close()

	if cfg.AutoMigrate {
		if err := postgres.RunMigrations(cfg.DB.DSN()); err != nil {
			log.Fatal("database migrations failed", zap.Error(err))
		}
		log.Info("database migrations applied")
	}

	// create all layers: repository -> service -> gRPC handler
	rateRepo := postgres.NewRateRepository(pool)
	exchangeClient := binance.New(cfg.Binance)
	calc := calculator.New()
	rateService := service.New(exchangeClient, calc, rateRepo, log, service.Config{Symbol: cfg.Symbol})

	handler := internalgrpc.NewHandler(rateService, log)
	grpcServer := internalgrpc.NewServer(handler, log)

	// separate HTTP server for Prometheus
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", telemetry.MetricsHandler())
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.MetricsPort),
		Handler: metricsMux,
		// protection from slow clients (gosec G112)
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("metrics server listening", zap.String("addr", metricsServer.Addr))
		if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("metrics server failed", zap.Error(err))
		}
	}()

	// gRPC server works in the background, errors come through the channel
	serverErrors := make(chan error, 1)
	go func() {
		if err := grpcServer.Start(cfg.GRPCPort); err != nil {
			serverErrors <- err
		}
	}()

	// wait for Ctrl+C or SIGTERM (docker stop)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatal("grpc server error", zap.Error(err))
	case sig := <-shutdown:
		log.Info("received shutdown signal", zap.String("signal", sig.String()))

		// graceful shutdown: give active requests 10 seconds to finish
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := grpcServer.Stop(shutdownCtx); err != nil {
			log.Error("error during grpc shutdown", zap.Error(err))
		}

		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			log.Error("error during metrics shutdown", zap.Error(err))
		}

		log.Info("service exited")
	}
}
