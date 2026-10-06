package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	ratev1 "github.com/kazah/golang-task-03/gen/rate/v1"
)

// Server wraps grpc.Server together with the standard health service.
type Server struct {
	server       *grpc.Server
	healthServer *health.Server
	logger       *zap.Logger
}

// NewServer creates the gRPC server and registers all services.
func NewServer(handler *Handler, logger *zap.Logger) *Server {
	// stats handler adds OpenTelemetry traces to every gRPC call
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)

	ratev1.RegisterRateServiceServer(grpcServer, handler)
	// reflection allows to use grpcurl without proto files
	reflection.Register(grpcServer)

	return &Server{
		server:       grpcServer,
		healthServer: healthServer,
		logger:       logger,
	}
}

// Start listens on the port and blocks until the server is stopped.
func (s *Server) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.logger.Info("gRPC server listening", zap.String("address", addr))

	if err := s.server.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("gRPC server error: %w", err)
	}

	return nil
}

// Stop shuts the server down gracefully. If ctx expires, the server is stopped by force.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("shutting down gRPC server...")
	s.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)

	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		s.logger.Info("gRPC server stopped")
		return nil
	case <-ctx.Done():
		s.logger.Warn("gRPC server stop timeout, forcing kill")
		s.server.Stop()
		return ctx.Err()
	}
}
