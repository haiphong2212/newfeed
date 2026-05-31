package platform

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
)

func NewFiber(service string) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      service,
		ErrorHandler: FiberErrorHandler,
	})
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
