package main

import (
	"context"
	"log"
	"time"

	authv1 "github.com/newfeed/community-news/gen/auth/v1"
	authgrpc "github.com/newfeed/community-news/services/auth-service/internal/auth/delivery/grpc"
	authhttp "github.com/newfeed/community-news/services/auth-service/internal/auth/delivery/http"
	"github.com/newfeed/community-news/services/auth-service/internal/auth/repository"
	"github.com/newfeed/community-news/services/auth-service/internal/auth/usecase"
	"github.com/newfeed/community-news/services/auth-service/internal/platform/security"
	"github.com/newfeed/community-news/services/shared/env"
	"github.com/newfeed/community-news/services/shared/platform"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	cfg := platform.Load("auth-service", ":8001", ":50051")
	logg := platform.Logger(cfg.ServiceName, cfg.Env)

	db, err := platform.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	users := repository.NewPostgresUserRepository(db)
	refreshTokens := repository.NewPostgresRefreshTokenRepository(db)
	passwords := security.NewPasswordHasher()
	tokens := security.NewJWTSigner(env.String("JWT_SECRET", "dev-secret-change-me"), env.Duration("JWT_EXPIRY", 15*time.Minute))
	auth := usecase.NewService(users, refreshTokens, passwords, tokens, env.Duration("REFRESH_TOKEN_EXPIRY", 7*24*time.Hour))
	grpcServer, err := platform.StartGRPC(cfg.GRPCAddr, cfg.ServiceName, logg, func(server *grpc.Server) {
		authv1.RegisterAuthServiceServer(server, authgrpc.NewServer(auth))
	})
	if err != nil {
		log.Fatal(err)
	}
	defer grpcServer.GracefulStop()

	handler := authhttp.NewHandler(auth)
	app := platform.NewFiber(cfg.ServiceName)
	handler.RegisterRoutes(app)
	if err := platform.ListenFiber(app, cfg.HTTPAddr, logg); err != nil {
		log.Fatal(err)
	}
}
