package grpc

import (
	"context"

	authv1 "github.com/newfeed/community-news/gen/auth/v1"
	"github.com/newfeed/community-news/services/auth-service/internal/auth/domain"
	"github.com/newfeed/community-news/services/auth-service/internal/auth/usecase"
)

type Server struct {
	authv1.UnimplementedAuthServiceServer
	auth *usecase.Service
}

func NewServer(auth *usecase.Service) *Server {
	return &Server{auth: auth}
}

func (s *Server) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.UserResponse, error) {
	user, err := s.auth.Register(ctx, usecase.RegisterInput{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
	})
	if err != nil {
		return nil, err
	}
	return &authv1.UserResponse{Id: user.ID, Email: user.Email, Role: string(user.Role)}, nil
}

func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.TokenResponse, error) {
	tokens, err := s.auth.Login(ctx, usecase.LoginInput{Email: req.GetEmail(), Password: req.GetPassword()})
	if err != nil {
		return nil, err
	}
	return tokenResponse(tokens), nil
}

func (s *Server) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.TokenResponse, error) {
	tokens, err := s.auth.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}
	return tokenResponse(tokens), nil
}

func (s *Server) Validate(ctx context.Context, req *authv1.ValidateRequest) (*authv1.ValidateResponse, error) {
	claims, err := s.auth.ValidateAccessToken(ctx, req.GetAccessToken())
	if err != nil {
		return nil, err
	}
	return &authv1.ValidateResponse{UserId: claims.UserID, Email: claims.Email, Role: claims.Role}, nil
}

func tokenResponse(tokens *domain.TokenPair) *authv1.TokenResponse {
	return &authv1.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	}
}
