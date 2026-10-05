package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/kazah/golang-task-03/internal/domain"
	"github.com/kazah/golang-task-03/internal/repository"
)

// rateRepository stores rates in PostgreSQL.
type rateRepository struct {
	pool   *pgxpool.Pool
	tracer trace.Tracer
}

// NewRateRepository creates a repository on top of the connection pool.
func NewRateRepository(pool *pgxpool.Pool) repository.RateRepository {
	return &rateRepository{
		pool:   pool,
		tracer: otel.Tracer("postgres-repository"),
	}
}

// Save inserts the rate and fills rate.ID.
func (r *rateRepository) Save(ctx context.Context, rate *domain.Rate) error {
	ctx, span := r.tracer.Start(ctx, "RateRepository.Save",
		trace.WithAttributes(
			attribute.Float64("rate.ask", rate.Ask),
			attribute.Float64("rate.bid", rate.Bid),
			attribute.String("rate.method", string(rate.Method)),
		),
	)
	defer span.End()

	// params are stored in the JSONB column
	paramsJSON, err := json.Marshal(rate.Params)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to marshal rate params: %w", err)
	}

	if rate.CreatedAt.IsZero() {
		rate.CreatedAt = time.Now().UTC()
	}

	// RETURNING gives us the generated id
	query := `
		INSERT INTO rates (ask, bid, exchange_timestamp, calculation_method, params, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err = r.pool.QueryRow(ctx, query,
		rate.Ask,
		rate.Bid,
		rate.Timestamp,
		string(rate.Method),
		paramsJSON,
		rate.CreatedAt,
	).Scan(&rate.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to insert rate: %w", err)
	}

	span.SetAttributes(attribute.Int64("rate.id", rate.ID))
	return nil
}

// GetLatest returns the last saved rate or ErrRateNotFound.
func (r *rateRepository) GetLatest(ctx context.Context) (*domain.Rate, error) {
	ctx, span := r.tracer.Start(ctx, "RateRepository.GetLatest")
	defer span.End()

	query := `
		SELECT id, ask, bid, exchange_timestamp, calculation_method, params, created_at
		FROM rates
		ORDER BY id DESC
		LIMIT 1
	`

	var (
		rate       domain.Rate
		paramsJSON []byte
		methodStr  string
	)

	err := r.pool.QueryRow(ctx, query).Scan(
		&rate.ID,
		&rate.Ask,
		&rate.Bid,
		&rate.Timestamp,
		&methodStr,
		&paramsJSON,
		&rate.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRateNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to query latest rate: %w", err)
	}

	rate.Method = domain.CalculationMethod(methodStr)
	if len(paramsJSON) > 0 {
		_ = json.Unmarshal(paramsJSON, &rate.Params)
	}

	return &rate, nil
}

// List returns rates from newest to oldest.
func (r *rateRepository) List(ctx context.Context, limit, offset int) ([]domain.Rate, error) {
	ctx, span := r.tracer.Start(ctx, "RateRepository.List",
		trace.WithAttributes(
			attribute.Int("query.limit", limit),
			attribute.Int("query.offset", offset),
		),
	)
	defer span.End()

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, ask, bid, exchange_timestamp, calculation_method, params, created_at
		FROM rates
		ORDER BY id DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to query rates list: %w", err)
	}
	defer rows.Close()

	var rates []domain.Rate
	for rows.Next() {
		var (
			rate       domain.Rate
			paramsJSON []byte
			methodStr  string
		)
		err := rows.Scan(
			&rate.ID,
			&rate.Ask,
			&rate.Bid,
			&rate.Timestamp,
			&methodStr,
			&paramsJSON,
			&rate.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rate row: %w", err)
		}
		rate.Method = domain.CalculationMethod(methodStr)
		if len(paramsJSON) > 0 {
			_ = json.Unmarshal(paramsJSON, &rate.Params)
		}
		rates = append(rates, rate)
	}

	return rates, rows.Err()
}

// Ping checks that the database is alive.
func (r *rateRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
