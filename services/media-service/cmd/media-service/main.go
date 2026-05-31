package main

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/newfeed/community-news/services/media-service/internal/media/repository"
	"github.com/newfeed/community-news/services/media-service/internal/media/usecase"
	"github.com/newfeed/community-news/services/shared/platform"
)

func main() {
	ctx := context.Background()
	cfg := platform.Load("media-service", ":8008", ":50057")
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
	if err := os.MkdirAll(cfg.ObjectRoot, 0755); err != nil {
		log.Fatal(err)
	}
	media := usecase.NewService(repository.NewPostgresRepository(db))

	app := platform.NewFiber(cfg.ServiceName)
	app.Post("/v1/media/upload", func(c *fiber.Ctx) error {
		ownerID := c.FormValue("owner_id")
		bucket := c.FormValue("bucket", "community-news")
		file, err := c.FormFile("file")
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "file is required")
		}
		source, err := file.Open()
		if err != nil {
			return err
		}
		defer source.Close()
		key := strings.Trim(filepath.Clean(file.Filename), `\./`)
		targetDir := filepath.Join(cfg.ObjectRoot, bucket)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return err
		}
		targetPath := filepath.Join(targetDir, key)
		target, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		size, copyErr := io.Copy(target, source)
		closeErr := target.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		object, err := media.SaveObject(c.UserContext(), ownerID, bucket, key, file.Header.Get("Content-Type"), size)
		if err != nil {
			return err
		}
		return c.Status(fiber.StatusCreated).JSON(object)
	})
	app.Get("/objects/:bucket/:key", func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(cfg.ObjectRoot, c.Params("bucket"), filepath.Clean(c.Params("key"))))
	})
	if err := platform.ListenFiber(app, cfg.HTTPAddr, logger); err != nil {
		log.Fatal(err)
	}
}
