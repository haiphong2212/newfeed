package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/newfeed/community-news/services/auth-service/internal/auth/domain"
	"github.com/newfeed/community-news/services/auth-service/internal/auth/repository"
	"github.com/newfeed/community-news/services/auth-service/internal/platform/security"
)

func TestRegisterLoginRefreshValidate(t *testing.T) {
	users := repository.NewMemoryUserRepository()
	refreshTokens := repository.NewMemoryRefreshTokenRepository()
	passwords := security.NewPasswordHasher()
	tokens := security.NewJWTSigner("test-secret", time.Minute)
	auth := NewService(users, refreshTokens, passwords, tokens, time.Hour)

	user, err := auth.Register(context.Background(), RegisterInput{
		Email:     "User@Example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if user.Email != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", user.Email)
	}

	pair, err := auth.Login(context.Background(), LoginInput{Email: "user@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("expected token pair")
	}

	claims, err := auth.ValidateAccessToken(context.Background(), pair.AccessToken)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if claims.UserID != user.ID || claims.Role != string(domain.RoleReader) {
		t.Fatalf("unexpected claims: %+v", claims)
	}

	refreshed, err := auth.Refresh(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshed.RefreshToken == pair.RefreshToken {
		t.Fatal("refresh token was not rotated")
	}
}

func TestLoginRejectsBadPassword(t *testing.T) {
	auth := NewService(
		repository.NewMemoryUserRepository(),
		repository.NewMemoryRefreshTokenRepository(),
		security.NewPasswordHasher(),
		security.NewJWTSigner("test-secret", time.Minute),
		time.Hour,
	)
	_, err := auth.Register(context.Background(), RegisterInput{Email: "a@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	_, err = auth.Login(context.Background(), LoginInput{Email: "a@example.com", Password: "wrong"})
	if err != domain.ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}
