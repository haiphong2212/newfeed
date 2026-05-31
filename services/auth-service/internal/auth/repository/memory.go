package repository

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/newfeed/community-news/services/auth-service/internal/auth/domain"
)

type MemoryUserRepository struct {
	mu      sync.RWMutex
	byID    map[string]*domain.User
	byEmail map[string]*domain.User
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{byID: map[string]*domain.User{}, byEmail: map[string]*domain.User{}}
}

func (r *MemoryUserRepository) Create(_ context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	email := strings.ToLower(user.Email)
	if _, exists := r.byEmail[email]; exists {
		return domain.ErrUserAlreadyExists
	}
	copy := *user
	r.byID[user.ID] = &copy
	r.byEmail[email] = &copy
	return nil
}

func (r *MemoryUserRepository) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.byEmail[strings.ToLower(email)]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	copy := *user
	return &copy, nil
}

func (r *MemoryUserRepository) FindByID(_ context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	copy := *user
	return &copy, nil
}

type MemoryRefreshTokenRepository struct {
	mu       sync.RWMutex
	byHash   map[string]*domain.RefreshToken
	idToHash map[string]string
}

func NewMemoryRefreshTokenRepository() *MemoryRefreshTokenRepository {
	return &MemoryRefreshTokenRepository{byHash: map[string]*domain.RefreshToken{}, idToHash: map[string]string{}}
}

func (r *MemoryRefreshTokenRepository) Save(_ context.Context, token domain.RefreshToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := token
	r.byHash[token.TokenHash] = &copy
	r.idToHash[token.ID] = token.TokenHash
	return nil
}

func (r *MemoryRefreshTokenRepository) FindByHash(_ context.Context, hash string) (*domain.RefreshToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	token, ok := r.byHash[hash]
	if !ok {
		return nil, domain.ErrRefreshNotFound
	}
	copy := *token
	return &copy, nil
}

func (r *MemoryRefreshTokenRepository) Revoke(_ context.Context, id string, revokedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	hash, ok := r.idToHash[id]
	if !ok {
		return domain.ErrRefreshNotFound
	}
	r.byHash[hash].RevokedAt = &revokedAt
	return nil
}
