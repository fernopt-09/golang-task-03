package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/kazah/golang-task-03/internal/calculator"
	"github.com/kazah/golang-task-03/internal/domain"
	"github.com/kazah/golang-task-03/internal/repository"
	"github.com/kazah/golang-task-03/internal/service"
)

type mockBinanceClient struct {
	OrderBook *domain.OrderBook
	Err       error
}

func (m *mockBinanceClient) GetDepth(_ context.Context, _ string) (*domain.OrderBook, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.OrderBook, nil
}

func sampleOrderBook() *domain.OrderBook {
	return &domain.OrderBook{
		Timestamp: 1712000000,
		Asks: []domain.PriceLevel{
			{Price: 95.0, Volume: 1.0},
			{Price: 96.0, Volume: 2.0},
		},
		Bids: []domain.PriceLevel{
			{Price: 94.0, Volume: 1.0},
			{Price: 93.0, Volume: 2.0},
		},
	}
}

func TestRateService_GetRates_Success(t *testing.T) {
	mockClient := &mockBinanceClient{OrderBook: sampleOrderBook()}
	calc := calculator.New()
	mockRepo := repository.NewMockRateRepository()
	logger := zap.NewNop()

	svc := service.New(mockClient, calc, mockRepo, logger, service.Config{Symbol: "USDTTRY"})

	ctx := context.Background()
	rate, err := svc.GetRates(ctx, domain.CalculationParams{
		Method: domain.MethodTopN,
		N:      1,
	})

	require.NoError(t, err)
	require.NotNil(t, rate)
	assert.Equal(t, 95.0, rate.Ask)
	assert.Equal(t, 94.0, rate.Bid)
	assert.Equal(t, domain.MethodTopN, rate.Method)
	assert.Equal(t, int64(1), rate.ID)
	assert.Equal(t, time.Unix(1712000000, 0).UTC(), rate.Timestamp)

	latest, err := mockRepo.GetLatest(ctx)
	require.NoError(t, err)
	assert.Equal(t, rate.ID, latest.ID)
	assert.Equal(t, 95.0, latest.Ask)
}

func TestRateService_GetRates_ClientError(t *testing.T) {
	mockClient := &mockBinanceClient{Err: errors.New("network down")}
	calc := calculator.New()
	mockRepo := repository.NewMockRateRepository()
	logger := zap.NewNop()

	svc := service.New(mockClient, calc, mockRepo, logger, service.Config{})

	_, err := svc.GetRates(context.Background(), domain.CalculationParams{
		Method: domain.MethodTopN,
		N:      1,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch order book")
	assert.Len(t, mockRepo.Rates, 0)
}

func TestRateService_GetRates_CalculationError(t *testing.T) {
	mockClient := &mockBinanceClient{OrderBook: &domain.OrderBook{}}
	calc := calculator.New()
	mockRepo := repository.NewMockRateRepository()
	logger := zap.NewNop()

	svc := service.New(mockClient, calc, mockRepo, logger, service.Config{})

	_, err := svc.GetRates(context.Background(), domain.CalculationParams{
		Method: domain.MethodTopN,
		N:      1,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "calculation error")
	assert.Len(t, mockRepo.Rates, 0)
}

func TestRateService_GetRates_RepoSaveError(t *testing.T) {
	mockClient := &mockBinanceClient{OrderBook: sampleOrderBook()}
	calc := calculator.New()
	mockRepo := repository.NewMockRateRepository()
	mockRepo.SaveErr = errors.New("disk full")
	logger := zap.NewNop()

	svc := service.New(mockClient, calc, mockRepo, logger, service.Config{})

	_, err := svc.GetRates(context.Background(), domain.CalculationParams{
		Method: domain.MethodTopN,
		N:      1,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database save error")
}

func TestRateService_HealthCheck(t *testing.T) {
	mockClient := &mockBinanceClient{OrderBook: sampleOrderBook()}
	calc := calculator.New()
	logger := zap.NewNop()

	t.Run("Healthy", func(t *testing.T) {
		mockRepo := repository.NewMockRateRepository()
		svc := service.New(mockClient, calc, mockRepo, logger, service.Config{})

		ok, err := svc.HealthCheck(context.Background())
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("Unhealthy", func(t *testing.T) {
		mockRepo := repository.NewMockRateRepository()
		mockRepo.SaveErr = errors.New("db disconnected")
		svc := service.New(mockClient, calc, mockRepo, logger, service.Config{})

		ok, err := svc.HealthCheck(context.Background())
		require.Error(t, err)
		assert.False(t, ok)
	})
}
