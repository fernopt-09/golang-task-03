package service

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/kazah/golang-task-03/internal/calculator"
	"github.com/kazah/golang-task-03/internal/client/binance"
	"github.com/kazah/golang-task-03/internal/domain"
	"github.com/kazah/golang-task-03/internal/repository"
	"github.com/kazah/golang-task-03/internal/telemetry"
)

// RateService is the business logic of the app.
type RateService interface {
	GetRates(ctx context.Context, params domain.CalculationParams) (*domain.Rate, error)
	HealthCheck(ctx context.Context) (bool, error)
}

// rateService is the implementation of RateService.
type rateService struct {
	client     binance.Client
	calculator calculator.Calculator
	repo       repository.RateRepository
	symbol     string
	logger     *zap.Logger
	tracer     trace.Tracer
}

// Config is the service settings.
type Config struct {
	Symbol string
}

// New creates the service. If the symbol is empty, the default one is used.
func New(
	client binance.Client,
	calc calculator.Calculator,
	repo repository.RateRepository,
	logger *zap.Logger,
	cfg Config,
) RateService {
	symbol := cfg.Symbol
	if symbol == "" {
		symbol = binance.DefaultSymbol
	}

	return &rateService{
		client:     client,
		calculator: calc,
		repo:       repo,
		symbol:     symbol,
		logger:     logger,
		tracer:     otel.Tracer("rate-service"),
	}
}

// GetRates gets the order book, calculates the rate and saves it to the database.
func (s *rateService) GetRates(ctx context.Context, params domain.CalculationParams) (*domain.Rate, error) {
	start := time.Now()
	ctx, span := s.tracer.Start(ctx, "RateService.GetRates",
		trace.WithAttributes(
			attribute.String("rate.method", string(params.Method)),
			attribute.Int64("rate.param_n", int64(params.N)),
			attribute.Int64("rate.param_m", int64(params.M)),
		),
	)
	defer span.End()

	methodStr := string(params.Method)

	// 1. get the order book from the exchange
	apiStart := time.Now()
	ob, err := s.client.GetDepth(ctx, s.symbol)
	apiDuration := time.Since(apiStart).Seconds()

	if err != nil {
		telemetry.ExternalAPIDuration.WithLabelValues("binance", "error").Observe(apiDuration)
		telemetry.RateRequestsTotal.WithLabelValues(methodStr, "error_api").Inc()
		s.logger.Error("failed to get depth from Binance", zap.Error(err))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to fetch order book: %w", err)
	}
	telemetry.ExternalAPIDuration.WithLabelValues("binance", "success").Observe(apiDuration)

	// 2. calculate ask and bid by the chosen method
	ask, bid, err := s.calculator.Calculate(ob, params)
	if err != nil {
		telemetry.RateRequestsTotal.WithLabelValues(methodStr, "error_calc").Inc()
		s.logger.Error("failed to calculate rate", zap.Error(err), zap.Any("params", params))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("calculation error: %w", err)
	}

	// 3. save the result, it is required for every call
	rate := &domain.Rate{
		Ask:       ask,
		Bid:       bid,
		Timestamp: time.Unix(ob.Timestamp, 0).UTC(),
		Method:    params.Method,
		Params: map[string]any{
			"n": params.N,
			"m": params.M,
		},
		CreatedAt: time.Now().UTC(),
	}

	dbStart := time.Now()
	err = s.repo.Save(ctx, rate)
	dbDuration := time.Since(dbStart).Seconds()

	if err != nil {
		telemetry.DBSaveDuration.WithLabelValues("error").Observe(dbDuration)
		telemetry.RateRequestsTotal.WithLabelValues(methodStr, "error_db").Inc()
		s.logger.Error("failed to save rate to database", zap.Error(err))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("database save error: %w", err)
	}
	telemetry.DBSaveDuration.WithLabelValues("success").Observe(dbDuration)

	totalDuration := time.Since(start).Seconds()
	telemetry.RateRequestsTotal.WithLabelValues(methodStr, "success").Inc()
	telemetry.RateRequestDuration.WithLabelValues(methodStr).Observe(totalDuration)

	s.logger.Info("rate fetched and saved",
		zap.Int64("id", rate.ID),
		zap.Float64("ask", rate.Ask),
		zap.Float64("bid", rate.Bid),
		zap.Time("timestamp", rate.Timestamp),
		zap.String("method", string(rate.Method)),
	)

	return rate, nil
}

// HealthCheck checks the database connection.
func (s *rateService) HealthCheck(ctx context.Context) (bool, error) {
	_, span := s.tracer.Start(ctx, "RateService.HealthCheck")
	defer span.End()

	if err := s.repo.Ping(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return false, fmt.Errorf("database health check failed: %w", err)
	}

	return true, nil
}
