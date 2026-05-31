package events

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/newfeed/community-news/services/shared/platform"
)

type RabbitArticlePublisher struct {
	rabbit *platform.RabbitPublisher
}

func NewRabbitArticlePublisher(rabbit *platform.RabbitPublisher) *RabbitArticlePublisher {
	return &RabbitArticlePublisher{rabbit: rabbit}
}

func (p *RabbitArticlePublisher) Publish(ctx context.Context, eventName string, payload any) error {
	routingKey := "article.published"
	if eventName != "ArticlePublished" {
		routingKey = "article.event"
	}
	body, ok := payload.(map[string]any)
	if !ok {
		body = map[string]any{"data": payload}
	}
	return p.rabbit.Publish(ctx, routingKey, platform.EventEnvelope{
		EventID:    newID(),
		EventName:  eventName,
		OccurredAt: time.Now().UTC(),
		Payload:    body,
	})
}

func newID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return hex.EncodeToString(buf[0:4]) + "-" + hex.EncodeToString(buf[4:6]) + "-" + hex.EncodeToString(buf[6:8]) + "-" + hex.EncodeToString(buf[8:10]) + "-" + hex.EncodeToString(buf[10:])
}
