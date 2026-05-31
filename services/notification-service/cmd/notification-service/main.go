package main

import (
	"context"
	"log"

	"github.com/newfeed/community-news/services/notification-service/internal/notification/repository"
	"github.com/newfeed/community-news/services/notification-service/internal/notification/usecase"
	"github.com/newfeed/community-news/services/shared/platform"
)

func main() {
	ctx := context.Background()
	cfg := platform.Load("notification-service", ":8005", ":50055")
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
	consumer, err := platform.NewRabbitConsumer(cfg.RabbitMQURL, db)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()
	notifications := usecase.NewService(repository.NewPostgresRepository(db))
	if err := consumer.Consume(ctx, "notification-service.article-published", "article.published", notifications.HandleEvent); err != nil {
		log.Fatal(err)
	}
	app := platform.NewFiber(cfg.ServiceName)
	if err := platform.ListenFiber(app, cfg.HTTPAddr, logger); err != nil {
		log.Fatal(err)
	}
}
