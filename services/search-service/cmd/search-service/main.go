package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	searchv1 "github.com/newfeed/community-news/gen/search/v1"
	searchgrpc "github.com/newfeed/community-news/services/search-service/internal/search/delivery/grpc"
	"github.com/newfeed/community-news/services/search-service/internal/search/domain"
	"github.com/newfeed/community-news/services/search-service/internal/search/repository"
	"github.com/newfeed/community-news/services/search-service/internal/search/usecase"
	"github.com/newfeed/community-news/services/shared/platform"
	"google.golang.org/grpc"
)

func main() {
	cfg := platform.Load("search-service", ":8004", ":50056")
	logger := platform.Logger(cfg.ServiceName, cfg.Env)
	db, err := platform.NewPostgres(context.Background(), cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	es, err := platform.NewElasticsearch(cfg.ElasticsearchURL)
	if err != nil {
		log.Fatal(err)
	}
	searchRepo := repository.NewElasticsearchRepository(es)
	if err := searchRepo.EnsureArticleIndex(context.Background()); err != nil {
		log.Fatal(err)
	}
	service := usecase.NewService(searchRepo)
	grpcServer, err := platform.StartGRPC(cfg.GRPCAddr, cfg.ServiceName, logger, func(server *grpc.Server) {
		searchv1.RegisterSearchServiceServer(server, searchgrpc.NewServer(service))
	})
	if err != nil {
		log.Fatal(err)
	}
	defer grpcServer.GracefulStop()
	consumer, err := platform.NewRabbitConsumer(cfg.RabbitMQURL, db)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()
	if err := consumer.Consume(context.Background(), "search-service.article-published", "article.published", service.HandleEvent); err != nil {
		log.Fatal(err)
	}

	app := platform.NewFiber(cfg.ServiceName)
	app.Post("/v1/search/articles", func(c *fiber.Ctx) error {
		var doc domain.Document
		if err := c.BodyParser(&doc); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
		}
		if err := service.IndexArticle(context.Background(), doc); err != nil {
			return err
		}
		return c.SendStatus(fiber.StatusAccepted)
	})
	app.Get("/v1/search/articles", func(c *fiber.Ctx) error {
		query := domain.Query{
			Text:     c.Query("q"),
			Title:    c.Query("title"),
			Tag:      c.Query("tag"),
			Category: c.Query("category"),
			Limit:    c.QueryInt("limit", 20),
			Cursor:   c.Query("cursor"),
		}
		results, err := service.Search(c.UserContext(), query)
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{"results": results})
	})
	if err := platform.ListenFiber(app, cfg.HTTPAddr, logger); err != nil {
		log.Fatal(err)
	}
}
