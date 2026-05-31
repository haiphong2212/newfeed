package main

import (
	"context"
	"log"
	"time"

	newsv1 "github.com/newfeed/community-news/gen/news/v1"
	articlegrpc "github.com/newfeed/community-news/services/news-service/internal/article/delivery/grpc"
	articlehttp "github.com/newfeed/community-news/services/news-service/internal/article/delivery/http"
	"github.com/newfeed/community-news/services/news-service/internal/article/repository"
	"github.com/newfeed/community-news/services/news-service/internal/article/usecase"
	"github.com/newfeed/community-news/services/news-service/internal/platform/events"
	"github.com/newfeed/community-news/services/shared/platform"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	cfg := platform.Load("news-service", ":8003", ":50053")
	logger := platform.Logger(cfg.ServiceName, cfg.Env)
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
	rabbit, err := platform.NewRabbitPublisher(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbit.Close()

	pgRepo := repository.NewPostgresRepository(db)
	cachedRepo := repository.NewCachedRepository(pgRepo, redisClient, 10*time.Minute)
	articles := usecase.NewService(cachedRepo, events.NewRabbitArticlePublisher(rabbit))
	grpcServer, err := platform.StartGRPC(cfg.GRPCAddr, cfg.ServiceName, logger, func(server *grpc.Server) {
		newsv1.RegisterNewsServiceServer(server, articlegrpc.NewServer(articles))
	})
	if err != nil {
		log.Fatal(err)
	}
	defer grpcServer.GracefulStop()
	handler := articlehttp.NewHandler(articles)
	app := platform.NewFiber(cfg.ServiceName)
	handler.RegisterRoutes(app)
	if err := platform.ListenFiber(app, cfg.HTTPAddr, logger); err != nil {
		log.Fatal(err)
	}
}
