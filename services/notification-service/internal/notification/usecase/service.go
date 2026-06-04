package usecase

import (
	"context"
	"fmt"

	"github.com/newfeed/community-news/services/notification-service/internal/notification/domain"
	"github.com/newfeed/community-news/services/shared/platform"
)

type Repository interface {
	FollowersForTopic(ctx context.Context, topic string) ([]string, error)
	Save(ctx context.Context, notification domain.Notification, eventID string) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (Service) FromEvent(event domain.EventName, userID string) domain.Notification {
	return domain.Notification{UserID: userID, Type: event}
}

func (s Service) HandleEvent(ctx context.Context, event platform.EventEnvelope) error {
	switch domain.EventName(event.EventName) {
	case domain.EventArticlePublished:
		return s.articlePublished(ctx, event)
	default:
		return nil
	}
}

func (s Service) articlePublished(ctx context.Context, event platform.EventEnvelope) error {
	category := text(event.Payload["category"])
	title := text(event.Payload["title"])
	if category == "" || title == "" {
		return fmt.Errorf("article published event missing category or title")
	}
	userIDs, err := s.repo.FollowersForTopic(ctx, category)
	if err != nil {
		return err
	}
	for _, userID := range userIDs {
		if err := s.repo.Save(ctx, domain.Notification{
			UserID: userID,
			Type:   domain.EventArticlePublished,
			Title:  "New article in " + category,
			Body:   title,
		}, event.EventID); err != nil {
			return err
		}
	}
	return nil
}

func text(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}
