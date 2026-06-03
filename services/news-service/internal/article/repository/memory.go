package repository

import (
	"context"
	"sync"
	"time"

	"github.com/newfeed/community-news/services/news-service/internal/article/domain"
)

type MemoryRepository struct {
	mu       sync.RWMutex
	articles map[string]domain.Article
}

func (r *MemoryRepository) ListByAuthor(_ context.Context, authorID string, limit int, cursor time.Time) ([]domain.Article, error) {
	return []domain.Article{}, nil
}

func (r *MemoryRepository) CreateComment(_ context.Context, comment domain.Comment) (domain.Comment, error) {
	return comment, nil
}

func (r *MemoryRepository) ListComments(_ context.Context, articleID string, limit int, cursor time.Time) ([]domain.Comment, error) {
	return []domain.Comment{}, nil
}

func (r *MemoryRepository) ShareArticle(_ context.Context, share domain.Share) (domain.Share, error) {
	return share, nil
}

func (r *MemoryRepository) ListSharesByUser(_ context.Context, userID string, limit int, cursor time.Time) ([]domain.Share, error) {
	return []domain.Share{}, nil
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
