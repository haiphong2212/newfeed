package main

import (
	"log"

	gatewayhttp "github.com/newfeed/community-news/services/api-gateway/internal/gateway/delivery/http"
	gatewaygrpc "github.com/newfeed/community-news/services/api-gateway/internal/gateway/grpc"
	"github.com/newfeed/community-news/services/shared/env"
	"github.com/newfeed/community-news/services/shared/platform"
)

func main() {
	cfg := platform.Load("api-gateway", ":8000", ":50050")
	logger := platform.Logger(cfg.ServiceName, cfg.Env)
	grpcServer, err := platform.StartGRPCHealth(cfg.GRPCAddr, cfg.ServiceName, logger)
	if err != nil {
		log.Fatal(err)
	}
	defer grpcServer.GracefulStop()
	targets := map[string]string{
		"auth-service":         env.String("AUTH_SERVICE_URL", "auth-service:50051"),
		"user-service":         env.String("USER_SERVICE_URL", "user-service:50052"),
		"news-service":         env.String("NEWS_SERVICE_URL", "news-service:50053"),
		"chat-service":         env.String("CHAT_SERVICE_URL", "chat-service:50054"),
		"notification-service": env.String("NOTIFICATION_SERVICE_URL", "notification-service:50055"),
		"search-service":       env.String("SEARCH_SERVICE_URL", "search-service:50056"),
		"media-service":        env.String("MEDIA_SERVICE_URL", "media-service:50057"),
		"analytics-service":    env.String("ANALYTICS_SERVICE_URL", "analytics-service:50058"),
	}
	clients, err := gatewaygrpc.NewClients(targets)
	if err != nil {
		log.Fatal(err)
	}
	defer clients.Close()
	httpTargets := gatewayhttp.HTTPTargets{
		User:   env.String("USER_HTTP_URL", "http://user-service:8002"),
		News:   env.String("NEWS_HTTP_URL", "http://news-service:8003"),
		Search: env.String("SEARCH_HTTP_URL", "http://search-service:8004"),
		Media:  env.String("MEDIA_HTTP_URL", "http://media-service:8008"),
	}
	handler := gatewayhttp.NewHandler(gatewaygrpc.NewHealthClient(targets), clients, httpTargets)
	app := platform.NewFiber(cfg.ServiceName)
	handler.RegisterRoutes(app)
	if err := platform.ListenFiber(app, cfg.HTTPAddr, logger); err != nil {
		log.Fatal(err)
	}
}
