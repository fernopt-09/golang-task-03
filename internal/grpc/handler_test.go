package grpc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	ratev1 "github.com/kazah/golang-task-03/gen/rate/v1"
	"github.com/kazah/golang-task-03/internal/domain"
	internalgrpc "github.com/kazah/golang-task-03/internal/grpc"
)

type mockService struct {
	rate *domain.Rate
	err  error
	ok   bool
}

func (m *mockService) GetRates(_ context.Context, _ domain.CalculationParams) (*domain.Rate, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.rate, nil
}

func (m *mockService) HealthCheck(_ context.Context) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.ok, nil
}

func TestHandler_GetRates(t *testing.T) {
	logger := zap.NewNop()

	t.Run("Success TopN", func(t *testing.T) {
		svc := &mockService{
			rate: &domain.Rate{
				ID:        1,
				Ask:       95.0,
				Bid:       94.0,
				Timestamp: time.Unix(1712000000, 0),
				Method:    domain.MethodTopN,
			},
		}

		handler := internalgrpc.NewHandler(svc, logger)
		n := uint32(2)
		req := &ratev1.GetRatesRequest{
			Method: ratev1.CalculationMethod_CALCULATION_METHOD_TOP_N,
			N:      &n,
		}

		resp, err := handler.GetRates(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 95.0, resp.Ask)
		assert.Equal(t, 94.0, resp.Bid)
		assert.Equal(t, int64(1712000000), resp.Timestamp)
		assert.Equal(t, ratev1.CalculationMethod_CALCULATION_METHOD_TOP_N, resp.Method)
		assert.Equal(t, uint32(2), resp.N)
	})

	t.Run("Success AvgNM", func(t *testing.T) {
		svc := &mockService{
			rate: &domain.Rate{
				ID:        2,
				Ask:       96.5,
				Bid:       93.5,
				Timestamp: time.Unix(1712000000, 0),
				Method:    domain.MethodAvgNM,
			},
		}

		handler := internalgrpc.NewHandler(svc, logger)
		n := uint32(1)
		m := uint32(3)
		req := &ratev1.GetRatesRequest{
			Method: ratev1.CalculationMethod_CALCULATION_METHOD_AVG_NM,
			N:      &n,
			M:      &m,
		}

		resp, err := handler.GetRates(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, 96.5, resp.Ask)
		assert.Equal(t, 93.5, resp.Bid)
		assert.Equal(t, ratev1.CalculationMethod_CALCULATION_METHOD_AVG_NM, resp.Method)
		assert.Equal(t, uint32(1), resp.N)
		assert.Equal(t, uint32(3), resp.M)
	})

	t.Run("Service Error", func(t *testing.T) {
		svc := &mockService{
			err: errors.New("upstream failed"),
		}

		handler := internalgrpc.NewHandler(svc, logger)
		_, err := handler.GetRates(context.Background(), &ratev1.GetRatesRequest{})
		require.Error(t, err)
	})
}

func TestHandler_HealthCheck(t *testing.T) {
	logger := zap.NewNop()

	t.Run("Healthy", func(t *testing.T) {
		svc := &mockService{ok: true}
		handler := internalgrpc.NewHandler(svc, logger)

		resp, err := handler.HealthCheck(context.Background(), &ratev1.HealthCheckRequest{})
		require.NoError(t, err)
		assert.Equal(t, ratev1.HealthCheckResponse_SERVING_STATUS_SERVING, resp.Status)
	})

	t.Run("Unhealthy", func(t *testing.T) {
		svc := &mockService{err: errors.New("database unreachable")}
		handler := internalgrpc.NewHandler(svc, logger)

		resp, err := handler.HealthCheck(context.Background(), &ratev1.HealthCheckRequest{})
		require.Error(t, err)
		assert.Equal(t, ratev1.HealthCheckResponse_SERVING_STATUS_NOT_SERVING, resp.Status)
	})
}
