package main

import (
	"context"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/newfeed/community-news/services/analytics-service/internal/analytics/domain"
	"github.com/newfeed/community-news/services/analytics-service/internal/analytics/repository"
	"github.com/newfeed/community-news/services/analytics-service/internal/analytics/usecase"
	"github.com/newfeed/community-news/services/shared/platform"
)

func main() {
	ctx := context.Background()
	cfg := platform.Load("analytics-service", ":8007", ":50058")
	logger := platform.Logger(cfg.ServiceName, cfg.Env)
	grpcServer, err := platform.StartGRPCHealth(cfg.GRPCAddr, cfg.ServiceName, logger)
	if err != nil {
		log.Fatal(err)
	}
	defer grpcServer.GracefulStop()
	redisClient, err := platform.NewRedis(ctx, cfg.RedisAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()
	db, err := platform.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	service := usecase.NewService(repository.NewRedisRepository(redisClient))
	consumer, err := platform.NewRabbitConsumer(cfg.RabbitMQURL, db)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()
	if err := consumer.Consume(ctx, "analytics-service.article-published", "article.published", service.HandleEvent); err != nil {
		log.Fatal(err)
	}

	app := platform.NewFiber(cfg.ServiceName)
	app.Post("/v1/analytics/articles/:id/metrics", func(c *fiber.Ctx) error {
		var metrics domain.ArticleMetrics
		if err := c.BodyParser(&metrics); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
		}
		metrics.ArticleID = c.Params("id")
		if err := service.Track(c.UserContext(), metrics); err != nil {
			return err
		}
		return c.JSON(fiber.Map{"article_id": metrics.ArticleID, "score": service.Score(metrics)})
	})
	app.Get("/v1/analytics/trending", func(c *fiber.Ctx) error {
		limit, _ := strconv.ParseInt(c.Query("limit", "20"), 10, 64)
		ids, err := service.Trending(c.UserContext(), limit)
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{"article_ids": ids})
	})
	if err := platform.ListenFiber(app, cfg.HTTPAddr, logger); err != nil {
		log.Fatal(err)
	}
}
