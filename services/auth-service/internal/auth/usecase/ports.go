package usecase

import (
	"context"
	"time"

	"github.com/newfeed/community-news/services/auth-service/internal/auth/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
}

type RefreshTokenRepository interface {
	Save(ctx context.Context, token domain.RefreshToken) error
	FindByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, id string, revokedAt time.Time) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) bool
}

type TokenSigner interface {
	SignAccessToken(user domain.User) (string, time.Time, error)
	ParseAccessToken(token string) (*Claims, error)
}

type Claims = domain.Claims
