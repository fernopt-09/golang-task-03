package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ratev1 "github.com/kazah/golang-task-03/gen/rate/v1"
	"github.com/kazah/golang-task-03/internal/domain"
	"github.com/kazah/golang-task-03/internal/service"
)

// Handler implements the RateService gRPC API.
type Handler struct {
	ratev1.UnimplementedRateServiceServer
	service service.RateService
	logger  *zap.Logger
}

// NewHandler creates the handler.
func NewHandler(svc service.RateService, logger *zap.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger,
	}
}

// GetRates returns the current USDT rate. topN with position 1 is used by default.
func (h *Handler) GetRates(ctx context.Context, req *ratev1.GetRatesRequest) (*ratev1.GetRatesResponse, error) {
	method := domain.MethodTopN
	respMethod := ratev1.CalculationMethod_CALCULATION_METHOD_TOP_N
	var n uint32 = 1
	var m uint32 = 1

	if req != nil {
		if req.Method == ratev1.CalculationMethod_CALCULATION_METHOD_AVG_NM {
			method = domain.MethodAvgNM
			respMethod = ratev1.CalculationMethod_CALCULATION_METHOD_AVG_NM
		}

		if req.N != nil && *req.N > 0 {
			n = *req.N
		}
		if req.M != nil && *req.M > 0 {
			m = *req.M
		}
	}

	rate, err := h.service.GetRates(ctx, domain.CalculationParams{
		Method: method,
		N:      n,
		M:      m,
	})
	if err != nil {
		h.logger.Error("GetRates failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get rates: %v", err)
	}

	return &ratev1.GetRatesResponse{
		Ask:       rate.Ask,
		Bid:       rate.Bid,
		Timestamp: rate.Timestamp.Unix(),
		Method:    respMethod,
		N:         n,
		M:         m,
	}, nil
}

// HealthCheck reports SERVING only if the database is reachable.
func (h *Handler) HealthCheck(ctx context.Context, _ *ratev1.HealthCheckRequest) (*ratev1.HealthCheckResponse, error) {
	healthy, err := h.service.HealthCheck(ctx)
	if err != nil || !healthy {
		return &ratev1.HealthCheckResponse{
			Status: ratev1.HealthCheckResponse_SERVING_STATUS_NOT_SERVING,
		}, status.Errorf(codes.Unavailable, "service unhealthy: %v", err)
	}

	return &ratev1.HealthCheckResponse{
		Status: ratev1.HealthCheckResponse_SERVING_STATUS_SERVING,
	}, nil
}
