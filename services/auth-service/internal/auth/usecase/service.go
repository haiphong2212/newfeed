package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/newfeed/community-news/services/auth-service/internal/auth/domain"
)

type Service struct {
	users         UserRepository
	refreshStore  RefreshTokenRepository
	passwords     PasswordHasher
	tokens        TokenSigner
	refreshExpiry time.Duration
}

func NewService(users UserRepository, refreshStore RefreshTokenRepository, passwords PasswordHasher, tokens TokenSigner, refreshExpiry time.Duration) *Service {
	return &Service{users: users, refreshStore: refreshStore, passwords: passwords, tokens: tokens, refreshExpiry: refreshExpiry}
}

type RegisterInput struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*domain.User, error) {
	if len(input.Password) < 8 {
		return nil, domain.ErrInvalidPassword
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if existing, err := s.users.FindByEmail(ctx, email); err == nil && existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}
	hash, err := s.passwords.Hash(input.Password)
	if err != nil {
		return nil, err
	}
	user, err := domain.NewUser(newID(), email, hash, input.FirstName, input.LastName)
	if err != nil {
		return nil, err
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*domain.TokenPair, error) {
	user, err := s.users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if !s.passwords.Compare(user.PasswordHash, input.Password) {
		return nil, domain.ErrInvalidCredentials
	}
	return s.issueTokens(ctx, *user)
}

func (s *Service) Refresh(ctx context.Context, rawRefreshToken string) (*domain.TokenPair, error) {
	stored, err := s.refreshStore.FindByHash(ctx, tokenHash(rawRefreshToken))
	if err != nil {
		return nil, domain.ErrRefreshNotFound
	}
	now := time.Now().UTC()
	if !stored.Active(now) {
		return nil, domain.ErrRefreshNotFound
	}
	user, err := s.users.FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}
	if err := s.refreshStore.Revoke(ctx, stored.ID, now); err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, *user)
}

func (s *Service) ValidateAccessToken(_ context.Context, raw string) (*Claims, error) {
	return s.tokens.ParseAccessToken(raw)
}

func (s *Service) issueTokens(ctx context.Context, user domain.User) (*domain.TokenPair, error) {
	access, accessExpiresAt, err := s.tokens.SignAccessToken(user)
	if err != nil {
		return nil, err
	}
	refresh := randomToken()
	now := time.Now().UTC()
	err = s.refreshStore.Save(ctx, domain.RefreshToken{
		ID:        newID(),
		UserID:    user.ID,
		TokenHash: tokenHash(refresh),
		ExpiresAt: now.Add(s.refreshExpiry),
		CreatedAt: now,
	})
	if err != nil {
		return nil, err
	}
	return &domain.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(accessExpiresAt).Seconds()),
	}, nil
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomToken() string {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func newID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	buf[6] = (buf[6] & 0x0f) | 0x40
	buf[8] = (buf[8] & 0x3f) | 0x80
	return hex.EncodeToString(buf[0:4]) + "-" + hex.EncodeToString(buf[4:6]) + "-" + hex.EncodeToString(buf[6:8]) + "-" + hex.EncodeToString(buf[8:10]) + "-" + hex.EncodeToString(buf[10:])
}
