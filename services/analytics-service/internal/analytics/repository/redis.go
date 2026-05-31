package repository

import (
	"context"

	"github.com/newfeed/community-news/services/analytics-service/internal/analytics/domain"
	"github.com/redis/go-redis/v9"
)

const TrendingArticlesKey = "trending_articles"

type RedisRepository struct {
	redis *redis.Client
}

func NewRedisRepository(redis *redis.Client) *RedisRepository {
	return &RedisRepository{redis: redis}
}

func (r *RedisRepository) SaveTrendingScore(ctx context.Context, metrics domain.ArticleMetrics) error {
	return r.redis.ZAdd(ctx, TrendingArticlesKey, redis.Z{
		Score:  metrics.TrendingScore(),
		Member: metrics.ArticleID,
	}).Err()
}

func (r *RedisRepository) TopTrending(ctx context.Context, limit int64) ([]string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return r.redis.ZRevRange(ctx, TrendingArticlesKey, 0, limit-1).Result()
}
