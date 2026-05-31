package usecase

import (
	"context"
	"fmt"

	"github.com/newfeed/community-news/services/search-service/internal/search/domain"
	"github.com/newfeed/community-news/services/shared/platform"
)

type Repository interface {
	IndexArticle(ctx context.Context, doc domain.Document) error
	Search(ctx context.Context, query domain.Query) ([]domain.Result, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return Service{repo: repo}
}

func (Service) Normalize(query domain.Query) domain.Query {
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 20
	}
	return query
}

func (s Service) IndexArticle(ctx context.Context, doc domain.Document) error {
	return s.repo.IndexArticle(ctx, doc)
}

func (s Service) Search(ctx context.Context, query domain.Query) ([]domain.Result, error) {
	return s.repo.Search(ctx, s.Normalize(query))
}

func (s Service) HandleEvent(ctx context.Context, event platform.EventEnvelope) error {
	if event.EventName != "ArticlePublished" {
		return nil
	}
	doc := domain.Document{
		ArticleID: text(event.Payload["article_id"]),
		Title:     text(event.Payload["title"]),
		Content:   text(event.Payload["content"]),
		Category:  text(event.Payload["category"]),
		Tags:      stringSlice(event.Payload["tags"]),
	}
	if doc.ArticleID == "" || doc.Title == "" {
		return fmt.Errorf("article published event missing article_id or title")
	}
	return s.IndexArticle(ctx, doc)
}

func text(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

func stringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
