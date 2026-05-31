package platform

import (
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func StartGRPCHealth(addr, service string, logger *slog.Logger) (*grpc.Server, error) {
	return StartGRPC(addr, service, logger, nil)
}

func StartGRPC(addr, service string, logger *slog.Logger, register func(*grpc.Server)) (*grpc.Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	server := grpc.NewServer()
	healthServer := health.NewServer()
	healthServer.SetServingStatus(service, healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(server, healthServer)
	if register != nil {
		register(server)
	}
	go func() {
		logger.Info("grpc server started", "addr", addr)
		if err := server.Serve(listener); err != nil {
			logger.Error("grpc server stopped", "error", err)
		}
	}()
	return server, nil
}
