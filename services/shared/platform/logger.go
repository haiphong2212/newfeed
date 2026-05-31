package platform

import (
	"log/slog"
	"os"
)

func Logger(service, env string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", service, "env", env)
}
