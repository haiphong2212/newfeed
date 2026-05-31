package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	userv1 "github.com/newfeed/community-news/gen/user/v1"
	"github.com/newfeed/community-news/services/shared/platform"
	profilegrpc "github.com/newfeed/community-news/services/user-service/internal/profile/delivery/grpc"
	"github.com/newfeed/community-news/services/user-service/internal/profile/domain"
	"github.com/newfeed/community-news/services/user-service/internal/profile/repository"
	"github.com/newfeed/community-news/services/user-service/internal/profile/usecase"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	cfg := platform.Load("user-service", ":8002", ":50052")
	logger := platform.Logger(cfg.ServiceName, cfg.Env)
	db, err := platform.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	profiles := usecase.NewService(repository.NewPostgresRepository(db))
	grpcServer, err := platform.StartGRPC(cfg.GRPCAddr, cfg.ServiceName, logger, func(server *grpc.Server) {
		userv1.RegisterUserServiceServer(server, profilegrpc.NewServer(profiles))
	})
	if err != nil {
		log.Fatal(err)
	}
	defer grpcServer.GracefulStop()
	app := platform.NewFiber(cfg.ServiceName)
	app.Put("/v1/users/:id/profile", func(c *fiber.Ctx) error {
		var profile domain.Profile
		if err := c.BodyParser(&profile); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
		}
		profile.UserID = c.Params("id")
		if err := profiles.UpsertProfile(c.UserContext(), profile); err != nil {
			return err
		}
		return c.SendStatus(fiber.StatusNoContent)
	})
	app.Post("/v1/users/:id/following/:target_id", func(c *fiber.Ctx) error {
		if err := profiles.FollowUser(c.UserContext(), c.Params("id"), c.Params("target_id")); err != nil {
			return err
		}
		return c.SendStatus(fiber.StatusNoContent)
	})
	app.Post("/v1/users/:id/topics/:topic", func(c *fiber.Ctx) error {
		if err := profiles.FollowTopicByUser(c.UserContext(), c.Params("id"), c.Params("topic")); err != nil {
			return err
		}
		return c.SendStatus(fiber.StatusNoContent)
	})
	if err := platform.ListenFiber(app, cfg.HTTPAddr, logger); err != nil {
		log.Fatal(err)
	}
}
