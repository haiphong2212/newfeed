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
	}
	redis *redis.Client
	ttl   time.Duration
}

func NewCachedRepository(next interface {
	Save(context.Context, domain.Article) error
	FindByID(context.Context, string) (*domain.Article, error)
}, redis *redis.Client, ttl time.Duration) *CachedRepository {
	return &CachedRepository{next: next, redis: redis, ttl: ttl}
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
