package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/newfeed/community-news/services/news-service/internal/article/domain"
)

type Repository interface {
	Save(ctx context.Context, article domain.Article) error
	FindByID(ctx context.Context, id string) (*domain.Article, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, eventName string, payload any) error
}

type Service struct {
	repo   Repository
	events EventPublisher
}

func NewService(repo Repository, events EventPublisher) *Service {
	return &Service{repo: repo, events: events}
}

func (s *Service) Publish(ctx context.Context, article domain.Article) error {
	if err := article.ValidateForPublish(); err != nil {
		return err
	}
	now := time.Now().UTC()
	if article.Slug == "" {
		article.Slug = slugify(article.Title)
	}
	article.Status = domain.StatusPublished
	article.UpdatedAt = now
	if article.CreatedAt.IsZero() {
		article.CreatedAt = now
	}
	if err := s.repo.Save(ctx, article); err != nil {
		return err
	}
	return s.events.Publish(ctx, "ArticlePublished", map[string]any{
		"article_id":           article.ID,
		"title":                article.Title,
		"content":              article.Content,
		"category":             article.Category,
		"tags":                 article.Tags,
		"discussion_room_name": article.DiscussionRoomName(),
	})
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		if r == ' ' || r == '-' || r == '_' {
			return '-'
		}
		return -1
	}, value)
	for strings.Contains(value, "--") {
		value = strings.ReplaceAll(value, "--", "-")
	}
	return strings.Trim(value, "-")
}
