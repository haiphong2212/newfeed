package main

import (
	"context"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

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
		files, err := uploadedFiles(c)
		if err != nil {
			return err
		}
		if err := validateUploadBatch(files); err != nil {
			return err
		}
		results := make([]fiber.Map, 0, len(files))
		for _, file := range files {
			source, err := file.Open()
			if err != nil {
				return err
			}
			key := objectKey(file.Filename)
			stored, putErr := store.Put(c.UserContext(), bucket, key, file.Header.Get("Content-Type"), source)
			closeErr := source.Close()
			if putErr != nil {
				return putErr
			}
			if closeErr != nil {
				return closeErr
			}
			size := stored.Size
			if size == 0 {
				size = file.Size
			}
			object, err := media.SaveObject(c.UserContext(), ownerID, stored.Bucket, stored.Key, file.Header.Get("Content-Type"), size)
			if err != nil {
				return err
			}
			results = append(results, fiber.Map{"object": object, "url": stored.URL})
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"objects": results})
	})
	app.Get("/objects/:bucket/*", func(c *fiber.Ctx) error {
		bucket := safePathPart(c.Params("bucket"))
		key := safePathPart(c.Params("*"))
		if bucket == "" || key == "" {
			return fiber.NewError(fiber.StatusBadRequest, "invalid object path")
		}
		return c.SendFile(filepath.Join(cfg.ObjectRoot, bucket, key))
	})
	if err := platform.ListenFiber(app, cfg.HTTPAddr, logger); err != nil {
		log.Fatal(err)
	}
}

func uploadedFiles(c *fiber.Ctx) ([]*multipart.FileHeader, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "multipart form is required")
	}
	files := append([]*multipart.FileHeader{}, form.File["files"]...)
	files = append(files, form.File["file"]...)
	if len(files) == 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "file is required")
	}
	return files, nil
}

func validateUploadBatch(files []*multipart.FileHeader) error {
	const maxImages = 3
	const maxVideoBytes = 20 * 1024 * 1024
	imageCount := 0
	for _, file := range files {
		contentType := file.Header.Get("Content-Type")
		switch {
		case strings.HasPrefix(contentType, "image/"):
			imageCount++
			if imageCount > maxImages {
				return fiber.NewError(fiber.StatusBadRequest, "maximum 3 images are allowed")
			}
		case strings.HasPrefix(contentType, "video/"):
			if file.Size > maxVideoBytes {
				return fiber.NewError(fiber.StatusBadRequest, "video files must be 20MB or smaller")
			}
		default:
			return fiber.NewError(fiber.StatusBadRequest, "only image and video uploads are allowed")
		}
	}
	return nil
}

func objectKey(filename string) string {
	name := strings.Trim(filepath.ToSlash(filepath.Clean(filename)), "/.\\")
	if name == "" {
		name = "upload"
	}
	return time.Now().UTC().Format("20060102/150405.000000000") + "-" + filepath.Base(name)
}

func safePathPart(value string) string {
	cleaned := strings.Trim(filepath.ToSlash(filepath.Clean(value)), "/")
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return ""
	}
	return cleaned
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
