package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type HealthClient struct {
	targets map[string]string
}

func NewHealthClient(targets map[string]string) HealthClient {
	return HealthClient{targets: targets}
}

func (c HealthClient) Check(ctx context.Context, service string) (string, error) {
	target := c.targets[service]
	if target == "" {
		return "unknown", nil
	}
	dialCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(dialCtx, target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", err
	}
	defer conn.Close()
	res, err := healthpb.NewHealthClient(conn).Check(dialCtx, &healthpb.HealthCheckRequest{Service: service})
	if err != nil {
		return "", err
	}
	return res.Status.String(), nil
}
