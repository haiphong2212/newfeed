package main

import (
	"context"
	"log"

	chathttp "github.com/newfeed/community-news/services/chat-service/internal/room/delivery/http"
	chatws "github.com/newfeed/community-news/services/chat-service/internal/room/delivery/websocket"
	"github.com/newfeed/community-news/services/chat-service/internal/room/repository"
	"github.com/newfeed/community-news/services/chat-service/internal/room/usecase"
	"github.com/newfeed/community-news/services/shared/platform"
)

func main() {
	ctx := context.Background()
	cfg := platform.Load("chat-service", ":8006", ":50054")
	logger := platform.Logger(cfg.ServiceName, cfg.Env)
	grpcServer, err := platform.StartGRPCHealth(cfg.GRPCAddr, cfg.ServiceName, logger)
	if err != nil {
		log.Fatal(err)
	}
	defer grpcServer.GracefulStop()
	db, err := platform.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	redisClient, err := platform.NewRedis(ctx, cfg.RedisAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()
	app := platform.NewFiber(cfg.ServiceName)
	repo := repository.NewPostgresRepository(db)
	rooms := usecase.NewService(repo)
	chathttp.NewHandler(rooms).RegisterRoutes(app)
	hub := chatws.NewHub(repo, redisClient)
	hub.RegisterRoutes(app)
	if err := platform.ListenFiber(app, cfg.HTTPAddr, logger); err != nil {
		log.Fatal(err)
	}
}
