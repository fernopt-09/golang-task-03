package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazah/golang-task-03/internal/domain"
	"github.com/kazah/golang-task-03/internal/repository"
	"github.com/kazah/golang-task-03/internal/repository/postgres"
)

func TestConfig_DSN(t *testing.T) {
	cfg := postgres.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "secretpassword",
		DBName:   "ratedb",
		SSLMode:  "disable",
	}

	expected := "postgres://postgres:secretpassword@localhost:5432/ratedb?sslmode=disable"
	assert.Equal(t, expected, cfg.DSN())
}

func TestMockRateRepository(t *testing.T) {
	repo := repository.NewMockRateRepository()
	ctx := context.Background()

	_, err := repo.GetLatest(ctx)
	require.ErrorIs(t, err, domain.ErrRateNotFound)

	rate1 := &domain.Rate{
		Ask:       95.5,
		Bid:       95.0,
		Timestamp: time.Now(),
		Method:    domain.MethodTopN,
		Params:    map[string]any{"n": 1},
	}

	err = repo.Save(ctx, rate1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), rate1.ID)

	rate2 := &domain.Rate{
		Ask:       96.0,
		Bid:       95.2,
		Timestamp: time.Now(),
		Method:    domain.MethodAvgNM,
		Params:    map[string]any{"n": 1, "m": 3},
	}

	err = repo.Save(ctx, rate2)
	require.NoError(t, err)
	assert.Equal(t, int64(2), rate2.ID)

	latest, err := repo.GetLatest(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), latest.ID)
	assert.Equal(t, 96.0, latest.Ask)

	list, err := repo.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}
