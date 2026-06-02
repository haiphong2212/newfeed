package main

import (
	"context"
	"log"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/newfeed/community-news/services/media-service/internal/media/repository"
	mediastorage "github.com/newfeed/community-news/services/media-service/internal/media/storage"
	"github.com/newfeed/community-news/services/media-service/internal/media/usecase"
	"github.com/newfeed/community-news/services/shared/env"
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
	media := usecase.NewService(repository.NewPostgresRepository(db))
	store := buildStorage(cfg.ObjectRoot)

	app := platform.NewFiber(cfg.ServiceName)
	app.Post("/v1/media/upload", func(c *fiber.Ctx) error {
		ownerID := c.FormValue("owner_id")
		bucket := c.FormValue("bucket", env.String("MEDIA_BUCKET", env.String("R2_BUCKET", "community-news")))
		file, err := c.FormFile("file")
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "file is required")
		}
		source, err := file.Open()
		if err != nil {
			return err
		}
		defer source.Close()
		key := strings.Trim(filepath.ToSlash(filepath.Clean(file.Filename)), "/.\\")
		stored, err := store.Put(c.UserContext(), bucket, key, file.Header.Get("Content-Type"), source)
		if err != nil {
			return err
		}
		size := stored.Size
		if size == 0 {
			size = file.Size
		}
		object, err := media.SaveObject(c.UserContext(), ownerID, stored.Bucket, stored.Key, file.Header.Get("Content-Type"), size)
		if err != nil {
			return err
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"object": object, "url": stored.URL})
	})
	app.Get("/objects/:bucket/:key", func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(cfg.ObjectRoot, c.Params("bucket"), filepath.Clean(c.Params("key"))))
	})
	if err := platform.ListenFiber(app, cfg.HTTPAddr, logger); err != nil {
		log.Fatal(err)
	}
}

func buildStorage(localRoot string) mediastorage.ObjectStorage {
	if env.String("OBJECT_STORAGE_DRIVER", "local") == "r2" {
		return mediastorage.NewR2Storage(mediastorage.R2Config{
			Endpoint:        env.String("R2_ENDPOINT", ""),
			AccessKeyID:     env.String("R2_ACCESS_KEY_ID", ""),
			SecretAccessKey: env.String("R2_SECRET_ACCESS_KEY", ""),
			PublicURL:       env.String("R2_PUBLIC_URL", ""),
		})
	}
	return mediastorage.NewLocalStorage(localRoot, env.String("MEDIA_PUBLIC_URL", ""))
}
