package events

import (
	"context"
	"log/slog"
)

type LogPublisher struct {
	logger *slog.Logger
}

func NewLogPublisher(logger *slog.Logger) *LogPublisher {
	return &LogPublisher{logger: logger}
}

func (p *LogPublisher) Publish(_ context.Context, eventName string, payload any) error {
	p.logger.Info("event published", "event", eventName, "payload", payload)
	return nil
}
