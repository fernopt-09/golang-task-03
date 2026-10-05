package repository

import (
	"context"

	"github.com/kazah/golang-task-03/internal/domain"
)

// RateRepository is the storage for calculated rates.
type RateRepository interface {
	Save(ctx context.Context, rate *domain.Rate) error
	GetLatest(ctx context.Context) (*domain.Rate, error)
	List(ctx context.Context, limit, offset int) ([]domain.Rate, error)
	Ping(ctx context.Context) error
}
