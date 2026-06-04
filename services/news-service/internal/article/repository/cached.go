package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/newfeed/community-news/services/news-service/internal/article/domain"
	"github.com/redis/go-redis/v9"
)

type CachedRepository struct {
	next interface {
		Save(context.Context, domain.Article) error
		FindByID(context.Context, string) (*domain.Article, error)
		ListByAuthor(context.Context, string, int, time.Time) ([]domain.Article, error)
		CreateComment(context.Context, domain.Comment) (domain.Comment, error)
		ListComments(context.Context, string, int, time.Time) ([]domain.Comment, error)
		ShareArticle(context.Context, domain.Share) (domain.Share, error)
		ListSharesByUser(context.Context, string, int, time.Time) ([]domain.Share, error)
	}
	redis *redis.Client
	ttl   time.Duration
}

func NewCachedRepository(next interface {
	Save(context.Context, domain.Article) error
	FindByID(context.Context, string) (*domain.Article, error)
	ListByAuthor(context.Context, string, int, time.Time) ([]domain.Article, error)
	CreateComment(context.Context, domain.Comment) (domain.Comment, error)
	ListComments(context.Context, string, int, time.Time) ([]domain.Comment, error)
	ShareArticle(context.Context, domain.Share) (domain.Share, error)
	ListSharesByUser(context.Context, string, int, time.Time) ([]domain.Share, error)
}, redis *redis.Client, ttl time.Duration) *CachedRepository {
	return &CachedRepository{next: next, redis: redis, ttl: ttl}
}

func (r *CachedRepository) ListByAuthor(ctx context.Context, authorID string, limit int, cursor time.Time) ([]domain.Article, error) {
	return r.next.ListByAuthor(ctx, authorID, limit, cursor)
}

func (r *CachedRepository) CreateComment(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
	return r.next.CreateComment(ctx, comment)
}

func (r *CachedRepository) ListComments(ctx context.Context, articleID string, limit int, cursor time.Time) ([]domain.Comment, error) {
	return r.next.ListComments(ctx, articleID, limit, cursor)
}

func (r *CachedRepository) ShareArticle(ctx context.Context, share domain.Share) (domain.Share, error) {
	return r.next.ShareArticle(ctx, share)
}

func (r *CachedRepository) ListSharesByUser(ctx context.Context, userID string, limit int, cursor time.Time) ([]domain.Share, error) {
	return r.next.ListSharesByUser(ctx, userID, limit, cursor)
}

func (r *CachedRepository) Save(ctx context.Context, article domain.Article) error {
	if err := r.next.Save(ctx, article); err != nil {
		return err
	}
	data, _ := json.Marshal(article)
	_ = r.redis.Set(ctx, "article:"+article.ID, data, r.ttl).Err()
	return nil
}

func (r *CachedRepository) FindByID(ctx context.Context, id string) (*domain.Article, error) {
	if data, err := r.redis.Get(ctx, "article:"+id).Bytes(); err == nil {
		var article domain.Article
		if json.Unmarshal(data, &article) == nil {
			return &article, nil
		}
	}
	article, err := r.next.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(article)
	_ = r.redis.Set(ctx, "article:"+id, data, r.ttl).Err()
	return article, nil
}
