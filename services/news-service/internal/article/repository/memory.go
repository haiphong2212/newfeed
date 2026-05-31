package repository

import (
	"context"
	"sync"

	"github.com/newfeed/community-news/services/news-service/internal/article/domain"
)

type MemoryRepository struct {
	mu       sync.RWMutex
	articles map[string]domain.Article
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{articles: map[string]domain.Article{}}
}

func (r *MemoryRepository) Save(_ context.Context, article domain.Article) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.articles[article.ID] = article
	return nil
}

func (r *MemoryRepository) FindByID(_ context.Context, id string) (*domain.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	article, ok := r.articles[id]
	if !ok {
		return nil, domain.ErrInvalidArticle
	}
	return &article, nil
}
