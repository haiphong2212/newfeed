package usecase

import (
	"context"
	"strconv"

	"github.com/newfeed/community-news/services/analytics-service/internal/analytics/domain"
	"github.com/newfeed/community-news/services/shared/platform"
)

type TrendingRepository interface {
	SaveTrendingScore(ctx context.Context, metrics domain.ArticleMetrics) error
	TopTrending(ctx context.Context, limit int64) ([]string, error)
}

type Service struct {
	trending TrendingRepository
}

func NewService(trending TrendingRepository) Service {
	return Service{trending: trending}
}

func (Service) Score(metrics domain.ArticleMetrics) float64 {
	return metrics.TrendingScore()
}

func (s Service) Track(ctx context.Context, metrics domain.ArticleMetrics) error {
	return s.trending.SaveTrendingScore(ctx, metrics)
}

func (s Service) Trending(ctx context.Context, limit int64) ([]string, error) {
	return s.trending.TopTrending(ctx, limit)
}

func (s Service) HandleEvent(ctx context.Context, event platform.EventEnvelope) error {
	metrics := domain.ArticleMetrics{ArticleID: text(event.Payload["article_id"])}
	switch event.EventName {
	case "ArticlePublished":
		metrics.ViewCount = number(event.Payload["view_count"])
		metrics.CommentCount = number(event.Payload["comment_count"])
		metrics.ChatActivity = number(event.Payload["chat_activity"])
		metrics.ReactionCount = number(event.Payload["reaction_count"])
	case "CommentCreated":
		metrics.CommentCount = number(event.Payload["comment_count"])
	case "ChatMessageCreated":
		metrics.ChatActivity = number(event.Payload["chat_activity"])
	default:
		return nil
	}
	if metrics.ArticleID == "" {
		return nil
	}
	return s.Track(ctx, metrics)
}

func text(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

func number(value any) int64 {
	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case int64:
		return typed
	case string:
		n, _ := strconv.ParseInt(typed, 10, 64)
		return n
	default:
		return 0
	}
}
