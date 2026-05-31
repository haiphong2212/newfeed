package http

import (
	"github.com/gofiber/fiber/v2"
	authv1 "github.com/newfeed/community-news/gen/auth/v1"
	newsv1 "github.com/newfeed/community-news/gen/news/v1"
	searchv1 "github.com/newfeed/community-news/gen/search/v1"
	userv1 "github.com/newfeed/community-news/gen/user/v1"
	gatewaygrpc "github.com/newfeed/community-news/services/api-gateway/internal/gateway/grpc"
)

type Handler struct {
	health gatewaygrpc.HealthClient
	grpc   *gatewaygrpc.Clients
}

func NewHandler(health gatewaygrpc.HealthClient, clients *gatewaygrpc.Clients) Handler {
	return Handler{health: health, grpc: clients}
}

func (h Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/v1/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "gateway-ready"})
	})
	app.Get("/v1/internal/:service/health", func(c *fiber.Ctx) error {
		status, err := h.health.Check(c.UserContext(), c.Params("service"))
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{"service": c.Params("service"), "grpc_status": status})
	})
	app.Post("/v1/auth/register", h.register)
	app.Post("/v1/auth/login", h.login)
	app.Post("/v1/auth/refresh", h.refresh)
	app.Get("/v1/auth/validate", h.validate)
	app.Post("/v1/articles/publish", h.publishArticle)
	app.Put("/v1/users/:id/profile", h.upsertProfile)
	app.Post("/v1/users/:id/following/:target_id", h.followUser)
	app.Post("/v1/users/:id/topics/:topic", h.followTopic)
	app.Get("/v1/search/articles", h.searchArticles)
	app.Post("/v1/search/articles", h.indexArticle)
}

func (h Handler) register(c *fiber.Ctx) error {
	var req authv1.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	res, err := h.grpc.Auth.Register(c.UserContext(), &req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h Handler) login(c *fiber.Ctx) error {
	var req authv1.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	res, err := h.grpc.Auth.Login(c.UserContext(), &req)
	if err != nil {
		return err
	}
	return c.JSON(res)
}

func (h Handler) refresh(c *fiber.Ctx) error {
	var req authv1.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	res, err := h.grpc.Auth.Refresh(c.UserContext(), &req)
	if err != nil {
		return err
	}
	return c.JSON(res)
}

func (h Handler) validate(c *fiber.Ctx) error {
	res, err := h.grpc.Auth.Validate(c.UserContext(), &authv1.ValidateRequest{AccessToken: c.Query("access_token")})
	if err != nil {
		return err
	}
	return c.JSON(res)
}

func (h Handler) publishArticle(c *fiber.Ctx) error {
	var req newsv1.PublishArticleRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	res, err := h.grpc.News.PublishArticle(c.UserContext(), &req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusAccepted).JSON(res)
}

func (h Handler) upsertProfile(c *fiber.Ctx) error {
	var req userv1.UpsertProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	req.UserId = c.Params("id")
	_, err := h.grpc.User.UpsertProfile(c.UserContext(), &req)
	return err
}

func (h Handler) followUser(c *fiber.Ctx) error {
	_, err := h.grpc.User.FollowUser(c.UserContext(), &userv1.FollowUserRequest{FollowerId: c.Params("id"), FollowedId: c.Params("target_id")})
	return err
}

func (h Handler) followTopic(c *fiber.Ctx) error {
	_, err := h.grpc.User.FollowTopic(c.UserContext(), &userv1.FollowTopicRequest{UserId: c.Params("id"), Topic: c.Params("topic")})
	return err
}

func (h Handler) searchArticles(c *fiber.Ctx) error {
	res, err := h.grpc.Search.SearchArticles(c.UserContext(), &searchv1.SearchArticlesRequest{
		Query:    c.Query("q"),
		Title:    c.Query("title"),
		Tag:      c.Query("tag"),
		Category: c.Query("category"),
		Limit:    int32(c.QueryInt("limit", 20)),
	})
	if err != nil {
		return err
	}
	return c.JSON(res)
}

func (h Handler) indexArticle(c *fiber.Ctx) error {
	var req searchv1.IndexArticleRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid json body")
	}
	_, err := h.grpc.Search.IndexArticle(c.UserContext(), &req)
	return err
}
