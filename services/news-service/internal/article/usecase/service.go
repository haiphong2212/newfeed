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
	ListByAuthor(ctx context.Context, authorID string, limit int, cursor time.Time) ([]domain.Article, error)
	CreateComment(ctx context.Context, comment domain.Comment) (domain.Comment, error)
	ListComments(ctx context.Context, articleID string, limit int, cursor time.Time) ([]domain.Comment, error)
	ShareArticle(ctx context.Context, share domain.Share) (domain.Share, error)
	ListSharesByUser(ctx context.Context, userID string, limit int, cursor time.Time) ([]domain.Share, error)
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

func (s *Service) ListPublishedByAuthor(ctx context.Context, authorID string, limit int, cursor time.Time) ([]domain.Article, error) {
	return s.repo.ListByAuthor(ctx, authorID, limit, cursor)
}

func (s *Service) CreateComment(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
	comment.Body = strings.TrimSpace(comment.Body)
	if comment.ArticleID == "" || comment.UserID == "" || comment.Body == "" {
		return domain.Comment{}, domain.ErrInvalidArticle
	}
	return s.repo.CreateComment(ctx, comment)
}

func (s *Service) ListComments(ctx context.Context, articleID string, limit int, cursor time.Time) ([]domain.Comment, error) {
	return s.repo.ListComments(ctx, articleID, limit, cursor)
}

func (s *Service) ShareArticle(ctx context.Context, share domain.Share) (domain.Share, error) {
	share.Caption = strings.TrimSpace(share.Caption)
	if share.ArticleID == "" || share.UserID == "" {
		return domain.Share{}, domain.ErrInvalidArticle
	}
	if share.Visibility == "" {
		share.Visibility = "public"
	}
	return s.repo.ShareArticle(ctx, share)
}

func (s *Service) ListSharesByUser(ctx context.Context, userID string, limit int, cursor time.Time) ([]domain.Share, error) {
	return s.repo.ListSharesByUser(ctx, userID, limit, cursor)
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
