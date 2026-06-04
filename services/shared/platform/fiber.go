package platform

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/newfeed/community-news/services/shared/env"
)

func NewFiber(service string) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      service,
		BodyLimit:    env.Int("HTTP_BODY_LIMIT_BYTES", 64*1024*1024),
		ErrorHandler: FiberErrorHandler,
	})
	app.Use(helmet.New(helmet.Config{
		XSSProtection:             "1; mode=block",
		ContentTypeNosniff:        "nosniff",
		XFrameOptions:             "DENY",
		ReferrerPolicy:            "no-referrer",
		CrossOriginEmbedderPolicy: "require-corp",
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginResourcePolicy: "same-origin",
	}))
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"service": service, "status": "ok"})
	})
	return app
}

func FiberErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if typed, ok := err.(*fiber.Error); ok {
		code = typed.Code
	}
	return c.Status(code).JSON(fiber.Map{"error": err.Error()})
}

func ListenFiber(app *fiber.App, addr string, logger *slog.Logger) error {
	errCh := make(chan error, 1)
	go func() {
		logger.Info("fiber server started", "addr", addr)
		errCh <- app.Listen(addr)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-stop:
		logger.Info("fiber server stopping", "addr", addr)
		return app.ShutdownWithContext(context.Background())
	}
}
